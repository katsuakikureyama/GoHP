package runtime

import (
	"math"
	"reflect"
	"testing"
)

func TestArrayKeyLastOrderTypesAndPreservation(t *testing.T) {
	for _, input := range []Array{Array{}, NewArray()} {
		if got := ArrayKeyLast(input); got != nil {
			t.Fatalf("empty key = %#v, want nil", got)
		}
	}
	for _, key := range []any{int64(0), int64(-5), int64(math.MinInt64), int64(math.MaxInt64), "", "02", "日本語", "a\x00\xff"} {
		input := NewArray(Entry{Key: "first", Value: nil}, Entry{Key: key, Value: NewArray(Entry{Value: false})})
		before := Clone(input)
		if got := ArrayKeyLast(input); got != key {
			t.Fatalf("last key = %#v, want %#v", got, key)
		}
		if !reflect.DeepEqual(input, before) {
			t.Fatal("last-key lookup changed input/metadata")
		}
	}
	input := NewArray(Entry{Key: "8", Value: false}, Entry{Key: int64(-2), Value: nil}, Entry{Key: int64(8), Value: "replacement"})
	if got := ArrayKeyLast(input); got != int64(-2) {
		t.Fatalf("replacement moved key or lookup sorted: %#v", got)
	}
	if got := ArrayKeyLast(NewArray(Entry{Key: "2", Value: nil})); got != int64(2) {
		t.Fatalf("normalized key lost integer type: %#v", got)
	}
	updated := ArraySet(input, int64(8), "updated")
	if got := ArrayKeyLast(updated); got != int64(-2) {
		t.Fatalf("existing-key write moved key: %#v", got)
	}
	// Lookup must not advance, reset or otherwise disturb automatic-key state.
	negative := NewArray(Entry{Key: int64(-5), Value: false})
	ArrayKeyLast(negative)
	if got := ArrayKeyLast(ArrayAppend(negative, nil)); got != int64(-4) {
		t.Fatalf("negative append metadata changed: %#v", got)
	}
	blocked := NewArray(Entry{Key: int64(math.MaxInt64), Value: nil})
	before := Clone(blocked)
	ArrayKeyLast(blocked)
	if !reflect.DeepEqual(blocked, before) {
		t.Fatal("overflow metadata changed")
	}
}

func TestArrayKeyLastRejectsNonArrays(t *testing.T) {
	for _, value := range []any{nil, false, int64(0), "", &struct{}{}, (*Array)(nil), []any(nil)} {
		func() {
			defer func() {
				if got := recover(); got != "array_key_last requires array" {
					t.Fatalf("nonarray %#v rejection = %v", value, got)
				}
			}()
			ArrayKeyLast(value)
		}()
	}
}
