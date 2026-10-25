package runtime

import (
	"math"
	"reflect"
	"testing"
)

func TestArrayKeysOrderTypesAndIsolation(t *testing.T) {
	input := NewArray(
		Entry{Key: "2", Value: "first"},
		Entry{Key: "02", Value: nil},
		Entry{Key: int64(2), Value: "replacement"},
		Entry{Key: int64(-3), Value: NewArray(Entry{Value: "nested"})},
		Entry{Key: "日本語", Value: false},
		Entry{Key: int64(math.MaxInt64), Value: true},
		Entry{Key: "a\x00\xff", Value: int64(1)},
	)
	before := Clone(input)
	got, ok := ArrayKeys(input).(Array)
	want := []Entry{
		{Key: int64(0), Value: int64(2)},
		{Key: int64(1), Value: "02"},
		{Key: int64(2), Value: int64(-3)},
		{Key: int64(3), Value: "日本語"},
		{Key: int64(4), Value: int64(math.MaxInt64)},
		{Key: int64(5), Value: "a\x00\xff"},
	}
	if !ok || !reflect.DeepEqual(got.Entries, want) {
		t.Fatalf("array_keys result = %#v, want typed entries %#v", got, want)
	}
	if !reflect.DeepEqual(input, before) {
		t.Fatal("array_keys modified its input")
	}
	got.Entries[0].Value = "changed result"
	if !reflect.DeepEqual(input, before) {
		t.Fatal("result aliases input storage")
	}
	input.Entries[1].Key = "changed input"
	if got.Entries[1].Value != "02" {
		t.Fatal("input mutation changed the result")
	}
	empty, ok := ArrayKeys(NewArray()).(Array)
	if !ok || Count(empty) != int64(0) {
		t.Fatalf("empty input did not return an empty array: %#v", empty)
	}
}

func TestArrayKeysRejectsNonArray(t *testing.T) {
	for _, tt := range []struct {
		name  string
		value any
	}{
		{"integer", int64(1)}, {"string", ""}, {"boolean", false},
		{"null", nil}, {"object", &struct{}{}}, {"float", 1.5},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if got := recover(); got != "array_keys requires array" {
					t.Fatalf("non-array was not explicitly rejected: %v", got)
				}
			}()
			ArrayKeys(tt.value)
		})
	}
}
