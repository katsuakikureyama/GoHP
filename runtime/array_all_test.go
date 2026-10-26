package runtime

import (
	"reflect"
	"testing"
)

func TestArrayAllEmptyDoesNotCallCallback(t *testing.T) {
	calls := 0
	if !ArrayAll(NewArray(), func(any, any) any {
		calls++
		return false
	}) {
		t.Fatal("empty array_all = false, want true")
	}
	if calls != 0 {
		t.Fatalf("empty array_all made %d callback calls, want none", calls)
	}
}

func TestArrayAllVisitsInsertionOrderWithNormalizedKeys(t *testing.T) {
	input := NewArray(
		Entry{Key: "-2", Value: "first"},
		Entry{Key: "08", Value: "second"},
		Entry{Key: "label", Value: "third"},
	)
	before := Clone(input)
	var calls []Entry
	if !ArrayAll(input, func(value, key any) any {
		calls = append(calls, Entry{Key: key, Value: value})
		return true
	}) {
		t.Fatal("array_all = false when every callback result was true")
	}
	wantCalls := []Entry{
		{Key: int64(-2), Value: "first"},
		{Key: "08", Value: "second"},
		{Key: "label", Value: "third"},
	}
	if !reflect.DeepEqual(calls, wantCalls) {
		t.Fatalf("callback calls = %#v, want %#v", calls, wantCalls)
	}
	if _, ok := calls[0].Key.(int64); !ok {
		t.Fatalf("first callback key type = %T, want int64", calls[0].Key)
	}
	if _, ok := calls[1].Key.(string); !ok {
		t.Fatalf("second callback key type = %T, want string", calls[1].Key)
	}
	if !reflect.DeepEqual(input, before) {
		t.Fatalf("array_all changed input: got %#v, want %#v", input, before)
	}
}

func TestArrayAllShortCircuitsAtFirstFalseResult(t *testing.T) {
	for _, tt := range []struct {
		name     string
		input    Array
		falseKey any
		want     []any
	}{
		{
			name:     "first",
			input:    NewArray(Entry{Key: "first", Value: int64(1)}, Entry{Key: "second", Value: int64(2)}),
			falseKey: "first",
			want:     []any{"first"},
		},
		{
			name:     "middle",
			input:    NewArray(Entry{Key: int64(0), Value: int64(1)}, Entry{Key: "middle", Value: int64(2)}, Entry{Key: "last", Value: int64(3)}),
			falseKey: "middle",
			want:     []any{int64(0), "middle"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var visited []any
			if ArrayAll(tt.input, func(_ any, key any) any {
				visited = append(visited, key)
				return key != tt.falseKey
			}) {
				t.Fatal("array_all = true after a false callback result")
			}
			if !reflect.DeepEqual(visited, tt.want) {
				t.Fatalf("visited keys = %#v, want %#v", visited, tt.want)
			}
		})
	}
}

func TestArrayAllClonesNestedValuesForCallback(t *testing.T) {
	input := NewArray(
		Entry{Key: "nested", Value: NewArray(Entry{Key: "state", Value: "original"})},
		Entry{Key: "tail", Value: "tail"},
	)
	before := Clone(input)
	if !ArrayAll(input, func(value, key any) any {
		if key == "nested" {
			nested, ok := value.(Array)
			if !ok {
				t.Fatalf("nested callback value type = %T, want runtime.Array", value)
			}
			nested.Entries[0].Value = "changed in callback"
		}
		return true
	}) {
		t.Fatal("array_all unexpectedly returned false")
	}
	if !reflect.DeepEqual(input, before) {
		t.Fatalf("callback mutation aliased input: got %#v, want %#v", input, before)
	}
	if got := Get(input.Entries[0].Value, "state"); got != "original" {
		t.Fatalf("nested input value = %#v, want original", got)
	}
}

func TestArrayAllRejectsUnsupportedArgumentsAndResults(t *testing.T) {
	for _, tt := range []struct {
		name string
		call func()
		want string
	}{
		{
			name: "non_array",
			call: func() { ArrayAll("not an array", func(any, any) any { return true }) },
			want: "array_all requires an array",
		},
		{
			name: "nil_callback_even_for_empty_array",
			call: func() { ArrayAll(NewArray(), nil) },
			want: "array_all requires a callback",
		},
		{
			name: "string_result",
			call: func() { ArrayAll(NewArray(Entry{Value: "value"}), func(any, any) any { return "true" }) },
			want: "array_all callback must return bool",
		},
		{
			name: "nil_result",
			call: func() { ArrayAll(NewArray(Entry{Value: "value"}), func(any, any) any { return nil }) },
			want: "array_all callback must return bool",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if got := recover(); got != tt.want {
					t.Fatalf("panic = %v, want %q", got, tt.want)
				}
			}()
			tt.call()
		})
	}
}
