package runtime

import "testing"

func TestEmptyValueStringAndArrayDomain(t *testing.T) {
	for _, s := range []string{"", "0"} {
		if !EmptyValue(s) {
			t.Fatalf("string %q must be empty", s)
		}
	}
	for _, s := range []string{"00", "0.0", " 0", "0 ", " ", "\n", "-0", "0\x00", "\x00", "é", "\xff"} {
		if EmptyValue(s) {
			t.Fatalf("string %q must be nonempty", s)
		}
	}
	if !EmptyValue(NewArray()) || EmptyValue(NewArray(Entry{Value: nil})) {
		t.Fatal("array outer count changed")
	}
}

func TestEmptyValueRejectsOtherActualDomains(t *testing.T) {
	for _, value := range []any{nil, false, true, int64(0), int(0), float64(0), struct{}{}, (*Array)(nil), NewCallable(0, func(...any) any { return nil })} {
		func() {
			defer func() {
				if got := recover(); got != "empty requires array or string" {
					t.Errorf("domain %#v: %v", value, got)
				}
			}()
			EmptyValue(value)
		}()
	}
}
