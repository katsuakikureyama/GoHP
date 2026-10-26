package runtime

import (
	"math"
	"testing"
)

func TestSubstrBounds(t *testing.T) {
	for _, tc := range []struct {
		name   string
		input  string
		offset int64
		length any
		want   string
	}{
		{"empty", "", 0, nil, ""},
		{"zero", "0", 0, nil, "0"},
		{"tail", "abcd", 1, nil, "bcd"},
		{"negative offset", "abcd", -2, nil, "cd"},
		{"negative offset clamps", "abcd", -9, nil, "abcd"},
		{"offset at end", "abcd", 4, nil, ""},
		{"offset after end", "abcd", 9, nil, ""},
		{"positive length", "abcd", 1, int64(2), "bc"},
		{"zero length", "abcd", 1, int64(0), ""},
		{"negative length", "abcd", 1, int64(-1), "bc"},
		{"negative offset and length", "abcd", -3, int64(-2), "b"},
		{"negative length before start", "abcd", 3, int64(-2), ""},
		{"nul", "a\x00b", 1, int64(1), "\x00"},
		{"unicode is bytes", "\xc3\xa9", 1, int64(1), "\xa9"},
		{"minimum offset", "abcd", math.MinInt64, nil, "abcd"},
		{"maximum offset", "abcd", math.MaxInt64, nil, ""},
		{"maximum length", "abcd", 0, int64(math.MaxInt64), "abcd"},
		{"minimum length", "abcd", 0, int64(math.MinInt64), ""},
		{"minimum values", "abcd", math.MinInt64, int64(math.MinInt64), ""},
		{"maximum values", "abcd", math.MaxInt64, int64(math.MaxInt64), ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := Substr(tc.input, tc.offset, tc.length); got != tc.want {
				t.Fatalf("Substr(%q, %d, %v) = %q, want %q", tc.input, tc.offset, tc.length, got, tc.want)
			}
		})
	}
}

func TestSubstrRejectsInvalidDomains(t *testing.T) {
	for _, tc := range []struct {
		name string
		call func()
		want string
	}{
		{"value", func() { Substr(int64(0), int64(0), nil) }, "substr requires string"},
		{"offset", func() { Substr("", "0", nil) }, "substr offset requires int"},
		{"length", func() { Substr("", int64(0), false) }, "substr length requires int or null"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if got := recover(); got != tc.want {
					t.Fatalf("panic = %v, want %q", got, tc.want)
				}
			}()
			tc.call()
		})
	}
}
