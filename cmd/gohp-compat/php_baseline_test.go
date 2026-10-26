package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/compat"
)

func TestProbePHPBaselineRejectsInvalidOracleOutput(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("PHP probe stubs require a POSIX shell")
	}
	valid := baselineJSON(t, "8.3.36")
	missingProjection := baselineForTest(t, "8.3.36")
	delete(missingProjection.INI, compat.PHPINIKeys()[0])
	missingJSON, err := json.Marshal(missingProjection)
	if err != nil {
		t.Fatal(err)
	}
	unknownJSON := strings.TrimSuffix(valid, "}") + `,"unexpected":true}`
	for _, tt := range []struct {
		name   string
		output string
	}{
		{name: "malformed_json", output: "{"},
		{name: "trailing_json", output: valid + "{}"},
		{name: "unknown_field", output: unknownJSON},
		{name: "missing_projection_key", output: string(missingJSON)},
		{name: "oversized_output", output: valid + strings.Repeat(" ", 128<<10)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			withPHPStub(t, "printf '%s' '"+tt.output+"'\n")
			if baseline, err := probePHPBaseline(); err == nil || baseline != nil {
				t.Fatalf("invalid probe output accepted: baseline=%+v err=%v", baseline, err)
			}
		})
	}
	for _, tt := range []struct {
		name, body string
	}{
		{name: "nonzero_exit", body: "printf '%s' '" + valid + "'\nprintf 'startup failure' >&2\nexit 3\n"},
		{name: "stderr_diagnostic", body: "printf '%s' '" + valid + "'\nprintf 'startup notice' >&2\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			withPHPStub(t, tt.body)
			if baseline, err := probePHPBaseline(); err == nil || baseline != nil {
				t.Fatalf("failing probe accepted: baseline=%+v err=%v", baseline, err)
			}
		})
	}
}

func TestProbePHPBaselinePreservesProjectionAndHash(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("PHP probe stubs require a POSIX shell")
	}
	baseline := baselineForTest(t, "8.3.36")
	empty := ""
	baseline.INI["precision"] = &empty
	baseline.Extensions = []compat.PHPExtension{{Name: "Core", Version: nil}}
	encoded, err := json.Marshal(baseline)
	if err != nil {
		t.Fatal(err)
	}
	withPHPStub(t, "printf '%s' '"+string(encoded)+"'\n")
	got, err := probePHPBaseline()
	if err != nil {
		t.Fatal(err)
	}
	if got.PHPVersion != "8.3.36" {
		t.Fatalf("PHP version = %q", got.PHPVersion)
	}
	if got.INI["precision"] == nil || *got.INI["precision"] != "" {
		t.Fatalf("empty INI value was not preserved: %#v", got.INI["precision"])
	}
	if len(got.Extensions) != 1 || got.Extensions[0].Version != nil {
		t.Fatalf("nullable extension version was not preserved: %#v", got.Extensions)
	}
	if gotHash, wantHash := compat.PHPBaselineHash(got), compat.PHPBaselineHash(&baseline); gotHash != wantHash {
		t.Fatalf("baseline hash = %s, want %s", gotHash, wantHash)
	}
}

func TestProbePHPBaselineNativePHP(t *testing.T) {
	if _, err := exec.LookPath("php"); err != nil {
		t.Skip("PHP executable is unavailable")
	}
	baseline, err := probePHPBaseline()
	if err != nil {
		t.Fatal(err)
	}
	if err := baseline.Validate(baseline.PHPVersion); err != nil {
		t.Fatalf("actual PHP baseline is invalid: %v", err)
	}
	if baseline.PHPVersion == "" || len(baseline.INI) != len(compat.PHPINIKeys()) || len(baseline.Locales) != len(compat.PHPLocaleKeys()) {
		t.Fatalf("native probe did not return its complete projection: %+v", baseline)
	}
}

func baselineForTest(t *testing.T, version string) compat.PHPBaseline {
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
	return compat.PHPBaseline{
		Scope:          "cli-startup-baseline",
		PHPVersion:     version,
		SAPI:           "cli",
		IntSize:        8,
		Timezone:       "UTC",
		INI:            ini,
		Locales:        locales,
		Extensions:     []compat.PHPExtension{},
		ZendExtensions: []compat.PHPExtension{},
	}
}

func withPHPStub(t *testing.T, body string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "php")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}
