package runtime

import "testing"

type objectMarker struct{ valid bool }

func (o *objectMarker) PHPObject() bool { return o != nil && o.valid }

type getterOnly struct{}

func (getterOnly) Get(string) any { return "host getter" }

func TestIsObjectValueDomain(t *testing.T) {
	for _, tt := range []struct {
		name  string
		value any
	}{
		{"marked_pointer", &objectMarker{valid: true}},
		{"generated_arrow", NewCallable(0, func(...any) any { return nil })},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if !IsObject(tt.value) {
				t.Fatalf("is_object(%T) = false, want true", tt.value)
			}
		})
	}

	for _, tt := range []struct {
		name  string
		value any
	}{
		{"nil", nil},
		{"false", false},
		{"integer", int64(1)},
		{"string", "object"},
		{"array", NewArray()},
		{"foreign_struct", struct{}{}},
		{"getter_only", getterOnly{}},
		{"marked_false", &objectMarker{valid: false}},
		{"marked_nil", (*objectMarker)(nil)},
		{"callable_nil", (*Callable)(nil)},
		{"callable_invalid_arity", &Callable{arity: -1, call: func(...any) any { return nil }}},
		{"callable_nil_function", &Callable{arity: 0}},
		{"slice", []any{}},
		{"map", map[string]any{}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if IsObject(tt.value) {
				t.Fatalf("is_object(%T) = true, want false", tt.value)
			}
		})
	}
}
