package runtime

import "github.com/katsuakikureyama/GoHP/internal/storage"

// ScalarFrame is the generated-code ABI for scalar local variables in one PHP
// scope. Its zero value is open and empty. It must be used by pointer, must
// not be copied, and is not safe for concurrent use.
//
// Generated scope functions create one frame and defer Close on every exit
// path. Close releases only the frame's local bindings; it has no finalizer or
// garbage-collection semantics. ImportScalarGlobal is a guarded scalar-only
// host operation; generated PHP does not use it yet. No raw reference, slot,
// array, object or callable API is exposed.
type ScalarFrame struct {
	frame     storage.Frame
	execution *Execution
}

// NewScalarFrame returns an open frame for generated scalar local bindings.
func NewScalarFrame() *ScalarFrame {
	return &ScalarFrame{}
}

func (frame *ScalarFrame) storageFrame() *storage.Frame {
	if frame == nil {
		panic("runtime scalar frame requires a frame")
	}
	if frame.execution != nil {
		RequireExecution(frame.execution)
	}
	return &frame.frame
}

// Read returns a defined scalar local value. Reading an undefined local is a
// generated-code error, whereas a defined PHP null is returned as nil.
func (frame *ScalarFrame) Read(name string) any {
	value, present := frame.storageFrame().Read(name)
	if !present {
		panic("undefined scalar frame variable")
	}
	return value
}

// Write assigns an exact scalar value. Validation and storage ownership are
// delegated to storage.Frame so this ABI cannot diverge from the canonical
// scalar domain.
func (frame *ScalarFrame) Write(name string, value any) {
	frame.storageFrame().Write(name, value)
}

// Close releases local bindings. It is idempotent; later reads and writes are
// rejected by the canonical frame.
func (frame *ScalarFrame) Close() {
	if frame == nil {
		panic("runtime scalar frame requires a frame")
	}
	frame.frame.Close()
	if frame.execution != nil {
		delete(frame.execution.scalarFrames, frame)
	}
}
