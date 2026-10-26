// Package arraykey is the dependency-leaf key/index policy shared by the
// value runtime and internal storage. It implements the existing int64/string
// subset for the active PHP 8.3 and 8.4 targets, not general PHP key coercion.
package arraykey

import (
	"math"
	"strconv"
)

// Normalize accepts exact int64/string keys. Only canonical in-range decimal
// integer strings become integer keys; signs, padding, whitespace and overflow
// that do not round-trip are preserved as strings. Null/bool/float coercion is
// deliberately outside the supported key domain.
func Normalize(value any) any {
	switch key := value.(type) {
	case int64:
		return key
	case string:
		if n, err := strconv.ParseInt(key, 10, 64); err == nil && strconv.FormatInt(n, 10) == key {
			return n
		}
		return key
	default:
		panic("unsupported array key type")
	}
}

// Index tracks automatic integer keys independently of string entries. It is
// a value: copying an array container must copy this state as well. A zero
// Index is uninitialized, so Next starts at zero but the first explicit negative
// key establishes its own successor. No deletion/reset policy is provided.
type Index struct {
	next        int64
	initialized bool
	blocked     bool
}

// Observe records an inserted integer key; lower/replacement keys never rewind
// the sequence. MaxInt64 can be stored, but later automatic insertion rejects.
func (index *Index) Observe(key int64) {
	if index.blocked || (index.initialized && key < index.next) {
		return
	}
	index.initialized = true
	if key == math.MaxInt64 {
		index.blocked = true
	} else {
		index.next = key + 1
	}
}

// Next is a non-mutating preflight. The caller records the returned key with
// Observe only after all operation inputs are valid, avoiding partial writes.
// Overflow rejection preserves the existing runtime error, not PHP exception
// or diagnostic equivalence.
func (index Index) Next() int64 {
	if index.blocked {
		panic("array key overflow unsupported")
	}
	return index.next
}
