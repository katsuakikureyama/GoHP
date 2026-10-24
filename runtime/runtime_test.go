package runtime

import (
	"math"
	"testing"
)

func TestArrayOrderingAndReplacement(t *testing.T) {
	a := NewArray(Entry{Key: "1", Value: int64(2)}, Entry{Key: int64(1), Value: int64(3)}, Entry{Value: "next"})
	if len(a.Entries) != 2 || a.Entries[0].Value != int64(3) || a.Entries[1].Key != int64(2) {
		t.Fatalf("wrong PHP array semantics: %#v", a)
	}
}
func TestCloneDoesNotAliasNestedArrays(t *testing.T) {
	a := NewArray(Entry{Value: NewArray(Entry{Value: int64(7)})})
	b := Clone(a).(Array)
	nested := b.Entries[0].Value.(Array)
	nested.Entries[0].Value = int64(9)
	if Get(Get(a, int64(0)), int64(0)) != int64(7) {
		t.Fatal("array copy aliased")
	}
}
func TestOverflowFailsExplicitly(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("overflow silently accepted")
		}
	}()
	Add(int64(9223372036854775807), int64(1))
}

func TestExplicitMaxArrayKeyOnlyBlocksFollowingAutomaticKey(t *testing.T) {
	a := NewArray(Entry{Key: int64(math.MaxInt64), Value: "edge"})
	if got := Get(a, int64(math.MaxInt64)); got != "edge" {
		t.Fatalf("explicit maximum key was not preserved: %v", got)
	}
	defer func() {
		if recover() == nil {
			t.Fatal("automatic key after maximum integer did not fail")
		}
	}()
	NewArray(Entry{Key: int64(math.MaxInt64), Value: "edge"}, Entry{Value: "overflow"})
}
