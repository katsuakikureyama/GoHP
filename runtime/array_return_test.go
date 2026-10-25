package runtime

import (
	"reflect"
	"testing"
)

func TestArrayReturnCopiesArrayValuesAndMetadata(t *testing.T) {
	input := NewArray(
		Entry{Key: int64(-3), Value: false},
		Entry{Key: "empty", Value: ""},
		Entry{Key: "null", Value: nil},
		Entry{Key: "nested", Value: NewArray(Entry{Key: "child", Value: "original"})},
	)
	before := Clone(input).(Array)

	returned, ok := ArrayReturn(input).(Array)
	if !ok {
		t.Fatal("array return did not return an Array")
	}
	if !reflect.DeepEqual(returned, input) {
		t.Fatalf("array return = %#v, want %#v", returned, input)
	}
	if returned.next != input.next || returned.nextInitialized != input.nextInitialized || returned.nextBlocked != input.nextBlocked {
		t.Fatalf("array return lost automatic-key metadata: got %#v, want %#v", returned, input)
	}

	returned.Entries[0].Value = true
	returned.Entries[3].Value.(Array).Entries[0].Value = "changed"
	if !reflect.DeepEqual(input, before) {
		t.Fatalf("array return aliases input storage: got %#v, want %#v", input, before)
	}
	if input.Entries[0].Value != false || input.Entries[1].Value != "" || input.Entries[2].Value != nil {
		t.Fatalf("array return changed falsy values: %#v", input.Entries)
	}
}

func TestArrayReturnRejectsNonArrays(t *testing.T) {
	for _, value := range []any{nil, false, int64(0), "", "text"} {
		t.Run("non_array", func(t *testing.T) {
			defer func() {
				if got := recover(); got != "array return requires array" {
					t.Fatalf("panic = %v, want %q", got, "array return requires array")
				}
			}()
			ArrayReturn(value)
		})
	}
}
