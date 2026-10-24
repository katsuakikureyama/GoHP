package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunRequiresExactlyOneProjectDirectory(t *testing.T) {
	for _, args := range [][]string{nil, {"one", "two"}} {
		t.Run(strings.Join(args, ","), func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if got := run(args, &stdout, &stderr); got != 2 {
				t.Fatalf("exit code = %d, want 2; stderr=%q", got, stderr.String())
			}
			if stdout.Len() != 0 {
				t.Fatalf("stdout = %q, want empty", stdout.String())
			}
		})
	}
}

func TestRunRejectsInvalidFormatBeforeOutput(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if got := run([]string{"-format", "yaml", "project"}, &stdout, &stderr); got != 2 {
		t.Fatalf("exit code = %d, want 2; stderr=%q", got, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestRunReadErrorWritesNoPartialReport(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing-project")
	var stdout, stderr bytes.Buffer
	if got := run([]string{missing}, &stdout, &stderr); got != 1 {
		t.Fatalf("exit code = %d, want 1; stderr=%q", got, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "read Composer metadata") {
		t.Fatalf("stderr = %q, want read error", stderr.String())
	}
}
