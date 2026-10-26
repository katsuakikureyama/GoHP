package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/compat"
)

// Exercise the actual pre-run metadata gate with a startup-baseline stub. No
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
		{"php84_floor", "8.4", "8.4.27", true},
		{"php84_next_patch", "8.4", "8.4.28", true},
		{"php84_old_patch", "8.4", "8.4.26", false},
		{"unverified_series", "8.5", "8.5.0", false},
		{"no_auto_enable", "", "8.4.20", false},
		{"mismatch", "8.3", "8.4.20", false},
		{"cross_series_mismatch", "8.4", "8.3.36", false},
		{"missing", "8.3", "", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writePHPStub(t, dir, baselineJSON(t, tt.version))
			t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
			t.Setenv("GOHP_COMMIT", strings.Repeat("a", 40))
			t.Setenv("GOHP_TREE", strings.Repeat("b", 40))
			t.Setenv("GOHP_DIRTY", "clean")
			t.Setenv("GOHP_PHP_TARGET", tt.target)
			meta, err := metadata()
			if (err == nil) != tt.accepted {
				t.Fatalf("admission of target=%q oracle=%q: %v", tt.target, tt.version, err)
			}
			expectedTarget := tt.target
			if expectedTarget == "" {
				expectedTarget = "8.3"
			}
			if tt.accepted && (meta.PHPTarget != expectedTarget || meta.PHPVersion != tt.version) {
				t.Fatalf("actual oracle identity lost: %+v", meta)
			}
			if tt.accepted && (meta.PHPBaseline == nil || meta.PHPBaseline.PHPVersion != meta.PHPVersion || meta.PHPBaselineSHA256 != compat.PHPBaselineHash(meta.PHPBaseline)) {
				t.Fatal("admitted oracle lost its startup baseline binding")
			}
		})
	}
}

func baselineJSON(t *testing.T, version string) string {
	t.Helper()
	empty := ""
	ini := make(map[string]*string, len(compat.PHPINIKeys()))
	for _, key := range compat.PHPINIKeys() {
		ini[key] = &empty
	}
	locales := make(map[string]*string, len(compat.PHPLocaleKeys()))
	for _, key := range compat.PHPLocaleKeys() {
		locales[key] = nil
	}
	b, err := json.Marshal(compat.PHPBaseline{
		Scope:          "cli-startup-baseline",
		PHPVersion:     version,
		SAPI:           "cli",
		IntSize:        8,
		Timezone:       "UTC",
		INI:            ini,
		Locales:        locales,
		Extensions:     []compat.PHPExtension{},
		ZendExtensions: []compat.PHPExtension{},
	})
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func writePHPStub(t *testing.T, dir, output string) {
	t.Helper()
	stub := "#!/bin/sh\nprintf '%s' '" + output + "'\n"
	if err := os.WriteFile(filepath.Join(dir, "php"), []byte(stub), 0700); err != nil {
		t.Fatal(err)
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
