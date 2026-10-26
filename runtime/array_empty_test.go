package runtime

import (
	"reflect"
	"testing"
)

func TestArrayEmptyOuterCountAndPreservation(t *testing.T) {
	for _, input := range []Array{Array{}, NewArray()} {
		if !ArrayEmpty(input) {
			t.Fatal("empty array must return true")
		}
	}
	for _, value := range []any{nil, false, int64(0), "", "0", NewArray(), &struct{}{}, NewCallable(0, func(...any) any { return nil })} {
		input := NewArray(Entry{Key: int64(-5), Value: value})
		before := Clone(input).(Array)
		if ArrayEmpty(input) {
			t.Fatalf("singleton array is not empty: %#v", value)
		}
		if len(input.Entries) != 1 || input.Entries[0].Key != before.Entries[0].Key || input.index != before.index {
			t.Fatal("empty changed input structure/index metadata")
		}
	}
	input := NewArray(Entry{Key: "nested", Value: NewArray(Entry{Key: "x", Value: nil})})
	before := Clone(input)
	ArrayEmpty(input)
	if !reflect.DeepEqual(input, before) {
		t.Fatal("empty changed nested input")
	}
}

func TestArrayEmptyRejectsNonArrays(t *testing.T) {
	for _, value := range []any{nil, false, true, int64(0), "", "0", &struct{}{}, (*Array)(nil), []any(nil)} {
		func() {
			defer func() {
				if got := recover(); got != "empty requires array" {
					t.Fatalf("nonarray %#v rejection = %v", value, got)
				}
			}()
			ArrayEmpty(value)
		}()
	}
}
