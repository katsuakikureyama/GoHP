package runtime

// scalarSlot is an internal handle to one binding, not a PHP-visible value.
// Copying a Go handle keeps the same binding identity and adds no reference owner.
// Use newScalarSlot and bind to create a distinct, independently unsettable alias.
// This scalar-only foundation is not used by generated code yet. It must not be
// extended to arrays via Clone: reference-bearing arrays require shared storage
// and separation-time ownership rules (docs/reference-value-model.md).
// Operations are synchronous and not safe for concurrent mutation.
type scalarSlot struct {
	binding *scalarBinding
}

type scalarBinding struct {
	defined bool
	value   any
	ref     *scalarRefCell
}

// owners counts distinct live bindings, not Go handles or Go GC reachability.
// Owners must explicitly unset or rebind; there are no finalizers. Arrays,
// objects, frame cleanup, escaping temporaries and cycles are outside this slice.
type scalarRefCell struct {
	value  any
	owners int
}

func newScalarSlot() scalarSlot {
	return scalarSlot{binding: &scalarBinding{}}
}

func (s scalarSlot) requireBinding() *scalarBinding {
	if s.binding == nil {
		panic("uninitialized scalar slot handle")
	}
	return s.binding
}

func (s scalarSlot) isDefined() bool {
	return s.requireBinding().defined
}

// read returns a scalar value, never a slot/cell. Undefined is not present null.
func (s scalarSlot) read() any {
	b := s.requireBinding()
	if !b.defined {
		panic("undefined scalar slot")
	}
	if b.ref != nil {
		return b.ref.value
	}
	return b.value
}

// assign writes through any existing reference; it does not rebind the slot.
// Reject unsupported values before changing either the value or its binding.
func (s scalarSlot) assign(value any) {
	b := s.requireBinding()
	switch value.(type) {
	case nil, int64, string, bool:
	default:
		panic("scalar slot requires int64, string, bool or null")
	}
	if b.ref != nil {
		b.ref.value = value
	} else {
		b.value = value
	}
	b.defined = true
}

// bind changes only the destination binding. Binding an undefined source is
// deliberately rejected until reference-creation/address semantics are added.
// Self-binding and rebinding within one alias set must not add logical owners.
func (s scalarSlot) bind(source scalarSlot) {
	dst := s.requireBinding()
	src := source.requireBinding()
	if !src.defined {
		panic("cannot bind undefined scalar slot")
	}
	if dst == src || (src.ref != nil && dst.ref == src.ref) {
		return
	}
	if src.ref == nil {
		src.ref = &scalarRefCell{value: src.value, owners: 1}
		src.value = nil
	}
	s.unset()
	src.ref.owners++
	dst.ref = src.ref
	dst.defined = true
}

// unset removes this binding only. Repeated unset is harmless, and a subsequent
// assign recreates independent storage without relinking surviving aliases.
func (s scalarSlot) unset() {
	b := s.requireBinding()
	if b.ref != nil {
		b.ref.owners--
		if b.ref.owners == 0 {
			b.ref.value = nil
		}
	}
	b.ref = nil
	b.value = nil
	b.defined = false
}
