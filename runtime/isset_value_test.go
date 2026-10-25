package runtime

import "testing"

func TestIssetValueDomain(t *testing.T) {
	if IssetValue(nil) {
		t.Fatal("IssetValue(nil) = true, want false")
	}
	for _, value := range []any{
		false, true, int64(0), int64(-1), int64(1), "", "0", "text",
		NewArray(), NewArray(Entry{Key: "null", Value: nil}), struct{}{},
		NewCallable(0, func(...any) any { return nil }),
		(*Array)(nil), (*Callable)(nil), []any(nil), map[string]any(nil),
	} {
		if !IssetValue(value) {
			t.Fatalf("IssetValue(%T: %v) = false, want true without coercion", value, value)
		}
	}
}
