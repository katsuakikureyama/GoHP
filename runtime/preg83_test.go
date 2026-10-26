//go:build !gohp_pcre2 || !cgo

package runtime

import (
	"errors"
	"math"
	"strings"
	"testing"
)

func TestPreg83ReportsUnavailableWithoutOptionalBackend(t *testing.T) {
	execution := &Execution{}
	result, err := execution.PregMatch83(`/a/s`, "a", 0, preg83TestLimits())
	if !errors.Is(err, ErrPreg83Unavailable) {
		t.Fatalf("PregMatch83() error = %v, want errors.Is(_, ErrPreg83Unavailable)", err)
	}
	if result.Return != nil || len(result.Matches.Entries) != 0 || result.EngineCode != 0 {
		t.Fatalf("unavailable result = %#v, want zero result", result)
	}
	if got := execution.PregLastError83(); got != 0 {
		t.Fatalf("PregLastError83() = %d, want 0", got)
	}
}

func TestPreg83RejectsDomainsBeforeOptionalBackend(t *testing.T) {
	tooLongPattern := "/" + strings.Repeat("a", 4097) + "/s"
	tooLongSubject := strings.Repeat("a", 4097)
	for _, tt := range []struct {
		name    string
		pattern string
		subject string
		offset  int64
		limits  Preg83Limits
	}{
		{"missing delimiters", "a", "a", 0, preg83TestLimits()},
		{"missing required modifier", `/a/`, "a", 0, preg83TestLimits()},
		{"unsupported modifier", `/a/i`, "a", 0, preg83TestLimits()},
		{"unescaped inner delimiter", `/a/b/s`, "a/b", 0, preg83TestLimits()},
		{"too long pattern", tooLongPattern, "", 0, preg83TestLimits()},
		{"too long subject", `/a/s`, tooLongSubject, 0, preg83TestLimits()},
		{"negative offset below subject", `/a/s`, "a", -2, preg83TestLimits()},
		{"positive offset after subject", `/a/s`, "a", 2, preg83TestLimits()},
		{"minimum offset", `/a/s`, "a", math.MinInt64, preg83TestLimits()},
		{"zero match limit", `/a/s`, "a", 0, Preg83Limits{Depth: 1, HeapKiB: 1}},
		{"zero depth limit", `/a/s`, "a", 0, Preg83Limits{Match: 1, HeapKiB: 1}},
		{"zero heap limit", `/a/s`, "a", 0, Preg83Limits{Match: 1, Depth: 1}},
		{"match limit above bridge ceiling", `/a/s`, "a", 0, Preg83Limits{Match: 100001, Depth: 1, HeapKiB: 1}},
		{"depth limit above bridge ceiling", `/a/s`, "a", 0, Preg83Limits{Match: 1, Depth: 10001, HeapKiB: 1}},
		{"heap limit above bridge ceiling", `/a/s`, "a", 0, Preg83Limits{Match: 1, Depth: 1, HeapKiB: 131073}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			execution := &Execution{}
			_, err := execution.PregMatch83(tt.pattern, tt.subject, tt.offset, tt.limits)
			if !errors.Is(err, ErrPreg83Input) {
				t.Fatalf("PregMatch83() error = %v, want errors.Is(_, ErrPreg83Input)", err)
			}
			if got := execution.PregLastError83(); got != 0 {
				t.Fatalf("invalid input changed PregLastError83() to %d, want 0", got)
			}
		})
	}
}

func TestPreg83RequiresExecutionWithoutOptionalBackend(t *testing.T) {
	var execution *Execution
	preg83RequirePanic(t, func() {
		_, _ = execution.PregMatch83(`/a/s`, "a", 0, preg83TestLimits())
	})
	preg83RequirePanic(t, func() { _ = execution.PregLastError83() })
}
