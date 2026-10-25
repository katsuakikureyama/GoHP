package runtime

import "testing"

func TestBoolReturnActualBooleans(t *testing.T) {
	for _, value := range []bool{false, true} {
		if got := BoolReturn(value); got != value {
			t.Fatalf("BoolReturn(%v) = %#v", value, got)
		}
	}
}

func TestBoolReturnRejectsCoercion(t *testing.T) {
	for _, value := range []any{nil, int64(0), int64(1), "", "0", "true", NewArray(), struct{}{}} {
		func() {
			defer func() {
				if got := recover(); got != "bool return requires bool" {
					t.Errorf("BoolReturn(%#v) panic = %v", value, got)
				}
			}()
			BoolReturn(value)
		}()
	}
}
