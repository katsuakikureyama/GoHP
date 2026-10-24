package runtime

import (
	"math"
	"reflect"
	"testing"
)

func TestArrayKeyExistsPresenceAndNormalization(t *testing.T) {
	input := NewArray(
		Entry{Key: "null", Value: nil},
		Entry{Key: "false", Value: false},
		Entry{Key: "zero", Value: int64(0)},
		Entry{Key: "empty", Value: ""},
		Entry{Key: "2", Value: "replaced"},
		Entry{Key: int64(2), Value: nil},
		Entry{Key: "02", Value: "distinct"},
		Entry{Key: "", Value: nil},
		Entry{Key: int64(-3), Value: true},
		Entry{Key: int64(math.MinInt64), Value: nil},
		Entry{Key: int64(math.MaxInt64), Value: nil},
		Entry{Key: "nested", Value: NewArray(Entry{Key: "inside", Value: nil})},
		Entry{Key: "日本語", Value: nil},
		Entry{Key: "a\x00\xff", Value: nil},
	)
	before := Clone(input)
	for _, tt := range []struct {
		name string
		key  any
		want bool
	}{
		{"null_value", "null", true}, {"false_value", "false", true},
		{"zero_value", "zero", true}, {"empty_value", "empty", true},
		{"missing", "missing", false}, {"integer", int64(2), true},
		{"numeric_string", "2", true}, {"leading_zero", "02", true},
		{"plus_is_distinct", "+2", false}, {"space_is_distinct", " 2", false},
		{"empty_key", "", true}, {"negative", int64(-3), true},
		{"negative_string", "-3", true}, {"min", "-9223372036854775808", true},
		{"max", "9223372036854775807", true}, {"out_of_range", "9223372036854775808", false},
		{"outer", "nested", true}, {"not_recursive", "inside", false},
		{"unicode", "日本語", true}, {"binary", "a\x00\xff", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := ArrayKeyExists(tt.key, input); got != tt.want {
				t.Fatalf("array_key_exists(%#v) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
	if !reflect.DeepEqual(input, before) {
		t.Fatal("key lookup mutated the array or its nested values")
	}
	if ArrayKeyExists("", NewArray()) || ArrayKeyExists(int64(0), NewArray()) {
		t.Fatal("empty array reported an existing key")
	}
}

func TestArrayKeyExistsRejectsInvalidDomains(t *testing.T) {
	for _, tt := range []struct {
		name  string
		value any
	}{
		{"null", nil}, {"boolean", false}, {"float", 1.5},
		{"object", &struct{}{}}, {"array", NewArray()},
	} {
		t.Run("key_"+tt.name, func(t *testing.T) {
			defer func() {
				if got := recover(); got != "array_key_exists requires integer or string key" {
					t.Fatalf("unsupported key was not explicitly rejected: %v", got)
				}
			}()
			ArrayKeyExists(tt.value, NewArray())
		})
	}
	for _, tt := range []struct {
		name  string
		value any
	}{
		{"null", nil}, {"boolean", false}, {"float", 1.5},
		{"object", &struct{}{}}, {"integer", int64(1)}, {"string", "x"},
	} {
		t.Run("array_"+tt.name, func(t *testing.T) {
			defer func() {
				if got := recover(); got != "array_key_exists requires array" {
					t.Fatalf("non-array was not explicitly rejected: %v", got)
				}
			}()
			ArrayKeyExists("x", tt.value)
		})
	}
}
