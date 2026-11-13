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

func TestVariableEmptyRejectsUnsupportedValues(t *testing.T) {
	type foreignString string
	for _, input := range []any{
		nil, false, true, int64(0), int64(1), 0, 0.0, foreignString(""),
		[]byte{}, []any{}, &struct{}{}, (*Array)(nil), (*string)(nil),
		NewCallable(0, func(...any) any { return nil }),
	} {
		func() {
			defer func() {
				if got := recover(); got != "empty requires array or string" {
					t.Fatalf("unsupported %T rejection: %v", input, got)
				}
			}()
			VariableEmpty(input)
		}()
	}
}
