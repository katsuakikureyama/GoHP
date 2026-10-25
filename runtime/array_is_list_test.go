package runtime

import (
	"math"
	"reflect"
	"testing"
)

func TestArrayIsListOrderAndValues(t *testing.T) {
	for _, tt := range []struct {
		name  string
		array Array
		want  bool
	}{
		{"empty", NewArray(), true},
		{"zero_value", Array{}, true},
		{"implicit", NewArray(Entry{Value: nil}, Entry{Value: false}, Entry{Value: ""}), true},
		{"explicit", NewArray(Entry{Key: int64(0), Value: "a"}, Entry{Key: int64(1), Value: "b"}), true},
		{"numeric_strings", NewArray(Entry{Key: "0", Value: nil}, Entry{Key: "1", Value: nil}), true},
		{"leading_zero", NewArray(Entry{Key: "00", Value: nil}), false},
		{"plus", NewArray(Entry{Key: "+0", Value: nil}), false},
		{"nonzero_start", NewArray(Entry{Key: int64(1), Value: nil}), false},
		{"gap", NewArray(Entry{Key: int64(0), Value: nil}, Entry{Key: int64(2), Value: nil}), false},
		{"order", NewArray(Entry{Key: int64(1), Value: nil}, Entry{Key: int64(0), Value: nil}), false},
		{"mixed", NewArray(Entry{Key: int64(0), Value: nil}, Entry{Key: "name", Value: nil}), false},
		{"negative", NewArray(Entry{Key: int64(-1), Value: nil}), false},
		{"max", NewArray(Entry{Key: int64(math.MaxInt64), Value: nil}), false},
		{"binary", NewArray(Entry{Key: "\x00\xff", Value: nil}), false},
		{"nested_assoc_value", NewArray(Entry{Value: NewArray(Entry{Key: "nested", Value: nil})}), true},
		{"replacement_order", NewArray(Entry{Key: int64(0), Value: "old"}, Entry{Key: int64(1), Value: nil}, Entry{Key: "0", Value: "new"}), true},
		{"replacement_out_of_order", NewArray(Entry{Key: int64(1), Value: "old"}, Entry{Key: int64(0), Value: nil}, Entry{Key: "1", Value: "new"}), false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			before := Clone(tt.array)
			if got := ArrayIsList(tt.array); got != tt.want {
				t.Fatalf("array_is_list(%#v) = %v, want %v", tt.array, got, tt.want)
			}
			if !reflect.DeepEqual(tt.array, before) {
				t.Fatal("list check changed array order, values or nested state")
			}
		})
	}
}

func TestArrayIsListRejectsNonArray(t *testing.T) {
	for _, tt := range []struct {
		name  string
		value any
	}{
		{"integer", int64(1)}, {"string", ""}, {"boolean", false},
		{"null", nil}, {"object", &struct{}{}}, {"float", 1.5},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if got := recover(); got != "array_is_list requires array" {
					t.Fatalf("non-array was not explicitly rejected: %v", got)
				}
			}()
			ArrayIsList(tt.value)
		})
	}
}
