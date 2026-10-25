package runtime

import "testing"

func TestIsArrayValueDomain(t *testing.T) {
	arrays := []struct {
		name  string
		value any
	}{
		{"zero_value", Array{}},
		{"empty", NewArray()},
		{"mixed_keys", NewArray(Entry{Key: int64(4), Value: "four"}, Entry{Key: "name", Value: false})},
		{"nested", NewArray(Entry{Value: NewArray(Entry{Key: "inner", Value: nil})})},
	}
	for _, tt := range arrays {
		t.Run(tt.name, func(t *testing.T) {
			if !IsArray(tt.value) {
				t.Fatalf("is_array(%T) = false, want true", tt.value)
			}
		})
	}

	notArrays := []struct {
		name  string
		value any
	}{
		{"nil", nil},
		{"bool", true},
		{"integer", int64(1)},
		{"string", "array"},
		{"slice", []any{}},
		{"map", map[string]any{}},
		{"array_pointer", &Array{}},
	}
	for _, tt := range notArrays {
		t.Run(tt.name, func(t *testing.T) {
			if IsArray(tt.value) {
				t.Fatalf("is_array(%T) = true, want false", tt.value)
			}
		})
	}
}
