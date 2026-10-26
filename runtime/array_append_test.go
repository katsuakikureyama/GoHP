package runtime

import (
	"math"
	"reflect"
	"testing"
)

func TestArrayAppendKeysAndValueIsolation(t *testing.T) {
	rhs := NewArray(Entry{Key: "deep", Value: "rhs"})
	source := NewArray(
		Entry{Key: "label", Value: "source"},
		Entry{Key: "nested", Value: NewArray(Entry{Key: "deep", Value: "source nested"})},
	)
	before := Clone(source)

	appended, ok := ArrayAppend(source, rhs).(Array)
	if !ok {
		t.Fatal("array append did not return an array")
	}
	want := []Entry{
		{Key: "label", Value: "source"},
		{Key: "nested", Value: NewArray(Entry{Key: "deep", Value: "source nested"})},
		{Key: int64(0), Value: NewArray(Entry{Key: "deep", Value: "rhs"})},
	}
	if !reflect.DeepEqual(appended.Entries, want) {
		t.Fatalf("entries = %#v, want %#v", appended.Entries, want)
	}
	if !reflect.DeepEqual(source, before) {
		t.Fatalf("append changed source: got %#v, want %#v", source, before)
	}

	// The appended RHS and the returned root each own their nested storage.
	rhs.Entries[0].Value = "changed rhs"
	if got := Get(Get(appended, int64(0)), "deep"); got != "rhs" {
		t.Fatalf("append result aliases RHS: %v", got)
	}
	appended.Entries[1].Value.(Array).Entries[0].Value = "changed result"
	if got := Get(Get(source, "nested"), "deep"); got != "source nested" {
		t.Fatalf("append result aliases source: %v", got)
	}

	negative := ArrayAppend(NewArray(Entry{Key: "-5", Value: "negative"}), "next").(Array)
	if got := negative.Entries[1]; !reflect.DeepEqual(got, Entry{Key: int64(-4), Value: "next"}) {
		t.Fatalf("negative append = %#v, want key -4", got)
	}
}

func TestArrayAppendSelfSnapshotAndMaximum(t *testing.T) {
	self := NewArray(Entry{Key: "entry", Value: NewArray(Entry{Key: "deep", Value: "before"})})
	appended := ArrayAppend(self, self).(Array)
	// Inspect stored storage directly: Get would clone and hide an aliasing bug.
	copyOfSelf := appended.Entries[1].Value.(Array)
	if len(copyOfSelf.Entries) != 1 {
		t.Fatal("self append did not retain the pre-write snapshot")
	}
	copyOfSelf.Entries[0].Value.(Array).Entries[0].Value = "changed appended snapshot"
	if got := Get(Get(appended, "entry"), "deep"); got != "before" {
		t.Fatalf("self append aliases its root value: %v", got)
	}
	if got := Get(Get(self, "entry"), "deep"); got != "before" {
		t.Fatalf("self append changed source: %v", got)
	}

	nearMaximum := NewArray(Entry{Key: int64(math.MaxInt64 - 1), Value: "before maximum"})
	atMaximum := ArrayAppend(nearMaximum, "maximum").(Array)
	if got := atMaximum.Entries[1]; !reflect.DeepEqual(got, Entry{Key: int64(math.MaxInt64), Value: "maximum"}) {
		t.Fatalf("maximum append = %#v, want maximum key", got)
	}
	assertArrayAppendPanic(t, func() { ArrayAppend(atMaximum, "overflow") }, "array key overflow unsupported")
}

func TestArrayAppendRejectsNonArrays(t *testing.T) {
	for _, tt := range []struct {
		name string
		base any
	}{
		{"nil", nil},
		{"integer", int64(1)},
		{"string", "array"},
		{"boolean", false},
		{"object", struct{}{}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assertArrayAppendPanic(t, func() { ArrayAppend(tt.base, "value") }, "array append requires array")
		})
	}
}

func assertArrayAppendPanic(t *testing.T, call func(), want string) {
	t.Helper()
	defer func() {
		if got := recover(); got != want {
			t.Fatalf("panic = %v, want %q", got, want)
		}
	}()
	call()
}
