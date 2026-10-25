package runtime

import (
	"math"
	"reflect"
	"testing"
)

func TestArrayKeyFirstOrderTypesAndPreservation(t *testing.T) {
	for _, array := range []Array{Array{}, NewArray()} {
		if got := ArrayKeyFirst(array); got != nil {
			t.Fatalf("empty result = %#v, want nil", got)
		}
	}
	for _, key := range []any{int64(0), int64(-5), int64(math.MinInt64), int64(math.MaxInt64), "", "02", "日本語", "a\x00\xff"} {
		array := NewArray(Entry{Key: key, Value: NewArray(Entry{Value: nil})}, Entry{Key: "tail", Value: false})
		before := Clone(array)
		if got := ArrayKeyFirst(array); got != key {
			t.Fatalf("first key = %#v, want %#v", got, key)
		}
		if !reflect.DeepEqual(array, before) {
			t.Fatal("first-key lookup changed contents or index metadata")
		}
	}
	array := NewArray(Entry{Key: "8", Value: nil}, Entry{Key: int64(-2), Value: false}, Entry{Key: int64(8), Value: "replaced"})
	if got := ArrayKeyFirst(array); got != int64(8) {
		t.Fatalf("first key sorted, moved or lost numeric normalization: %#v", got)
	}
	updated := ArraySet(array, int64(8), "updated")
	appended := ArrayAppend(updated, nil)
	if ArrayKeyFirst(appended) != int64(8) || ArrayKeyLast(appended) != int64(9) {
		t.Fatalf("replacement/append changed first position: %#v", appended)
	}
	negative := NewArray(Entry{Key: int64(-5), Value: false})
	ArrayKeyFirst(negative)
	appended = ArrayAppend(negative, nil)
	if ArrayKeyFirst(appended) != int64(-5) || ArrayKeyLast(appended) != int64(-4) {
		t.Fatal("lookup changed negative append metadata")
	}
	empty := NewArray()
	ArrayKeyFirst(empty)
	if got := ArrayKeyFirst(ArrayAppend(empty, nil)); got != int64(0) {
		t.Fatalf("empty lookup changed first append key: %#v", got)
	}
}

func TestArrayKeyFirstRejectsNonArrays(t *testing.T) {
	for _, value := range []any{nil, false, int64(0), "", &struct{}{}, (*Array)(nil), []any(nil)} {
		func() {
			defer func() {
				if got := recover(); got != "array_key_first requires array" {
					t.Fatalf("nonarray %#v rejection = %v", value, got)
				}
			}()
			ArrayKeyFirst(value)
		}()
	}
}
