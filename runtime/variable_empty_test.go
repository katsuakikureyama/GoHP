package runtime

import (
	"reflect"
	"testing"
)

func TestVariableEmptyValues(t *testing.T) {
	for _, test := range []struct {
		input string
		want  bool
	}{
		{"", true}, {"0", true}, {"00", false}, {"0.0", false}, {"-0", false},
		{"+0", false}, {" 0", false}, {"0 ", false}, {" ", false}, {"false", false},
		{"text", false}, {"é", false}, {"０", false}, {"\x00", false}, {"0\x00", false}, {"\xff", false},
	} {
		before := test.input
		if got := VariableEmpty(test.input); got != test.want || test.input != before {
			t.Fatalf("string %q: got %v want %v, preservation=%v", test.input, got, test.want, test.input == before)
		}
	}
	for _, test := range []struct {
		input Array
		want  bool
	}{
		{Array{}, true}, {NewArray(), true}, {NewArray(Entry{Value: nil}), false},
		{NewArray(Entry{Value: false}), false}, {NewArray(Entry{Value: "0"}), false},
		{NewArray(Entry{Value: NewArray()}), false},
		{NewArray(Entry{Key: int64(-5), Value: NewArray(Entry{Key: "nested", Value: "0"})}), false},
	} {
		before := Clone(test.input)
		if got := VariableEmpty(test.input); got != test.want {
			t.Fatalf("array outer count: got %v for %#v", got, test.input)
		}
		if !reflect.DeepEqual(test.input, before) {
			t.Fatal("empty mutated array values, keys or append metadata")
		}
	}
}

func TestVariableEmptyScalars(t *testing.T) {
	// null is empty
	if !VariableEmpty(nil) {
		t.Fatal("empty(null) should be true")
	}
	// false is empty
	if !VariableEmpty(false) {
		t.Fatal("empty(false) should be true")
	}
	// true is not empty
	if VariableEmpty(true) {
		t.Fatal("empty(true) should be false")
	}
	// int64(0) is empty
	if !VariableEmpty(int64(0)) {
		t.Fatal("empty(0) should be true")
	}
	// int64(1) is not empty
	if VariableEmpty(int64(1)) {
		t.Fatal("empty(1) should be false")
	}
}

// Foreign Go values (floats, plain ints, byte slices, typed nil pointers,
// foreign string aliases) stay outside the closed PHP value domain.
func TestVariableEmptyRejectsUnsupportedValues(t *testing.T) {
	type foreignString string
	for _, input := range []any{
		0, 0.0, foreignString(""), []byte{}, []any{}, &struct{}{}, (*Array)(nil), (*string)(nil),
	} {
		func() {
			defer func() {
				if got := recover(); got != "empty requires a supported PHP value" {
					t.Fatalf("unsupported %T rejection: %v", input, got)
				}
			}()
			VariableEmpty(input)
		}()
	}
}

func TestVariableEmptyObjectsAreNotEmpty(t *testing.T) {
	if VariableEmpty(NewCallable(0, func(...any) any { return nil })) {
		t.Fatal("a generated arrow (PHP Closure) is never empty")
	}
}
