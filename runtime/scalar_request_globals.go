package runtime

// requireScalarGlobalName limits this host ABI to ordinary literal identifier
// names. It is not PHP variable-variable, superglobal or $GLOBALS admission.
func requireScalarGlobalName(name string) {
	letter := func(b byte) bool { return b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' }
	if name == "" || !letter(name[0]) || name == "this" || name == "GLOBALS" {
		panic("scalar global requires an ordinary ASCII identifier")
	}
	for i := 1; i < len(name); i++ {
		b := name[i]
		if !letter(b) && !(b >= '0' && b <= '9') && b != '_' {
			panic("scalar global requires an ordinary ASCII identifier")
		}
	}
}

// ReadScalarGlobal returns presence separately from null without creating a
// binding. Values are exactly nil, bool, int64 or string, never a Slot handle.
func (execution *Execution) ReadScalarGlobal(name string) (any, bool) {
	RequireExecution(execution)
	requireScalarGlobalName(name)
	return execution.scalarGlobals.Read(name)
}

// WriteScalarGlobal assigns a scalar through the current global binding.
// Canonical storage validates the value before allocating or mutating a slot.
func (execution *Execution) WriteScalarGlobal(name string, value any) {
	RequireExecution(execution)
	requireScalarGlobalName(name)
	execution.scalarGlobals.Write(name, value)
}

// UnsetScalarGlobal removes only this name. Existing imported locals retain
// the old cell; a recreated global name acquires an independent binding.
func (execution *Execution) UnsetScalarGlobal(name string) {
	RequireExecution(execution)
	requireScalarGlobalName(name)
	execution.scalarGlobals.Unset(name)
}

// ImportScalarGlobal replaces a same-named local with an alias of the request
// global. An absent global becomes present null. A frame is attached to one
// Execution for its remaining lifetime and cannot move between requests.
// This bounded host ABI does not admit PHP global/reference source syntax.
func (frame *ScalarFrame) ImportScalarGlobal(execution *Execution, name string) {
	local := frame.storageFrame()
	RequireExecution(execution)
	requireScalarGlobalName(name)
	if frame.execution != nil && frame.execution != execution {
		panic("scalar frame belongs to a different execution context")
	}
	// Validate the canonical local lifetime before creating a global cell.
	// A read is noncreating and is also valid for an absent local.
	local.Read(name)
	source := execution.scalarGlobals.Reference(name)
	defer source.Unset()
	local.Bind(name, source)
	if frame.execution == nil {
		frame.execution = execution
		if execution.scalarFrames == nil {
			execution.scalarFrames = make(map[*ScalarFrame]struct{})
		}
		execution.scalarFrames[frame] = struct{}{}
	}
}

// Unset removes a local binding without removing the request-global name.
// An ordinary subsequent Write is local until another explicit import.
func (frame *ScalarFrame) Unset(name string) {
	frame.storageFrame().Unset(name)
}

// Close ends this host execution. It releases scalar globals and imported
// scalar frames, clears existing tokenizer/error state and rejects stale use.
// It is idempotent. This is not an HTTP lifecycle, a PHP destructor order or
// ownership enforcement for arbitrary objects/callbacks held by the host.
func (execution *Execution) Close() {
	if execution == nil {
		panic("execution context is required")
	}
	if execution.closed {
		return
	}
	execution.closed = true
	execution.scalarGlobals.Close()
	for frame := range execution.scalarFrames {
		frame.Close()
	}
	execution.scalarFrames = nil
	execution.strtokText = ""
	execution.strtokPos = 0
	execution.strtokReady = false
	execution.preg83Error = 0
	execution.jsonError = 0
}
