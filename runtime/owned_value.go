package runtime

import (
	"github.com/katsuakikureyama/GoHP/internal/arraykey"
	"github.com/katsuakikureyama/GoHP/internal/storage"
)

// ValueSlot is an opt-in, value-only runtime boundary over canonical storage.
// It accepts exact scalars and canonical flat scalar Arrays, not arbitrary PHP
// values. Its zero value is open but undefined. Use by pointer, never copy or
// access concurrently, and Close explicitly. Generated use needs a closed proof.
// No internal handle, binding or PHP reference operation is exposed.
type ValueSlot struct {
	value  storage.Slot
	closed bool
}

func NewValueSlot(value any) *ValueSlot {
	slot := &ValueSlot{}
	slot.Write(value)
	return slot
}

func (slot *ValueSlot) requireOpen() {
	if slot == nil || slot.closed {
		panic("runtime value operation requires an open slot")
	}
}

// Write imports a detached host value before replacing the current value.
// Invalid inputs leave the destination unchanged; imported owners are released
// on every exit. Existing runtime.Array operations are not changed by this API.
func (slot *ValueSlot) Write(value any) {
	slot.requireOpen()
	temporary := captureRuntimeValue(value)
	defer temporary.Close()
	temporary.MoveTo(&slot.value)
}

// Read exports a fresh legacy value. In particular, Array.Entries never shares
// a backing slice with internal storage or another exported array.
func (slot *ValueSlot) Read() any {
	temporary := slot.Snapshot()
	defer temporary.Close()
	return temporary.Take()
}

// Snapshot retains the current value, not this slot's binding. It can outlive
// replacement or Close. An undefined slot rejects rather than becoming null.
func (slot *ValueSlot) Snapshot() *ValueTemporary {
	slot.requireOpen()
	temporary, present := slot.value.Snapshot()
	if !present {
		panic("undefined runtime value slot")
	}
	return &ValueTemporary{value: temporary}
}

// ArrayWrite/ArrayAppend operate on the current owned array without exporting
// a read copy. Only scalar elements and bounded int64/string keys are admitted.
func (slot *ValueSlot) ArrayWrite(key, value any) {
	slot.requireOpen()
	slot.value.ArrayWrite(key, value)
}

func (slot *ValueSlot) ArrayAppend(value any) int64 {
	slot.requireOpen()
	return slot.value.ArrayAppend(value)
}

// Close releases the owned value once. All other operations reject thereafter.
func (slot *ValueSlot) Close() {
	if slot == nil {
		panic("runtime close requires a value slot")
	}
	if !slot.closed {
		slot.closed = true
		slot.value.Unset()
	}
}

// ValueTemporary is an owned snapshot, not a PHP reference or writable location.
// Consume it with MoveTo/Take or Close it at the operation's lifetime boundary.
// A deferred Close is a failure guard, not permission to retain values until
// function exit. It is pointer-only/non-concurrent; its zero value is inactive.
type ValueTemporary struct {
	value *storage.Temporary
}

func NewValueTemporary(value any) *ValueTemporary {
	return &ValueTemporary{value: captureRuntimeValue(value)}
}

func (temporary *ValueTemporary) requireValue() *storage.Temporary {
	if temporary == nil || temporary.value == nil {
		panic("runtime operation requires a value temporary")
	}
	return temporary.value
}

// MoveTo consumes only after successful assignment. A nil/closed destination
// leaves the temporary available for another destination or cleanup.
func (temporary *ValueTemporary) MoveTo(destination *ValueSlot) {
	value := temporary.requireValue()
	destination.requireOpen()
	value.MoveTo(&destination.value)
}

// Take consumes and exports a detached legacy value, never an internal owned
// handle. The caller need not Close its result. Deferred temporary.Close remains
// safe, and internal array ownership is released even if export panics.
func (temporary *ValueTemporary) Take() any {
	value := temporary.requireValue().Take()
	if array, ok := value.(*storage.ScalarArray); ok {
		defer array.Close()
		keys := array.Keys()
		entries := make([]Entry, 0, len(keys))
		for _, key := range keys {
			element, _ := array.Read(key)
			entries = append(entries, Entry{Key: key, Value: element})
		}
		// Import requires an index reconstructible from ordered keys; the
		// only exposed mutations are write/append, never delete/reorder.
		return NewArray(entries...)
	}
	return value
}

func (temporary *ValueTemporary) Close() {
	if temporary == nil {
		panic("runtime close requires a value temporary")
	}
	if temporary.value != nil {
		temporary.value.Close()
	}
}

// captureRuntimeValue is the only legacy-to-storage adapter. It preflights the
// entire flat array before allocating owned storage. Public Entries may have
// been edited by a host: do not silently normalize, collapse duplicate keys or
// discard hidden index history. Non-reconstructible metadata is outside this
// opt-in boundary, not a new restriction on existing Array APIs.
func captureRuntimeValue(value any) *storage.Temporary {
	switch input := value.(type) {
	case nil, bool, int64, string:
		return storage.NewTemporary(input)
	case Array:
		seen := make(map[any]bool, len(input.Entries))
		var index arraykey.Index
		for _, entry := range input.Entries {
			key := arraykey.Normalize(entry.Key)
			if key != entry.Key || seen[key] {
				panic("runtime value array requires canonical unique keys")
			}
			seen[key] = true
			switch entry.Value.(type) {
			case nil, bool, int64, string:
			default:
				panic("runtime value array requires scalar elements")
			}
			if n, ok := key.(int64); ok {
				index.Observe(n)
			}
		}
		if index != input.index {
			panic("runtime value array has non-reconstructible index")
		}
		array := storage.NewScalarArray()
		defer array.Close()
		for _, entry := range input.Entries {
			array.Write(entry.Key, entry.Value)
		}
		return storage.NewTemporary(array)
	default:
		panic("runtime value requires scalar or canonical flat array")
	}
}
