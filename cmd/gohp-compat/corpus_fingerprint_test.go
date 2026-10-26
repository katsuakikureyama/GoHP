package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSourceFingerprintIncludesCorpusManifest(t *testing.T) {
	root := t.TempDir()
	manifest := filepath.Join(root, "corpora", "synthetic", "manifest.json")
	if err := os.MkdirAll(filepath.Dir(manifest), 0o700); err != nil {
		t.Fatal(err)
	}
	write := func(name, content string) {
		t.Helper()
		if err := os.WriteFile(name, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	fingerprint := func() string {
		t.Helper()
		got, err := sourceFingerprint(root)
		if err != nil {
			t.Fatal(err)
		}
		return got
	}
	write(manifest, `{"id":"first-selection"}`)
	before := fingerprint()
	write(manifest, `{"id":"other-selection"}`)
	after := fingerprint()
	if before == after {
		t.Fatal("changing the pinned selection must change the evidence source fingerprint")
	}
	write(filepath.Join(root, "corpora", "synthetic", "report.json"), `{"status":"unverified"}`)
	if fingerprint() != after {
		t.Fatal("an inspection report is output, not a pinned source input")
	}
	if err := os.Remove(manifest); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("report.json", manifest); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if _, err := sourceFingerprint(root); err == nil {
		t.Fatal("symlinked manifest must not be accepted as a source fingerprint input")
	}
}
