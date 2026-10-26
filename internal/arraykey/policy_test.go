package arraykey

import (
	"math"
	"testing"
)

func TestArrayKeyNormalize(t *testing.T) {
	for _, tc := range []struct{ input, want any }{
		{int64(0), int64(0)}, {int64(-5), int64(-5)},
		{"0", int64(0)}, {"1", int64(1)}, {"-5", int64(-5)},
		{"9223372036854775807", int64(math.MaxInt64)},
		{"-9223372036854775808", int64(math.MinInt64)},
		{"9223372036854775808", "9223372036854775808"},
		{"-9223372036854775809", "-9223372036854775809"},
		{"", ""}, {"+1", "+1"}, {"01", "01"}, {"-0", "-0"}, {"-01", "-01"},
		{" 1", " 1"}, {"1\n", "1\n"}, {"1.0", "1.0"}, {"1e2", "1e2"},
		{"0x10", "0x10"}, {"1_0", "1_0"}, {"１", "１"}, {"1\x00", "1\x00"},
	} {
		if got := Normalize(tc.input); got != tc.want {
			t.Fatalf("Normalize(%#v) = %#v, want %#v", tc.input, got, tc.want)
		}
	}
	for _, key := range []any{nil, true, int(1), uint64(1), float64(1), []int{1}, map[string]int{"1": 1}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("Normalize(%#v) accepted unsupported key", key)
				}
			}()
			Normalize(key)
		}()
	}
}

func TestArrayNextIndex(t *testing.T) {
	for _, tc := range []struct {
		name string
		keys []int64
		want int64
	}{
		{"empty", nil, 0},
		{"negative", []int64{-5}, -4},
		{"lower and replacement", []int64{-5, -8, -5}, -4},
		{"advance through zero", []int64{-2, -1, 0}, 1},
		{"zero is initialized", []int64{0, -5}, 1},
		{"minimum", []int64{math.MinInt64}, math.MinInt64 + 1},
		{"maximum available", []int64{math.MaxInt64 - 1}, math.MaxInt64},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var index Index
			for _, key := range tc.keys {
				index.Observe(key)
			}
			for i := 0; i < 2; i++ {
				if got := index.Next(); got != tc.want {
					t.Fatalf("Next = %d, want %d; reads must not advance", got, tc.want)
				}
			}
		})
	}

	var original Index
	original.Observe(-5)
	copy := original
	copy.Observe(9)
	if original.Next() != -4 || copy.Next() != 10 {
		t.Fatal("copied index state did not evolve independently")
	}
	copy.Observe(math.MaxInt64)
	copy.Observe(-10) // Explicit lower keys cannot unblock exhausted state.
	blockedCopy := copy
	for _, index := range []Index{copy, blockedCopy} {
		func() {
			defer func() {
				if got := recover(); got != "array key overflow unsupported" {
					t.Errorf("exhausted Next panic = %#v", got)
				}
			}()
			index.Next()
		}()
	}
	if original.Next() != -4 {
		t.Fatal("exhausting a copy changed the original index")
	}
}
