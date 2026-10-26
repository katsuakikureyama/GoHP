// Package storage contains the internal reference-storage foundation. It is
// exposed to closed scalar generated functions through runtime.ScalarFrame;
// reference bindings remain internal until address IR is integrated. Only
// null, bool, int64 and string values are admitted.
package storage

// Slot is one binding, not a PHP value. Slots must only be used by pointer and
// must not be copied or accessed concurrently. Owners must call Unset at the end
// of a binding's lifetime; Go garbage collection is not semantic release.
type Slot struct {
	defined bool
	value   any
	ref     *cell
}

// cell counts bindings, not Go pointers. A single remaining owner stays valid;
// future array separation, not scalar reads, will decide when to unwrap cells.
type cell struct {
	value  any
	owners int
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

// Read returns a scalar value, never the reference binding. A missing binding
// is distinct from a present null and reading it does not create storage.
func (slot *Slot) Read() (any, bool) {
	if slot.ref != nil {
		return slot.ref.value, true
	}
	return slot.value, slot.defined
}

// Write assigns a value through the existing binding. Validate before touching
// state so unsupported arrays/objects cannot enter or corrupt an alias group.
func (slot *Slot) Write(value any) {
	scalar(value)
	if slot.ref != nil {
		slot.ref.value = value
		return
	}
	slot.value, slot.defined = value, true
}

// Bind aliases source, initializing an undefined source to null. Rebinding this
// slot releases only its old binding, not the bindings of its previous peers.
func (slot *Slot) Bind(source *Slot) {
	if slot == nil || source == nil {
		panic("storage reference requires a slot")
	}
	if !source.defined {
		source.Write(nil)
	}
	if slot == source || slot.ref != nil && slot.ref == source.ref {
		return
	}
	if source.ref == nil {
		source.ref = &cell{value: source.value, owners: 1}
		source.value = nil
	}
	// Retain before releasing the destination; lifetime must not depend on GC.
	next := source.ref
	next.owners++
	slot.Unset()
	slot.ref, slot.defined = next, true
}

// Unset drops only this binding. Reusing the slot creates an independent value.
// It is also the explicit release operation for local/temporary bindings.
func (slot *Slot) Unset() {
	if slot.ref != nil {
		slot.ref.owners--
		if slot.ref.owners == 0 {
			slot.ref.value = nil
		}
	}
	slot.ref, slot.value, slot.defined = nil, nil, false
}
