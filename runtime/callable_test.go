package runtime

import (
	"reflect"
	"testing"
)

func TestGeneratedCallableCopiesArgumentsAndResult(t *testing.T) {
	input := NewArray(Entry{Key: int64(-2), Value: NewArray(Entry{Key: "nested", Value: "original"})})
	before := Clone(input).(Array)
	var received Array
	callback := NewCallable(2, func(args ...any) any {
		if len(args) != 2 || args[1] != "marker" {
			t.Fatalf("callback arguments = %#v, want array then marker", args)
		}
		received = args[0].(Array)
		received.Entries[0].Value.(Array).Entries[0].Value = "changed by callback"
		return received
	})

	got, ok := Invoke(callback, input, "marker").(Array)
	if !ok {
		t.Fatal("callable result did not remain an array")
	}
	if gotValue := Get(Get(got, int64(-2)), "nested"); gotValue != "changed by callback" {
		t.Fatalf("callable result = %v, want detached callback change", gotValue)
	}
	if !reflect.DeepEqual(input, before) {
		t.Fatalf("callable argument aliases input: got %#v, want %#v", input, before)
	}
	got.Entries[0].Value.(Array).Entries[0].Value = "changed result"
	if gotValue := Get(Get(received, int64(-2)), "nested"); gotValue != "changed by callback" {
		t.Fatalf("callable result aliases callback return: got %v", gotValue)
	}
}

func TestGeneratedCallablePreservesArgumentCountAndOrder(t *testing.T) {
	var calls [][]any
	callback := NewCallable(2, func(args ...any) any {
		calls = append(calls, append([]any(nil), args...))
		return int64(len(calls))
	})

	if got := Invoke(callback, "first", int64(2)); got != int64(1) {
		t.Fatalf("first invocation = %#v, want 1", got)
	}
	if got := Invoke(callback, "second", int64(3)); got != int64(2) {
		t.Fatalf("second invocation = %#v, want 2", got)
	}
	want := [][]any{{"first", int64(2)}, {"second", int64(3)}}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("callback calls = %#v, want %#v", calls, want)
	}
}

func TestGeneratedCallableRejectsInvalidValuesAndArityBeforeCall(t *testing.T) {
	for _, tt := range []struct {
		name  string
		value any
	}{
		{"null", nil},
		{"boolean", true},
		{"integer", int64(1)},
		{"string", "callback"},
		{"array", NewArray()},
		{"object", struct{}{}},
		{"function", func(...any) any { return nil }},
		{"typed_nil", (*Callable)(nil)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if got := recover(); got != "callable requires generated arrow" {
					t.Fatalf("panic = %v, want %q", got, "callable requires generated arrow")
				}
			}()
			CallableParameter(tt.value)
		})
	}

	calls := 0
	callback := NewCallable(1, func(args ...any) any {
		calls++
		return nil
	})
	defer func() {
		if got := recover(); got != "callable argument count mismatch" {
			t.Fatalf("panic = %v, want %q", got, "callable argument count mismatch")
		}
		if calls != 0 {
			t.Fatalf("callback ran %d times before arity rejection", calls)
		}
	}()
	Invoke(callback)
}

func TestGeneratedCallableIsAnOpaqueCloneValue(t *testing.T) {
	callback := NewCallable(0, func(...any) any { return true })
	if got := Clone(callback); got != callback {
		t.Fatalf("Clone(callable) = %#v, want the same generated callable", got)
	}
	if got := CallableParameter(callback); got != callback {
		t.Fatalf("callable parameter = %#v, want the same generated callable", got)
	}
}
