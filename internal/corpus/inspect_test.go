package corpus

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type corpusFixtureFile struct {
	path string
	role string
	data string
}

func TestCorpusIdentity(t *testing.T) {
	root := t.TempDir()
	entries := []corpusFixtureFile{
		{path: "src/z.php", role: "subject", data: "<?php echo 'z';\n"},
		{path: "NOTICE", role: "provenance", data: "synthetic provenance\n"},
		{path: "src/a.php", role: "subject", data: "<?php echo 'a';\n"},
	}
	writeFixtureFiles(t, root, entries)

	manifest := validManifest(t, entries)
	manifestPath := writeManifest(t, manifest)
	report, err := Inspect(manifestPath, root)
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}

	if report.Schema != ReportSchema {
		t.Errorf("Schema = %q, want %q", report.Schema, ReportSchema)
	}
	if report.Status != "source-identity-match" {
		t.Errorf("Status = %q, want source-identity-match", report.Status)
	}
	if report.Compatibility != "unverified" {
		t.Errorf("Compatibility = %q, want unverified", report.Compatibility)
	}
	if report.AnalysisStatus != "not-measured" {
		t.Errorf("AnalysisStatus = %q, want not-measured", report.AnalysisStatus)
	}
	wantManifestHash := sha256Hex(manifest)
	if report.ManifestSHA256 != wantManifestHash {
		t.Errorf("ManifestSHA256 = %q, want %q", report.ManifestSHA256, wantManifestHash)
	}

	if got, want := manifestPaths(report.Manifest.Files), []string{"src/z.php", "NOTICE", "src/a.php"}; !reflect.DeepEqual(got, want) {
		t.Errorf("manifest file order = %v, want %v", got, want)
	}
	if got, want := observationPaths(report.Files), []string{"NOTICE", "src/a.php", "src/z.php"}; !reflect.DeepEqual(got, want) {
		t.Errorf("observed file order = %v, want sorted %v", got, want)
	}

	wantSubjectBytes := int64(len(entries[0].data) + len(entries[2].data))
	if report.Measurements.SubjectFiles != 2 || report.Measurements.SubjectBytes != wantSubjectBytes {
		t.Errorf("subject measurements = %+v, want 2 files / %d bytes", report.Measurements, wantSubjectBytes)
	}
	wantProvenanceBytes := int64(len(entries[1].data))
	if report.Measurements.ProvenanceFiles != 1 || report.Measurements.ProvenanceBytes != wantProvenanceBytes {
		t.Errorf("provenance measurements = %+v, want 1 file / %d bytes", report.Measurements, wantProvenanceBytes)
	}
	if report.Measurements.FunctionUnits != nil || report.Measurements.SyntaxNodes != nil || report.Measurements.CallSites != nil || report.Measurements.CompiledUnits != nil || report.Measurements.ExecutedCases != nil {
		t.Errorf("analysis measurements must remain nil: %+v", report.Measurements)
	}

	reportJSON, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("marshal report: %v", err)
	}
	if strings.Contains(string(reportJSON), root) {
		t.Errorf("report must not echo source root %q", root)
	}
}

func TestCorpusManifestRejections(t *testing.T) {
	root := t.TempDir()
	entries := []corpusFixtureFile{{path: "subject.php", role: "subject", data: "<?php\n"}}
	writeFixtureFiles(t, root, entries)
	base := string(validManifest(t, entries))

	tooLarge := base + strings.Repeat(" ", 1<<20)
	tooDeep := strings.Repeat("[", 65) + "0" + strings.Repeat("]", 65)
	cases := []struct {
		name string
		raw  string
	}{
		{name: "unknown field", raw: strings.Replace(base, "\n}", ",\n  \"unexpected\": true\n}", 1)},
		{name: "case-folded field alias", raw: strings.Replace(base, "\"project\":", "\"Project\":", 1)},
		{name: "nested unknown field", raw: strings.Replace(base, "\"revision\": \""+strings.Repeat("a", 40)+"\"", "\"revision\": \""+strings.Repeat("a", 40)+"\", \"unexpected\": true", 1)},
		{name: "duplicate key", raw: strings.Replace(base, "\n  \"id\": \"synthetic-corpus\",", "\n  \"id\": \"synthetic-corpus\",\n  \"id\": \"duplicate\",", 1)},
		{name: "trailing JSON", raw: base + " {}"},
		{name: "null required field", raw: strings.Replace(base, "\"project\": \"Synthetic Project\"", "\"project\": null", 1)},
		{name: "blank project", raw: strings.Replace(base, "Synthetic Project", "", 1)},
		{name: "non-printable version", raw: strings.Replace(base, "0.0.0-test", "0.0.0\\n", 1)},
		{name: "nesting beyond limit", raw: tooDeep},
		{name: "manifest larger than limit", raw: tooLarge},
		{name: "bad schema", raw: strings.Replace(base, ManifestSchema, "gohp.source-corpus/v9", 1)},
		{name: "bad ID", raw: strings.Replace(base, "synthetic-corpus", "Synthetic corpus", 1)},
		{name: "repository credentials", raw: strings.Replace(base, "https://github.com/example/synthetic", "https://user:pass@github.com/example/synthetic", 1)},
		{name: "repository query", raw: strings.Replace(base, "https://github.com/example/synthetic", "https://github.com/example/synthetic?q=x", 1)},
		{name: "repository fragment", raw: strings.Replace(base, "https://github.com/example/synthetic", "https://github.com/example/synthetic#tag", 1)},
		{name: "repository is not HTTPS", raw: strings.Replace(base, "https://github.com/example/synthetic", "http://github.com/example/synthetic", 1)},
		{name: "bad revision", raw: strings.Replace(base, strings.Repeat("a", 40), strings.Repeat("A", 40), 1)},
		{name: "non-whole selection", raw: strings.Replace(base, "\"unit\": \"whole-file\"", "\"unit\": \"snippet\"", 1)},
		{name: "resolved closure", raw: strings.Replace(base, "\"dependency_closure\": \"unresolved\"", "\"dependency_closure\": \"resolved\"", 1)},
		{name: "blank description", raw: strings.Replace(base, "Synthetic whole-file input", "", 1)},
		{name: "duplicate target", raw: strings.Replace(base, "[\n    \"8.3\",\n    \"8.4\"\n  ]", "[\"8.3\", \"8.3\"]", 1)},
		{name: "no targets", raw: strings.Replace(base, "[\n    \"8.3\",\n    \"8.4\"\n  ]", "[]", 1)},
		{name: "no files", raw: replaceFiles(t, base, "[]")},
		{name: "no subject", raw: strings.Replace(base, "\"role\": \"subject\"", "\"role\": \"provenance\"", 1)},
		{name: "duplicate path", raw: duplicateFirstFile(t, base)},
		{name: "parent path", raw: strings.Replace(base, "subject.php", "../subject.php", 1)},
		{name: "absolute path", raw: strings.Replace(base, "subject.php", "/subject.php", 1)},
		{name: "backslash path", raw: strings.Replace(base, "subject.php", "dir\\\\subject.php", 1)},
		{name: "non-ASCII path", raw: strings.Replace(base, "subject.php", "日本語.php", 1)},
		{name: "invalid role", raw: strings.Replace(base, "\"role\": \"subject\"", "\"role\": \"other\"", 1)},
		{name: "zero bytes", raw: strings.Replace(base, "\"bytes\": 6", "\"bytes\": 0", 1)},
		{name: "oversized file declaration", raw: strings.Replace(base, "\"bytes\": 6", "\"bytes\": 16777217", 1)},
		{name: "uppercase file hash", raw: strings.Replace(base, sha256Hex([]byte(entries[0].data)), strings.Repeat("A", 64), 1)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertReject(t, []byte(tc.raw), root)
		})
	}

	t.Run("more than 64 files", func(t *testing.T) {
		many := make([]corpusFixtureFile, 65)
		for i := range many {
			many[i] = corpusFixtureFile{path: "many/" + twoDigit(i) + ".php", role: "subject", data: "x"}
		}
		manyRoot := t.TempDir()
		writeFixtureFiles(t, manyRoot, many)
		assertReject(t, validManifest(t, many), manyRoot)
	})

	t.Run("manifest path must be regular", func(t *testing.T) {
		assertRejectPath(t, t.TempDir(), root)
		actual := writeManifest(t, []byte(base))
		link := filepath.Join(t.TempDir(), "manifest-link.json")
		if err := os.Symlink(actual, link); err != nil {
			t.Skipf("symlinks unavailable: %v", err)
		}
		assertRejectPath(t, link, root)
	})
}

func TestCorpusSourceRejections(t *testing.T) {
	entries := []corpusFixtureFile{
		{path: "a.php", role: "subject", data: "<?php echo 'a';\n"},
		{path: "b.php", role: "subject", data: "<?php echo 'b';\n"},
		{path: "SOURCE", role: "provenance", data: "pinned source\n"},
	}

	t.Run("source root missing or not a directory", func(t *testing.T) {
		root := t.TempDir()
		manifestPath := writeManifest(t, validManifest(t, entries))
		assertReject(t, mustReadFile(t, manifestPath), filepath.Join(root, "missing"))
		fileRoot := filepath.Join(root, "not-a-directory")
		if err := os.WriteFile(fileRoot, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
		assertReject(t, mustReadFile(t, manifestPath), fileRoot)
	})

	for _, tc := range []struct {
		name   string
		mutate func(t *testing.T, root string)
	}{
		{
			name: "same-size content drift", mutate: func(t *testing.T, root string) {
				if err := os.WriteFile(filepath.Join(root, "a.php"), []byte(entries[1].data), 0o600); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "content drift", mutate: func(t *testing.T, root string) {
				if err := os.WriteFile(filepath.Join(root, "a.php"), []byte("changed"), 0o600); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "missing listed file", mutate: func(t *testing.T, root string) {
				if err := os.Remove(filepath.Join(root, "a.php")); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "listed directory", mutate: func(t *testing.T, root string) {
				path := filepath.Join(root, "a.php")
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(path, 0o700); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "late failure has no partial report", mutate: func(t *testing.T, root string) {
				if err := os.WriteFile(filepath.Join(root, "b.php"), []byte("changed late"), 0o600); err != nil {
					t.Fatal(err)
				}
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			writeFixtureFiles(t, root, entries)
			tc.mutate(t, root)
			assertReject(t, validManifest(t, entries), root)
		})
	}

	t.Run("leaf and component symlinks", func(t *testing.T) {
		root := t.TempDir()
		writeFixtureFiles(t, root, entries)
		if err := os.Symlink(filepath.Join(root, "a.php"), filepath.Join(root, "linked.php")); err != nil {
			t.Skipf("symlinks unavailable: %v", err)
		}
		leaf := []corpusFixtureFile{{path: "linked.php", role: "subject", data: entries[0].data}}
		assertReject(t, validManifest(t, leaf), root)

		if err := os.Mkdir(filepath.Join(root, "real"), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "real", "inner.php"), []byte(entries[0].data), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(filepath.Join(root, "real"), filepath.Join(root, "alias")); err != nil {
			t.Fatal(err)
		}
		component := []corpusFixtureFile{{path: "alias/inner.php", role: "subject", data: entries[0].data}}
		assertReject(t, validManifest(t, component), root)
	})
}

func validManifest(t *testing.T, entries []corpusFixtureFile) []byte {
	t.Helper()
	files := make([]File, 0, len(entries))
	for _, entry := range entries {
		files = append(files, File{
			Path:   entry.path,
			Role:   entry.role,
			Bytes:  int64(len(entry.data)),
			SHA256: sha256Hex([]byte(entry.data)),
		})
	}
	manifest := struct {
		Schema     string    `json:"schema"`
		ID         string    `json:"id"`
		Project    string    `json:"project"`
		Version    string    `json:"version"`
		Source     Origin    `json:"source"`
		Selection  Selection `json:"selection"`
		PHPTargets []string  `json:"php_targets"`
		Files      []File    `json:"files"`
	}{
		Schema:  ManifestSchema,
		ID:      "synthetic-corpus",
		Project: "Synthetic Project",
		Version: "0.0.0-test",
		Source: Origin{
			Repository: "https://github.com/example/synthetic",
			Revision:   strings.Repeat("a", 40),
		},
		Selection: Selection{
			Unit:              "whole-file",
			Description:       "Synthetic whole-file input",
			DependencyClosure: "unresolved",
		},
		PHPTargets: []string{"8.3", "8.4"},
		Files:      files,
	}
	raw, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	return raw
}

func writeFixtureFiles(t *testing.T, root string, entries []corpusFixtureFile) {
	t.Helper()
	for _, entry := range entries {
		path := filepath.Join(root, filepath.FromSlash(entry.path))
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatalf("mkdir %s: %v", entry.path, err)
		}
		if err := os.WriteFile(path, []byte(entry.data), 0o600); err != nil {
			t.Fatalf("write %s: %v", entry.path, err)
		}
	}
}

func writeManifest(t *testing.T, raw []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "manifest.json")
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	return path
}

func assertReject(t *testing.T, raw []byte, root string) {
	t.Helper()
	assertRejectPath(t, writeManifest(t, raw), root)
}

func assertRejectPath(t *testing.T, manifestPath, root string) {
	t.Helper()
	report, err := Inspect(manifestPath, root)
	if err == nil {
		t.Fatal("Inspect() error = nil, want rejection")
	}
	if !reflect.DeepEqual(report, Report{}) {
		t.Errorf("Inspect() returned partial report on error: %+v", report)
	}
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func manifestPaths(files []File) []string {
	paths := make([]string, len(files))
	for i, file := range files {
		paths[i] = file.Path
	}
	return paths
}

func observationPaths(files []Observation) []string {
	paths := make([]string, len(files))
	for i, file := range files {
		paths[i] = file.Path
	}
	return paths
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func replaceFiles(t *testing.T, raw, replacement string) string {
	t.Helper()
	var value map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		t.Fatal(err)
	}
	value["files"] = json.RawMessage(replacement)
	updated, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return string(updated)
}

func duplicateFirstFile(t *testing.T, raw string) string {
	t.Helper()
	var value struct {
		Files []File `json:"files"`
	}
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		t.Fatal(err)
	}
	value.Files = append(value.Files, value.Files[0])
	encoded, err := json.Marshal(value.Files)
	if err != nil {
		t.Fatal(err)
	}
	return replaceFiles(t, raw, string(encoded))
}

func twoDigit(value int) string {
	const digits = "0123456789"
	return string([]byte{digits[value/10], digits[value%10]})
}
