package corpus

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerify(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"src/example.php": "<?php file_put_contents(__DIR__ . '/../EXECUTED', 'this source must not run');\n",
		"NOTICE":          "example notice\n",
		"PROVENANCE":      "selected for a test\n",
	}
	for name, contents := range files {
		filename := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filename, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	manifest := validManifest(files)
	manifestPath := writeManifest(t, root, manifest)
	report, err := Verify(manifestPath, root)
	if err != nil {
		t.Fatal(err)
	}
	if report.Schema != verificationSchema || report.Status != "source-verified" || report.Compatibility != "unverified" {
		t.Fatalf("unexpected report: %#v", report)
	}
	if report.SourceFileCount != 1 || report.SourceBytes != int64(len(files["src/example.php"])) {
		t.Fatalf("unexpected source totals: %#v", report)
	}
	if report.Measurements.FunctionUnits != nil || report.Measurements.Callsites != nil || report.Measurements.CompileRatio != nil || report.Measurements.ExecutionRatio != nil {
		t.Fatal("measurements must remain explicitly unmeasured")
	}
	if len(report.Targets) != 2 || report.Targets[0].PHPTarget != "8.3" || report.Targets[1].Execution != "not-measured" {
		t.Fatalf("unexpected targets: %#v", report.Targets)
	}
	bytes, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	want := sha256.Sum256(bytes)
	if report.ManifestSHA256 != hex.EncodeToString(want[:]) {
		t.Fatal("manifest digest did not bind exact bytes")
	}
	if !strings.Contains(report.Notice, "does not verify upstream authenticity") {
		t.Fatalf("missing boundary notice: %q", report.Notice)
	}
	if _, err := os.Stat(filepath.Join(root, "EXECUTED")); !os.IsNotExist(err) {
		t.Fatalf("source execution sentinel exists or could not be checked: %v", err)
	}
}

func TestVerifyRejectsInvalidManifests(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{"src/example.php": "<?php\n", "NOTICE": "notice\n", "PROVENANCE": "provenance\n"}
	for name, contents := range files {
		filename := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filename, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	valid := validManifest(files)
	invalid := validManifest(files)
	invalid["php_targets"] = []string{"8.3", "8.3"}
	badTargets, err := json.Marshal(invalid)
	if err != nil {
		t.Fatal(err)
	}
	missingProvenance := validManifest(files)
	delete(missingProvenance["upstream"].(map[string]any), "revision")
	unknownNested := validManifest(files)
	unknownNested["upstream"].(map[string]any)["unknown"] = true
	badSchema := validManifest(files)
	badSchema["schema"] = "gohp.corpus.manifest/v9"
	emptyFiles := validManifest(files)
	emptyFiles["files"] = []map[string]any{}
	emptyTargets := validManifest(files)
	emptyTargets["php_targets"] = []string{}
	missingNotice := validManifest(files)
	missingNotice["files"].([]map[string]any)[1]["role"] = "provenance"
	missingSource := validManifest(files)
	missingSource["files"].([]map[string]any)[0]["role"] = "notice"
	duplicateFiles := validManifest(files)
	duplicateFiles["files"] = append(duplicateFiles["files"].([]map[string]any), fileEntry("NOTICE", "notice", files["NOTICE"]))
	disabledTarget := validManifest(files)
	disabledTarget["php_targets"] = []string{"8.2"}
	badNumericSize := validManifest(files)
	badNumericSize["files"].([]map[string]any)[0]["bytes"] = 1.5
	controlPath := validManifest(files)
	controlPath["files"].([]map[string]any)[0]["path"] = "src/\x00example.php"
	overFileLimit := validManifest(files)
	overFileLimit["files"].([]map[string]any)[0]["bytes"] = int64(16<<20) + 1
	overAggregateLimit := validManifest(files)
	for i := 0; i < 4; i++ {
		overAggregateLimit["files"] = append(overAggregateLimit["files"].([]map[string]any), map[string]any{"path": "src/large" + string(rune('a'+i)) + ".php", "role": "source", "bytes": int64(16 << 20), "sha256": strings.Repeat("0", 64)})
	}
	overFileCount := validManifest(files)
	for i := 0; i < 62; i++ {
		overFileCount["files"] = append(overFileCount["files"].([]map[string]any), map[string]any{"path": "src/many" + string(rune('a'+i)) + ".php", "role": "source", "bytes": int64(1), "sha256": strings.Repeat("0", 64)})
	}
	cases := []struct{ name, contents, want string }{
		{"duplicate key", `{"schema":"gohp.corpus.manifest/v0.1","schema":"gohp.corpus.manifest/v0.1"}`, "duplicate JSON"},
		{"nested duplicate key", strings.Replace(string(marshalManifest(t, valid)), `"repository":"example/repository"`, `"repository":"example/repository","repository":"other"`, 1), "duplicate JSON"},
		{"trailing value", string(marshalManifest(t, valid)) + ` {}`, "trailing JSON"},
		{"unknown root", strings.Replace(string(marshalManifest(t, valid)), `"id":"sample"`, `"id":"sample","unknown":true`, 1), "unknown"},
		{"unknown nested", string(marshalManifest(t, unknownNested)), "unknown"},
		{"bad schema", string(marshalManifest(t, badSchema)), "schema must be"},
		{"duplicate targets", string(badTargets), "duplicate target"},
		{"disabled target", string(marshalManifest(t, disabledTarget)), "disabled target"},
		{"empty files", string(marshalManifest(t, emptyFiles)), "files must contain"},
		{"empty targets", string(marshalManifest(t, emptyTargets)), "php_targets must be nonempty"},
		{"missing notice", string(marshalManifest(t, missingNotice)), "include a notice"},
		{"missing source", string(marshalManifest(t, missingSource)), "include a source"},
		{"duplicate file paths", string(marshalManifest(t, duplicateFiles)), "duplicate paths"},
		{"path traversal", strings.Replace(string(marshalManifest(t, valid)), `src/example.php`, `../example.php`, 1), "canonical"},
		{"backslash path", strings.Replace(string(marshalManifest(t, valid)), `src/example.php`, `src\\example.php`, 1), "canonical"},
		{"absolute path", strings.Replace(string(marshalManifest(t, valid)), `src/example.php`, `/example.php`, 1), "canonical"},
		{"control path", string(marshalManifest(t, controlPath)), "canonical"},
		{"bad numeric size", string(marshalManifest(t, badNumericSize)), "positive integer"},
		{"over file limit", string(marshalManifest(t, overFileLimit)), "byte limit"},
		{"over aggregate limit", string(marshalManifest(t, overAggregateLimit)), "aggregate limit"},
		{"over file count", string(marshalManifest(t, overFileCount)), "files must contain"},
		{"missing provenance", string(marshalManifest(t, missingProvenance)), "missing or unknown"},
		{"invalid utf8", string([]byte{0xff}), "valid UTF-8"},
		{"too deep", strings.Repeat("[", maxJSONDepth+2) + strings.Repeat("]", maxJSONDepth+2), "nested levels"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "manifest.json")
			if err := os.WriteFile(path, []byte(test.contents), 0o644); err != nil {
				t.Fatal(err)
			}
			report, err := Verify(path, root)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Verify error = %v, want %q", err, test.want)
			}
			assertZeroReport(t, report)
		})
	}
}

func TestVerifyRejectsSourceFiles(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{"src/example.php": "<?php\n", "NOTICE": "notice\n", "PROVENANCE": "provenance\n"}
	for name, contents := range files {
		filename := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filename, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	cases := []struct {
		name   string
		mutate func(map[string]any)
		want   string
	}{
		{"missing", func(m map[string]any) {
			if err := os.Remove(filepath.Join(root, "NOTICE")); err != nil {
				t.Fatal(err)
			}
		}, "no such file"},
		{"size", func(m map[string]any) { entries := m["files"].([]map[string]any); entries[0]["bytes"] = int64(1) }, "size mismatch"},
		{"hash", func(m map[string]any) {
			entries := m["files"].([]map[string]any)
			entries[0]["sha256"] = strings.Repeat("0", 64)
		}, "sha256 mismatch"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if err := os.WriteFile(filepath.Join(root, "NOTICE"), []byte(files["NOTICE"]), 0o644); err != nil {
				t.Fatal(err)
			}
			m := validManifest(files)
			test.mutate(m)
			path := writeManifest(t, root, m)
			report, err := Verify(path, root)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Verify error = %v, want %q", err, test.want)
			}
			assertZeroReport(t, report)
		})
	}
	outside := filepath.Join(t.TempDir(), "outside.php")
	if err := os.WriteFile(outside, []byte("outside"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "src", "escaped.php")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	m := validManifest(files)
	entries := m["files"].([]map[string]any)
	entries[0] = fileEntry("src/escaped.php", "source", "outside")
	if report, err := Verify(writeManifest(t, root, m), root); err == nil || !strings.Contains(err.Error(), "escapes source root") {
		t.Fatalf("Verify symlink error = %v", err)
	} else {
		assertZeroReport(t, report)
	}
	nonregular := validManifest(files)
	nonregular["files"].([]map[string]any)[0] = map[string]any{"path": "src", "role": "source", "bytes": int64(1), "sha256": strings.Repeat("0", 64)}
	if report, err := Verify(writeManifest(t, root, nonregular), root); err == nil || !strings.Contains(err.Error(), "not regular") {
		t.Fatalf("Verify nonregular error = %v", err)
	} else {
		assertZeroReport(t, report)
	}
}

func assertZeroReport(t *testing.T, report Report) {
	t.Helper()
	if report.Schema != "" || report.Status != "" || report.Manifest.ID != "" || report.SourceFileCount != 0 || report.SourceBytes != 0 || report.ManifestSHA256 != "" || report.Compatibility != "" || report.Notice != "" || len(report.Targets) != 0 {
		t.Fatalf("failed verification returned partial report: %#v", report)
	}
}

func validManifest(files map[string]string) map[string]any {
	return map[string]any{
		"schema": "gohp.corpus.manifest/v0.1", "id": "sample",
		"upstream":    map[string]any{"repository": "example/repository", "revision": strings.Repeat("a", 40), "version": "1.2.3"},
		"selection":   map[string]any{"unit": "whole-file", "exclusions": "dependencies not selected", "dependencies": "unresolved"},
		"php_targets": []string{"8.3", "8.4"},
		"files":       []map[string]any{fileEntry("src/example.php", "source", files["src/example.php"]), fileEntry("NOTICE", "notice", files["NOTICE"]), fileEntry("PROVENANCE", "provenance", files["PROVENANCE"])},
	}
}

func fileEntry(name, role, contents string) map[string]any {
	digest := sha256.Sum256([]byte(contents))
	return map[string]any{"path": name, "role": role, "bytes": int64(len(contents)), "sha256": hex.EncodeToString(digest[:])}
}

func writeManifest(t *testing.T, root string, manifest map[string]any) string {
	t.Helper()
	path := filepath.Join(root, "manifest.json")
	if err := os.WriteFile(path, marshalManifest(t, manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func marshalManifest(t *testing.T, manifest map[string]any) []byte {
	t.Helper()
	b, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
