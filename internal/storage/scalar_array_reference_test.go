package storage

import "testing"

func requireScalarArrayRead(t *testing.T, array *ScalarArray, key, want any, wantPresent bool) {
	t.Helper()
	got, present := array.Read(key)
	if present != wantPresent {
		t.Fatalf("Read(%#v) presence = %v, want %v (value %#v)", key, present, wantPresent, got)
	}
	if got != want {
		t.Fatalf("Read(%#v) value = %#v, want %#v", key, got, want)
	}
}

func TestScalarArrayCopyTiming(t *testing.T) {
	t.Run("last external owner released before separation snapshots the element", func(t *testing.T) {
		external := NewSlot(int64(1))
		defer external.Unset()
		a := NewScalarArray()
		defer a.Close()
		a.Bind("ref", external)
		b := a.Copy()
		defer b.Close()

		external.Unset()
		b.Write("ref", int64(2))
		requireScalarArrayRead(t, a, "ref", int64(1), true)
		requireScalarArrayRead(t, b, "ref", int64(2), true)
	})

	t.Run("separation while an external owner survives retains the alias", func(t *testing.T) {
		external := NewSlot(int64(1))
		defer external.Unset()
		a := NewScalarArray()
		defer a.Close()
		a.Bind("ref", external)
		b := a.Copy()
		defer b.Close()

		b.Write("plain", int64(0))
		external.Unset()
		b.Write("ref", int64(2))
		requireScalarArrayRead(t, a, "ref", int64(2), true)
		requireScalarArrayRead(t, b, "ref", int64(2), true)
	})

	t.Run("live external owner remains aliased after copy separation", func(t *testing.T) {
		external := NewSlot(int64(1))
		defer external.Unset()
		a := NewScalarArray()
		defer a.Close()
		a.Bind(int64(1), external)
		b := a.Copy()
		defer b.Close()

		b.Write(int64(1), int64(2))
		requireRead(t, external, int64(2), true)
		requireScalarArrayRead(t, a, int64(1), int64(2), true)
		requireScalarArrayRead(t, b, int64(1), int64(2), true)
	})

	t.Run("multiple keys that share a cell retain their internal aliases", func(t *testing.T) {
		external := NewSlot(int64(1))
		defer external.Unset()
		a := NewScalarArray()
		defer a.Close()
		a.Bind("first", external)
		a.Bind("second", external)
		b := a.Copy()
		defer b.Close()
		external.Unset() // The two array bindings alone must preserve this alias.

		b.Write("plain", int64(0))
		b.Write("first", int64(9))
		for _, array := range []*ScalarArray{a, b} {
			requireScalarArrayRead(t, array, "first", int64(9), true)
			requireScalarArrayRead(t, array, "second", int64(9), true)
		}
		requireRead(t, external, nil, false)
	})
}

func TestScalarArrayReferenceSeparation(t *testing.T) {
	var a ScalarArray
	defer a.Close()
	if value, present := a.Read("missing"); present || value != nil {
		t.Fatalf("zero Read(missing) = (%#v, %v), want (nil, false)", value, present)
	}
	if keys := a.Keys(); keys != nil {
		t.Fatalf("zero Keys() = %#v, want nil", keys)
	}
	a.Write("plain", int64(1))
	b := a.Copy()
	defer b.Close()

	escaped := b.Reference("plain")
	escaped.Write(int64(2))
	requireScalarArrayRead(t, &a, "plain", int64(1), true)
	requireScalarArrayRead(t, b, "plain", int64(2), true)
	escaped.Unset()
	requireScalarArrayRead(t, b, "plain", int64(2), true)

	c := a.Copy()
	defer c.Close()
	missing := c.Reference("missing")
	defer missing.Unset()
	requireRead(t, missing, nil, true)
	requireScalarArrayRead(t, &a, "missing", nil, false)
	requireScalarArrayRead(t, c, "missing", nil, true)

	b.Write("bound", "inside")
	rebound := b.Reference("bound")
	replacement := NewSlot("outside")
	defer replacement.Unset()
	rebound.Bind(replacement)
	rebound.Write("replacement-value")
	requireScalarArrayRead(t, b, "bound", "inside", true)
	requireRead(t, replacement, "replacement-value", true)
	rebound.Unset()
	requireScalarArrayRead(t, b, "bound", "inside", true)
}

func TestScalarArrayRebinding(t *testing.T) {
	oldExternal := NewSlot("old")
	defer oldExternal.Unset()
	a := NewScalarArray()
	defer a.Close()
	a.Bind("ref", oldExternal)
	a.Bind("peer", oldExternal)
	b := a.Copy()
	defer b.Close()

	replacement := NewSlot("replacement")
	defer replacement.Unset()
	b.Bind("ref", replacement)
	requireScalarArrayRead(t, a, "ref", "old", true)
	requireScalarArrayRead(t, a, "peer", "old", true)
	requireScalarArrayRead(t, b, "peer", "old", true)
	requireScalarArrayRead(t, b, "ref", "replacement", true)

	replacement.Write("from-external")
	requireScalarArrayRead(t, b, "ref", "from-external", true)
	replacement.Unset()
	b.Write("ref", "after-release")
	requireScalarArrayRead(t, b, "ref", "after-release", true)
	requireScalarArrayRead(t, a, "ref", "old", true)
	requireScalarArrayRead(t, a, "peer", "old", true)

	undefined := new(Slot)
	defer undefined.Unset()
	b.Bind("null", undefined)
	requireRead(t, undefined, nil, true)
	requireScalarArrayRead(t, b, "null", nil, true)
}
