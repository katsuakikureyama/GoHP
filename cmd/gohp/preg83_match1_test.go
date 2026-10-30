package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/ir"
)

// TestMain in main_test.go builds the CLI once. This test confines itself to
// compiler selection and artifact publication; it never builds or runs the
// optional native-PCRE2 generated output.
func TestPreg83Match1CLIProfileAndPublication(t *testing.T) {
	for _, tt := range []struct {
		name          string
		profile       string
		wantProfile   string
		wantCall      string
		forbiddenCall string
	}{
		{
			name:          "explicit match-one profile",
			profile:       ir.Preg83Match1Profile,
			wantProfile:   ir.Preg83Match1Profile,
			wantCall:      ".PregMatch83LocalMatch1(",
			forbiddenCall: ".PregMatch83Local(",
		},
		{
			name:          "explicit default profile remains default",
			profile:       ir.Preg83Profile,
			wantProfile:   ir.Preg83Profile,
			wantCall:      ".PregMatch83Local(",
			forbiddenCall: ".PregMatch83LocalMatch1(",
		},
		{
			name:          "implicit default profile remains default",
			wantProfile:   ir.Preg83Profile,
			wantCall:      ".PregMatch83Local(",
			forbiddenCall: ".PregMatch83LocalMatch1(",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			source := writePHP(t, dir, "preg83.php", preg83CLIProgram)
			output := filepath.Join(dir, "generated.go")
			report := filepath.Join(dir, "report.json")
			args := []string{
				"-in", source, "-out", output, "-report", report,
				"-php-version", "8.3", "-experimental-pcre2",
			}
			if tt.profile != "" {
				args = append(args, "-experimental-pcre2-profile", tt.profile)
			}
			stdout, stderr, err := runCLI(t, args...)
			if err != nil {
				t.Fatalf("CLI failed: %v\nstderr: %s", err, stderr)
			}
			if stdout != "" {
				t.Fatalf("stdout = %q, want empty when publishing artifacts", stdout)
			}
			generated := string(readFile(t, output))
			if !strings.Contains(generated, tt.wantCall) {
				t.Fatalf("generated source lacks %q:\n%s", tt.wantCall, generated)
			}
			if strings.Contains(generated, tt.forbiddenCall) {
				t.Fatalf("generated source unexpectedly contains %q:\n%s", tt.forbiddenCall, generated)
			}
			var provenance struct {
				PHPTarget    string `json:"php_target"`
				RegexProfile string `json:"regex_profile"`
				Status       string `json:"status"`
			}
			if err := json.Unmarshal(readFile(t, report), &provenance); err != nil {
				t.Fatalf("decode report: %v", err)
			}
			if provenance.PHPTarget != "8.3" || provenance.RegexProfile != tt.wantProfile || provenance.Status != "generated" {
				t.Fatalf("profile provenance = %+v, want target=8.3 profile=%q status=generated", provenance, tt.wantProfile)
			}
			assertNoStagedArtifacts(t, dir)
		})
	}

	for _, tt := range []struct {
		name string
		args []string
	}{
		{
			name: "profile requires opt-in",
			args: []string{
				"-php-version", "8.3", "-experimental-pcre2-profile", ir.Preg83Match1Profile,
			},
		},
		{
			name: "unknown profile is rejected",
			args: []string{
				"-php-version", "8.3", "-experimental-pcre2", "-experimental-pcre2-profile", "unknown-profile",
			},
		},
		{
			name: "match-one profile requires php83",
			args: []string{
				"-php-version", "8.4", "-experimental-pcre2", "-experimental-pcre2-profile", ir.Preg83Match1Profile,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			source := writePHP(t, dir, "preg83.php", preg83CLIProgram)
			output := filepath.Join(dir, "generated.go")
			report := filepath.Join(dir, "report.json")
			args := append([]string{"-in", source, "-out", output, "-report", report}, tt.args...)
			stdout, stderr, err := runCLI(t, args...)
			if err == nil {
				t.Fatalf("CLI unexpectedly succeeded; stdout: %s; stderr: %s", stdout, stderr)
			}
			if stdout != "" {
				t.Fatalf("failed publication wrote stdout: %q", stdout)
			}
			assertDoesNotExist(t, output)
			assertDoesNotExist(t, report)
			assertNoStagedArtifacts(t, dir)
		})
	}
}
