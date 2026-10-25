package runtime

import (
	"math"
	"reflect"
	"testing"
)

func TestArraySetCopyAndOrder(t *testing.T) {
	input := NewArray(
		Entry{Key: int64(8), Value: "first"},
		Entry{Key: "2", Value: "numeric key"},
		Entry{Key: "nested", Value: NewArray(Entry{Key: "old", Value: "original"})},
	)
	before := Clone(input)
	replacement := NewArray(Entry{Key: "child", Value: "replacement"})

	updated, ok := ArraySet(input, "2", replacement).(Array)
	if !ok {
		t.Fatal("array assignment did not return an array")
	}
	wantUpdated := []Entry{
		{Key: int64(8), Value: "first"},
		{Key: int64(2), Value: replacement},
		{Key: "nested", Value: NewArray(Entry{Key: "old", Value: "original"})},
	}
	if !reflect.DeepEqual(updated.Entries, wantUpdated) {
		t.Fatalf("replacement entries = %#v, want %#v", updated.Entries, wantUpdated)
	}
	if !reflect.DeepEqual(input, before) {
		t.Fatalf("array assignment changed source: got %#v, want %#v", input, before)
	}

	inserted := ArraySet(updated, "later", NewArray(Entry{Key: "inside", Value: "new"})).(Array)
	if got := inserted.Entries[len(inserted.Entries)-1]; !reflect.DeepEqual(got, Entry{Key: "later", Value: NewArray(Entry{Key: "inside", Value: "new"})}) {
		t.Fatalf("insertion = %#v, want final ordered entry", got)
	}
	if !reflect.DeepEqual(updated.Entries, wantUpdated) {
		t.Fatal("insertion changed the prior array value")
	}

	// Both a supplied RHS and the returned array own their nested storage.
	replacement.Entries[0].Value = "changed source rhs"
	if got := Get(updated.Entries[1].Value, "child"); got != "replacement" {
		t.Fatalf("updated array aliases RHS: %v", got)
	}
	inserted.Entries[2].Value.(Array).Entries[0].Value = "changed result"
	if got := Get(updated.Entries[2].Value, "old"); got != "original" {
		t.Fatalf("inserted array aliases its prior value: %v", got)
	}
	updated.Entries[2].Value.(Array).Entries[0].Value = "changed prior result"
	if !reflect.DeepEqual(input, before) {
		t.Fatal("updated array aliases the original nested storage")
	}
	if updated.next != input.next || updated.nextBlocked != input.nextBlocked {
		t.Fatal("replacement changed next-index metadata")
	}
	higher := ArraySet(input, int64(12), "higher").(Array)
	if higher.next != 13 || higher.nextBlocked {
		t.Fatal("inserting a larger key lost next-index metadata")
	}
	maximum := ArraySet(input, int64(math.MaxInt64), "maximum").(Array)
	maximum = ArraySet(maximum, int64(math.MaxInt64), "replaced").(Array)
	if !maximum.nextBlocked || Get(maximum, int64(math.MaxInt64)) != "replaced" {
		t.Fatal("maximum-key replacement lost its blocked next index")
	}
	negative := ArraySet(NewArray(), int64(-5), "negative").(Array)
	if negative.next != -4 {
		t.Fatal("PHP 8.3 empty-array negative-key metadata was lost")
	}
}

func TestArraySetRejectsInvalidDomain(t *testing.T) {
	for _, tt := range []struct {
		name string
		call func()
		want string
	}{
		{"null_base", func() { ArraySet(nil, "key", "value") }, "array assignment requires array"},
		{"scalar_base", func() { ArraySet(int64(1), "key", "value") }, "array assignment requires array"},
		{"boolean_key", func() { ArraySet(NewArray(), true, "value") }, "unsupported array key type"},
		{"null_key", func() { ArraySet(NewArray(), nil, "value") }, "unsupported array key type"},
		{"float_key", func() { ArraySet(NewArray(), 1.5, "value") }, "unsupported array key type"},
		{"array_key", func() { ArraySet(NewArray(), NewArray(), "value") }, "unsupported array key type"},
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
