package main

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var gohpBinary string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "gohp-cli-")
	if err != nil {
		panic(err)
	}
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("cannot locate cmd/gohp")
	}
	gohpBinary = filepath.Join(dir, "gohp")
	if runtime.GOOS == "windows" {
		gohpBinary += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", gohpBinary, ".")
	cmd.Dir = filepath.Dir(file)
	cmd.Env = os.Environ()
	if output, err := cmd.CombinedOutput(); err != nil {
		panic("build gohp: " + err.Error() + "\n" + string(output))
	}
	code := m.Run()
	if err := os.RemoveAll(dir); err != nil {
		panic(err)
	}
	os.Exit(code)
}

func TestCLIWritesFreshOutputAndReport(t *testing.T) {
	for _, tt := range []struct {
		name      string
		outputDir string
		reportDir string
	}{
		{name: "same directory", outputDir: "artifacts", reportDir: "artifacts"},
		{name: "different directories", outputDir: "generated", reportDir: "provenance"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			source := writePHP(t, dir, "source.php", validPHP)
			outputDir := mkdir(t, dir, tt.outputDir)
			reportDir := outputDir
			if tt.reportDir != tt.outputDir {
				reportDir = mkdir(t, dir, tt.reportDir)
			}
			output := filepath.Join(outputDir, "generated.go")
			report := filepath.Join(reportDir, "report.json")

			stdout, stderr, err := runCLI(t, "-in", source, "-out", output, "-report", report)
			if err != nil {
				t.Fatalf("gohp failed: %v\nstderr: %s", err, stderr)
			}
			if stdout != "" {
				t.Fatalf("stdout = %q, want empty when -out is used", stdout)
			}
			generated := readFile(t, output)
			if !strings.Contains(string(generated), "package main") {
				t.Fatalf("generated source lacks package declaration: %q", generated)
			}
			var provenance map[string]any
			if err := json.Unmarshal(readFile(t, report), &provenance); err != nil {
				t.Fatalf("decode report: %v", err)
			}
			if got, want := provenance["source_sha256"], sha256Hex([]byte(validPHP)); got != want {
				t.Errorf("source_sha256 = %v, want %s", got, want)
			}
			if got, want := provenance["generated_sha256"], sha256Hex(generated); got != want {
				t.Errorf("generated_sha256 = %v, want %s", got, want)
			}
			if got := provenance["status"]; got != "generated" {
				t.Errorf("status = %v, want generated", got)
			}
			if got := readFile(t, source); string(got) != validPHP {
				t.Errorf("source changed to %q", got)
			}
			assertNoStagedArtifacts(t, outputDir, reportDir)
		})
	}
}

func TestCLIRejectsExistingOrAliasedArtifactsWithoutClobbering(t *testing.T) {
	dir := t.TempDir()
	source := writePHP(t, dir, "source.php", validPHP)
	sourceBytes := readFile(t, source)

	tests := []struct {
		name    string
		prepare func(t *testing.T) (out, report string, preserved map[string][]byte)
		verify  func(t *testing.T, out, report string)
	}{
		{
			name: "source as output",
			prepare: func(t *testing.T) (string, string, map[string][]byte) {
				return source, "", map[string][]byte{source: sourceBytes}
			},
		},
		{
			name: "source as report",
			prepare: func(t *testing.T) (string, string, map[string][]byte) {
				out := filepath.Join(dir, "source-report.go")
				return out, source, map[string][]byte{source: sourceBytes}
			},
		},
		{
			name: "existing output",
			prepare: func(t *testing.T) (string, string, map[string][]byte) {
				out := writeFile(t, dir, "existing.go", "keep output")
				return out, "", map[string][]byte{out: []byte("keep output")}
			},
		},
		{
			name: "existing report prevents output",
			prepare: func(t *testing.T) (string, string, map[string][]byte) {
				out := filepath.Join(dir, "blocked.go")
				report := writeFile(t, dir, "existing.json", "keep report")
				return out, report, map[string][]byte{report: []byte("keep report")}
			},
		},
		{
			name: "symlink to source",
			prepare: func(t *testing.T) (string, string, map[string][]byte) {
				alias := filepath.Join(dir, "source-symlink.go")
				if err := os.Symlink(source, alias); err != nil {
					t.Skipf("symlinks unavailable: %v", err)
				}
				return alias, "", map[string][]byte{source: sourceBytes, alias: sourceBytes}
			},
			verify: func(t *testing.T, out, _ string) {
				t.Helper()
				info, err := os.Lstat(out)
				if err != nil {
					t.Fatalf("inspect source symlink: %v", err)
				}
				if info.Mode()&os.ModeSymlink == 0 {
					t.Errorf("source alias was replaced with mode %v", info.Mode())
				}
				if target, err := os.Readlink(out); err != nil || target != source {
					t.Errorf("source alias target = %q, %v; want %q", target, err, source)
				}
			},
		},
		{
			name: "hard link to source",
			prepare: func(t *testing.T) (string, string, map[string][]byte) {
				out := filepath.Join(dir, "hardlink-output.go")
				alias := filepath.Join(dir, "source-hardlink.json")
				if err := os.Link(source, alias); err != nil {
					t.Skipf("hard links unavailable: %v", err)
				}
				return out, alias, map[string][]byte{source: sourceBytes, alias: sourceBytes}
			},
			verify: func(t *testing.T, _, report string) {
				t.Helper()
				sourceInfo, err := os.Stat(source)
				if err != nil {
					t.Fatalf("stat source: %v", err)
				}
				aliasInfo, err := os.Stat(report)
				if err != nil {
					t.Fatalf("stat hard link: %v", err)
				}
				if !os.SameFile(sourceInfo, aliasInfo) {
					t.Errorf("hard-link source alias was replaced")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, report, preserved := tt.prepare(t)
			args := []string{"-in", source, "-out", out}
			if report != "" {
				args = append(args, "-report", report)
			}
			_, stderr, err := runCLI(t, args...)
			if err == nil {
				t.Fatalf("gohp unexpectedly succeeded; stderr: %s", stderr)
			}
			for path, want := range preserved {
				if got := readFile(t, path); string(got) != string(want) {
					t.Errorf("%s was clobbered: got %q, want %q", path, got, want)
				}
			}
			if _, tracked := preserved[out]; !tracked {
				assertDoesNotExist(t, out)
			}
			if tt.verify != nil {
				tt.verify(t, out, report)
			}
			assertNoStagedArtifacts(t, dir)
		})
	}
}

func TestCLIFailuresLeaveNoArtifacts(t *testing.T) {
	t.Run("missing report directory", func(t *testing.T) {
		dir := t.TempDir()
		source := writePHP(t, dir, "source.php", validPHP)
		out := filepath.Join(dir, "generated.go")
		report := filepath.Join(dir, "missing", "report.json")
		_, stderr, err := runCLI(t, "-in", source, "-out", out, "-report", report)
		if err == nil {
			t.Fatalf("gohp unexpectedly succeeded; stderr: %s", stderr)
		}
		assertDoesNotExist(t, out)
		assertDoesNotExist(t, report)
		assertNoStagedArtifacts(t, dir)
	})

	t.Run("unsupported syntax", func(t *testing.T) {
		dir := t.TempDir()
		source := writePHP(t, dir, "unsupported.php", "<?php\n$callback = function ($x) { return $x; };\n")
		out := filepath.Join(dir, "generated.go")
		report := filepath.Join(dir, "report.json")
		_, stderr, err := runCLI(t, "-in", source, "-out", out, "-report", report)
		if err == nil {
			t.Fatalf("gohp unexpectedly succeeded; stderr: %s", stderr)
		}
		assertDoesNotExist(t, out)
		assertDoesNotExist(t, report)
		assertNoStagedArtifacts(t, dir)
	})

	t.Run("report requires output", func(t *testing.T) {
		dir := t.TempDir()
		source := writePHP(t, dir, "source.php", validPHP)
		report := filepath.Join(dir, "report.json")
		_, stderr, err := runCLI(t, "-in", source, "-report", report)
		if err == nil {
			t.Fatalf("gohp unexpectedly succeeded; stderr: %s", stderr)
		}
		assertDoesNotExist(t, report)
		assertNoStagedArtifacts(t, dir)
	})

	t.Run("output and report share a fresh path", func(t *testing.T) {
		dir := t.TempDir()
		source := writePHP(t, dir, "source.php", validPHP)
		candidate := filepath.Join(dir, "candidate")
		_, stderr, err := runCLI(t, "-in", source, "-out", candidate, "-report", candidate)
		if err == nil {
			t.Fatalf("gohp unexpectedly succeeded; stderr: %s", stderr)
		}
		assertDoesNotExist(t, candidate)
		assertNoStagedArtifacts(t, dir)
	})
}

func TestCLIWritesToStdoutWithoutReport(t *testing.T) {
	dir := t.TempDir()
	source := writePHP(t, dir, "source.php", validPHP)
	stdout, stderr, err := runCLI(t, "-in", source)
	if err != nil {
		t.Fatalf("gohp failed: %v\nstderr: %s", err, stderr)
	}
	if !strings.Contains(stdout, "package main") {
		t.Fatalf("stdout lacks generated Go source: %q", stdout)
	}
	assertNoStagedArtifacts(t, dir)
}

func runCLI(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	cmd := exec.Command(gohpBinary, args...)
	var out, errOut strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err = cmd.Run()
	return out.String(), errOut.String(), err
}

func writePHP(t *testing.T, dir, name, content string) string {
	return writeFile(t, dir, name, content)
}

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return b
}

func mkdir(t *testing.T, parent, name string) string {
	t.Helper()
	path := filepath.Join(parent, name)
	if err := os.Mkdir(path, 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func assertDoesNotExist(t *testing.T, path string) {
	t.Helper()
	_, err := os.Lstat(path)
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("%s exists or could not be inspected: %v", path, err)
	}
}

func assertNoStagedArtifacts(t *testing.T, dirs ...string) {
	t.Helper()
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			t.Fatalf("read output directory %s: %v", dir, err)
		}
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), ".gohp-") {
				t.Errorf("staged artifact remains: %s", filepath.Join(dir, entry.Name()))
			}
		}
	}
}

func sha256Hex(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }

const validPHP = "<?php\necho 'ok';\n"
