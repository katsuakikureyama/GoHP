package storage

import "testing"

func TestTemporarySnapshot(t *testing.T) {
	t.Run("absence null and scalar snapshot", func(t *testing.T) {
		undefined := new(Slot)
		if temporary, present := undefined.Snapshot(); present || temporary != nil {
			t.Fatalf("undefined Snapshot() = (%#v, %v), want (nil, false)", temporary, present)
		}

		null := NewSlot(nil)
		temporary, present := null.Snapshot()
		if !present || temporary == nil {
			t.Fatal("present null did not produce a temporary")
		}
		if value := temporary.Take(); value != nil {
			t.Fatalf("null temporary Take() = %#v, want nil", value)
		}
		temporary.Close()

		source := NewSlot("before")
		peer := NewSlot(nil)
		peer.Bind(source)
		snapshot, present := source.Snapshot()
		if !present || source.ref.owners != 2 {
			t.Fatalf("scalar Snapshot() changed outer binding ownership: present=%v owners=%d", present, source.ref.owners)
		}
		source.Write("after")
		destination := NewSlot(nil)
		snapshot.MoveTo(destination)
		requireRead(t, destination, "before", true)
		requireRead(t, peer, "after", true)
		destination.Unset()
		peer.Unset()
		source.Unset()
		null.Unset()
	})

	t.Run("array snapshot retains value but not outer binding", func(t *testing.T) {
		external := NewSlot("seed")
		payload := NewScalarArray()
		payload.Bind("ref", external)
		payload.Write("plain", "before")
		source := NewSlot(payload)
		payload.Close()

		temporary, present := source.Snapshot()
		if !present {
			t.Fatal("array Snapshot() unexpectedly absent")
		}
		source.ArrayWrite("plain", "after capture")
		replacement := NewSlot("replacement")
		source.Bind(replacement)
		source.Unset()
		replacement.Unset()

		captured := temporary.Take().(*ScalarArray)
		requireOwnedArrayValue(t, captured, "plain", "before")
		requireOwnedArrayValue(t, captured, "ref", "seed")
		external.Write("through existing element reference")
		requireOwnedArrayValue(t, captured, "ref", "through existing element reference")
		captured.Close()
		temporary.Close()
		external.Unset()
	})
}

func TestTemporaryTransfer(t *testing.T) {
	t.Run("MoveTo resolves current outer binding and consumes once", func(t *testing.T) {
		root := NewSlot("root")
		alias := NewSlot(nil)
		alias.Bind(root)
		NewTemporary("through alias").MoveTo(alias)
		requireRead(t, root, "through alias", true)

		replacement := NewSlot("replacement")
		pending := NewTemporary("current replacement")
		alias.Bind(replacement)
		pending.MoveTo(alias)
		requireRead(t, root, "through alias", true)
		requireRead(t, replacement, "current replacement", true)
		requirePanic(t, func() { pending.Take() })
		requirePanic(t, func() { pending.MoveTo(alias) })

		alias.Unset()
		replacement.Unset()
		root.Unset()
	})

	t.Run("self root array move retains before release", func(t *testing.T) {
		payload := NewScalarArray()
		payload.Write(int64(0), "zero")
		root := NewSlot(payload)
		payload.Close()
		temporary, _ := root.Snapshot()
		temporary.MoveTo(root)
		stored := readOwnedArray(t, root)
		requireOwnedArrayValue(t, stored, int64(0), "zero")
		stored.Close()
		root.Unset()
	})

	t.Run("Take transfers array ownership and escaped alias", func(t *testing.T) {
		payload := NewScalarArray()
		payload.Write(int64(0), "zero")
		source := NewSlot("seed")
		payload.Bind("ref", source)
		temporary := NewTemporary(payload)
		payload.Close()

		array := func() *ScalarArray {
			defer temporary.Close()
			return temporary.Take().(*ScalarArray)
		}()
		if key := array.Append("one"); key != 1 {
			t.Fatalf("taken array Append() key = %d, want 1", key)
		}
		alias := array.Reference("ref")
		array.Close()
		alias.Write("survives array release")
		requireRead(t, alias, "survives array release", true)
		alias.Unset()
		source.Unset()
	})
}

func TestTemporaryRelease(t *testing.T) {
	t.Run("explicit close releases one retained array owner", func(t *testing.T) {
		payload := NewScalarArray()
		payload.Write("value", "kept")
		root := NewSlot(payload)
		payload.Close()
		data := root.value.(*ScalarArray).data
		temporary, _ := root.Snapshot()
		root.Unset()
		if data.owners != 1 {
			t.Fatalf("temporary array owner count = %d, want 1", data.owners)
		}
		temporary.Close()
		if data.owners != 0 || data.entries != nil {
			t.Fatalf("temporary Close() did not release data: %#v", data)
		}
		temporary.Close()
	})

	t.Run("defer releases on normal and panic exits", func(t *testing.T) {
		for _, panicLater := range []bool{false, true} {
			payload := NewScalarArray()
			payload.Write("value", "kept")
			root := NewSlot(payload)
			payload.Close()
			data := root.value.(*ScalarArray).data

			if panicLater {
				requirePanic(t, func() {
					temporary, _ := root.Snapshot()
					defer temporary.Close()
					root.Unset()
					panic("later operand")
				})
			} else {
				func() {
					temporary, _ := root.Snapshot()
					defer temporary.Close()
					root.Unset()
				}()
			}
			if data.owners != 0 || data.entries != nil {
				t.Fatalf("deferred Close(panic=%v) leaked data: %#v", panicLater, data)
			}
		}
	})
}

func TestTemporaryCopyTiming(t *testing.T) {
	for _, releaseBeforeSeparation := range []bool{true, false} {
		t.Run(map[bool]string{true: "release old owner first", false: "keep old owner through separation"}[releaseBeforeSeparation], func(t *testing.T) {
			external := NewSlot(int64(1))
			payload := NewScalarArray()
			payload.Bind("ref", external)
			root := NewSlot(payload)
			payload.Close()

			held, _ := root.Snapshot()
			root.ArrayWrite("plain", int64(0)) // Separate while external remains live.
			external.Unset()
			value := readOwnedArray(t, root)
			other := NewSlot(value)
			value.Close()
			if releaseBeforeSeparation {
				held.Close()
			}
			other.ArrayWrite("ref", int64(2))
			if !releaseBeforeSeparation {
				held.Close()
			}

			rootValue := readOwnedArray(t, root)
			otherValue := readOwnedArray(t, other)
			if releaseBeforeSeparation {
				requireOwnedArrayValue(t, rootValue, "ref", int64(1))
			} else {
				requireOwnedArrayValue(t, rootValue, "ref", int64(2))
			}
			requireOwnedArrayValue(t, otherValue, "ref", int64(2))
			rootValue.Close()
			otherValue.Close()
			other.Unset()
			root.Unset()
		})
	}
}

func TestTemporaryFailure(t *testing.T) {
	t.Run("invalid construction and inactive operations", func(t *testing.T) {
		closed := NewScalarArray()
		closed.Close()
		for _, value := range []any{(*ScalarArray)(nil), closed, []any{"foreign"}, new(Slot)} {
			requirePanic(t, func() { NewTemporary(value) })
		}
		var nilTemporary *Temporary
		requirePanic(t, func() { nilTemporary.Close() })
		requirePanic(t, func() { nilTemporary.Take() })
		requirePanic(t, func() { (*Slot)(nil).Snapshot() })
		zero := new(Temporary)
		destination := NewSlot(nil)
		requirePanic(t, func() { nilTemporary.MoveTo(destination) })
		requirePanic(t, func() { zero.Take() })
		requirePanic(t, func() { zero.MoveTo(destination) })
		zero.Close()
		pending := NewTemporary("retry after nil destination")
		requirePanic(t, func() { pending.MoveTo(nil) })
		pending.MoveTo(destination)
		requireRead(t, destination, "retry after nil destination", true)
		pending.Close()
		destination.Unset()
	})

	t.Run("failed constrained move preserves temporary and destination", func(t *testing.T) {
		payload := NewScalarArray()
		payload.Write("value", "kept")
		temporary := NewTemporary(payload)
		payload.Close()
		destination := newScalarSlot()
		destination.Write("unchanged")

		requirePanic(t, func() { temporary.MoveTo(destination) })
		requireRead(t, destination, "unchanged", true)
		array := temporary.Take().(*ScalarArray)
		requireOwnedArrayValue(t, array, "value", "kept")
		array.Close()
		temporary.Close()
		requirePanic(t, func() { temporary.Take() })
		requirePanic(t, func() { temporary.MoveTo(destination) })
		destination.Unset()
	})
}
