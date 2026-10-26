//go:build gohp_pcre2 && cgo

package pcre2

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"
)

const (
	testMatchLimit = MaxMatchLimit
	testDepthLimit = MaxDepthLimit
	testHeapKiB    = MaxHeapKiB
)

func testRequest(pattern, subject []byte) Request {
	return Request{
		Pattern: pattern,
		Subject: subject,
		Profile: ByteDotAll,
		Limits: Limits{
			Match:   testMatchLimit,
			Depth:   testDepthLimit,
			HeapKiB: testHeapKiB,
		},
	}
}

func TestMatchReportsNumericCapturesAndNativeNameTableOrder(t *testing.T) {
	t.Parallel()

	got, err := Match(testRequest([]byte("(?<z>a)(b)(?<a>)"), []byte("ab")))
	if err != nil {
		t.Fatalf("Match() error = %v", err)
	}
	if got.ReturnCode != 4 || got.CaptureCount != 3 {
		t.Fatalf("Match() return/capture count = %d/%d, want 4/3", got.ReturnCode, got.CaptureCount)
	}
	wantCaptures := []Capture{
		{Number: 0, Present: true, Start: 0, End: 2, Bytes: []byte("ab")},
		{Number: 1, Present: true, Start: 0, End: 1, Bytes: []byte("a")},
		{Number: 2, Present: true, Start: 1, End: 2, Bytes: []byte("b")},
		{Number: 3, Present: true, Start: 2, End: 2, Bytes: []byte{}},
	}
	if !reflect.DeepEqual(got.Captures, wantCaptures) {
		t.Fatalf("Match() captures = %#v, want %#v", got.Captures, wantCaptures)
	}
	// PCRE2's name table is in native lexical order, distinct from capture order.
	wantNames := []Name{{Number: 3, Name: "a"}, {Number: 1, Name: "z"}}
	if !reflect.DeepEqual(got.Names, wantNames) {
		t.Fatalf("Match() names = %#v, want %#v", got.Names, wantNames)
	}
}

func TestMatchDistinguishesUnmatchedEmptyAndTrailingCaptures(t *testing.T) {
	t.Parallel()

	got, err := Match(testRequest([]byte("(a)?()(b)?"), []byte{}))
	if err != nil {
		t.Fatalf("Match() error = %v", err)
	}
	if got.ReturnCode != 3 || got.CaptureCount != 3 {
		t.Fatalf("Match() return/capture count = %d/%d, want 3/3", got.ReturnCode, got.CaptureCount)
	}
	want := []Capture{
		{Number: 0, Present: true, Start: 0, End: 0, Bytes: []byte{}},
		{Number: 1, Present: false, Start: -1, End: -1, Bytes: []byte{}},
		{Number: 2, Present: true, Start: 0, End: 0, Bytes: []byte{}},
	}
	if !reflect.DeepEqual(got.Captures, want) {
		t.Fatalf("Match() captures = %#v, want %#v", got.Captures, want)
	}
	for _, capture := range got.Captures {
		if capture.Number == 3 {
			t.Fatalf("Match() returned trailing unmatched capture: %#v", got.Captures)
		}
	}
}

func TestMatchUsesRawBytesAndByteOffsets(t *testing.T) {
	t.Parallel()

	subject := []byte{0xc3, 0xa9, 0x00, 0xff, 'x'}
	request := testRequest([]byte{0x00, '(', 0xff, ')'}, subject)
	request.Offset = 2
	got, err := Match(request)
	if err != nil {
		t.Fatalf("Match() error = %v", err)
	}
	want := []Capture{
		{Number: 0, Present: true, Start: 2, End: 4, Bytes: []byte{0x00, 0xff}},
		{Number: 1, Present: true, Start: 3, End: 4, Bytes: []byte{0xff}},
	}
	if !reflect.DeepEqual(got.Captures, want) {
		t.Fatalf("Match() captures = %#v, want %#v", got.Captures, want)
	}

	insideUTF8 := testRequest([]byte("."), []byte{0xc3, 0xa9, 'a'})
	insideUTF8.Offset = 1
	got, err = Match(insideUTF8)
	if err != nil {
		t.Fatalf("Match(offset inside UTF-8) error = %v", err)
	}
	if got.ReturnCode != 1 || len(got.Captures) != 1 || got.Captures[0].Start != 1 || got.Captures[0].End != 2 || !bytes.Equal(got.Captures[0].Bytes, []byte{0xa9}) {
		t.Fatalf("Match(offset inside UTF-8) = %#v, want byte at [1:2]", got)
	}

	empty, err := Match(testRequest([]byte("^$"), []byte{}))
	if err != nil {
		t.Fatalf("Match(empty) error = %v", err)
	}
	if len(empty.Captures) != 1 || !empty.Captures[0].Present || empty.Captures[0].Start != 0 || empty.Captures[0].End != 0 {
		t.Fatalf("Match(empty) captures = %#v, want an empty whole match", empty.Captures)
	}
	emptyPattern, err := Match(testRequest([]byte{}, []byte{}))
	if err != nil || emptyPattern.ReturnCode != 1 || len(emptyPattern.Captures) != 1 || !emptyPattern.Captures[0].Present {
		t.Fatalf("Match(empty pattern) = %#v, %v; want one empty whole match", emptyPattern, err)
	}
}

func TestMatchUsesFixedDotAllProfile(t *testing.T) {
	t.Parallel()

	got, err := Match(testRequest([]byte("a.b"), []byte("a\nb")))
	if err != nil {
		t.Fatalf("Match() error = %v", err)
	}
	if got.ReturnCode != 1 || len(got.Captures) != 1 || !bytes.Equal(got.Captures[0].Bytes, []byte("a\nb")) {
		t.Fatalf("Match() = %#v, want DOTALL whole match", got)
	}

	for _, pattern := range [][]byte{[]byte("(*UTF)a"), []byte("(*UCP)a"), []byte("(*CR)a"), []byte("(*BSR_ANYCRLF)a"), []byte("(*LIMIT_MATCH=1)a"), []byte("(*LIMIT_DEPTH=1)a"), []byte("(*LIMIT_HEAP=1)a")} {
		if _, err := Match(testRequest(pattern, []byte("a"))); err == nil {
			t.Fatalf("Match(%q) error = nil, want fixed-profile rejection", pattern)
		}
	}
}

func TestMatchReturnsNoMatchAndUnknownEngineFailuresDistinctly(t *testing.T) {
	t.Parallel()

	notFound, err := Match(testRequest([]byte("z"), []byte("a")))
	if err != nil {
		t.Fatalf("Match(no match) error = %v, want nil", err)
	}
	if notFound.ReturnCode != NoMatch || len(notFound.Captures) != 0 {
		t.Fatalf("Match(no match) = %#v, want raw NoMatch without captures", notFound)
	}

	depthLimited := testRequest([]byte("^(a+)+$"), []byte(strings.Repeat("a", 100)+"!"))
	depthLimited.Limits.Depth = 1
	result, err := Match(depthLimited)
	var engineErr *EngineError
	if !errors.As(err, &engineErr) {
		t.Fatalf("Match(depth limit) error = %v, want *EngineError", err)
	}
	if result.ReturnCode >= 0 || result.ReturnCode == NoMatch || result.ReturnCode == MatchLimit || engineErr.Code != result.ReturnCode || engineErr.Message == "" {
		t.Fatalf("Match(depth limit) result/error = %#v / %#v, want preserved unknown negative engine failure", result, engineErr)
	}
}

func TestMatchCopiesReturnedBytes(t *testing.T) {
	t.Parallel()

	pattern := []byte("(a)")
	subject := []byte("a")
	first, err := Match(testRequest(pattern, subject))
	if err != nil {
		t.Fatalf("first Match() error = %v", err)
	}
	pattern[1] = 'b'
	subject[0] = 'z'
	if !bytes.Equal(first.Captures[0].Bytes, []byte("a")) || !bytes.Equal(first.Captures[1].Bytes, []byte("a")) {
		t.Fatalf("first result changed after input mutation: %#v", first.Captures)
	}
	if _, err := Match(testRequest([]byte("(b)"), []byte("b"))); err != nil {
		t.Fatalf("second Match() error = %v", err)
	}
	if !bytes.Equal(first.Captures[0].Bytes, []byte("a")) || !bytes.Equal(first.Captures[1].Bytes, []byte("a")) {
		t.Fatalf("first result changed after a later call: %#v", first.Captures)
	}
}

func TestMatchReturnsTypedErrorsForMalformedAndInvalidInput(t *testing.T) {
	t.Parallel()

	_, err := Match(testRequest([]byte("("), []byte{}))
	var compileErr *CompileError
	if !errors.As(err, &compileErr) {
		t.Fatalf("Match(malformed pattern) error = %v, want *CompileError", err)
	}
	if compileErr.Code == 0 || compileErr.Offset < 0 || compileErr.Message == "" {
		t.Fatalf("CompileError = %#v, want code, non-negative offset, and message", compileErr)
	}

	invalid := []struct {
		name    string
		request Request
	}{
		{"negative offset", func() Request { r := testRequest([]byte("a"), []byte("a")); r.Offset = -1; return r }()},
		{"offset after subject", func() Request { r := testRequest([]byte("a"), []byte("a")); r.Offset = 2; return r }()},
		{"unknown profile", func() Request { r := testRequest([]byte("a"), []byte("a")); r.Profile = Profile("other"); return r }()},
		{"zero match limit", func() Request { r := testRequest([]byte("a"), []byte("a")); r.Limits.Match = 0; return r }()},
		{"match limit above profile cap", func() Request {
			r := testRequest([]byte("a"), []byte("a"))
			r.Limits.Match = testMatchLimit + 1
			return r
		}()},
		{"zero depth limit", func() Request { r := testRequest([]byte("a"), []byte("a")); r.Limits.Depth = 0; return r }()},
		{"depth limit above profile cap", func() Request {
			r := testRequest([]byte("a"), []byte("a"))
			r.Limits.Depth = testDepthLimit + 1
			return r
		}()},
		{"zero heap limit", func() Request { r := testRequest([]byte("a"), []byte("a")); r.Limits.HeapKiB = 0; return r }()},
		{"heap limit above profile cap", func() Request {
			r := testRequest([]byte("a"), []byte("a"))
			r.Limits.HeapKiB = testHeapKiB + 1
			return r
		}()},
		{"pattern larger than bound", testRequest(bytes.Repeat([]byte("a"), MaxPatternBytes+1), []byte{})},
		{"subject larger than bound", testRequest([]byte("a"), bytes.Repeat([]byte("a"), MaxSubjectBytes+1))},
		{"captures larger than bound", testRequest([]byte(strings.Repeat("()", MaxCaptures+1)), []byte{})},
	}
	for _, test := range invalid {
		t.Run(test.name, func(t *testing.T) {
			_, err := Match(test.request)
			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("Match() error = %v, want errors.Is(_, ErrInvalidInput)", err)
			}
		})
	}
}

func TestMatchHonorsKnownRawStatusAndMaximumCaptureBound(t *testing.T) {
	t.Parallel()

	limited := testRequest([]byte("^(a+)+$"), []byte(strings.Repeat("a", 100)+"!"))
	limited.Limits.Match = 1
	got, err := Match(limited)
	if err != nil {
		t.Fatalf("Match(match-limit) error = %v, want raw known status", err)
	}
	if got.ReturnCode != MatchLimit {
		t.Fatalf("Match(match-limit) return code = %d, want %d", got.ReturnCode, MatchLimit)
	}

	maximum, err := Match(testRequest([]byte(strings.Repeat("()", MaxCaptures)), []byte{}))
	if err != nil {
		t.Fatalf("Match(max captures) error = %v", err)
	}
	if maximum.CaptureCount != MaxCaptures || maximum.ReturnCode != MaxCaptures+1 || len(maximum.Captures) != MaxCaptures+1 {
		t.Fatalf("Match(max captures) = return %d, count %d, len %d; want %d, %d, %d", maximum.ReturnCode, maximum.CaptureCount, len(maximum.Captures), MaxCaptures+1, MaxCaptures, MaxCaptures+1)
	}
}
