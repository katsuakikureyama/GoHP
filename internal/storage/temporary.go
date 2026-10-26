package storage

// Temporary owns one value snapshot, never a source binding or writable
// location. Use only by pointer, without concurrent access or Go struct copies.
// A zero Temporary is inactive; construct a live null with NewTemporary(nil).
// Each live temporary must be consumed by MoveTo/Take or explicitly Closed.
// Go GC/finalizers are not semantic release.
type Temporary struct {
	value  any
	active bool
}

// NewTemporary retains an independent value owner. Array inputs remain owned
// by the caller; their container is shared, not eagerly separated. Invalid
// payloads reject through the same value-domain guard as Slot.Write.
func NewTemporary(value any) *Temporary {
	return &Temporary{value: retainValue(value), active: true}
}

// Snapshot captures the current value once without retaining the outer alias.
// Absence returns (nil, false), without creating a null binding or temporary.
// The caller owns a present result, even when it contains null. In particular,
// this does not call Read then retain again and leak an intermediate handle.
func (slot *Slot) Snapshot() (*Temporary, bool) {
	if slot == nil {
		panic("storage snapshot requires a slot")
	}
	value, present := slot.borrow()
	if !present {
		return nil, false
	}
	return NewTemporary(value), true
}

func (temporary *Temporary) requireActive() {
	if temporary == nil || !temporary.active {
		panic("storage operation requires an active temporary")
	}
}

// MoveTo writes the captured value through destination's CURRENT binding and
// consumes this temporary only on success. Write retains before releasing the
// previous destination, including when source and destination were the same
// Slot. A domain/receiver failure leaves the temporary owned by the caller for
// retry or deferred Close. It never binds destination to the snapshot's source.
func (temporary *Temporary) MoveTo(destination *Slot) {
	temporary.requireActive()
	if destination == nil {
		panic("storage temporary move requires a slot")
	}
	destination.Write(temporary.value)
	temporary.Close()
}

// Take consumes the temporary and transfers its already-owned payload, without
// copying or releasing it. An array result must be Closed by its new owner.
// A deferred Close on this temporary is safe after Take; no borrowed handle is
// exposed while the temporary remains active.
func (temporary *Temporary) Take() any {
	temporary.requireActive()
	value := temporary.value
	temporary.value, temporary.active = nil, false
	return value
}

// Close releases any remaining value owner exactly once. Close is idempotent
// for inactive/consumed temporaries, including the zero value; nil rejects.
// Use defer as a failure guard, but release at the operation's specified
// lifetime boundary: keeping an array temporary alive can affect later COW.
func (temporary *Temporary) Close() {
	if temporary == nil {
		panic("storage close requires a temporary")
	}
	if !temporary.active {
		return
	}
	value := temporary.value
	temporary.value, temporary.active = nil, false
	releaseValue(value)
}
