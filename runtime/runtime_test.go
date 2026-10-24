package runtime

import (
	"math"
	"strings"
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

func TestRTrimUsesBytesAndLiteralMask(t *testing.T) {
	if got := RTrim("path///", "/"); got != "path" {
		t.Fatalf("RTrim slash mask = %#v, want path", got)
	}
	if got := RTrim("stay", ""); got != "stay" {
		t.Fatalf("RTrim empty mask = %#v, want stay", got)
	}
	// The final byte of both UTF-8 encodings is 0xa9. Byte-wise PHP trimming
	// removes it, even though the remaining string is not valid UTF-8.
	if got := RTrim("é", "©"); got != string([]byte{0xc3}) {
		t.Fatalf("RTrim byte mask = % x, want c3", []byte(got.(string)))
	}
}

func TestRTrimRejectsDynamicNonStringAndRanges(t *testing.T) {
	for _, tt := range []struct {
		name string
		fn   func()
	}{
		{name: "non_string", fn: func() { RTrim(int64(1), "1") }},
		{name: "range", fn: func() { RTrim("x", "a..z") }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if r == nil || !strings.Contains(r.(string), "rtrim") {
					t.Fatalf("RTrim did not reject unsupported input: %v", r)
				}
			}()
			tt.fn()
		})
	}
}

func TestBoolRejectsUnsupportedValue(t *testing.T) {
	defer func() {
		if r := recover(); r == nil || !strings.Contains(r.(string), "boolean conversion") {
			t.Fatalf("Bool did not reject unsupported truthiness value: %v", r)
		}
	}()
	Bool(struct{}{})
}

func TestIdenticalRejectsArrayValues(t *testing.T) {
	defer func() {
		if r := recover(); r == nil || !strings.Contains(r.(string), "strict equality") {
			t.Fatalf("Identical did not reject array identity: %v", r)
		}
	}()
	Identical(NewArray(), NewArray())
}

func TestStrToLowerRejectsUnsupportedValue(t *testing.T) {
	defer func() {
		if r := recover(); r == nil || !strings.Contains(r.(string), "strtolower requires string") {
			t.Fatalf("StrToLower did not reject unsupported input: %v", r)
		}
	}()
	StrToLower(int64(1))
}
