package storage

import "github.com/katsuakikureyama/GoHP/internal/arraykey"

// ScalarArray is an owned handle to ordered, scalar-only array storage. Handles
// must be used by pointer, never copied as Go values, and are not concurrent.
// Copy explicitly retains shared storage; Close releases one handle. Go GC is
// not semantic release. The zero value is an open, empty handle.
//
// Keys use the runtime's bounded int64/string normalization and automatic index
// policy. Elements admit the same values as Slot. This internal
// foundation is not a runtime.Array replacement or a generated-code API yet.
type ScalarArray struct {
	data   *scalarArrayData
	closed bool
}

// A shared container owns each element binding once, independently of how many
// array handles own the container. Only separation duplicates element bindings.
type scalarArrayData struct {
	owners  int
	entries []scalarArrayEntry
	index   arraykey.Index
}

type scalarArrayEntry struct {
	key  any
	slot *Slot
}

func NewScalarArray() *ScalarArray { return &ScalarArray{} }

func (array *ScalarArray) requireOpen() {
	if array == nil || array.closed {
		panic("storage operation requires an open scalar array")
	}
}

func (array *ScalarArray) find(key any) *Slot {
	if array.data != nil {
		for _, entry := range array.data.entries {
			if entry.key == key {
				return entry.slot
			}
		}
	}
	return nil
}

// Read returns a value, not the element binding. Missing keys do not allocate,
// separate storage or become present null. Scalar values are immutable copies.
func (array *ScalarArray) Read(key any) (any, bool) {
	array.requireOpen()
	key = arraykey.Normalize(key)
	if slot := array.find(key); slot != nil {
		return slot.Read()
	}
	return nil, false
}

// Keys returns a detached insertion-order snapshot; replacement keeps position.
func (array *ScalarArray) Keys() []any {
	array.requireOpen()
	if array.data == nil {
		return nil
	}
	keys := make([]any, len(array.data.entries))
	for i, entry := range array.data.entries {
		keys[i] = entry.key
	}
	return keys
}

// Copy retains the container, not each element reference. Separating later,
// rather than eagerly copying now, preserves the timing of reference release.
func (array *ScalarArray) Copy() *ScalarArray {
	array.requireOpen()
	if array.data == nil {
		array.data = &scalarArrayData{owners: 1}
	}
	array.data.owners++
	return &ScalarArray{data: array.data}
}

// separate prepares exclusive container ownership. A singleton element
// reference becomes an ordinary value in the new container; a reference with
// other live bindings is retained. Scalar-only elements cannot form cycles.
func (array *ScalarArray) separate() {
	if array.data == nil {
		array.data = &scalarArrayData{owners: 1}
		return
	}
	if array.data.owners == 1 {
		return
	}
	old := array.data
	next := &scalarArrayData{owners: 1, entries: make([]scalarArrayEntry, 0, len(old.entries)), index: old.index}
	for _, entry := range old.entries {
		copy := &Slot{}
		if entry.slot.ref != nil && entry.slot.ref.owners > 1 {
			copy.Bind(entry.slot)
		} else {
			value, _ := entry.slot.Read() // Every stored element is defined.
			copy.Write(value)
		}
		next.entries = append(next.entries, scalarArrayEntry{key: entry.key, slot: copy})
	}
	old.owners--
	array.data = next
}

// local is private: callers must validate all inputs and separate first. Never
// hand this slot to a consumer, which could otherwise unset an owned binding.
func (array *ScalarArray) local(key any) *Slot {
	if slot := array.find(key); slot != nil {
		return slot
	}
	slot := &Slot{}
	array.data.entries = append(array.data.entries, scalarArrayEntry{key: key, slot: slot})
	if n, ok := key.(int64); ok {
		array.data.index.Observe(n)
	}
	return slot
}

// Write separates before assigning through the selected element's binding.
// Invalid input must not separate storage: doing so would change later alias
// behavior even when no visible element value was changed by the failed write.
func (array *ScalarArray) Write(key, value any) {
	array.requireOpen()
	key = arraykey.Normalize(key)
	scalar(value)
	array.separate()
	array.local(key).Write(value)
}

// Append inserts an ordinary scalar at the next automatic key and returns that
// key. Invalid values and exhausted indices must reject before separation,
// preserving both the index and reference sharing on failure.
func (array *ScalarArray) Append(value any) int64 {
	array.requireOpen()
	scalar(value)
	var index arraykey.Index
	if array.data != nil {
		index = array.data.index
	}
	key := index.Next()
	array.separate()
	array.local(key).Write(value)
	return key
}

// Bind replaces only this element binding with an alias of source. The source
// remains independently owned and an undefined source becomes present null.
func (array *ScalarArray) Bind(key any, source *Slot) {
	array.requireOpen()
	key = arraykey.Normalize(key)
	if source == nil {
		panic("storage reference requires a slot")
	}
	array.separate()
	array.local(key).Bind(source)
}

// Reference first separates the container, then returns a fresh owned alias.
// Missing keys become present null only in this container. The caller must
// Unset the returned Slot; rebinding it never replaces the element binding.
func (array *ScalarArray) Reference(key any) *Slot {
	array.requireOpen()
	key = arraykey.Normalize(key)
	array.separate()
	alias := &Slot{}
	alias.Bind(array.local(key))
	return alias
}

// Close releases this handle once. The last container owner releases element
// bindings; independently retained references can outlive every array handle.
func (array *ScalarArray) Close() {
	if array == nil {
		panic("storage close requires a scalar array")
	}
	if array.closed {
		return
	}
	array.closed = true
	if array.data != nil {
		array.data.owners--
		if array.data.owners == 0 {
			for _, entry := range array.data.entries {
				entry.slot.Unset()
			}
			array.data.entries = nil
		}
		array.data = nil
	}
}
