package storage

import "testing"

func requireRead(t *testing.T, slot *Slot, want any, wantPresent bool) {
	t.Helper()
	got, present := slot.Read()
	if present != wantPresent {
		t.Fatalf("Read() presence = %v, want %v (value %#[1]v)", present, wantPresent, got)
	}
	if got != want {
		t.Fatalf("Read() value = %#v, want %#v", got, want)
	}
}

func requirePanic(t *testing.T, action func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("action did not panic")
		}
	}()
	action()
}

func TestSlotZeroValueIsUndefined(t *testing.T) {
	var slot Slot
	requireRead(t, &slot, nil, false)
}

func TestSlotNewAndWriteScalarValues(t *testing.T) {
	for _, initial := range []any{nil, false, int64(-7), "seed"} {
		slot := NewSlot(initial)
		requireRead(t, slot, initial, true)
	}

	slot := NewSlot("before")
	for _, value := range []any{nil, true, int64(42), "after"} {
		slot.Write(value)
		requireRead(t, slot, value, true)
	}
}

func TestSlotRejectsUnsupportedValuesWithoutMutation(t *testing.T) {
	for _, invalid := range []any{int(1), float64(1), []any{"array"}, struct{ Name string }{"object"}} {
		slot := NewSlot("before")
		requirePanic(t, func() { slot.Write(invalid) })
		requireRead(t, slot, "before", true)

		requirePanic(t, func() { NewSlot(invalid) })
	}
}

func TestSlotBindSharesWritesAndRebindingLeavesFormerPeersBound(t *testing.T) {
	left := NewSlot("left")
	right := NewSlot("right")
	alias := NewSlot(nil)
	peer := NewSlot(nil)

	alias.Bind(left)
	peer.Bind(alias)
	alias.Write("first-change")
	requireRead(t, left, "first-change", true)
	requireRead(t, peer, "first-change", true)

	alias.Bind(right)
	alias.Write("second-change")
	requireRead(t, right, "second-change", true)
	requireRead(t, alias, "second-change", true)
	requireRead(t, left, "first-change", true)
	requireRead(t, peer, "first-change", true)
}

func TestSlotBindUndefinedSourceCreatesPresentNullAndSelfBindingWorks(t *testing.T) {
	undefined := new(Slot)
	alias := NewSlot("discarded")
	alias.Bind(undefined)
	requireRead(t, undefined, nil, true)
	requireRead(t, alias, nil, true)

	alias.Write("shared")
	requireRead(t, undefined, "shared", true)

	initialized := NewSlot("value")
	initialized.Bind(initialized)
	requireRead(t, initialized, "value", true)

	selfUndefined := new(Slot)
	selfUndefined.Bind(selfUndefined)
	requireRead(t, selfUndefined, nil, true)

	requirePanic(t, func() { alias.Bind(nil) })
	requireRead(t, alias, "shared", true)
}

func TestSlotUnsetDetachesOnlySelectedBinding(t *testing.T) {
	value := NewSlot(int64(1))
	alias := NewSlot(nil)
	alias.Bind(value)

	alias.Unset()
	requireRead(t, alias, nil, false)
	value.Write(int64(2))
	requireRead(t, value, int64(2), true)

	alias.Write(int64(3))
	requireRead(t, alias, int64(3), true)
	requireRead(t, value, int64(2), true)
}

func TestSlotValueAssignmentReadsValueWithoutBinding(t *testing.T) {
	source := NewSlot("before")
	destination := NewSlot(nil)

	value, present := source.Read()
	if !present {
		t.Fatal("source unexpectedly undefined")
	}
	destination.Write(value)
	source.Write("after")

	requireRead(t, source, "after", true)
	requireRead(t, destination, "before", true)
}

func TestSlotByReferenceParameterRebindingAndUnsetAreLocal(t *testing.T) {
	caller := NewSlot("before")
	local := new(Slot)
	local.Bind(caller)

	local.Write("mutated")
	requireRead(t, caller, "mutated", true)

	replacement := NewSlot("replacement")
	local.Bind(replacement)
	local.Write("rebound-local")
	requireRead(t, caller, "mutated", true)
	requireRead(t, replacement, "rebound-local", true)

	local.Unset()
	requireRead(t, local, nil, false)
	requireRead(t, caller, "mutated", true)
}
