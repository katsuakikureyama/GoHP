package runtime

import (
	"reflect"
	"testing"
)

func TestArrayValuesOrderTypesAndIsolation(t *testing.T) {
	input := NewArray(
		Entry{Key: int64(9), Value: int64(7)},
		Entry{Key: "name", Value: "日本語"},
		Entry{Key: int64(3), Value: true},
		Entry{Key: "empty", Value: nil},
		Entry{Key: "false", Value: false},
		Entry{Key: "nested", Value: NewArray(Entry{Key: "kept", Value: int64(4)})},
		Entry{Key: "9", Value: int64(8)},
	)
	before := Clone(input)
	got, ok := ArrayValues(input).(Array)
	want := []Entry{
		{Key: int64(0), Value: int64(8)},
		{Key: int64(1), Value: "日本語"},
		{Key: int64(2), Value: true},
		{Key: int64(3), Value: nil},
		{Key: int64(4), Value: false},
		{Key: int64(5), Value: NewArray(Entry{Key: "kept", Value: int64(4)})},
	}
	if !ok || !reflect.DeepEqual(got.Entries, want) {
		t.Fatalf("array_values result = %#v, want typed entries %#v", got, want)
	}
	if !reflect.DeepEqual(input, before) {
		t.Fatal("array_values modified its input")
	}
	got.Entries[0].Value = "changed result"
	got.Entries[5].Value.(Array).Entries[0].Value = "changed nested result"
	if !reflect.DeepEqual(input, before) {
		t.Fatal("result aliases input outer or nested array storage")
	}
	input.Entries[1].Value = "changed input"
	input.Entries[5].Value.(Array).Entries[0].Value = "changed nested input"
	if got.Entries[1].Value != "日本語" || Get(got.Entries[5].Value, "kept") != "changed nested result" {
		t.Fatal("input mutation changed result storage")
	}
	empty, ok := ArrayValues(NewArray()).(Array)
	if !ok || Count(empty) != int64(0) {
		t.Fatalf("empty input did not return an empty array: %#v", empty)
	}
}

func TestArrayValuesRejectsNonArray(t *testing.T) {
	for _, tt := range []struct {
		name  string
		value any
	}{
		{"integer", int64(1)}, {"string", ""}, {"boolean", false},
		{"null", nil}, {"object", &struct{}{}}, {"float", 1.5},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if got := recover(); got != "array_values requires array" {
					t.Fatalf("non-array was not explicitly rejected: %v", got)
				}
			}()
			ArrayValues(tt.value)
		})
	}
}
