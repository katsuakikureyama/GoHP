package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/corpus"
)

func TestRunUsageAndReadErrors(t *testing.T) {
	for _, tc := range []struct {
		args []string
		code int
	}{
		{nil, 2}, {[]string{"root"}, 2},
		{[]string{"-manifest", "manifest"}, 2},
		{[]string{"-manifest", "manifest", "a", "b"}, 2},
		{[]string{"-unknown"}, 2}, {[]string{"-help"}, 0},
		{[]string{"-manifest", filepath.Join(t.TempDir(), "missing"), t.TempDir()}, 1},
	} {
		var out, diagnostics bytes.Buffer
		if code := run(tc.args, &out, &diagnostics); code != tc.code || out.Len() != 0 || diagnostics.Len() == 0 {
			t.Fatalf("%v: code=%d stdout=%q stderr=%q", tc.args, code, out.String(), diagnostics.String())
		}
	}
}

func cliFixture(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	manifest := corpus.Manifest{
		Schema: "gohp.corpus.manifest/v0.1", ID: "cli-fixture",
		Upstream:   corpus.Upstream{Repository: "https://example.invalid/project", Revision: strings.Repeat("a", 40), Version: "fixture"},
		Selection:  corpus.Selection{Unit: "whole-file", Exclusions: "Other files are excluded", Dependencies: "unresolved"},
		PHPTargets: []string{"8.3", "8.4"},
	}
	for _, item := range []struct{ path, role, text string }{
		{"source.php", "source", "<?php exit(37);"}, {"NOTICE", "notice", "fixture notice"},
	} {
		data := []byte(item.text)
		if err := os.WriteFile(filepath.Join(root, item.path), data, 0600); err != nil {
			t.Fatal(err)
		}
		hash := sha256.Sum256(data)
		manifest.Files = append(manifest.Files, corpus.ManifestFile{Path: item.path, Role: item.role, Bytes: int64(len(data)), SHA256: hex.EncodeToString(hash[:])})
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(t.TempDir(), "manifest.json")
	if err := os.WriteFile(manifestPath, data, 0600); err != nil {
		t.Fatal(err)
	}
	return manifestPath, root
}

func TestRunVerifiesBeforeOutput(t *testing.T) {
	manifest, root := cliFixture(t)
	var out, diagnostics bytes.Buffer
	args := []string{"-manifest", manifest, root}
	if code := run(args, &out, &diagnostics); code != 0 || diagnostics.Len() != 0 {
		t.Fatalf("code=%d stderr=%q", code, diagnostics.String())
	}
	var report corpus.Report
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.Status != "source-verified" || report.Compatibility != "unverified" || report.SourceFileCount != 1 || report.Measurements.CompileRatio != nil || len(report.Targets) != 2 {
		t.Fatalf("unexpected report: %+v", report)
	}
	// A late notice mismatch must not leak an apparently successful source report.
	if err := os.WriteFile(filepath.Join(root, "NOTICE"), []byte("changed notice"), 0600); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if code := run(args, &out, &diagnostics); code != 1 || out.Len() != 0 || diagnostics.Len() == 0 {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, out.String(), diagnostics.String())
	}
}

type brokenWriter struct{ short bool }

func (w brokenWriter) Write(p []byte) (int, error) {
	if w.short {
		return len(p) - 1, nil
	}
	return 0, errors.New("destination unavailable")
}

func TestRunReportsWriteFailure(t *testing.T) {
	manifest, root := cliFixture(t)
	for _, short := range []bool{false, true} {
		var diagnostics bytes.Buffer
		if code := run([]string{"-manifest", manifest, root}, brokenWriter{short}, &diagnostics); code != 1 || !strings.Contains(diagnostics.String(), "write corpus verification") {
			t.Fatalf("code=%d stderr=%q", code, diagnostics.String())
		}
	}
}
