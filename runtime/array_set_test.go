package runtime

import (
	"math"
	"reflect"
	"testing"
)

func TestArraySetReplacesInOrderAndAppendsExplicitNormalizedKeys(t *testing.T) {
	input := NewArray(
		Entry{Key: int64(2), Value: "first"},
		Entry{Key: "name", Value: "kept"},
		Entry{Key: int64(-3), Value: false},
	)
	before := Clone(input)

	replaced, ok := ArraySet(input, "2", "replacement").(Array)
	if !ok {
		t.Fatal("array_set did not return an array")
	}
	wantReplaced := []Entry{
		{Key: int64(2), Value: "replacement"},
		{Key: "name", Value: "kept"},
		{Key: int64(-3), Value: false},
	}
	if !reflect.DeepEqual(replaced.Entries, wantReplaced) {
		t.Fatalf("replacement entries = %#v, want %#v", replaced.Entries, wantReplaced)
	}
	if !reflect.DeepEqual(input, before) {
		t.Fatal("array_set replacement modified the input")
	}

	inserted, ok := ArraySet(replaced, "8", true).(Array)
	if !ok {
		t.Fatal("array_set insertion did not return an array")
	}
	wantInserted := append(append([]Entry(nil), wantReplaced...), Entry{Key: int64(8), Value: true})
	if !reflect.DeepEqual(inserted.Entries, wantInserted) {
		t.Fatalf("insertion entries = %#v, want %#v", inserted.Entries, wantInserted)
	}
	if inserted.next != 9 || inserted.nextBlocked {
		t.Fatalf("explicit numeric insertion metadata = next %d blocked %v, want next 9 unblocked", inserted.next, inserted.nextBlocked)
	}
	if !reflect.DeepEqual(replaced.Entries, wantReplaced) {
		t.Fatal("array_set insertion modified its source array")
	}
}

func TestArraySetCopiesNestedInputAndAssignedValueIncludingSelfSnapshot(t *testing.T) {
	input := NewArray(
		Entry{Key: "nested", Value: NewArray(Entry{Key: "inside", Value: "original"})},
	)
	rightHandSide := NewArray(Entry{Key: "rhs", Value: NewArray(Entry{Key: "deep", Value: "value"})})
	beforeInput := Clone(input)
	beforeRHS := Clone(rightHandSide)

	got := ArraySet(input, "assigned", rightHandSide).(Array)
	got.Entries[0].Value.(Array).Entries[0].Value = "changed result input copy"
	got.Entries[1].Value.(Array).Entries[0].Value.(Array).Entries[0].Value = "changed result rhs copy"
	if !reflect.DeepEqual(input, beforeInput) {
		t.Fatal("result nested mutation changed source input")
	}
	if !reflect.DeepEqual(rightHandSide, beforeRHS) {
		t.Fatal("result nested mutation changed assigned value")
	}
	input.Entries[0].Value.(Array).Entries[0].Value = "changed input"
	rightHandSide.Entries[0].Value.(Array).Entries[0].Value = "changed rhs"
	if Get(got.Entries[0].Value, "inside") != "changed result input copy" {
		t.Fatal("source input mutation changed result nested value")
	}
	if Get(got.Entries[1].Value.(Array).Entries[0].Value, "deep") != "changed result rhs copy" {
		t.Fatal("assigned value mutation changed result nested value")
	}

	selfSource := NewArray(Entry{Key: "old", Value: NewArray(Entry{Key: "child", Value: "snapshot"})})
	selfResult := ArraySet(selfSource, "self", selfSource).(Array)
	selfSnapshot := Get(selfResult, "self").(Array)
	if ArrayKeyExists("self", selfSnapshot) || Get(selfSnapshot, "old").(Array).Entries[0].Value != "snapshot" {
		t.Fatalf("self assignment did not retain the pre-assignment snapshot: %#v", selfSnapshot)
	}
	selfSource.Entries[0].Value.(Array).Entries[0].Value = "changed source"
	selfSnapshot.Entries[0].Value.(Array).Entries[0].Value = "changed snapshot"
	if Get(selfResult.Entries[0].Value, "child") != "snapshot" {
		t.Fatal("self source mutation changed result outer copy")
	}
	if Get(selfSource.Entries[0].Value, "child") != "changed source" {
		t.Fatal("self snapshot mutation changed source")
	}
}

func TestArraySetPreservesAndUpdatesExplicitKeyMetadataAtMaxInt64(t *testing.T) {
	input := NewArray(Entry{Key: int64(math.MaxInt64 - 1), Value: "before max"})
	if input.next != math.MaxInt64 || input.nextBlocked {
		t.Fatalf("unexpected source metadata: next %d blocked %v", input.next, input.nextBlocked)
	}

	got := ArraySet(input, int64(math.MaxInt64), "at max").(Array)
	if got.next != math.MaxInt64 || !got.nextBlocked {
		t.Fatalf("max-key metadata = next %d blocked %v, want max and blocked", got.next, got.nextBlocked)
	}
	if input.next != math.MaxInt64 || input.nextBlocked {
		t.Fatal("max-key insertion changed source metadata")
	}
	preserved := ArraySet(got, "name", "later").(Array)
	if preserved.next != math.MaxInt64 || !preserved.nextBlocked {
		t.Fatalf("string-key assignment did not preserve max metadata: %#v", preserved)
	}
}

func TestArraySetRejectsInvalidArrayAndKeyTypes(t *testing.T) {
	for _, tt := range []struct {
		name  string
		array any
	}{
		{"null", nil}, {"boolean", false}, {"integer", int64(1)},
		{"string", "x"}, {"float", 1.5}, {"object", &struct{}{}},
	} {
		t.Run("array_"+tt.name, func(t *testing.T) {
			defer func() {
				if got := recover(); got != "array_set requires array" {
					t.Fatalf("non-array was not explicitly rejected: %v", got)
				}
			}()
			ArraySet(tt.array, "key", nil)
		})
	}
	for _, tt := range []struct {
		name string
		key  any
	}{
		{"null", nil}, {"boolean", true}, {"float", 1.5},
		{"array", NewArray()}, {"object", &struct{}{}},
	} {
		t.Run("key_"+tt.name, func(t *testing.T) {
			defer func() {
				if got := recover(); got != "unsupported array key type" {
					t.Fatalf("unsupported key was not explicitly rejected: %v", got)
				}
			}()
			ArraySet(NewArray(), tt.key, nil)
		})
	}
}
