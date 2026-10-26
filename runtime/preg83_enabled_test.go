//go:build gohp_pcre2 && cgo

package runtime

import (
	"errors"
	"math"
	"strings"
	"testing"
)

func TestPreg83BuildsOffsetCaptureMatchesInPHPOrder(t *testing.T) {
	execution := &Execution{}
	result, err := execution.PregMatch83(`/(?<first>a)(?<second>b)?/s`, "ab", 0, preg83TestLimits())
	if err != nil {
		t.Fatal(err)
	}
	preg83RequireReturn(t, result, int64(1))
	preg83RequireArray(t, result.Matches, []Entry{
		{Key: int64(0), Value: preg83Capture("ab", 0)},
		{Key: "first", Value: preg83Capture("a", 0)},
		{Key: int64(1), Value: preg83Capture("a", 0)},
		{Key: "second", Value: preg83Capture("b", 1)},
		{Key: int64(2), Value: preg83Capture("b", 1)},
	})
	if got := execution.PregLastError83(); got != 0 {
		t.Fatalf("PregLastError83() = %d, want 0 after match", got)
	}
}

func TestPreg83ShapesUnmatchedAndTrailingCaptures(t *testing.T) {
	execution := &Execution{}

	middle, err := execution.PregMatch83(`/(a)(b)?(c)/s`, "ac", 0, preg83TestLimits())
	if err != nil {
		t.Fatal(err)
	}
	preg83RequireReturn(t, middle, int64(1))
	preg83RequireArray(t, middle.Matches, []Entry{
		{Key: int64(0), Value: preg83Capture("ac", 0)},
		{Key: int64(1), Value: preg83Capture("a", 0)},
		{Key: int64(2), Value: preg83Capture("", -1)},
		{Key: int64(3), Value: preg83Capture("c", 1)},
	})

	trailing, err := execution.PregMatch83(`/(a)(b)?(c)?/s`, "a", 0, preg83TestLimits())
	if err != nil {
		t.Fatal(err)
	}
	preg83RequireReturn(t, trailing, int64(1))
	preg83RequireArray(t, trailing.Matches, []Entry{
		{Key: int64(0), Value: preg83Capture("a", 0)},
		{Key: int64(1), Value: preg83Capture("a", 0)},
	})
}

func TestPreg83ReturnsFreshNestedArrays(t *testing.T) {
	execution := &Execution{}
	first, err := execution.PregMatch83(`/(?<first>a)/s`, "a", 0, preg83TestLimits())
	if err != nil {
		t.Fatal(err)
	}
	firstNamed := first.Matches.Entries[1].Value.(Array)
	firstNumeric := first.Matches.Entries[2].Value.(Array)
	firstNamed.Entries[0].Value = "changed named"
	if got := firstNumeric.Entries[0].Value; got != "a" {
		t.Fatalf("named and numeric capture arrays alias: numeric value = %#v", got)
	}

	second, err := execution.PregMatch83(`/(?<first>a)/s`, "a", 0, preg83TestLimits())
	if err != nil {
		t.Fatal(err)
	}
	preg83RequireArray(t, second.Matches, []Entry{
		{Key: int64(0), Value: preg83Capture("a", 0)},
		{Key: "first", Value: preg83Capture("a", 0)},
		{Key: int64(1), Value: preg83Capture("a", 0)},
	})
	second.Matches.Entries[0].Value.(Array).Entries[0].Value = "changed later"
	third, err := execution.PregMatch83(`/(?<first>a)/s`, "a", 0, preg83TestLimits())
	if err != nil {
		t.Fatal(err)
	}
	if got := third.Matches.Entries[0].Value.(Array).Entries[0].Value; got != "a" {
		t.Fatalf("later call reused capture storage: %#v", got)
	}
}

func TestPreg83ResetsPerExecutionErrorState(t *testing.T) {
	execution := &Execution{}
	limit := Preg83Limits{Match: 1, Depth: 10000, HeapKiB: 131072}
	failed, err := execution.PregMatch83(`/(a+)+$/s`, strings.Repeat("a", 512)+"b", 0, limit)
	if err != nil {
		t.Fatal(err)
	}
	preg83RequireReturn(t, failed, false)
	if got := execution.PregLastError83(); got != 2 {
		t.Fatalf("PregLastError83() = %d, want 2 after match-limit failure", got)
	}

	success, err := execution.PregMatch83(`/a/s`, "a", 0, preg83TestLimits())
	if err != nil {
		t.Fatal(err)
	}
	preg83RequireReturn(t, success, int64(1))

	nomatch, err := execution.PregMatch83(`/a/s`, "b", 0, preg83TestLimits())
	if err != nil {
		t.Fatal(err)
	}
	preg83RequireReturn(t, nomatch, int64(0))
	if len(nomatch.Matches.Entries) != 0 {
		t.Fatalf("no-match captures = %#v, want empty", nomatch.Matches)
	}
}

func TestPreg83ExecutionIsolationAndNestedGoHelper(t *testing.T) {
	a, b := &Execution{}, &Execution{}
	limit := Preg83Limits{Match: 1, Depth: 10000, HeapKiB: 131072}
	failed, err := a.PregMatch83(`/(a+)+$/s`, strings.Repeat("a", 512)+"b", 0, limit)
	if err != nil {
		t.Fatal(err)
	}
	preg83RequireReturn(t, failed, false)
	if got := b.PregLastError83(); got != 0 {
		t.Fatalf("other execution error = %d, want 0", got)
	}

	result, err := preg83NestedGoHelper(b, `/b/s`, "b")
	if err != nil {
		t.Fatal(err)
	}
	preg83RequireReturn(t, result, int64(1))
	if got := b.PregLastError83(); got != 0 {
		t.Fatalf("nested Go helper state = %d, want 0", got)
	}
	_, err = preg83NestedGoHelper(a, `/a/s`, "z")
	if err != nil {
		t.Fatal(err)
	}
	if got := a.PregLastError83(); got != 0 {
		t.Fatalf("same execution did not reset through Go helper: %d", got)
	}
}

func TestPreg83PreservesStateForUnsupportedCalls(t *testing.T) {
	execution := &Execution{}
	limit := Preg83Limits{Match: 1, Depth: 10000, HeapKiB: 131072}
	_, err := execution.PregMatch83(`/(a+)+$/s`, strings.Repeat("a", 512)+"b", 0, limit)
	if err != nil {
		t.Fatal(err)
	}
	if got := execution.PregLastError83(); got != 2 {
		t.Fatalf("initial state = %d, want 2", got)
	}
	for _, pattern := range []string{`/(/s`, `/(?<same>a)(?<same>b)/s`} {
		_, err := execution.PregMatch83(pattern, "ab", 0, preg83TestLimits())
		if err == nil || errors.Is(err, ErrPreg83Input) {
			t.Fatalf("unsupported pattern %q error = %v, want backend compile error", pattern, err)
		}
		if got := execution.PregLastError83(); got != 2 {
			t.Fatalf("unsupported pattern %q changed state to %d, want 2", pattern, got)
		}
	}
	// These failures are not PHP backtrack-limit outcomes and must not reset
	// the last compatible state or turn into a successful no-match.
	for _, request := range []struct {
		pattern string
		limits  Preg83Limits
	}{
		{`/(?J)(?<same>a)(?<same>b)/s`, preg83TestLimits()},
		{`/^(a+)+$/s`, Preg83Limits{Match: 100000, Depth: 1, HeapKiB: 131072}},
		{`/a/i`, preg83TestLimits()},
	} {
		_, err := execution.PregMatch83(request.pattern, strings.Repeat("a", 100)+"!", 0, request.limits)
		if err == nil || execution.PregLastError83() != 2 {
			t.Fatalf("unsupported request changed state: err=%v state=%d", err, execution.PregLastError83())
		}
	}
}

func TestPreg83RejectsInvalidDomainsAndRequiresExecution(t *testing.T) {
	execution := &Execution{}
	for _, tt := range []struct {
		pattern string
		subject string
		offset  int64
		limits  Preg83Limits
	}{
		{"a", "a", 0, preg83TestLimits()},
		{`/a/`, "a", 0, preg83TestLimits()},
		{`/a/b/s`, "a/b", 0, preg83TestLimits()},
		{`/a/s`, "a", math.MinInt64, preg83TestLimits()},
		{"/" + strings.Repeat("a", 4097) + "/s", "", 0, preg83TestLimits()},
		{`/a/s`, strings.Repeat("a", 4097), 0, preg83TestLimits()},
		{`/a/s`, "a", 0, Preg83Limits{Match: 0, Depth: 1, HeapKiB: 1}},
	} {
		_, err := execution.PregMatch83(tt.pattern, tt.subject, tt.offset, tt.limits)
		if !errors.Is(err, ErrPreg83Input) {
			t.Fatalf("PregMatch83(%q) error = %v, want ErrPreg83Input", tt.pattern, err)
		}
	}
	var nilExecution *Execution
	preg83RequirePanic(t, func() {
		_, _ = nilExecution.PregMatch83(`/a/s`, "a", 0, preg83TestLimits())
	})
	preg83RequirePanic(t, func() { _ = nilExecution.PregLastError83() })
}

func preg83NestedGoHelper(execution *Execution, pattern, subject string) (Preg83Result, error) {
	return execution.PregMatch83(pattern, subject, 0, preg83TestLimits())
}

func preg83Capture(text string, offset int64) Array {
	return NewArray(Entry{Key: int64(0), Value: text}, Entry{Key: int64(1), Value: offset})
}

func preg83RequireReturn(t *testing.T, result Preg83Result, want any) {
	t.Helper()
	if result.Return != want {
		t.Fatalf("return = %#v (%T), want %#v (%T)", result.Return, result.Return, want, want)
	}
	switch want {
	case int64(1):
		if result.EngineCode <= 0 {
			t.Fatalf("match engine code = %d, want positive raw PCRE2 result", result.EngineCode)
		}
	case int64(0):
		if result.EngineCode != -1 {
			t.Fatalf("no-match engine code = %d, want -1", result.EngineCode)
		}
	case false:
		if result.EngineCode != -47 {
			t.Fatalf("match-limit engine code = %d, want -47", result.EngineCode)
		}
	}
}

func preg83RequireArray(t *testing.T, got Array, want []Entry) {
	t.Helper()
	if len(got.Entries) != len(want) {
		t.Fatalf("entry count = %d, want %d: %#v", len(got.Entries), len(want), got)
	}
	for i, wantEntry := range want {
		gotEntry := got.Entries[i]
		if gotEntry.Key != wantEntry.Key || !preg83SameValue(gotEntry.Value, wantEntry.Value) {
			t.Fatalf("entry %d = %#v, want %#v", i, gotEntry, wantEntry)
		}
	}
}

func preg83SameValue(got, want any) bool {
	gotArray, gotIsArray := got.(Array)
	wantArray, wantIsArray := want.(Array)
	if gotIsArray || wantIsArray {
		if !gotIsArray || !wantIsArray || len(gotArray.Entries) != len(wantArray.Entries) {
			return false
		}
		for i := range gotArray.Entries {
			if gotArray.Entries[i].Key != wantArray.Entries[i].Key || !preg83SameValue(gotArray.Entries[i].Value, wantArray.Entries[i].Value) {
				return false
			}
		}
		return true
	}
	return got == want
}
