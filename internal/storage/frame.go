package storage

// Frame owns named scalar bindings for one scope. Its zero value is open and
// empty. Frames must only be used by pointer, must not be copied, and are not
// concurrency-safe. The owner must call Close on every exit path (normally with
// defer); garbage collection does not release logical reference owners.
//
// Frame never exposes its owned slots. Reference returns a separate binding
// whose lifetime belongs to the caller, so it can outlive Close safely.
type Frame struct {
	slots  map[string]*Slot
	closed bool
}

func NewFrame() *Frame {
	return &Frame{}
}

func (frame *Frame) requireOpen() {
	if frame == nil || frame.closed {
		panic("storage operation requires an open frame")
	}
}

// local is only called after validating the frame and operation inputs.
func (frame *Frame) local(name string) *Slot {
	if slot := frame.slots[name]; slot != nil {
		return slot
	}
	if frame.slots == nil {
		frame.slots = make(map[string]*Slot)
	}
	slot := &Slot{}
	frame.slots[name] = slot
	return slot
}

// Read returns a value and presence without creating an absent binding.
func (frame *Frame) Read(name string) (any, bool) {
	frame.requireOpen()
	if slot := frame.slots[name]; slot != nil {
		return slot.Read()
	}
	return nil, false
}

// Write assigns through the local binding. Validate before creating storage,
// as well as before changing an existing ordinary value or reference group.
func (frame *Frame) Write(name string, value any) {
	frame.requireOpen()
	scalar(value)
	frame.local(name).Write(value)
}

// Bind aliases source through a frame-owned binding. Rebinding the local does
// not rebind source or its other aliases. An undefined source becomes null.
func (frame *Frame) Bind(name string, source *Slot) {
	frame.requireOpen()
	if source == nil {
		panic("storage reference requires a slot")
	}
	frame.local(name).Bind(source)
}

// Reference retains a new, caller-owned alias of the local binding. An absent
// name becomes present null. The caller must Unset the returned Slot when done;
// closing the frame releases only its own binding, not this returned alias.
func (frame *Frame) Reference(name string) *Slot {
	frame.requireOpen()
	alias := &Slot{}
	alias.Bind(frame.local(name))
	return alias
}

// Unset removes only the named binding. A later use of the name creates an
// independent binding, while existing exported references remain valid.
func (frame *Frame) Unset(name string) {
	frame.requireOpen()
	if slot := frame.slots[name]; slot != nil {
		slot.Unset()
		delete(frame.slots, name)
	}
}

// Close deterministically releases all frame-owned bindings. It is idempotent;
// every other operation rejects a closed frame. No PHP destructor order is
// implied: only scalars are admitted, and generated PHP does not use Frame yet.
func (frame *Frame) Close() {
	if frame == nil {
		panic("storage close requires a frame")
	}
	if frame.closed {
		return
	}
	frame.closed = true
	for _, slot := range frame.slots {
		slot.Unset()
	}
	frame.slots = nil
}
