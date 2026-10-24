package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEvidenceNeverOverwritesExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "run.json")
	if err := writeExclusive(path, []byte("old evidence")); err != nil {
		t.Fatal(err)
	}
	if err := writeExclusive(path, []byte("new evidence")); err == nil {
		t.Fatal("existing evidence overwritten")
	}
	b, err := os.ReadFile(path)
	if err != nil || string(b) != "old evidence" {
		t.Fatalf("old evidence changed: %q, %v", b, err)
	}
	alias := filepath.Join(filepath.Dir(path), "alias.json")
	if err := os.Symlink(path, alias); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if err := writeExclusive(alias, []byte("replacement")); err == nil {
		t.Fatal("symlink destination accepted")
	}
	b, err = os.ReadFile(path)
	if err != nil || string(b) != "old evidence" {
		t.Fatalf("aliased evidence changed: %q, %v", b, err)
	}
}
