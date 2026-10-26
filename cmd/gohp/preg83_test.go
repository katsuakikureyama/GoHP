package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/ir"
)

const preg83CLIProgram = `<?php
$subject = 'a';
$matches = null;
$found = preg_match('/a/s', $subject, $matches, PREG_OFFSET_CAPTURE, 0);
echo preg_last_error();
`

// TestMain in main_test.go builds the CLI once for this package. This test only
// checks frontend/profile selection and fresh artifact publication; it does not
// build or execute the generated optional-PCRE2 program.
func TestPreg83CLIProfileAndPublication(t *testing.T) {
	t.Run("opt-in php83 publishes profile", func(t *testing.T) {
		dir := t.TempDir()
		source := writePHP(t, dir, "preg83.php", preg83CLIProgram)
		output := filepath.Join(dir, "generated.go")
		report := filepath.Join(dir, "report.json")

		stdout, stderr, err := runCLI(t,
			"-in", source, "-out", output, "-report", report,
			"-php-version", "8.3", "-experimental-pcre2",
		)
		if err != nil {
			t.Fatalf("opt-in CLI failed: %v\nstderr: %s", err, stderr)
		}
		if stdout != "" {
			t.Fatalf("stdout = %q, want empty when publishing artifacts", stdout)
		}
		if generated := string(readFile(t, output)); !strings.Contains(generated, "package main") {
			t.Fatalf("generated source lacks package declaration: %q", generated)
		}
		var provenance struct {
			PHPTarget    string `json:"php_target"`
			RegexProfile string `json:"regex_profile"`
			Status       string `json:"status"`
		}
		if err := json.Unmarshal(readFile(t, report), &provenance); err != nil {
			t.Fatalf("decode report: %v", err)
		}
		if provenance.PHPTarget != "8.3" || provenance.RegexProfile != ir.Preg83Profile || provenance.Status != "generated" {
			t.Fatalf("profile provenance = %+v", provenance)
		}
		assertNoStagedArtifacts(t, dir)
	})

	for _, tt := range []struct {
		name string
		args []string
	}{
		{
			name: "default php83 rejects experimental source",
			args: []string{"-php-version", "8.3"},
		},
		{
			name: "opt-in php84 rejects profile",
			args: []string{"-php-version", "8.4", "-experimental-pcre2"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			source := writePHP(t, dir, "preg83.php", preg83CLIProgram)
			output := filepath.Join(dir, "generated.go")
			report := filepath.Join(dir, "report.json")
			args := append([]string{"-in", source, "-out", output, "-report", report}, tt.args...)
			_, stderr, err := runCLI(t, args...)
			if err == nil {
				t.Fatalf("CLI unexpectedly succeeded; stderr: %s", stderr)
			}
			assertDoesNotExist(t, output)
			assertDoesNotExist(t, report)
			assertNoStagedArtifacts(t, dir)
		})
	}

	t.Run("report still requires output", func(t *testing.T) {
		dir := t.TempDir()
		source := writePHP(t, dir, "preg83.php", preg83CLIProgram)
		report := filepath.Join(dir, "report.json")
		_, stderr, err := runCLI(t,
			"-in", source, "-report", report,
			"-php-version", "8.3", "-experimental-pcre2",
		)
		if err == nil {
			t.Fatalf("CLI unexpectedly succeeded; stderr: %s", stderr)
		}
		assertDoesNotExist(t, report)
		assertNoStagedArtifacts(t, dir)
	})
}
