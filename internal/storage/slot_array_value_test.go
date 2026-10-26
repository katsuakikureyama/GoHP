package storage

import "testing"

// readOwnedArray obtains the caller-owned copy returned by a general Slot.
// Each caller must Close the result after its assertions.
func readOwnedArray(t *testing.T, slot *Slot) *ScalarArray {
	t.Helper()
	value, present := slot.Read()
	if !present {
		t.Fatal("Read() unexpectedly returned an undefined slot")
	}
	array, ok := value.(*ScalarArray)
	if !ok {
		t.Fatalf("Read() value = %T, want *ScalarArray", value)
	}
	return array
}

func requireOwnedArrayValue(t *testing.T, array *ScalarArray, key, want any) {
	t.Helper()
	requireArrayRead(t, array, key, want, true)
}

func TestSlotArrayValueCopies(t *testing.T) {
	original := NewScalarArray()
	if key := original.Append("zero"); key != 0 {
		t.Fatalf("first Append() key = %d, want 0", key)
	}
	original.Write(int64(3), "three")

	slot := NewSlot(original)
	original.Close() // Write retains; the input handle remains caller-owned.

	copy := readOwnedArray(t, slot)
	if key := copy.Append("four"); key != 4 {
		t.Fatalf("copied Append() key = %d, want 4", key)
	}
	// A returned array is an independent caller handle until it is written back.
	unchanged := readOwnedArray(t, slot)
	if _, present := unchanged.Read(int64(4)); present {
		t.Fatal("unwritten returned copy changed the slot value")
	}
	unchanged.Close()

	slot.Write(copy) // Explicit write-back retains before releasing the old payload.
	copy.Close()

	same := readOwnedArray(t, slot)
	slot.Write(same) // Reassigning an equivalent shared handle cannot release it first.
	same.Close()

	preserved := readOwnedArray(t, slot)
	requireOwnedArrayValue(t, preserved, int64(0), "zero")
	requireOwnedArrayValue(t, preserved, int64(3), "three")
	requireOwnedArrayValue(t, preserved, int64(4), "four")
	if key := preserved.Append("five"); key != 5 {
		t.Fatalf("array index after slot copies = %d, want 5", key)
	}
	slot.Write(preserved)
	preserved.Close()

	temporary := readOwnedArray(t, slot)
	temporary.Close()
	// Closing a temporary read must not release the slot's retained payload.
	afterTemporary := readOwnedArray(t, slot)
	requireOwnedArrayValue(t, afterTemporary, int64(5), "five")
	escaped := afterTemporary.Reference("ref")
	afterTemporary.Close()

	// Conversely, Unset releases the slot's payload but not an escaped element
	// reference, which owns its own binding.
	survivingRead := readOwnedArray(t, slot)
	slot.Unset()
	requireOwnedArrayValue(t, survivingRead, int64(5), "five")
	survivingRead.Close()
	escaped.Write("survives-slot-unset")
	requireRead(t, escaped, "survives-slot-unset", true)
	escaped.Unset()
}

func TestSlotArrayReferenceOwnership(t *testing.T) {
	t.Run("outer binding replacement is shared but value copies are not", func(t *testing.T) {
		first := NewScalarArray()
		first.Write("version", "first")
		outer := NewSlot(first)
		first.Close()
		peer := NewSlot(nil)
		peer.Bind(outer)
		if owners := outer.ref.value.(*ScalarArray).data.owners; owners != 1 {
			t.Fatalf("Bind retained payload per alias: owners=%d, want 1", owners)
		}

		second := NewScalarArray()
		second.Write("version", "second")
		outer.Write(second)
		second.Close()
		peerValue := readOwnedArray(t, peer)
		requireOwnedArrayValue(t, peerValue, "version", "second")
		peerValue.Close()

		ordinaryCopy := readOwnedArray(t, outer)
		independent := NewSlot(ordinaryCopy)
		ordinaryCopy.Close()
		third := NewScalarArray()
		third.Write("version", "third")
		outer.Write(third)
		third.Close()

		peerValue = readOwnedArray(t, peer)
		requireOwnedArrayValue(t, peerValue, "version", "third")
		peerValue.Close()
		independentValue := readOwnedArray(t, independent)
		requireOwnedArrayValue(t, independentValue, "version", "second")
		independentValue.Close()
		independent.Unset()
		peer.Unset()
		outer.Unset()
	})

	t.Run("owned reads preserve internal reference copy timing", func(t *testing.T) {
		external := NewSlot(int64(1))
		inner := NewScalarArray()
		inner.Bind("ref", external)
		outer := NewSlot(inner)
		inner.Close()

		lateCopy := readOwnedArray(t, outer)
		external.Unset()
		lateCopy.Write("ref", int64(2))
		stored := readOwnedArray(t, outer)
		requireOwnedArrayValue(t, stored, "ref", int64(1))
		stored.Close()
		lateCopy.Close()
		outer.Unset()

		external = NewSlot(int64(1))
		inner = NewScalarArray()
		inner.Bind("ref", external)
		outer = NewSlot(inner)
		inner.Close()
		earlyCopy := readOwnedArray(t, outer)
		earlyCopy.Write("plain", int64(0)) // Separate while external is still live.
		external.Unset()
		earlyCopy.Write("ref", int64(2))
		stored = readOwnedArray(t, outer)
		requireOwnedArrayValue(t, stored, "ref", int64(2))
		stored.Close()
		earlyCopy.Close()
		outer.Unset()
	})

	t.Run("local rebinding preserves former array peers", func(t *testing.T) {
		old := NewScalarArray()
		old.Write("version", "old")
		caller := NewSlot(old)
		old.Close()
		local := NewSlot(nil)
		local.Bind(caller)

		replacement := NewScalarArray()
		replacement.Write("version", "replacement")
		rebound := NewSlot(replacement)
		replacement.Close()
		local.Bind(rebound)

		callerValue := readOwnedArray(t, caller)
		requireOwnedArrayValue(t, callerValue, "version", "old")
		callerValue.Close()
		localValue := readOwnedArray(t, local)
		requireOwnedArrayValue(t, localValue, "version", "replacement")
		localValue.Close()
		local.Unset()
		rebound.Unset()
		caller.Unset()
	})
}

func TestSlotArrayReleaseAndInvalidWrites(t *testing.T) {
	t.Run("invalid values do not replace ordinary or aliased payloads", func(t *testing.T) {
		original := NewScalarArray()
		original.Write("value", "kept")
		holder := NewSlot(original)
		original.Close()
		alias := NewSlot(nil)
		alias.Bind(holder)

		closed := NewScalarArray()
		closed.Close()
		ordinaryValue := readOwnedArray(t, holder)
		ordinary := NewSlot(ordinaryValue)
		ordinaryValue.Close()
		defer ordinary.Unset()
		for _, invalid := range []any{(*ScalarArray)(nil), closed, new(Slot), []any{"foreign"}} {
			requirePanic(t, func() { holder.Write(invalid) })
			requirePanic(t, func() { alias.Write(invalid) })
			requirePanic(t, func() { ordinary.Write(invalid) })
			for _, slot := range []*Slot{holder, alias, ordinary} {
				value := readOwnedArray(t, slot)
				requireOwnedArrayValue(t, value, "value", "kept")
				value.Close()
			}
		}
		requirePanic(t, func() { NewSlot((*ScalarArray)(nil)) })
		requirePanic(t, func() { NewSlot(closed) })
		alias.Unset()
		holder.Unset()
	})

	t.Run("replacement final release and rebind release once", func(t *testing.T) {
		first := NewScalarArray()
		first.Write("value", "first")
		holder := NewSlot(first)
		first.Close()
		storedFirst := holder.value.(*ScalarArray)
		firstData := storedFirst.data

		second := NewScalarArray()
		second.Write("value", "second")
		holder.Write(second)
		second.Close()
		if !storedFirst.closed || firstData.owners != 0 || firstData.entries != nil {
			t.Fatalf("replaced payload was not released once: closed=%v owners=%d entries=%#v", storedFirst.closed, firstData.owners, firstData.entries)
		}
		holder.Unset()
		if firstData.owners != 0 {
			t.Fatalf("second release changed old payload owners to %d, want 0", firstData.owners)
		}

		old := NewScalarArray()
		old.Write("value", "old")
		caller := NewSlot(old)
		old.Close()
		callerData := caller.value.(*ScalarArray).data
		local := NewSlot(nil)
		local.Bind(caller)
		newValue := NewScalarArray()
		newValue.Write("value", "new")
		replacement := NewSlot(newValue)
		newValue.Close()
		local.Bind(replacement)
		if callerData.owners != 1 {
			t.Fatalf("local rebind released caller payload: owners=%d, want 1", callerData.owners)
		}
		caller.Unset()
		if callerData.owners != 0 {
			t.Fatalf("last former peer release owners=%d, want 0", callerData.owners)
		}
		local.Unset()
		replacement.Unset()
	})

	t.Run("escaped element reference survives final slot release", func(t *testing.T) {
		external := NewSlot("seed")
		payload := NewScalarArray()
		payload.Bind("ref", external)
		escaped := payload.Reference("ref")
		holder := NewSlot(payload)
		payload.Close()
		stored := holder.value.(*ScalarArray)
		data := stored.data
		peer := NewSlot(nil)
		peer.Bind(holder)

		holder.Unset()
		if data.owners != 1 || stored.closed {
			t.Fatal("releasing a non-final outer alias closed its payload")
		}
		peer.Unset()
		if data.owners != 0 {
			t.Fatalf("final slot release owners=%d, want 0", data.owners)
		}
		external.Unset()
		escaped.Write("survives")
		requireRead(t, escaped, "survives", true)
		escaped.Unset()
	})

	t.Run("deferred temporary and slot cleanup on panic", func(t *testing.T) {
		payload := NewScalarArray()
		payload.Write("value", "kept")
		shared := payload.data
		requirePanic(t, func() {
			holder := NewSlot(payload)
			defer holder.Unset()
			payload.Close()
			temporary := readOwnedArray(t, holder)
			defer temporary.Close()
			panic("caller exit")
		})
		if shared.owners != 0 || shared.entries != nil {
			t.Fatalf("deferred exit leaked array owners: %#v", shared)
		}
	})
}
