//go:build !gohp_pcre2 || !cgo

package pcre2

import (
	"errors"
	"testing"
)

func TestMatchReportsUnavailableWithoutEnabledBackend(t *testing.T) {
	t.Parallel()

	_, err := Match(Request{
		Pattern: []byte("a"),
		Subject: []byte("a"),
		Profile: ByteDotAll,
		Limits: Limits{
			Match:   100000,
			Depth:   10000,
			HeapKiB: 131072,
		},
	})
	if err == nil {
		t.Fatal("Match() error = nil, want unavailable backend")
	}
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("Match() error = %v, want errors.Is(_, ErrUnavailable)", err)
	}
	var unavailable *BackendUnavailableError
	if !errors.As(err, &unavailable) {
		t.Fatalf("Match() error = %T, want *BackendUnavailableError", err)
	}
}
