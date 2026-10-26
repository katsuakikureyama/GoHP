package runtime

import (
	"math"
	"reflect"
	"testing"
)

func TestValueSlotRoundTrip(t *testing.T) {
	slot := NewValueSlot(nil)
	defer slot.Close()

	for _, value := range []any{nil, false, int64(-42), "text\x00value"} {
		slot.Write(value)
		if got := slot.Read(); got != value {
			t.Fatalf("Read() = %#v, want %#v", got, value)
		}
	}

	input := NewArray(
		Entry{Key: "label", Value: "first"},
		Entry{Key: int64(-2), Value: nil},
		Entry{Key: "1", Value: true},
	)
	slot.Write(input)
	input.Entries[0].Value = "changed input"

	want := []Entry{
		{Key: "label", Value: "first"},
		{Key: int64(-2), Value: nil},
		{Key: int64(1), Value: true},
	}
	got := requireOwnedValueArray(t, slot.Read())
	if !reflect.DeepEqual(got.Entries, want) {
		t.Fatalf("round trip entries = %#v, want %#v", got.Entries, want)
	}

	// Read exports a fresh legacy array; neither its backing slice nor its
	// values may become a writable path back to the slot.
	got.Entries[0].Value = "changed output"
	again := requireOwnedValueArray(t, slot.Read())
	if !reflect.DeepEqual(again.Entries, want) {
		t.Fatalf("output mutation changed slot: %#v", again.Entries)
	}

	var zero ValueSlot
	zero.Write(NewArray())
	if got := requireOwnedValueArray(t, zero.Read()); len(got.Entries) != 0 {
		t.Fatalf("empty array entries = %#v", got.Entries)
	}
	zero.Close()
}

func TestValueSlotArrayMutation(t *testing.T) {
	slot := NewValueSlot(NewArray(
		Entry{Key: int64(1), Value: "old"},
		Entry{Key: "label", Value: "stable"},
	))
	defer slot.Close()

	// Write observes normalized keys and works through the owned root rather
	// than through a legacy Read() copy.
	slot.ArrayWrite("1", "new")
	slot.ArrayWrite("01", "string key")
	if got := requireOwnedValueArray(t, slot.Read()); !reflect.DeepEqual(got.Entries, []Entry{
		{Key: int64(1), Value: "new"},
		{Key: "label", Value: "stable"},
		{Key: "01", Value: "string key"},
	}) {
		t.Fatalf("array writes = %#v", got.Entries)
	}

	// A snapshot owns the pre-write array value; it is not an address to the
	// source slot and remains stable after the source separates for mutation.
	snapshot := slot.Snapshot()
	slot.ArrayWrite("label", "changed")
	if got := requireOwnedValueArray(t, snapshot.Take()); Get(got, "label") != "stable" {
		t.Fatalf("snapshot changed with source: %#v", got)
	}
	if got := Get(slot.Read(), "label"); got != "changed" {
		t.Fatalf("source write missing: %#v", got)
	}
	snapshot.Close()

	for _, tt := range []struct {
		name string
		base Array
		want int64
	}{
		{"negative", NewArray(Entry{Key: int64(-5), Value: "negative"}), -4},
		{"minimum", NewArray(Entry{Key: int64(math.MinInt64), Value: "minimum"}), math.MinInt64 + 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			candidate := NewValueSlot(tt.base)
			defer candidate.Close()
			if got := candidate.ArrayAppend("next"); got != tt.want {
				t.Fatalf("ArrayAppend() key = %d, want %d", got, tt.want)
			}
			if got := Get(candidate.Read(), tt.want); got != "next" {
				t.Fatalf("appended value = %#v", got)
			}
		})
	}

	maximum := NewValueSlot(NewArray(Entry{Key: int64(math.MaxInt64), Value: "maximum"}))
	defer maximum.Close()
	requireOwnedValuePanic(t, func() { maximum.ArrayAppend("overflow") })
	if got := Get(maximum.Read(), int64(math.MaxInt64)); got != "maximum" {
		t.Fatalf("overflow append changed base: %#v", got)
	}
}

func TestValueTemporaryTransfer(t *testing.T) {
	source := NewValueSlot(NewArray(Entry{Key: "state", Value: "before"}))
	defer source.Close()

	snapshot := source.Snapshot()
	source.ArrayWrite("state", "after")
	target := NewValueSlot("old target")
	defer target.Close()
	snapshot.MoveTo(target)
	if got := Get(target.Read(), "state"); got != "before" {
		t.Fatalf("moved snapshot = %#v, want before", got)
	}
	requireOwnedValuePanic(t, func() { snapshot.Take() })

	// Moving a snapshot back to its current source replaces the source with
	// the captured value, rather than retaining a source binding/location.
	self := source.Snapshot()
	source.ArrayWrite("state", "different")
	self.MoveTo(source)
	if got := Get(source.Read(), "state"); got != "after" {
		t.Fatalf("self destination = %#v, want after", got)
	}

	kept := NewValueSlot(NewArray(Entry{Key: "kept", Value: "value"}))
	temporary := kept.Snapshot()
	kept.Close()
	taken := requireOwnedValueArray(t, temporary.Take())
	temporary.Close()
	if got := Get(taken, "kept"); got != "value" {
		t.Fatalf("taken output did not survive temporary close: %#v", got)
	}
	taken.Entries[0].Value = "caller mutation"
	if got := Get(taken, "kept"); got != "caller mutation" {
		t.Fatalf("taken output is not a normal detached array: %#v", got)
	}
}

func TestValueSlotGuards(t *testing.T) {
	slot := NewValueSlot(NewArray(Entry{Key: "stable", Value: "before"}))
	defer slot.Close()

	invalidArrays := []Array{
		NewArray(Entry{Key: "nested", Value: NewArray()}),
		{Entries: []Entry{{Key: int64(1), Value: "stale index"}}},
		{Entries: []Entry{{Key: "1", Value: "noncanonical key"}}},
		{Entries: []Entry{{Key: "same", Value: "first"}, {Key: "same", Value: "second"}}},
		{Entries: []Entry{{Key: int(1), Value: "foreign key"}}},
		{Entries: []Entry{{Key: "value", Value: int(1)}}},
	}
	for _, invalid := range append([]any{int(1), float64(1), struct{}{}, []any{}}, arraysToAny(invalidArrays)...) {
		requireOwnedValuePanic(t, func() { slot.Write(invalid) })
		if got := Get(slot.Read(), "stable"); got != "before" {
			t.Fatalf("invalid Write(%#v) changed value to %#v", invalid, got)
		}
	}

	for _, action := range []func(){
		func() { slot.ArrayWrite(int(1), "bad key") },
		func() { slot.ArrayWrite("stable", int(1)) },
		func() { slot.ArrayAppend(NewArray()) },
	} {
		requireOwnedValuePanic(t, action)
		if got := Get(slot.Read(), "stable"); got != "before" {
			t.Fatalf("invalid array mutation changed value to %#v", got)
		}
	}

	for _, value := range []any{nil, false, int64(1), "text"} {
		candidate := NewValueSlot(value)
		requireOwnedValuePanic(t, func() { candidate.ArrayWrite("key", "value") })
		requireOwnedValuePanic(t, func() { candidate.ArrayAppend("value") })
		candidate.Close()
	}

	var undefined ValueSlot
	requireOwnedValuePanic(t, func() { undefined.Read() })
	requireOwnedValuePanic(t, func() { undefined.Snapshot() })
	requireOwnedValuePanic(t, func() { undefined.ArrayWrite("key", "value") })
	requireOwnedValuePanic(t, func() { undefined.ArrayAppend("value") })
	undefined.Close()
	requireOwnedValuePanic(t, func() { undefined.Write(nil) })
	undefined.Close()

	slot.Close()
	requireOwnedValuePanic(t, func() { slot.Read() })
	requireOwnedValuePanic(t, func() { slot.Write(nil) })
	requireOwnedValuePanic(t, func() { slot.Snapshot() })
	requireOwnedValuePanic(t, func() { slot.ArrayWrite("key", "value") })
	requireOwnedValuePanic(t, func() { slot.ArrayAppend("value") })

	var nilSlot *ValueSlot
	requireOwnedValuePanic(t, func() { nilSlot.Read() })
	requireOwnedValuePanic(t, func() { nilSlot.Write(nil) })
	requireOwnedValuePanic(t, func() { nilSlot.Close() })
}

func TestValueTemporaryGuards(t *testing.T) {
	for _, invalid := range []any{int(1), float64(1), struct{}{}, NewArray(Entry{Key: "nested", Value: NewArray()})} {
		requireOwnedValuePanic(t, func() { NewValueTemporary(invalid) })
	}

	var zero ValueTemporary
	requireOwnedValuePanic(t, func() { zero.Take() })
	requireOwnedValuePanic(t, func() { zero.MoveTo(NewValueSlot(nil)) })
	zero.Close()

	consumed := NewValueTemporary("value")
	if got := consumed.Take(); got != "value" {
		t.Fatalf("Take() = %#v", got)
	}
	consumed.Close()
	requireOwnedValuePanic(t, func() { consumed.Take() })
	requireOwnedValuePanic(t, func() { consumed.MoveTo(NewValueSlot(nil)) })

	closedDestination := NewValueSlot(nil)
	closedDestination.Close()
	for _, destination := range []*ValueSlot{nil, closedDestination} {
		temporary := NewValueTemporary("retry")
		requireOwnedValuePanic(t, func() { temporary.MoveTo(destination) })
		if got := temporary.Take(); got != "retry" {
			t.Fatalf("failed MoveTo consumed temporary: %#v", got)
		}
	}

	// A deferred Close remains a correct failure cleanup path when a move
	// rejects before consuming the temporary.
	cleanup := NewValueTemporary("cleanup")
	func() {
		defer cleanup.Close()
		requireOwnedValuePanic(t, func() { cleanup.MoveTo(nil) })
	}()
	requireOwnedValuePanic(t, func() { cleanup.Take() })

	var nilTemporary *ValueTemporary
	requireOwnedValuePanic(t, func() { nilTemporary.Take() })
	requireOwnedValuePanic(t, func() { nilTemporary.MoveTo(NewValueSlot(nil)) })
	requireOwnedValuePanic(t, func() { nilTemporary.Close() })
}

func requireOwnedValueArray(t *testing.T, value any) Array {
	t.Helper()
	array, ok := value.(Array)
	if !ok {
		t.Fatalf("value = %#v, want runtime.Array", value)
	}
	return array
}

func requireOwnedValuePanic(t *testing.T, action func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("operation did not panic")
		}
	}()
	action()
}

func arraysToAny(arrays []Array) []any {
	values := make([]any, len(arrays))
	for i, array := range arrays {
		values[i] = array
	}
	return values
}
