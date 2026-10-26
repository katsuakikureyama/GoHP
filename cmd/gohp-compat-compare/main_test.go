package main

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/compat"
)

func TestRunRequiresExactlyTwoArchives(t *testing.T) {
	for _, args := range [][]string{nil, {"one"}, {"one", "two", "three"}} {
		var stdout, stderr bytes.Buffer
		if got := run(args, &stdout, &stderr); got != 2 {
			t.Fatalf("run(%q) exit code = %d, want 2; stderr=%q", args, got, stderr.String())
		}
		if stdout.Len() != 0 {
			t.Fatalf("run(%q) stdout = %q, want empty", args, stdout.String())
		}
		if !strings.Contains(stderr.String(), "usage: gohp-compat-compare") {
			t.Fatalf("run(%q) stderr = %q, want usage", args, stderr.String())
		}
	}
}

func TestRunRejectsInvalidFormatBeforeArchiveOutput(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if got := run([]string{"-format", "yaml", "left", "right"}, &stdout, &stderr); got != 2 {
		t.Fatalf("exit code = %d, want 2; stderr=%q", got, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestRunArchiveErrorWritesNoPartialComparison(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing-archive")
	var stdout, stderr bytes.Buffer
	if got := run([]string{missing, missing}, &stdout, &stderr); got != 1 {
		t.Fatalf("exit code = %d, want 1; stderr=%q", got, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "load left archive") {
		t.Fatalf("stderr = %q, want archive error", stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	load := func(path string) (compat.Report, error) {
		if path == "right" {
			return compat.Report{}, errors.New("invalid right archive")
		}
		return compat.Report{}, nil
	}
	if got := runWith([]string{"left", "right"}, &stdout, &stderr, load); got != 1 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "load right archive") {
		t.Fatal("right archive failure exposed a partial comparison")
	}
}

func TestRunReportsOutputWriteFailure(t *testing.T) {
	var stderr bytes.Buffer
	writer := failWriter{err: errors.New("write denied")}
	load := func(string) (compat.Report, error) { return compat.Report{}, nil }
	if got := runWith([]string{"left", "right"}, writer, &stderr, load); got != 1 {
		t.Fatalf("exit code = %d, want 1; stderr=%q", got, stderr.String())
	}
	if !strings.Contains(stderr.String(), "write output comparison") || !strings.Contains(stderr.String(), "write denied") {
		t.Fatalf("stderr = %q, want write error", stderr.String())
	}
}

type failWriter struct{ err error }

func (w failWriter) Write([]byte) (int, error) { return 0, w.err }
