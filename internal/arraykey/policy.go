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
// key establishes its own successor. Pop handles only a removed integer tail;
// it is not a general deletion/reset policy.
type Index struct {
	next        int64
	initialized bool
	blocked     bool // MaxInt64 is occupied, even if Pop rewinds next below it.
}

// Observe records an inserted integer key; lower/replacement keys never rewind
// the sequence. MaxInt64 can be stored; insertion at that key rejects while
// it remains occupied. Pop may make a lower predecessor available first.
func (index *Index) Observe(key int64) {
	if index.initialized && key < index.next {
		return
	}
	index.initialized = true
	if key == math.MaxInt64 {
		index.next = key
		index.blocked = true
	} else {
		index.next = key + 1
	}
}

// Pop observes the integer key of an entry just removed by array_pop. Only
// the immediate predecessor rewinds next; string tails never call this method.
// MaxInt64 uses a saturated next value plus occupancy, not an overflowing +1.
func (index *Index) Pop(key int64) {
	if key == math.MaxInt64 {
		index.blocked = false
		return
	}
	if index.initialized && key+1 == index.next {
		if key == math.MinInt64 {
			// PHP's minimum next-index sentinel resumes automatic keys at 0.
			*index = Index{}
		} else {
			index.next = key
		}
	}
}

// Next is a non-mutating preflight. The caller records the returned key with
// Observe only after all operation inputs are valid, avoiding partial writes.
// Overflow rejection preserves the existing runtime error, not PHP exception
// or diagnostic equivalence.
func (index Index) Next() int64 {
	if index.blocked && index.next == math.MaxInt64 {
		panic("array key overflow unsupported")
	}
	return index.next
}
