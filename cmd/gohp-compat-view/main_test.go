package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunRequiresArchive(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if got := run(nil, &stdout, &stderr); got != 2 {
		t.Fatalf("exit code = %d, want 2; stderr=%q", got, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "usage: gohp-compat-view") {
		t.Fatalf("stderr = %q, want usage", stderr.String())
	}
}

func TestRunRejectsInvalidFlagsBeforeOutput(t *testing.T) {
	for _, args := range [][]string{
		{"-target", "8.2", "archive"},
		{"-kind", "invalid", "archive"},
		{"-status", "pass", "archive"},
		{"-format", "yaml", "archive"},
	} {
		t.Run(strings.Join(args[:2], " "), func(t *testing.T) {
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

func TestRunArchiveErrorWritesNoPartialView(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing-archive")
	var stdout, stderr bytes.Buffer
	if got := run([]string{missing}, &stdout, &stderr); got != 1 {
		t.Fatalf("exit code = %d, want 1; stderr=%q", got, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "load archive") {
		t.Fatalf("stderr = %q, want archive error", stderr.String())
	}
}
