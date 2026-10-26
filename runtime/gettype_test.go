package runtime

import "testing"

type getTypeGetterOnly struct{ calls *int }

func (value getTypeGetterOnly) Get(string) any {
	(*value.calls)++
	return "host getter"
}

type getTypeObjectMarker struct{ valid bool }

func (value *getTypeObjectMarker) PHPObject() bool { return value != nil && value.valid }

func TestGetTypeSupportedValues(t *testing.T) {
	called := 0
	arrow := NewCallable(0, func(...any) any {
		called++
		return nil
	})
	contextualArrow := NewCallableWithExecution(0, func(*Execution, ...any) any {
		called++
		return nil
	})
	for _, tt := range []struct {
		name  string
		value any
		want  string
	}{
		{"null", nil, "NULL"},
		{"false", false, "boolean"},
		{"true", true, "boolean"},
		{"minimum_integer", int64(-1 << 63), "integer"},
		{"maximum_integer", int64(1<<63 - 1), "integer"},
		{"empty_string", "", "string"},
		{"string", "gohp", "string"},
		{"empty_array", NewArray(), "array"},
		{"nested_array", NewArray(Entry{Key: "nested", Value: NewArray(Entry{Value: nil})}), "array"},
		{"generated_object", &getTypeObjectMarker{valid: true}, "object"},
		{"arrow", arrow, "object"},
		{"contextual_arrow", contextualArrow, "object"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetType(tt.value); got != tt.want {
				t.Fatalf("GetType(%T) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
	if called != 0 {
		t.Fatalf("GetType invoked callable body %d times", called)
	}
}

func TestGetTypeRejectsForeignValues(t *testing.T) {
	getterCalls := 0
	for _, tt := range []struct {
		name  string
		value any
	}{
		{"array_pointer", &Array{}},
		{"typed_nil_object", (*getTypeObjectMarker)(nil)},
		{"typed_nil_callable", (*Callable)(nil)},
		{"invalid_callable", &Callable{}},
		{"false_object_marker", &getTypeObjectMarker{valid: false}},
		{"getter_only", getTypeGetterOnly{calls: &getterCalls}},
		{"go_integer", int(1)},
		{"go_float", float64(1)},
		{"slice", []any{}},
		{"map", map[string]any{}},
		{"struct", struct{}{}},
		{"function", func() {}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if got := recover(); got != "gettype requires supported PHP value" {
					t.Fatalf("panic = %v, want %q", got, "gettype requires supported PHP value")
				}
			}()
			GetType(tt.value)
		})
	}
	if getterCalls != 0 {
		t.Fatalf("GetType read getter %d times", getterCalls)
	}
}
