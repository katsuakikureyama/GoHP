package runtime

import (
	"math"
	"reflect"
	"testing"
)

func TestArrayReverseOrderKeysAndIsolation(t *testing.T) {
	input := NewArray(Entry{Key: "8", Value: nil}, Entry{Key: "02", Value: false}, Entry{Key: int64(-5), Value: NewArray(Entry{Value: "nested"})}, Entry{Key: "a\x00\xff", Value: "tail"})
	before := Clone(input)
	for _, preserve := range []bool{false, true} {
		got := ArrayReverse(input, preserve).(Array)
		keys := []any{"a\x00\xff", int64(0), "02", int64(1)}
		if preserve {
			keys = []any{"a\x00\xff", int64(-5), "02", int64(8)}
		}
		for i, entry := range got.Entries {
			if entry.Key != keys[i] || !reflect.DeepEqual(entry.Value, input.Entries[3-i].Value) {
				t.Fatalf("entry %d = %#v", i, entry)
			}
		}
		got.Entries[1].Value.(Array).Entries[0].Value = "changed"
		got.Entries[0].Value = "changed"
		if !reflect.DeepEqual(input, before) {
			t.Fatal("input or nested storage changed")
		}
		if got := ArrayReverse(NewArray(), preserve).(Array); len(got.Entries) != 0 || ArrayKeyLast(ArrayAppend(got, nil)) != int64(0) {
			t.Fatal("empty result metadata")
		}
	}
}

func TestArrayReverseAppendMetadata(t *testing.T) {
	for _, tc := range []struct {
		keys     []int64
		preserve bool
		next     int64
	}{
		{[]int64{-5, -10}, true, -4}, {[]int64{-5, -10}, false, 2},
		{[]int64{7, 1}, true, 8}, {[]int64{math.MaxInt64}, false, 1},
		{[]int64{math.MinInt64}, true, math.MinInt64 + 1},
	} {
		input := NewArray()
		for _, key := range tc.keys {
			input = ArraySet(input, key, nil).(Array)
		}
		result := ArrayReverse(input, tc.preserve)
		if got := ArrayKeyLast(ArrayAppend(result, nil)); got != tc.next {
			t.Fatalf("keys %v preserve %v: next %v, want %v", tc.keys, tc.preserve, got, tc.next)
		}
	}
	input := NewArray(Entry{Key: int64(math.MaxInt64), Value: nil})
	defer func() {
		if got := recover(); got != "array key overflow unsupported" {
			t.Fatalf("overflow guard: %v", got)
		}
	}()
	ArrayAppend(ArrayReverse(input, true), nil)
}

func TestArrayReverseRejectsInvalidDomains(t *testing.T) {
	for _, value := range []any{nil, false, int64(0), "", &struct{}{}, (*Array)(nil), []any(nil)} {
		func() {
			defer func() {
				if got := recover(); got != "array_reverse requires array" {
					t.Fatalf("array guard: %v", got)
				}
			}()
			ArrayReverse(value, false)
		}()
	}
	for _, flag := range []any{nil, int64(0), "", NewArray(), &struct{}{}} {
		func() {
			defer func() {
				if got := recover(); got != "array_reverse preserve_keys requires bool" {
					t.Fatalf("flag guard: %v", got)
				}
			}()
			ArrayReverse(NewArray(), flag)
		}()
	}
}
