package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/corpus"
)

func TestCorpusCLIOutputs(t *testing.T) {
	manifest, root := writeFixture(t)
	args := []string{"-manifest", manifest, root}
	var markdown, markdownErr bytes.Buffer
	if got := run(args, &markdown, &markdownErr); got != 0 {
		t.Fatalf("markdown exit code = %d, want 0; stderr=%q", got, markdownErr.String())
	}
	if !strings.Contains(markdown.String(), "Source identity:") ||
		!strings.Contains(markdown.String(), "Compatibility: **unverified**") ||
		!strings.Contains(markdown.String(), "Executed cases | n/a (not measured)") {
		t.Fatalf("markdown missing corpus boundaries:\n%s", markdown.String())
	}

	var jsonOne, jsonErr bytes.Buffer
	if got := run([]string{"-format", "json", "-manifest", manifest, root}, &jsonOne, &jsonErr); got != 0 {
		t.Fatalf("json exit code = %d, want 0; stderr=%q", got, jsonErr.String())
	}
	var report corpus.Report
	if err := json.Unmarshal(jsonOne.Bytes(), &report); err != nil {
		t.Fatalf("decode JSON: %v\n%s", err, jsonOne.String())
	}
	if report.Schema != corpus.ReportSchema || len(report.Files) != 2 {
		t.Fatalf("report = %#v, want schema %q and two files", report, corpus.ReportSchema)
	}
	var jsonTwo, secondErr bytes.Buffer
	if got := run([]string{"-format", "json", "-manifest", manifest, root}, &jsonTwo, &secondErr); got != 0 {
		t.Fatalf("second json exit code = %d, want 0; stderr=%q", got, secondErr.String())
	}
	if !bytes.Equal(jsonOne.Bytes(), jsonTwo.Bytes()) {
		t.Fatalf("JSON output is not deterministic\nfirst=%s\nsecond=%s", jsonOne.String(), jsonTwo.String())
	}
}

func TestCorpusCLIFailures(t *testing.T) {
	manifest, root := writeFixture(t)
	t.Run("usage", func(t *testing.T) {
		for _, args := range [][]string{nil, {"-manifest", manifest}, {"-format", "yaml", "-manifest", manifest, root}} {
			var stdout, stderr bytes.Buffer
			if got := run(args, &stdout, &stderr); got != 2 {
				t.Fatalf("run(%q) exit code = %d, want 2; stderr=%q", args, got, stderr.String())
			}
			if stdout.Len() != 0 {
				t.Fatalf("run(%q) stdout = %q, want empty", args, stdout.String())
			}
		}
	})
	t.Run("help", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		if got := run([]string{"-h"}, &stdout, &stderr); got != 0 {
			t.Fatalf("help exit code = %d, want 0; stderr=%q", got, stderr.String())
		}
		if stdout.Len() != 0 || !strings.Contains(stderr.String(), "usage: gohp-corpus") {
			t.Fatalf("help stdout=%q stderr=%q", stdout.String(), stderr.String())
		}
	})
	t.Run("inspection mismatch has no output", func(t *testing.T) {
		if err := os.WriteFile(filepath.Join(root, "subject.php"), []byte("<?php echo 'changed';\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		var stdout, stderr bytes.Buffer
		if got := run([]string{"-manifest", manifest, root}, &stdout, &stderr); got != 1 {
			t.Fatalf("mismatch exit code = %d, want 1; stderr=%q", got, stderr.String())
		}
		if stdout.Len() != 0 {
			t.Fatalf("mismatch stdout = %q, want empty", stdout.String())
		}
	})
	t.Run("write errors", func(t *testing.T) {
		manifest, root := writeFixture(t)
		for _, stdout := range []io.Writer{errorWriter{err: errors.New("broken pipe")}, shortWriter{}} {
			var stderr bytes.Buffer
			if got := run([]string{"-manifest", manifest, root}, stdout, &stderr); got != 1 {
				t.Fatalf("write failure exit code = %d, want 1; stderr=%q", got, stderr.String())
			}
			if !strings.Contains(stderr.String(), "write source corpus") {
				t.Fatalf("stderr = %q, want write error", stderr.String())
			}
		}
	})
}

func TestCorpusMarkdownEscapes(t *testing.T) {
	output := string(corpus.Markdown(corpus.Report{
		Status:         "matched",
		ManifestSHA256: "abc",
		Manifest: corpus.Manifest{
			ID:      "corpus",
			Project: "<project>|[x]`\nnext",
			Version: "1.0.0",
			Source:  corpus.Origin{Repository: "https://example.test/<repo>", Revision: "0123456789abcdef0123456789abcdef01234567"},
			Selection: corpus.Selection{
				Unit: "whole-file", Description: "a|b [c] `d`\n e", DependencyClosure: "unresolved",
			},
			PHPTargets: []string{"8.3", "8.4"},
		},
		Files: []corpus.Observation{{Path: "a|b.php", Role: "subject", Bytes: 1, SHA256: "hash", Status: "matched"}},
	}))
	for _, want := range []string{"&lt;project&gt;\\|\\[x\\]\\` next", "a\\|b \\[c\\] \\`d\\`  e", "Compatibility: **unverified**", "no GoHP compilation"} {
		if !strings.Contains(output, want) {
			t.Fatalf("Markdown missing escaped %q:\n%s", want, output)
		}
	}
	if strings.Contains(output, "| Function units | 0 |") || strings.Contains(output, "| Executed cases | 0 |") {
		t.Fatalf("unmeasured values must not render as zero:\n%s", output)
	}
}

type errorWriter struct{ err error }

func (w errorWriter) Write([]byte) (int, error) { return 0, w.err }

type shortWriter struct{}

func (shortWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	return len(p) - 1, nil
}

func writeFixture(t *testing.T) (manifestPath, sourceRoot string) {
	t.Helper()
	sourceRoot = t.TempDir()
	subject := []byte("<?php echo 'subject';\n")
	provenance := []byte("source notice\n")
	if err := os.WriteFile(filepath.Join(sourceRoot, "subject.php"), subject, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceRoot, "NOTICE"), provenance, 0o600); err != nil {
		t.Fatal(err)
	}
	manifest := corpus.Manifest{
		Schema:  corpus.ManifestSchema,
		ID:      "fixture-corpus",
		Project: "fixture project",
		Version: "1.0.0",
		Source: corpus.Origin{
			Repository: "https://github.com/example/fixture",
			Revision:   "0123456789abcdef0123456789abcdef01234567",
		},
		Selection: corpus.Selection{
			Unit:              "whole-file",
			Description:       "fixture source selection",
			DependencyClosure: "unresolved",
		},
		PHPTargets: []string{"8.3", "8.4"},
		Files: []corpus.File{
			{Path: "subject.php", Role: "subject", Bytes: int64(len(subject)), SHA256: hash(subject)},
			{Path: "NOTICE", Role: "provenance", Bytes: int64(len(provenance)), SHA256: hash(provenance)},
		},
	}
	encoded, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	manifestPath = filepath.Join(t.TempDir(), "corpus.json")
	if err := os.WriteFile(manifestPath, encoded, 0o600); err != nil {
		t.Fatal(err)
	}
	return manifestPath, sourceRoot
}

func hash(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
