//go:build gohp_pcre2 && cgo

package runtime

import (
	"strings"
	"testing"
)

func TestPreg83LocalMatch1UsesProfileLimitAndResetsState(t *testing.T) {
	pattern := `/^(a+)+$/s`
	subject := strings.Repeat("a", 4) + "b"

	matchOne := &Execution{}
	limited := matchOne.PregMatch83LocalMatch1(pattern, subject, int64(0))
	preg83RequireReturn(t, limited, false)
	if got := matchOne.PregLastError83(); got != 2 {
		t.Fatalf("match-one error state = %d, want 2", got)
	}

	reset := matchOne.PregMatch83LocalMatch1(`/a/s`, "a", int64(0))
	preg83RequireReturn(t, reset, int64(1))
	if got := matchOne.PregLastError83(); got != 0 {
		t.Fatalf("successful match-one call did not reset error state: %d", got)
	}

	defaultProfile := &Execution{}
	defaultResult := defaultProfile.PregMatch83Local(pattern, subject, int64(0))
	preg83RequireReturn(t, defaultResult, int64(0))
	if got := defaultProfile.PregLastError83(); got != 0 {
		t.Fatalf("default profile error state = %d, want 0", got)
	}
}
