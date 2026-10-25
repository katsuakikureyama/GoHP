package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Exercise the actual pre-run metadata gate with a version-probe stub. No
// legacy PHP binary is installed or executed, and no fixture/test run starts.
func TestMetadataOracleAdmission(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("version-probe stub requires a POSIX shell")
	}
	for _, tt := range []struct {
		name, target, version string
		accepted              bool
	}{
		{"baseline", "8.3", "8.3.36", true},
		{"default", "", "8.3.36", true},
		{"next_patch", "8.3", "8.3.37", true},
		{"old_patch", "8.3", "8.3.35", false},
		{"retired", "7.4", "7.4.33", false},
		{"unverified_series", "8.4", "8.4.20", false},
		{"no_auto_enable", "", "8.4.20", false},
		{"mismatch", "8.3", "8.4.20", false},
		{"missing", "8.3", "", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			stub := "#!/bin/sh\nprintf '%s' '" + tt.version + "'\n"
			if err := os.WriteFile(filepath.Join(dir, "php"), []byte(stub), 0700); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
			t.Setenv("GOHP_COMMIT", strings.Repeat("a", 40))
			t.Setenv("GOHP_TREE", strings.Repeat("b", 40))
			t.Setenv("GOHP_DIRTY", "clean")
			t.Setenv("GOHP_PHP_TARGET", tt.target)
			meta, err := metadata()
			if (err == nil) != tt.accepted {
				t.Fatalf("admission of target=%q oracle=%q: %v", tt.target, tt.version, err)
			}
			if tt.accepted && (meta.PHPTarget != "8.3" || meta.PHPVersion != tt.version) {
				t.Fatalf("actual oracle identity lost: %+v", meta)
			}
		})
	}
}

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
