package runtime

import (
	"errors"
	"fmt"
	"strings"

	"github.com/katsuakikureyama/GoHP/internal/pcre2"
)

// Preg83Limits bounds the optional experimental byte-oriented engine call.
// All fields must be positive and within the B1 bridge's documented ceilings.
type Preg83Limits struct{ Match, Depth, HeapKiB uint32 }

// Preg83Result is a detached value, not a PHP output-argument location.
// Use it only when err == nil. Return is exactly int64(1), int64(0), or false.
// EngineCode is the raw PCRE2 result, not PHP's return or last-error code.
type Preg83Result struct {
	Return     any
	Matches    Array
	EngineCode int
}

var (
	ErrPreg83Input       = errors.New("input outside experimental preg83 profile")
	ErrPreg83Unavailable = pcre2.ErrUnavailable
)

// PregLastError83 reads the experimental regex state for this execution only.
// No compiler builtin is admitted by this method. Like Strtok, it requires a
// non-nil Execution, shared through synchronous calls and never copied.
func (execution *Execution) PregLastError83() int64 {
	RequireExecution(execution)
	return execution.preg83Error
}

// PregMatch83 is the B2 experimental value adapter: slash-delimited /s pattern,
// raw byte strings, in-range signed byte offset, and PREG_OFFSET_CAPTURE only.
// This is NOT generated PHP preg_match support or a general PHP extension API.
// Unsupported domains/backends/engine errors preserve the previous error state;
// they are explicit Go errors, never fabricated PHP false or no-match results.
func (execution *Execution) PregMatch83(pattern, subject string, offset int64, limits Preg83Limits) (Preg83Result, error) {
	RequireExecution(execution)
	invalid := func(detail string) (Preg83Result, error) {
		return Preg83Result{}, fmt.Errorf("%w: %s", ErrPreg83Input, detail)
	}
	if len(pattern) < 3 || len(pattern) > pcre2.MaxPatternBytes+3 || pattern[0] != '/' || !strings.HasSuffix(pattern, "/s") {
		return invalid("requires bounded slash-delimited /s pattern")
	}
	body := pattern[1 : len(pattern)-2]
	// A slash escaped by an odd number of backslashes belongs to the body.
	// Never reinterpret an unescaped delimiter as literal regex syntax.
	for i := 0; i < len(body); i++ {
		if body[i] == '\\' {
			i++
			continue
		}
		if body[i] == '/' {
			return invalid("unescaped inner delimiter")
		}
	}
	length := int64(len(subject))
	if length > pcre2.MaxSubjectBytes || offset < -length || offset > length {
		return invalid("subject size or byte offset outside bounds")
	}
	if limits.Match == 0 || limits.Match > pcre2.MaxMatchLimit || limits.Depth == 0 || limits.Depth > pcre2.MaxDepthLimit || limits.HeapKiB == 0 || limits.HeapKiB > pcre2.MaxHeapKiB {
		return invalid("engine limits outside bounds")
	}
	if offset < 0 {
		offset += length
	}
	raw, err := pcre2.Match(pcre2.Request{Pattern: []byte(body), Subject: []byte(subject), Offset: int(offset), Profile: pcre2.ByteDotAll,
		Limits: pcre2.Limits{Match: limits.Match, Depth: limits.Depth, HeapKiB: limits.HeapKiB}})
	if err != nil {
		return Preg83Result{EngineCode: raw.ReturnCode}, err
	}
	// Duplicate-name and branch-reset alias shaping is not measured by this
	// first wrapper. Reject non-unique name/number mappings before state update.
	names := make(map[int]string, len(raw.Names))
	seen := make(map[string]bool, len(raw.Names))
	for _, name := range raw.Names {
		if name.Number < 1 || name.Number > raw.CaptureCount || name.Name == "" || seen[name.Name] || names[name.Number] != "" {
			return invalid("non-unique capture names outside profile")
		}
		seen[name.Name] = true
		names[name.Number] = name.Name
	}
	result := Preg83Result{Matches: NewArray(), EngineCode: raw.ReturnCode}
	var nextError int64
	switch {
	case raw.ReturnCode == pcre2.NoMatch:
		result.Return = int64(0)
	case raw.ReturnCode == pcre2.MatchLimit:
		result.Return = false
		nextError = 2
	case raw.ReturnCode > 0:
		result.Return = int64(1)
		entries := make([]Entry, 0, len(raw.Captures)*2)
		for _, capture := range raw.Captures {
			pair := NewArray(Entry{Key: int64(0), Value: string(capture.Bytes)}, Entry{Key: int64(1), Value: int64(capture.Start)})
			if name, ok := names[capture.Number]; ok {
				entries = append(entries, Entry{Key: name, Value: pair})
			}
			entries = append(entries, Entry{Key: int64(capture.Number), Value: pair})
		}
		// NewArray deep-clones pairs: numeric and named entries do not share
		// mutable nested entry slices, even for the same capture.
		result.Matches = NewArray(entries...)
	default:
		return Preg83Result{EngineCode: raw.ReturnCode}, fmt.Errorf("unexpected preg83 engine code %d", raw.ReturnCode)
	}
	// Publish state only after the complete result is constructed. Unknown
	// failures above are outside the compatible outcome set and do not reset it.
	execution.preg83Error = nextError
	return result, nil
}
