package runtime

import (
	"math"
	"reflect"
	"testing"
)

func TestArrayMapPreservesKeysOrderAndValueIsolation(t *testing.T) {
	input := NewArray(
		Entry{Key: int64(9), Value: "first"},
		Entry{Key: "2", Value: "normalized numeric key"},
		Entry{Key: "name", Value: NewArray(Entry{Key: "nested", Value: "original"})},
		Entry{Key: int64(3), Value: nil},
	)
	before := Clone(input)
	var calls []any
	var callbackReturn any
	got, ok := ArrayMap(func(value any) any {
		calls = append(calls, Clone(value))
		if nested, ok := value.(Array); ok {
			nested.Entries[0].Value = "changed by callback"
			callbackReturn = nested
			return nested
		}
		return value
	}, input).(Array)
	if !ok {
		t.Fatal("array_map did not return an array")
	}
	if got.index != input.index {
		t.Fatal("array_map lost next-index metadata")
	}
	want := []Entry{
		{Key: int64(9), Value: "first"},
		{Key: int64(2), Value: "normalized numeric key"},
		{Key: "name", Value: NewArray(Entry{Key: "nested", Value: "changed by callback"})},
		{Key: int64(3), Value: nil},
	}
	if !reflect.DeepEqual(got.Entries, want) {
		t.Fatalf("array_map entries = %#v, want %#v", got.Entries, want)
	}
	if !reflect.DeepEqual(calls, []any{"first", "normalized numeric key", NewArray(Entry{Key: "nested", Value: "original"}), nil}) {
		t.Fatalf("callback calls = %#v, want exactly one per input entry", calls)
	}
	if !reflect.DeepEqual(input, before) {
		t.Fatalf("array_map changed input: got %#v, want %#v", input, before)
	}
	callbackReturn.(Array).Entries[0].Value = "changed after callback"
	if Get(got.Entries[2].Value, "nested") != "changed by callback" {
		t.Fatal("array_map result aliases callback return storage")
	}
	got.Entries[2].Value.(Array).Entries[0].Value = "changed result"
	if !reflect.DeepEqual(input, before) {
		t.Fatal("array_map result aliases input storage")
	}
	maxKey := NewArray(Entry{Key: int64(math.MaxInt64), Value: "last"})
	maxMapped := ArrayMap(func(value any) any { return value }, maxKey).(Array)
	if !reflect.DeepEqual(maxMapped, maxKey) {
		t.Fatal("maximum integer key or blocked next-index metadata was lost")
	}
	assertArrayAppendPanic(t, func() { ArrayAppend(maxMapped, nil) }, "array key overflow unsupported")
}

func TestArrayMapEmptyAndInvalidArguments(t *testing.T) {
	calls := 0
	empty, ok := ArrayMap(func(any) any { calls++; return "unreachable" }, NewArray()).(Array)
	if !ok || len(empty.Entries) != 0 || calls != 0 {
		t.Fatalf("empty array_map = %#v with %d callback calls, want empty and zero", empty, calls)
	}
	for _, tt := range []struct {
		name string
		call func()
		want string
	}{
		{"non_array", func() { ArrayMap(func(any) any { return nil }, "not array") }, "array_map requires an array"},
		{"nil_callback", func() { ArrayMap(nil, NewArray()) }, "array_map requires a callback"},
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
