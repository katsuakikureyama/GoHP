package runtime

import (
	"reflect"
	"testing"
)

func TestArrayPathProbeValuesAndCopy(t *testing.T) {
	input := NewArray(
		Entry{Key: "branch", Value: NewArray(
			Entry{Key: int64(-5), Value: NewArray(Entry{Key: "child", Value: "original"})},
			Entry{Key: "false", Value: false},
			Entry{Key: "zero", Value: int64(0)},
			Entry{Key: "empty", Value: ""},
			Entry{Key: "nil", Value: nil},
		)},
	)
	before := Clone(input).(Array)

	got, present := ArrayPathProbe(input, "branch")
	array, ok := got.(Array)
	if !present || !ok {
		t.Fatalf("nested array probe = (%#v, %t), want present Array", got, present)
	}
	if array.next != -4 || !array.nextInitialized || array.nextBlocked {
		t.Fatalf("nested array metadata = %#v, want preserved negative-key successor", array)
	}
	array.Entries[0].Value.(Array).Entries[0].Value = "changed"
	if got := Get(Get(input, "branch"), int64(-5)); got.(Array).Entries[0].Value != "original" {
		t.Fatalf("returned nested array aliases input: %#v", got)
	}

	for _, tt := range []struct {
		name string
		key  any
		want any
	}{
		{"false", "false", false},
		{"zero", "zero", int64(0)},
		{"empty string", "empty", ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, present := ArrayPathProbe(input, "branch", tt.key)
			if !present || !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("probe = (%#v, %t), want (%#v, true)", got, present, tt.want)
			}
		})
	}
	for _, keys := range [][]any{{"branch", "nil"}, {"branch", "missing"}} {
		if got, present := ArrayPathProbe(input, keys...); got != nil || present {
			t.Fatalf("probe(%#v) = (%#v, %t), want (nil, false)", keys, got, present)
		}
	}
	if !reflect.DeepEqual(input, before) {
		t.Fatalf("array path probe changed input: got %#v, want %#v", input, before)
	}
}

func TestArrayPathProbeBoundaries(t *testing.T) {
	bareArray := NewArray(Entry{Key: int64(-2), Value: NewArray(Entry{Key: "child", Value: "original"})})
	got, present := ArrayPathProbe(bareArray)
	copy, ok := got.(Array)
	if !present || !ok || copy.next != bareArray.next || copy.nextInitialized != bareArray.nextInitialized || copy.nextBlocked != bareArray.nextBlocked {
		t.Fatalf("bare array probe = (%#v, %t), want detached present array", got, present)
	}
	copy.Entries[0].Value.(Array).Entries[0].Value = "changed"
	if Get(bareArray.Entries[0].Value, "child") != "original" {
		t.Fatal("bare array probe did not detach nested storage")
	}
	for _, tt := range []struct {
		name  string
		value any
		want  bool
	}{
		{"null", nil, false},
		{"false", false, true},
		{"zero", int64(0), true},
		{"empty string", "", true},
	} {
		t.Run("bare_"+tt.name, func(t *testing.T) {
			got, present := ArrayPathProbe(tt.value)
			if !reflect.DeepEqual(got, tt.value) || present != tt.want {
				t.Fatalf("bare probe = (%#v, %t), want (%#v, %t)", got, present, tt.value, tt.want)
			}
		})
	}

	input := NewArray(
		Entry{Key: "nil", Value: nil},
		Entry{Key: "scalar", Value: int64(1)},
		Entry{Key: "string", Value: "text"},
		Entry{Key: "boolean", Value: false},
		Entry{Key: "object", Value: struct{}{}},
	)
	for _, keys := range [][]any{{"missing", true}, {"nil", true}} {
		if got, present := ArrayPathProbe(input, keys...); got != nil || present {
			t.Fatalf("early path %#v = (%#v, %t), want (nil, false)", keys, got, present)
		}
	}
	if got, present := ArrayPathProbe(nil, true); got != nil || present {
		t.Fatalf("nil root = (%#v, %t), want (nil, false)", got, present)
	}
	for _, tt := range []struct {
		name string
		base any
		keys []any
	}{
		{"scalar root", int64(1), []any{"key"}},
		{"scalar intermediate", input, []any{"scalar", "child"}},
		{"string intermediate", input, []any{"string", "child"}},
		{"boolean intermediate", input, []any{"boolean", "child"}},
		{"object intermediate", input, []any{"object", "child"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if got := recover(); got != "array path requires array or null" {
					t.Fatalf("panic = %v, want %q", got, "array path requires array or null")
				}
			}()
			ArrayPathProbe(tt.base, tt.keys...)
		})
	}
}

func TestArrayPathProbeRejectsReachedInvalidKeys(t *testing.T) {
	input := NewArray(Entry{Key: "branch", Value: NewArray()})
	for _, tt := range []struct {
		name string
		key  any
	}{
		{"boolean", true},
		{"null", nil},
		{"float", 1.5},
		{"array", NewArray()},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if got := recover(); got != "array path requires integer or string key" {
					t.Fatalf("panic = %v, want %q", got, "array path requires integer or string key")
				}
			}()
			ArrayPathProbe(input, tt.key)
		})
	}
}

func TestArrayPathIssetUsesPathPresence(t *testing.T) {
	input := NewArray(Entry{Key: "branch", Value: NewArray(
		Entry{Key: "false", Value: false},
		Entry{Key: "zero", Value: int64(0)},
		Entry{Key: "empty", Value: ""},
		Entry{Key: "nil", Value: nil},
		Entry{Key: "array", Value: NewArray()},
	)})
	for _, tt := range []struct {
		keys []any
		want bool
	}{
		{[]any{"branch", "false"}, true},
		{[]any{"branch", "zero"}, true},
		{[]any{"branch", "empty"}, true},
		{[]any{"branch", "array"}, true},
		{[]any{"branch", "nil"}, false},
		{[]any{"branch", "missing"}, false},
	} {
		if got := ArrayPathIsset(input, tt.keys...); got != tt.want {
			t.Fatalf("ArrayPathIsset(%#v) = %t, want %t", tt.keys, got, tt.want)
		}
	}
}
