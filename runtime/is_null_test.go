package runtime

import "testing"

func TestIsNullValueDomain(t *testing.T) {
	if !IsNull(nil) {
		t.Fatal("is_null(nil) = false")
	}
	for _, value := range []any{
		false, true, int64(0), int64(-1), int64(1), "", "0", "null",
		NewArray(), NewArray(Entry{Key: "null", Value: nil}), struct{}{},
		NewCallable(0, func(...any) any { return nil }),
		(*Array)(nil), (*Callable)(nil), []any(nil), map[string]any(nil),
	} {
		if IsNull(value) {
			t.Fatalf("is_null(%T: %v) = true; must not coerce", value, value)
		}
	}
}
