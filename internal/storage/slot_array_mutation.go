package storage

// arrayForMutation borrows the CURRENT binding's owned payload for one
// synchronous operation. It must never escape, run user code, or survive a root
// Write/Bind/Unset. Slot/array operations are non-concurrent. No Read/Copy or
// reference promotion is needed: the root already owns the handle throughout
// the call. Copying here would introduce an artificial container owner and
// potentially change the timing of element-reference separation.
func (slot *Slot) arrayForMutation() *ScalarArray {
	if slot == nil {
		panic("storage array mutation requires a defined array slot")
	}
	value, present := slot.borrow()
	array, ok := value.(*ScalarArray)
	if !present || !ok {
		panic("storage array mutation requires a defined array slot")
	}
	array.requireOpen()
	return array
}

// ArrayWrite assigns a scalar through a one-level element of the current
// array-valued binding. Keys and values are validated before separation. Outer
// aliases observe the write; ordinary array value copies separate as needed.
// An absent/null/scalar root rejects rather than becoming an array.
//
// All arguments have already been evaluated when the operation resolves the
// root. These synchronous internal operations are not persistent locations or
// a PHP evaluation-order policy; generated address lowering remains separate.
func (slot *Slot) ArrayWrite(key, value any) {
	slot.arrayForMutation().Write(key, value)
}

// ArrayAppend appends a scalar to the current array-valued binding and returns
// its automatic key. Invalid values/index overflow reject before separation.
func (slot *Slot) ArrayAppend(value any) int64 {
	return slot.arrayForMutation().Append(value)
}

// ArrayBind rebinds only the selected element to source. Validation happens
// before separation or source initialization. The source remains independently
// owned; a valid undefined source becomes present null. No array-valued sources
// are admitted into this scalar-only container.
func (slot *Slot) ArrayBind(key any, source *Slot) {
	slot.arrayForMutation().Bind(key, source)
}

// ArrayReference separates the current container as needed and returns a fresh
// caller-owned alias, never its owned element Slot. A missing key becomes null.
// The alias can outlive the root; the caller must Unset it. Rebinding this alias
// does not rebind the array element (use ArrayBind for that operation).
func (slot *Slot) ArrayReference(key any) *Slot {
	return slot.arrayForMutation().Reference(key)
}
