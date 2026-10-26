// Package pcre2 exposes an opt-in, byte-oriented PCRE2 matching boundary.
//
// It deliberately is not a PHP builtin. In particular it has no PHP capture
// shaping, execution state, output-location, or compiler integration.
package pcre2

import (
	"errors"
	"fmt"
)

const (
	// MaxPatternBytes and MaxSubjectBytes bound conversion and result copying at
	// this first bridge boundary. They are deliberately independent of PCRE2's
	// own resource limits.
	MaxPatternBytes = 4096
	MaxSubjectBytes = 4096
	MaxCaptures     = 64
	MaxMatchLimit   = 100000
	MaxDepthLimit   = 10000
	MaxHeapKiB      = 131072

	// NoMatch and MatchLimit are PCRE2 return codes deliberately exposed as raw
	// successful engine outcomes.
	NoMatch    = -1
	MatchLimit = -47
)

// Profile identifies a fixed engine configuration. It is not a pass-through
// for arbitrary PCRE2 options.
type Profile string

const (
	// ByteDotAll is the initial 8-bit profile: DOTALL, LF newlines,
	// BSR_UNICODE, C character tables, and no UTF/UCP/JIT.
	ByteDotAll Profile = "byte-dotall-v1"
)

// Limits are per-call PCRE2 limits. Every field must be non-zero.
type Limits struct {
	Match   uint32
	Depth   uint32
	HeapKiB uint32
}

// Request is one synchronous, stateless match operation. Offset is a byte
// offset into Subject, not a rune offset.
type Request struct {
	Pattern []byte
	Subject []byte
	Offset  int
	Profile Profile
	Limits  Limits
}

// Name is a PCRE2 named capture. Names retain PCRE2's native name-table order.
type Name struct {
	Number int
	Name   string
}

// Capture is one numbered raw PCRE2 capture. An unmatched group has Present
// false, Start and End -1, and an empty Bytes slice. A zero-length match is
// instead Present with equal non-negative offsets.
type Capture struct {
	Number  int
	Present bool
	Start   int
	End     int
	Bytes   []byte
}

// Result contains raw engine data. CaptureCount is the pattern's total number
// of capturing groups, whereas Captures contains only the groups reported by a
// positive PCRE2 match return code (including group zero).
type Result struct {
	Version      string
	ReturnCode   int
	CaptureCount int
	Names        []Name
	Captures     []Capture
}

// BackendUnavailableError means that this binary was built without the opt-in
// PCRE2 backend. It is intentionally distinct from a no-match result.
type BackendUnavailableError struct{}

func (*BackendUnavailableError) Error() string { return "pcre2 backend unavailable" }

func (*BackendUnavailableError) Is(target error) bool {
	_, ok := target.(*BackendUnavailableError)
	return ok
}

// ErrUnavailable is matched with errors.Is when the optional backend is not
// compiled into the binary.
var ErrUnavailable error = &BackendUnavailableError{}

// ErrInvalidInput identifies a request outside this bridge's declared bounds
// or fixed profile.
var ErrInvalidInput = errors.New("invalid pcre2 bridge request")

// CompileError is returned when PCRE2 rejects a pattern. Offset is a byte
// offset reported by PCRE2.
type CompileError struct {
	Code    int
	Offset  int
	Message string
}

func (e *CompileError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("pcre2 compile error %d at byte %d", e.Code, e.Offset)
	}
	return fmt.Sprintf("pcre2 compile error %d at byte %d: %s", e.Code, e.Offset, e.Message)
}

// EngineError preserves a raw PCRE2 negative result (other than NoMatch and
// MatchLimit), or an internal bridge/profile failure whose code is documented
// by the returned error rather than reclassified as a match outcome.
type EngineError struct {
	Code    int
	Message string
}

func (e *EngineError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("pcre2 engine error %d", e.Code)
	}
	return fmt.Sprintf("pcre2 engine error %d: %s", e.Code, e.Message)
}
