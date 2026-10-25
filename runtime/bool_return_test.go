package runtime

import "testing"

func TestBoolReturnPreservesBooleanValues(t *testing.T) {
	for _, want := range []bool{true, false} {
		if got := BoolReturn(want); got != want {
			t.Fatalf("BoolReturn(%t) = %#v, want %t", want, got, want)
		}
	}
}

func TestBoolReturnRejectsNonBooleans(t *testing.T) {
	for _, value := range []any{nil, int64(0), "", "true", NewArray(), struct{}{}} {
		t.Run("non_bool", func(t *testing.T) {
			defer func() {
				if got := recover(); got != "bool return requires bool" {
					t.Fatalf("panic = %v, want %q", got, "bool return requires bool")
				}
			}()
			BoolReturn(value)
		})
	}
}
