package arraykey

import (
	"math"
	"testing"
)

func TestArrayPopIndex(t *testing.T) {
	for _, tc := range []struct {
		name string
		keys []int64
		pop  int64
		want int64
	}{
		{"negative", []int64{-5}, -5, -5},
		{"sparse last", []int64{0, 2}, 2, 2},
		{"out of order nonpredecessor", []int64{10, 2}, 2, 11},
		{"lower key does not rewind", []int64{4, 3}, 3, 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var index Index
			for _, key := range tc.keys {
				index.Observe(key)
			}
			index.Pop(tc.pop) // The caller has already removed this integer key.
			if got := index.Next(); got != tc.want {
				t.Fatalf("Next after Pop(%d) = %d, want %d", tc.pop, got, tc.want)
			}
		})
	}

	// String keys are never observed or popped by Index, so a string-only PHP
	// array retains the zero automatic-key default.
	var stringOnly Index
	if got := stringOnly.Next(); got != 0 {
		t.Fatalf("string-only next = %d, want 0", got)
	}

	var minimum Index
	minimum.Observe(math.MinInt64)
	minimum.Pop(math.MinInt64)
	if got := minimum.Next(); got != 0 {
		t.Fatalf("Next after popping MinInt64 = %d, want PHP sentinel 0", got)
	}

	var maximum Index
	maximum.Observe(math.MaxInt64)
	maximum.Pop(math.MaxInt64)
	if got := maximum.Next(); got != math.MaxInt64 {
		t.Fatalf("Next after popping sole MaxInt64 = %d, want MaxInt64", got)
	}

	// The high key remains occupied here. Removing only the lower final entry
	// makes that predecessor reusable, but must not permit overwriting MAX.
	var occupiedMaximum Index
	occupiedMaximum.Observe(math.MaxInt64)
	occupiedMaximum.Observe(math.MaxInt64 - 1)
	occupiedMaximum.Pop(math.MaxInt64 - 1)
	if got := occupiedMaximum.Next(); got != math.MaxInt64-1 {
		t.Fatalf("Next after popping MaxInt64-1 = %d, want MaxInt64-1", got)
	}
	occupiedMaximum.Observe(math.MaxInt64 - 1) // Simulate the permitted append.
	assertOverflowNext(t, occupiedMaximum)

	var original Index
	original.Observe(2)
	copy := original
	copy.Pop(2)
	if got := original.Next(); got != 3 {
		t.Fatalf("original Next after copying/pop = %d, want 3", got)
	}
	if got := copy.Next(); got != 2 {
		t.Fatalf("copy Next after Pop = %d, want 2", got)
	}
}

func assertOverflowNext(t *testing.T, index Index) {
	t.Helper()
	defer func() {
		if got := recover(); got != "array key overflow unsupported" {
			t.Fatalf("Next panic = %v, want %q", got, "array key overflow unsupported")
		}
	}()
	index.Next()
}
