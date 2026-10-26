package storage

import (
	"math"
	"testing"
)

func TestSlotArrayMutationValues(t *testing.T) {
	payload := NewScalarArray()
	payload.Write(int64(0), "zero")
	root := NewSlot(payload)
	payload.Close()

	// Read returns an ordinary value copy. Mutating that copy does not use the
	// root's writable path until a caller explicitly writes it back.
	snapshot := readOwnedArray(t, root)
	snapshot.Write("snapshot", "outside")
	snapshot.Close()
	stored := readOwnedArray(t, root)
	requireArrayRead(t, stored, "snapshot", nil, false)
	stored.Close()

	// An outer alias resolves the same current owned payload, so its writes are
	// visible through the root. The string key exercises the shared key policy.
	alias := NewSlot(nil)
	alias.Bind(root)
	alias.ArrayWrite("1", "canonical key")
	if key := alias.ArrayAppend("after"); key != 2 {
		t.Fatalf("ArrayAppend() key = %d, want 2", key)
	}
	stored = readOwnedArray(t, root)
	requireOwnedArrayValue(t, stored, int64(1), "canonical key")
	requireOwnedArrayValue(t, stored, int64(2), "after")
	stored.Close()

	// A separately owned ordinary array value still isolates direct mutations.
	ordinaryValue := readOwnedArray(t, root)
	ordinary := NewSlot(ordinaryValue)
	ordinaryValue.Close()
	ordinary.ArrayWrite(int64(1), "ordinary only")
	stored = readOwnedArray(t, root)
	requireOwnedArrayValue(t, stored, int64(1), "canonical key")
	stored.Close()
	ordinaryStored := readOwnedArray(t, ordinary)
	requireOwnedArrayValue(t, ordinaryStored, int64(1), "ordinary only")
	ordinaryStored.Close()

	ordinary.Unset()
	alias.Unset()
	root.Unset()
}

func TestSlotArrayMutationCopyTiming(t *testing.T) {
	t.Run("last external owner released before direct separation snapshots", func(t *testing.T) {
		external := NewSlot(int64(1))
		payload := NewScalarArray()
		payload.Bind("ref", external)
		left := NewSlot(payload)
		payload.Close()
		valueCopy := readOwnedArray(t, left)
		right := NewSlot(valueCopy)
		valueCopy.Close()

		external.Unset()
		right.ArrayWrite("ref", int64(2))
		leftValue := readOwnedArray(t, left)
		requireOwnedArrayValue(t, leftValue, "ref", int64(1))
		leftValue.Close()
		rightValue := readOwnedArray(t, right)
		requireOwnedArrayValue(t, rightValue, "ref", int64(2))
		rightValue.Close()

		right.Unset()
		left.Unset()
	})

	t.Run("direct separation before external release preserves alias", func(t *testing.T) {
		external := NewSlot(int64(1))
		payload := NewScalarArray()
		payload.Bind("ref", external)
		left := NewSlot(payload)
		payload.Close()
		valueCopy := readOwnedArray(t, left)
		right := NewSlot(valueCopy)
		valueCopy.Close()

		right.ArrayWrite("plain", int64(0))
		external.Unset()
		right.ArrayWrite("ref", int64(2))
		for _, slot := range []*Slot{left, right} {
			value := readOwnedArray(t, slot)
			requireOwnedArrayValue(t, value, "ref", int64(2))
			value.Close()
		}

		right.Unset()
		left.Unset()
	})
}

func TestSlotArrayMutationBindings(t *testing.T) {
	old := NewSlot("old")
	payload := NewScalarArray()
	payload.Bind("selected", old)
	payload.Bind("peer", old)
	root := NewSlot(payload)
	payload.Close()

	replacement := NewSlot("replacement")
	root.ArrayBind("selected", replacement)
	replacement.Write("from replacement")
	old.Write("from old")
	stored := readOwnedArray(t, root)
	requireOwnedArrayValue(t, stored, "selected", "from replacement")
	requireOwnedArrayValue(t, stored, "peer", "from old")
	stored.Close()

	// An exported element alias is a separate caller-owned binding. Rebinding it
	// must not rebind the stored element.
	returned := root.ArrayReference("selected")
	outside := NewSlot("outside")
	returned.Bind(outside)
	returned.Write("outside changed")
	stored = readOwnedArray(t, root)
	requireOwnedArrayValue(t, stored, "selected", "from replacement")
	stored.Close()
	requireRead(t, outside, "outside changed", true)
	returned.Unset()

	// Missing-element reference creation changes only the selected container.
	snapshot := readOwnedArray(t, root)
	missing := root.ArrayReference("8")
	requireRead(t, missing, nil, true)
	requireArrayRead(t, snapshot, int64(8), nil, false)
	snapshot.Close()
	stored = readOwnedArray(t, root)
	requireArrayRead(t, stored, int64(8), nil, true)
	stored.Close()
	missing.Unset()

	undefined := new(Slot)
	root.ArrayBind("initialized", undefined)
	requireRead(t, undefined, nil, true)
	undefined.Write("bound after initialization")
	stored = readOwnedArray(t, root)
	requireOwnedArrayValue(t, stored, "initialized", "bound after initialization")
	stored.Close()
	undefined.Unset()

	// A different returned alias outlives final ownership of the root array.
	surviving := root.ArrayReference("peer")
	root.Unset()
	old.Unset()
	surviving.Write("survives root unset")
	requireRead(t, surviving, "survives root unset", true)
	surviving.Unset()

	outside.Unset()
	replacement.Unset()
}

func TestSlotArrayMutationCurrentRoot(t *testing.T) {
	oldPayload := NewScalarArray()
	oldPayload.Write("version", "old")
	caller := NewSlot(oldPayload)
	oldPayload.Close()
	root := NewSlot(nil)
	root.Bind(caller)

	replacementPayload := NewScalarArray()
	replacementPayload.Write("version", "replacement")
	replacement := NewSlot(replacementPayload)
	replacementPayload.Close()

	// Go evaluates the RHS before entering ArrayWrite. The method must then use
	// root's current binding, not the array that root held before that evaluation.
	root.ArrayWrite("written", func() any {
		root.Bind(replacement)
		return "after rebind"
	}())
	oldValue := readOwnedArray(t, caller)
	requireOwnedArrayValue(t, oldValue, "version", "old")
	requireArrayRead(t, oldValue, "written", nil, false)
	oldValue.Close()
	replacementValue := readOwnedArray(t, replacement)
	requireOwnedArrayValue(t, replacementValue, "written", "after rebind")
	replacementValue.Close()

	// The same rule applies when the RHS replaces the current binding's value.
	newPayload := NewScalarArray()
	newPayload.Write("version", "new")
	if key := root.ArrayAppend(func() any {
		root.Write(newPayload)
		newPayload.Close()
		return "after replacement"
	}()); key != 0 {
		t.Fatalf("ArrayAppend() after RHS replacement key = %d, want 0", key)
	}
	callerValue := readOwnedArray(t, caller)
	requireOwnedArrayValue(t, callerValue, "version", "old")
	callerValue.Close()
	replacementValue = readOwnedArray(t, replacement)
	requireOwnedArrayValue(t, replacementValue, "version", "new")
	requireOwnedArrayValue(t, replacementValue, int64(0), "after replacement")
	replacementValue.Close()

	root.Unset()
	replacement.Unset()
	caller.Unset()
}

func TestSlotArrayMutationRejects(t *testing.T) {
	undefined := new(Slot)
	null := NewSlot(nil)
	scalar := NewSlot("scalar")
	source := NewSlot("source")
	for _, test := range []struct {
		name string
		root *Slot
	}{
		{"nil", nil},
		{"undefined", undefined},
		{"null", null},
		{"scalar", scalar},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := test.root
			requirePanic(t, func() { root.ArrayWrite("key", "value") })
			requirePanic(t, func() { root.ArrayAppend("value") })
			requirePanic(t, func() { root.ArrayBind("key", source) })
			requirePanic(t, func() { root.ArrayReference("key") })
		})
	}
	requireRead(t, undefined, nil, false)
	requireRead(t, null, nil, true)
	requireRead(t, scalar, "scalar", true)
	requireRead(t, source, "source", true)
	source.Unset()
	scalar.Unset()
	null.Unset()
	undefined.Unset()

	closed := NewScalarArray()
	closed.Close()
	closedRoot := &Slot{defined: true, value: closed}
	requirePanic(t, func() { closedRoot.ArrayWrite("key", "value") })
	closedRoot.Unset()

	// These rejected operations must not detach a shared array while the
	// external binding is live. The later write after release distinguishes
	// unchanged sharing from an accidental separation.
	external := NewSlot(int64(1))
	payload := NewScalarArray()
	payload.Bind("ref", external)
	left := NewSlot(payload)
	payload.Close()
	valueCopy := readOwnedArray(t, left)
	right := NewSlot(valueCopy)
	valueCopy.Close()
	arrayPayload := NewScalarArray()
	arraySource := NewSlot(arrayPayload)
	arrayPayload.Close()
	nested := NewScalarArray()
	for _, action := range []func(){
		func() { right.ArrayWrite(false, "bad key") },
		func() { right.ArrayWrite("ref", int(1)) },
		func() { right.ArrayWrite("ref", nested) },
		func() { right.ArrayAppend(nested) },
		func() { right.ArrayReference(false) },
		func() { right.ArrayBind(false, undefined) },
		func() { right.ArrayBind("missing", nil) },
		func() { right.ArrayBind("missing", arraySource) },
	} {
		requirePanic(t, action)
	}
	requireRead(t, undefined, nil, false)
	nested.Close()
	external.Unset()
	right.ArrayWrite("ref", int64(2))
	leftValue := readOwnedArray(t, left)
	requireOwnedArrayValue(t, leftValue, "ref", int64(1))
	leftValue.Close()
	rightValue := readOwnedArray(t, right)
	requireOwnedArrayValue(t, rightValue, "ref", int64(2))
	requireArrayRead(t, rightValue, "missing", nil, false)
	rightValue.Close()
	arraySource.Unset()
	right.Unset()
	left.Unset()

	// Overflow is also preflight-only: it must not separate a shared container.
	external = NewSlot(int64(1))
	payload = NewScalarArray()
	payload.Write(int64(math.MaxInt64), "maximum")
	payload.Bind("ref", external)
	left = NewSlot(payload)
	payload.Close()
	valueCopy = readOwnedArray(t, left)
	right = NewSlot(valueCopy)
	valueCopy.Close()
	requirePanic(t, func() { right.ArrayAppend("overflow") })
	external.Unset()
	right.ArrayWrite("ref", int64(2))
	leftValue = readOwnedArray(t, left)
	requireOwnedArrayValue(t, leftValue, "ref", int64(1))
	leftValue.Close()
	rightValue = readOwnedArray(t, right)
	requireOwnedArrayValue(t, rightValue, "ref", int64(2))
	rightValue.Close()
	right.Unset()
	left.Unset()
}
