package storage

import "testing"

func TestScalarBindingArrayGuards(t *testing.T) {
	payload := NewScalarArray()
	payload.Write("value", "kept")
	defer payload.Close()

	// An ordinary Slot may own an array. Reading it returns a separate owned
	// handle, rather than exposing the binding's handle.
	general := NewSlot(payload)
	defer general.Unset()
	owned := readOwnedArray(t, general)
	requireArrayRead(t, owned, "value", "kept", true)
	owned.Close()

	frame := NewFrame()
	defer frame.Close()
	frame.Write("frame", "scalar")
	elements := NewScalarArray()
	defer elements.Close()
	elements.Write("element", "scalar")

	// Direct writes through scalar-only owners reject an array. The same is
	// true through aliases exported from those owners.
	requirePanic(t, func() { frame.Write("frame", payload) })
	requirePanic(t, func() { elements.Write("element", payload) })
	frameAlias := frame.Reference("frame")
	defer frameAlias.Unset()
	elementAlias := elements.Reference("element")
	defer elementAlias.Unset()
	requirePanic(t, func() { frameAlias.Write(payload) })
	requirePanic(t, func() { elementAlias.Write(payload) })

	// A general source becomes constrained when a scalar-only owner binds it.
	// This prevents aliases from introducing an indirect nested-array cycle.
	source := NewSlot("scalar")
	defer source.Unset()
	frame.Bind("source", source)
	requirePanic(t, func() { source.Write(payload) })
	requirePanic(t, func() { elements.Bind("indirect", general) })
	requirePanic(t, func() { elements.Write("self", elements) })
	requirePanic(t, func() { elementAlias.Write(elements) })
	requirePanic(t, func() { payload.Bind("self", general) })

	var nilArray *ScalarArray
	closed := NewScalarArray()
	closed.Close()
	for _, invalid := range []any{nilArray, closed, new(Slot), []any{"foreign"}} {
		requirePanic(t, func() { general.Write(invalid) })
		requirePanic(t, func() { frame.Write("frame", invalid) })
		requirePanic(t, func() { elements.Write("element", invalid) })
	}
	requireRead(t, frameAlias, "scalar", true)
	requireRead(t, elementAlias, "scalar", true)
	requireRead(t, source, "scalar", true)
	requireArrayRead(t, elements, "element", "scalar", true)
	requireArrayRead(t, payload, "self", nil, false)
}

func TestScalarBindingRestrictionsFollowOwners(t *testing.T) {
	source := NewSlot("seed")
	defer source.Unset()
	frame := NewFrame()
	defer frame.Close()
	frame.Bind("frame", source)

	elements := NewScalarArray()
	elements.Bind("element", source)
	copy := elements.Copy()
	defer copy.Close()

	if source.ref == nil || source.ref.owners != 3 || source.ref.scalarOwners != 2 {
		t.Fatalf("initial binding counts = %#v, want 3 owners and 2 scalar owners", source.ref)
	}
	// Retaining a container does not create another element binding.
	if elements.data != copy.data || source.ref.owners != 3 {
		t.Fatal("Copy() added an element binding")
	}
	copy.Write("separate", "copy")
	if source.ref.owners != 4 || source.ref.scalarOwners != 3 {
		t.Fatalf("separation counts = %#v, want 4 owners and 3 scalar owners", source.ref)
	}

	payload := NewScalarArray()
	defer payload.Close()
	requirePanic(t, func() { source.Write(payload) })

	// Each release affects only the selected constrained binding.
	frame.Unset("frame")
	if source.ref.owners != 3 || source.ref.scalarOwners != 2 {
		t.Fatalf("frame Unset counts = %#v, want 3 owners and 2 scalar owners", source.ref)
	}
	requirePanic(t, func() { source.Write(payload) })
	elements.Close()
	if source.ref.owners != 2 || source.ref.scalarOwners != 1 {
		t.Fatalf("array Close counts = %#v, want 2 owners and 1 scalar owner", source.ref)
	}
	requirePanic(t, func() { source.Write(payload) })

	// Rebinding the final scalar-only entry frees the surviving general alias
	// to own an array, without disturbing the replacement binding.
	replacement := NewSlot("replacement")
	defer replacement.Unset()
	copy.Bind("element", replacement)
	if source.ref.owners != 1 || source.ref.scalarOwners != 0 {
		t.Fatalf("final element rebind counts = %#v, want one unrestricted owner", source.ref)
	}
	source.Write(payload)
	owned := readOwnedArray(t, source)
	owned.Close()

	// A Frame's scalar domain belongs to each local binding, including a new
	// binding created after Unset and an existing binding that is rebound.
	first := NewSlot("first")
	defer first.Unset()
	second := NewSlot("second")
	defer second.Unset()
	domain := NewFrame()
	defer domain.Close()
	domain.Bind("value", first)
	domain.Bind("value", second)
	requirePanic(t, func() { second.Write(payload) })
	domain.Unset("value")
	third := NewSlot("third")
	defer third.Unset()
	domain.Bind("value", third)
	requirePanic(t, func() { third.Write(payload) })

	reused := newScalarSlot()
	reused.Write("before")
	reused.Unset()
	requirePanic(t, func() { reused.Write(payload) })
	reused.Bind(third)
	reused.Unset()
	requirePanic(t, func() { reused.Bind(source) }) // source now owns an array.

	// An exported alias is not permanently restricted by its former owner.
	escapingFrame := NewFrame()
	escapingFrame.Write("value", "before close")
	escaped := escapingFrame.Reference("value")
	defer escaped.Unset()
	escapingFrame.Close()
	escaped.Write(payload)
	escapedValue := readOwnedArray(t, escaped)
	escapedValue.Close()
}

func TestScalarArrayBindingRejectsBeforeSeparation(t *testing.T) {
	external := NewSlot(int64(1))
	defer external.Unset()
	left := NewScalarArray()
	defer left.Close()
	left.Bind("ref", external)
	right := left.Copy()
	defer right.Close()

	payload := NewScalarArray()
	payload.Write("value", "source stays intact")
	defer payload.Close()
	arraySource := NewSlot(payload)
	defer arraySource.Unset()

	shared := left.data
	requirePanic(t, func() { left.Bind("ref", arraySource) })
	requirePanic(t, func() { left.Bind("missing", arraySource) })
	if left.data != shared || right.data != shared {
		t.Fatal("rejected array Bind separated shared storage")
	}
	if _, present := left.Read("missing"); present {
		t.Fatal("rejected missing-key Bind created an element")
	}

	// Frame validates the source before local() creates a missing binding.
	frame := NewFrame()
	defer frame.Close()
	requirePanic(t, func() { frame.Bind("missing", arraySource) })
	if frame.slots != nil {
		t.Fatal("rejected Frame.Bind created a missing binding")
	}

	// The source remained an owned array value after all rejected bindings.
	owned := readOwnedArray(t, arraySource)
	requireArrayRead(t, owned, "value", "source stays intact", true)
	owned.Close()

	// If either invalid Bind had separated left while external was live, the
	// later release would retain the element reference in both copies and this
	// write would also change right. The required result proves no detachment.
	external.Unset()
	left.Write("ref", int64(2))
	requireArrayRead(t, left, "ref", int64(2), true)
	requireArrayRead(t, right, "ref", int64(1), true)
}
