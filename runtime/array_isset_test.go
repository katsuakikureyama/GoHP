package runtime

import (
	"reflect"
	"testing"
)

func TestArrayIssetPresenceNilAndFalsyValues(t *testing.T) {
	input := NewArray(
		Entry{Key: "nil", Value: nil},
		Entry{Key: "false", Value: false},
		Entry{Key: "zero", Value: int64(0)},
		Entry{Key: "string-zero", Value: "0"},
		Entry{Key: "empty-string", Value: ""},
		Entry{Key: "empty-array", Value: NewArray()},
		Entry{Key: "nested", Value: NewArray(Entry{Key: "child", Value: nil})},
	)
	before := Clone(input)

	for _, tt := range []struct {
		key  any
		want bool
	}{
		{"missing", false},
		{"nil", false},
		{"false", true},
		{"zero", true},
		{"string-zero", true},
		{"empty-string", true},
		{"empty-array", true},
		{"nested", true},
	} {
		if got := ArrayIsset(input, tt.key); got != tt.want {
			t.Errorf("ArrayIsset(%#v) = %t, want %t", tt.key, got, tt.want)
		}
	}
	if !reflect.DeepEqual(input, before) {
		t.Fatalf("ArrayIsset changed input: got %#v, want %#v", input, before)
	}
}

func TestArrayIssetNormalizesIntegerStringKeys(t *testing.T) {
	input := NewArray(
		Entry{Key: int64(2), Value: "two"},
		Entry{Key: "02", Value: "leading zero"},
		Entry{Key: int64(-5), Value: "negative"},
		Entry{Key: "nul\x00key", Value: "nul"},
		Entry{Key: "日本語", Value: "utf8"},
	)

	for _, tt := range []struct {
		name string
		key  any
		want bool
	}{
		{"canonical numeric string", "2", true},
		{"integer numeric key", int64(2), true},
		{"noncanonical leading zero", "02", true},
		{"negative numeric string", "-5", true},
		{"negative integer", int64(-5), true},
		{"nul key", "nul\x00key", true},
		{"utf8 key", "日本語", true},
		{"unrelated key", "3", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := ArrayIsset(input, tt.key); got != tt.want {
				t.Fatalf("ArrayIsset(%#v) = %t, want %t", tt.key, got, tt.want)
			}
		})
	}
	if got := ArrayIsset(NewArray(Entry{Key: "02", Value: "leading zero only"}), int64(2)); got {
		t.Fatal("noncanonical numeric string key unexpectedly normalized")
	}
}

func TestArrayIssetRejectsOutsideBoundedDomain(t *testing.T) {
	for _, tt := range []struct {
		name string
		call func()
		want string
	}{
		{"nil base", func() { ArrayIsset(nil, "key") }, "isset requires array"},
		{"scalar base", func() { ArrayIsset(int64(1), "key") }, "isset requires array"},
		{"boolean key", func() { ArrayIsset(NewArray(), true) }, "isset requires integer or string key"},
		{"nil key", func() { ArrayIsset(NewArray(), nil) }, "isset requires integer or string key"},
		{"float key", func() { ArrayIsset(NewArray(), 1.5) }, "isset requires integer or string key"},
		{"array key", func() { ArrayIsset(NewArray(), NewArray()) }, "isset requires integer or string key"},
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
