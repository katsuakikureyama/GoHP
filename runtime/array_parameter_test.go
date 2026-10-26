package runtime

import (
	"reflect"
	"testing"
)

func TestArrayParameterClonesArrayAndMetadata(t *testing.T) {
	input := NewArray(
		Entry{Key: int64(-5), Value: NewArray(Entry{Key: "child", Value: "original"})},
		Entry{Key: "label", Value: true},
		Entry{Key: int64(3), Value: "tail"},
	)
	before := Clone(input).(Array)

	got, ok := ArrayParameter(input).(Array)
	if !ok {
		t.Fatal("array parameter did not return an array")
	}
	if !reflect.DeepEqual(got, input) {
		t.Fatalf("array parameter = %#v, want %#v", got, input)
	}
	if got.index != input.index {
		t.Fatal("array parameter lost next-index metadata")
	}

	got.Entries[0].Value.(Array).Entries[0].Value = "changed"
	got.Entries[1].Value = false
	if !reflect.DeepEqual(input, before) {
		t.Fatalf("array parameter aliases input: got %#v, want %#v", input, before)
	}
}

func TestArrayParameterRejectsNonArrays(t *testing.T) {
	for _, tt := range []struct {
		name  string
		value any
	}{
		{"null", nil},
		{"boolean", true},
		{"integer", int64(1)},
		{"string", "items"},
		{"object", struct{}{}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if got := recover(); got != "array parameter requires array" {
					t.Fatalf("panic = %v, want %q", got, "array parameter requires array")
				}
			}()
			ArrayParameter(tt.value)
		})
	}
}
