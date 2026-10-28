package runtime

import (
	"reflect"
	"testing"
)

func TestListValuesSelectsNumericKeysAndCopiesArrays(t *testing.T) {
	object := &struct{ Label string }{Label: "shared"}
	input := NewArray(
		Entry{Key: int64(4), Value: "four"},
		Entry{Key: "extra", Value: "ignored"},
		Entry{Key: int64(1), Value: nil},
		Entry{Key: int64(3), Value: object},
		Entry{Key: int64(0), Value: NewArray(Entry{Key: "child", Value: "source"})},
		Entry{Key: int64(2), Value: "two"},
	)
	before := Clone(input)

	values := ListValues(input, 5)
	if len(values) != 5 {
		t.Fatalf("value count = %d, want 5", len(values))
	}
	if nested, ok := values[0].(Array); !ok || Get(nested, "child") != "source" {
		t.Fatalf("key zero = %#v", values[0])
	}
	if values[1] != nil || values[2] != "two" || values[4] != "four" {
		t.Fatalf("numeric selections = %#v", values)
	}
	if values[3] != object {
		t.Fatalf("object pointer was copied: %#v", values[3])
	}

	values[0].(Array).Entries[0].Value = "changed"
	object.Label = "changed-through-pointer"
	if got := Get(Get(input, int64(0)), "child"); got != "source" {
		t.Fatalf("nested array aliased input: %#v", got)
	}
	if got := Get(input, int64(3)); got != object || object.Label != "changed-through-pointer" {
		t.Fatalf("object pointer identity changed: %#v", got)
	}
	if !reflect.DeepEqual(input, before) {
		// The object itself is intentionally shared, so its field mutation is
		// visible through both values. Array storage must stay detached.
		t.Fatalf("ListValues changed input array storage: %#v", input)
	}
}

func TestListValuesRejectsOutsideBoundedDomain(t *testing.T) {
	for _, tt := range []struct {
		name  string
		value any
		count int
		want  string
	}{
		{"nonarray", nil, 1, "list assignment requires array"},
		{"string", "items", 1, "list assignment requires array"},
		{"zero_targets", NewArray(Entry{Value: "present"}), 0, "list assignment requires targets"},
		{"negative_targets", NewArray(Entry{Value: "present"}), -1, "list assignment requires targets"},
		{"empty", NewArray(), 1, "list assignment requires present numeric keys"},
		{"sparse", NewArray(Entry{Key: int64(1), Value: "one"}), 2, "list assignment requires present numeric keys"},
		{"string_only", NewArray(Entry{Key: "first", Value: "one"}), 1, "list assignment requires present numeric keys"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if got := recover(); got != tt.want {
					t.Fatalf("panic = %#v, want %q", got, tt.want)
				}
			}()
			ListValues(tt.value, tt.count)
		})
	}
}
