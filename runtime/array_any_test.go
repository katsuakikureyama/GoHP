package runtime

import (
	"reflect"
	"testing"
)

func TestArrayAnyVisitsInsertionOrderAndShortCircuits(t *testing.T) {
	input := NewArray(
		Entry{Key: int64(-2), Value: "first"},
		Entry{Key: "08", Value: "second"},
		Entry{Key: "label", Value: "unreachable"},
	)
	before := Clone(input)
	var calls []Entry
	got := ArrayAny(input, func(value, key any) any {
		calls = append(calls, Entry{Key: key, Value: value})
		return value == "second"
	})
	if !got {
		t.Fatal("array_any did not return true for the matching entry")
	}
	wantCalls := []Entry{
		{Key: int64(-2), Value: "first"},
		{Key: "08", Value: "second"},
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
		t.Fatalf("array_any changed input: got %#v, want %#v", input, before)
	}
}

func TestArrayAnyFalseForEmptyAndNoMatch(t *testing.T) {
	calls := 0
	if ArrayAny(NewArray(), func(any, any) any {
		calls++
		return true
	}) {
		t.Fatal("empty array_any = true, want false")
	}
	if calls != 0 {
		t.Fatalf("empty array_any made %d callback calls, want none", calls)
	}

	input := NewArray(Entry{Key: "first", Value: int64(1)}, Entry{Key: int64(4), Value: nil})
	if ArrayAny(input, func(value, key any) any {
		calls++
		return false
	}) {
		t.Fatal("array_any = true when every callback result was false")
	}
	if calls != 2 {
		t.Fatalf("non-matching array_any made %d callback calls, want 2", calls)
	}
}

func TestArrayAnyClonesNestedValuesForCallback(t *testing.T) {
	input := NewArray(
		Entry{Key: "nested", Value: NewArray(Entry{Key: "state", Value: "original"})},
		Entry{Key: "tail", Value: "tail"},
	)
	before := Clone(input)
	if ArrayAny(input, func(value, key any) any {
		if key == "nested" {
			nested, ok := value.(Array)
			if !ok {
				t.Fatalf("nested callback value type = %T, want runtime.Array", value)
			}
			nested.Entries[0].Value = "changed in callback"
		}
		return false
	}) {
		t.Fatal("array_any unexpectedly matched")
	}
	if !reflect.DeepEqual(input, before) {
		t.Fatalf("callback mutation aliased input: got %#v, want %#v", input, before)
	}
	if got := Get(input.Entries[0].Value, "state"); got != "original" {
		t.Fatalf("nested input value = %#v, want original", got)
	}
}

func TestArrayAnyRejectsUnsupportedArgumentsAndResults(t *testing.T) {
	for _, tt := range []struct {
		name string
		call func()
		want string
	}{
		{
			name: "non_array",
			call: func() { ArrayAny("not an array", func(any, any) any { return false }) },
			want: "array_any requires an array",
		},
		{
			name: "nil_callback_even_for_empty_array",
			call: func() { ArrayAny(NewArray(), nil) },
			want: "array_any requires a callback",
		},
		{
			name: "string_result",
			call: func() { ArrayAny(NewArray(Entry{Value: "value"}), func(any, any) any { return "true" }) },
			want: "array_any callback must return bool",
		},
		{
			name: "nil_result",
			call: func() { ArrayAny(NewArray(Entry{Value: "value"}), func(any, any) any { return nil }) },
			want: "array_any callback must return bool",
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
