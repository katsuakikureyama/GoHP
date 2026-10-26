// Package storage contains the internal reference-storage foundation. It is
// exposed to closed scalar generated functions through runtime.ScalarFrame;
// reference bindings remain internal until address IR is integrated. General
// slots own scalars or scalar-element arrays. Frame and ScalarArray bindings
// remain scalar-only, including writes through their exported aliases.
package storage

// Slot is one binding, not a PHP value. Slots must only be used by pointer and
// must not be copied or accessed concurrently. Owners must call Unset at the end
// of a binding's lifetime; Go garbage collection is not semantic release.
type Slot struct {
	defined bool
	value   any
	ref     *cell
	// Immutable binding domain; retained across Unset and rebinding.
	scalarOnly bool
}

// cell counts bindings, not Go pointers. A single remaining owner stays valid;
// array separation, not scalar reads, decides when to unwrap cells.
type cell struct {
	value        any
	owners       int
	scalarOwners int
}

func newScalarSlot() *Slot { return &Slot{scalarOnly: true} }

// scalarSource validates a constrained Bind before any destination allocation
// or array separation. Borrowing here never creates an owned temporary.
func scalarSource(source *Slot) {
	if source == nil {
		panic("storage reference requires a slot")
	}
	value, _ := source.borrow()
	scalar(value)
}

func retainValue(value any) any {
	if array, ok := value.(*ScalarArray); ok {
		return array.Copy() // Rejects nil/closed handles before retaining.
	}
	scalar(value)
	return value
}

func releaseValue(value any) {
	if array, ok := value.(*ScalarArray); ok {
		array.Close()
	}
}

func scalar(value any) {
	switch value.(type) {
	case nil, bool, int64, string:
	default:
		panic("storage slot requires null, bool, int64 or string")
	}
}

func NewSlot(value any) *Slot {
	slot := &Slot{}
	slot.Write(value)
	return slot
}

// borrow is internal only: it must not expose an owned array payload handle.
func (slot *Slot) borrow() (any, bool) {
	if slot.ref != nil {
		return slot.ref.value, true
	}
	return slot.value, slot.defined
}

// Read dereferences without exporting the binding. An array result is a fresh
// caller-owned handle that must be Closed; scalar results need no release.
// Missing differs from present null and does not allocate storage.
func (slot *Slot) Read() (any, bool) {
	value, present := slot.borrow()
	if !present {
		return nil, false
	}
	return retainValue(value), true
}

// Write retains the RHS before releasing the previous owned payload. Array
// handles are copied, not consumed. Any live scalar-only binding constrains all
// writes to its cell, preventing aliases from introducing nesting or cycles.
func (slot *Slot) Write(value any) {
	if slot.scalarOnly || slot.ref != nil && slot.ref.scalarOwners > 0 {
		scalar(value)
	}
	next := retainValue(value)
	previous := slot.value
	if slot.ref != nil {
		previous = slot.ref.value
		slot.ref.value = next
	} else {
		slot.value, slot.defined = next, true
	}
	releaseValue(previous)
}

// Bind aliases source, initializing an undefined source to null. Rebinding this
// slot releases only its old binding, not the bindings of its previous peers.
func (slot *Slot) Bind(source *Slot) {
	if slot == nil || source == nil {
		panic("storage reference requires a slot")
	}
	if slot.scalarOnly {
		scalarSource(source)
	}
	if !source.defined {
		source.Write(nil)
	}
	if slot == source || slot.ref != nil && slot.ref == source.ref {
		return
	}
	if source.ref == nil {
		source.ref = &cell{value: source.value, owners: 1}
		if source.scalarOnly {
			source.ref.scalarOwners = 1
		}
		source.value = nil
	}
	// Retain before releasing the destination; lifetime must not depend on GC.
	next := source.ref
	next.owners++
	if slot.scalarOnly {
		next.scalarOwners++
	}
	slot.Unset()
	slot.ref, slot.defined = next, true
}

// Unset drops only this binding. Reusing the slot creates an independent value.
// It is also the explicit release operation for local/temporary bindings.
func (slot *Slot) Unset() {
	previous := slot.value
	if slot.ref != nil {
		previous = nil
		slot.ref.owners--
		if slot.scalarOnly {
			slot.ref.scalarOwners--
		}
		if slot.ref.owners == 0 {
			previous = slot.ref.value
			slot.ref.value = nil
		}
	}
	slot.ref, slot.value, slot.defined = nil, nil, false
	releaseValue(previous)
}
