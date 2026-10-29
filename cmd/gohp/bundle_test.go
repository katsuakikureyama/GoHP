package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/codegen"
	"github.com/katsuakikureyama/GoHP/internal/corpus"
	"github.com/katsuakikureyama/GoHP/internal/ir"
	"github.com/katsuakikureyama/GoHP/internal/sourceclosure"
)

func TestCLIBundleFreshReport(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"app/entry.php": "<?php\nrequire_once __DIR__ . '/dep.php';\necho bundle_message();\n$box = new BundleBox();\necho $box->value();\n",
		"app/dep.php":   "<?php\nfunction bundle_message() {\n  return 'hello';\n}\nclass BundleBox {\n  public function value() {\n    return 'box';\n  }\n}\n",
		"NOTICE":        "bundle test notice\n",
		"PROVENANCE":    "bundle test provenance\n",
	}
	root, manifest := writeBundleCorpus(t, dir, files)
	outputDir := mkdir(t, dir, "generated")
	reportDir := mkdir(t, dir, "reports")
	output := filepath.Join(outputDir, "bundle.go")
	report := filepath.Join(reportDir, "bundle.json")

	stdout, stderr, err := runCLI(t,
		"-manifest", manifest, "-source-root", root, "-entry", "app/entry.php",
		"-out", output, "-report", report,
	)
	if err != nil {
		t.Fatalf("bundle failed: %v\nstderr: %s", err, stderr)
	}
	if stdout != "" {
		t.Fatalf("stdout = %q, want empty when -out is used", stdout)
	}
	generated := readFile(t, output)
	if !strings.Contains(string(generated), "func F_bundle_message") || !strings.Contains(string(generated), "type C_BundleBox") {
		t.Fatalf("generated source omitted dependency declarations:\n%s", generated)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(readFile(t, report), &raw); err != nil {
		t.Fatalf("decode report: %v", err)
	}
	if _, ok := raw["source_sha256"]; ok {
		t.Fatal("bundle report must not present one source_sha256 as a closure fingerprint")
	}
	if _, ok := raw["input"]; ok {
		t.Fatal("bundle report must not present one input as the closure")
	}
	var got struct {
		Schema          string               `json:"schema"`
		ManifestSHA256  string               `json:"manifest_sha256"`
		Entry           string               `json:"entry"`
		Sources         []ir.SourceFile      `json:"sources"`
		LoadingGraph    sourceclosure.Report `json:"loading_graph"`
		PHPTarget       string               `json:"php_target"`
		RegexProfile    string               `json:"regex_profile"`
		GeneratedSHA256 string               `json:"generated_sha256"`
		Output          string               `json:"output"`
		Status          string               `json:"status"`
		Compatibility   string               `json:"compatibility"`
		SourceMap       codegen.SourceMap    `json:"source_map"`
	}
	if err := json.Unmarshal(readFile(t, report), &got); err != nil {
		t.Fatalf("decode typed report: %v", err)
	}
	manifestBytes := readFile(t, manifest)
	if want := sha256Hex(manifestBytes); got.Schema != "gohp.build/v0.2" || got.ManifestSHA256 != want || got.Entry != "app/entry.php" || got.PHPTarget != phpTarget(t) || got.RegexProfile != "" || got.Output != output || got.Status != "generated" || got.Compatibility != "not established by compilation" || got.GeneratedSHA256 != sha256Hex(generated) {
		t.Fatalf("unexpected bundle report: %#v", got)
	}
	assertBundleSources(t, got.Sources, files, "app/entry.php", "app/dep.php")
	if got.LoadingGraph.Schema != "gohp.source-closure/v0.1" || got.LoadingGraph.Status != "graph-checked" || got.LoadingGraph.ManifestSHA256 != got.ManifestSHA256 || got.LoadingGraph.Entry != got.Entry || len(got.LoadingGraph.Edges) != 1 {
		t.Fatalf("unexpected loading graph: %#v", got.LoadingGraph)
	}
	edge := got.LoadingGraph.Edges[0]
	if edge.From != "app/entry.php" || edge.Line != 2 || edge.Suffix != "/dep.php" || edge.To != "app/dep.php" || edge.Reused {
		t.Fatalf("unexpected loading edge: %#v", edge)
	}
	if got.SourceMap.Schema != "gohp.sourcemap/v0.2" || got.SourceMap.Precision != "statement-start-file-line" {
		t.Fatalf("unexpected bundle source map header: %#v", got.SourceMap)
	}
	assertBundleSources(t, got.SourceMap.Sources, files, "app/entry.php", "app/dep.php")
	seenMapID := map[string]bool{}
	for _, mapping := range got.SourceMap.Mappings {
		seenMapID[mapping.SourceID] = true
	}
	if !seenMapID["app/entry.php"] || !seenMapID["app/dep.php"] {
		t.Fatalf("source map lost original file identities: %#v", got.SourceMap.Mappings)
	}
	for path, want := range files {
		if got := string(readFile(t, filepath.Join(root, filepath.FromSlash(path)))); got != want {
			t.Errorf("source %s changed: got %q, want %q", path, got, want)
		}
	}
	assertNoStagedArtifacts(t, outputDir, reportDir)
}

func TestCLIBundleFailures(t *testing.T) {
	base := map[string]string{
		"app/entry.php": "<?php\nrequire_once __DIR__ . '/dep.php';\necho bundle_message();\n",
		"app/dep.php":   "<?php\nfunction bundle_message() { return 'ok'; }\n",
		"NOTICE":        "bundle test notice\n",
		"PROVENANCE":    "bundle test provenance\n",
	}

	t.Run("mixed and partial modes", func(t *testing.T) {
		dir := t.TempDir()
		root, manifest := writeBundleCorpus(t, dir, base)
		legacy := filepath.Join(root, "app", "entry.php")
		cases := []struct {
			name string
			args []string
		}{
			{"mixed legacy input", []string{"-in", legacy, "-manifest", manifest, "-source-root", root, "-entry", "app/entry.php"}},
			{"manifest only", []string{"-manifest", manifest}},
			{"missing entry", []string{"-manifest", manifest, "-source-root", root}},
			{"missing manifest", []string{"-source-root", root, "-entry", "app/entry.php"}},
		}
		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				out := filepath.Join(dir, strings.ReplaceAll(tt.name, " ", "-")+".go")
				report := filepath.Join(dir, strings.ReplaceAll(tt.name, " ", "-")+".json")
				args := append(append([]string{}, tt.args...), "-out", out, "-report", report)
				_, stderr, err := runCLI(t, args...)
				if err == nil {
					t.Fatalf("bundle unexpectedly succeeded; stderr: %s", stderr)
				}
				assertDoesNotExist(t, out)
				assertDoesNotExist(t, report)
				assertNoStagedArtifacts(t, dir)
			})
		}
	})

	failureCases := []struct {
		name       string
		files      map[string]string
		entry      string
		mutate     func(t *testing.T, root string)
		wantStderr string
	}{
		{
			name:  "manifest drift",
			files: base,
			entry: "app/entry.php",
			mutate: func(t *testing.T, root string) {
				writeFile(t, filepath.Join(root, "app"), "dep.php", "<?php\nfunction bundle_message() { return 'changed'; }\n")
			},
			wantStderr: "size or sha256 mismatch",
		},
		{
			name: "unregistered dependency",
			files: map[string]string{
				"app/entry.php": "<?php\nrequire_once __DIR__ . '/missing.php';\n",
				"NOTICE":        "bundle test notice\n",
				"PROVENANCE":    "bundle test provenance\n",
			},
			entry:      "app/entry.php",
			wantStderr: "not a selected source",
		},
		{
			name: "executable dependency",
			files: map[string]string{
				"app/entry.php": "<?php\nrequire_once __DIR__ . '/dep.php';\n",
				"app/dep.php":   "<?php\necho 'side effect';\n",
				"NOTICE":        "bundle test notice\n",
				"PROVENANCE":    "bundle test provenance\n",
			},
			entry:      "app/entry.php",
			wantStderr: "app/dep.php",
		},
		{
			name: "late load",
			files: map[string]string{
				"app/entry.php": "<?php\necho 'driver';\nrequire_once __DIR__ . '/dep.php';\n",
				"app/dep.php":   "<?php\nfunction dep() {}\n",
				"NOTICE":        "bundle test notice\n",
				"PROVENANCE":    "bundle test provenance\n",
			},
			entry:      "app/entry.php",
			wantStderr: "app/entry.php",
		},
		{
			name: "conditional load",
			files: map[string]string{
				"app/entry.php": "<?php\nif (true) { require_once __DIR__ . '/dep.php'; }\n",
				"app/dep.php":   "<?php\nfunction dep() {}\n",
				"NOTICE":        "bundle test notice\n",
				"PROVENANCE":    "bundle test provenance\n",
			},
			entry:      "app/entry.php",
			wantStderr: "app/entry.php",
		},
		{
			name: "dynamic load",
			files: map[string]string{
				"app/entry.php": "<?php\nrequire_once $path;\n",
				"app/dep.php":   "<?php\nfunction dep() {}\n",
				"NOTICE":        "bundle test notice\n",
				"PROVENANCE":    "bundle test provenance\n",
			},
			entry:      "app/entry.php",
			wantStderr: "app/entry.php",
		},
		{
			name: "dependency semantic error",
			files: map[string]string{
				"app/entry.php": "<?php\nrequire_once __DIR__ . '/dep.php';\necho 'driver';\n",
				"app/dep.php":   "<?php\nfunction invalid_bundle_parameter(MissingBundleClass $value) {}\n",
				"NOTICE":        "bundle test notice\n",
				"PROVENANCE":    "bundle test provenance\n",
			},
			entry:      "app/entry.php",
			wantStderr: "app/dep.php",
		},
	}
	for _, tt := range failureCases {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			root, manifest := writeBundleCorpus(t, dir, tt.files)
			if tt.mutate != nil {
				tt.mutate(t, root)
			}
			out := filepath.Join(dir, "generated.go")
			report := filepath.Join(dir, "report.json")
			_, stderr, err := runCLI(t,
				"-manifest", manifest, "-source-root", root, "-entry", tt.entry,
				"-out", out, "-report", report,
			)
			if err == nil {
				t.Fatalf("bundle unexpectedly succeeded; stderr: %s", stderr)
			}
			if !strings.Contains(stderr, tt.wantStderr) {
				t.Fatalf("stderr = %q, want source-local failure containing %q", stderr, tt.wantStderr)
			}
			assertDoesNotExist(t, out)
			assertDoesNotExist(t, report)
			assertNoStagedArtifacts(t, dir)
		})
	}
}

func TestCLIBundlePreservesArtifacts(t *testing.T) {
	for _, tt := range []struct {
		name    string
		prepare func(t *testing.T, root, dir string) (out, report string, preserved map[string][]byte)
	}{
		{
			name: "existing output",
			prepare: func(t *testing.T, _, dir string) (string, string, map[string][]byte) {
				out := writeFile(t, dir, "existing.go", "keep output")
				return out, filepath.Join(dir, "report.json"), map[string][]byte{out: []byte("keep output")}
			},
		},
		{
			name: "existing report",
			prepare: func(t *testing.T, _, dir string) (string, string, map[string][]byte) {
				out := filepath.Join(dir, "generated.go")
				report := writeFile(t, dir, "existing.json", "keep report")
				return out, report, map[string][]byte{report: []byte("keep report")}
			},
		},
		{
			name: "source as output",
			prepare: func(t *testing.T, root, dir string) (string, string, map[string][]byte) {
				entry := filepath.Join(root, "app", "entry.php")
				return entry, filepath.Join(dir, "report.json"), map[string][]byte{entry: readFile(t, entry)}
			},
		},
		{
			name: "hard linked source as report",
			prepare: func(t *testing.T, root, dir string) (string, string, map[string][]byte) {
				entry := filepath.Join(root, "app", "entry.php")
				report := filepath.Join(dir, "source-alias.json")
				if err := os.Link(entry, report); err != nil {
					t.Skipf("hard links unavailable: %v", err)
				}
				return filepath.Join(dir, "generated.go"), report, map[string][]byte{entry: readFile(t, entry), report: readFile(t, report)}
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			root, manifest := writeBundleCorpus(t, dir, map[string]string{
				"app/entry.php": "<?php\nrequire_once __DIR__ . '/dep.php';\necho bundle_message();\n",
				"app/dep.php":   "<?php\nfunction bundle_message() { return 'ok'; }\n",
				"NOTICE":        "bundle test notice\n",
				"PROVENANCE":    "bundle test provenance\n",
			})
			out, report, preserved := tt.prepare(t, root, dir)
			_, stderr, err := runCLI(t,
				"-manifest", manifest, "-source-root", root, "-entry", "app/entry.php",
				"-out", out, "-report", report,
			)
			if err == nil {
				t.Fatalf("bundle unexpectedly succeeded; stderr: %s", stderr)
			}
			for path, want := range preserved {
				if got := readFile(t, path); string(got) != string(want) {
					t.Errorf("%s was clobbered: got %q, want %q", path, got, want)
				}
			}
			if _, kept := preserved[out]; !kept {
				assertDoesNotExist(t, out)
			}
			assertNoStagedArtifacts(t, dir)
		})
	}
}

func writeBundleCorpus(t *testing.T, dir string, files map[string]string) (root, manifest string) {
	t.Helper()
	root = mkdir(t, dir, "source")
	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	entries := make([]corpus.ManifestFile, 0, len(paths))
	for _, path := range paths {
		contents := []byte(files[path])
		filename := filepath.Join(root, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filename, contents, 0o644); err != nil {
			t.Fatal(err)
		}
		role := "source"
		switch path {
		case "NOTICE":
			role = "notice"
		case "PROVENANCE":
			role = "provenance"
		}
		digest := sha256.Sum256(contents)
		entries = append(entries, corpus.ManifestFile{Path: path, Role: role, Bytes: int64(len(contents)), SHA256: hex.EncodeToString(digest[:])})
	}
	b, err := json.Marshal(corpus.Manifest{
		Schema: "gohp.corpus.manifest/v0.1", ID: "bundle-cli-test",
		Upstream:   corpus.Upstream{Repository: "example/bundle", Revision: strings.Repeat("a", 40), Version: "test"},
		Selection:  corpus.Selection{Unit: "whole-file", Exclusions: "synthetic CLI test corpus", Dependencies: "unresolved"},
		PHPTargets: []string{"8.3", "8.4"}, Files: entries,
	})
	if err != nil {
		t.Fatal(err)
	}
	manifest = writeFile(t, dir, "manifest.json", string(b))
	return root, manifest
}

func assertBundleSources(t *testing.T, sources []ir.SourceFile, files map[string]string, wantIDs ...string) {
	t.Helper()
	if len(sources) != len(wantIDs) {
		t.Fatalf("source count = %d, want %d: %#v", len(sources), len(wantIDs), sources)
	}
	byID := make(map[string]ir.SourceFile, len(sources))
	for _, source := range sources {
		byID[source.ID] = source
	}
	for _, id := range wantIDs {
		got, ok := byID[id]
		if !ok {
			t.Fatalf("missing source identity %q in %#v", id, sources)
		}
		want := []byte(files[id])
		if got.Bytes != int64(len(want)) || got.SHA256 != sha256Hex(want) {
			t.Errorf("source identity for %s = %#v, want bytes/hash for %q", id, got, string(want))
		}
	}
}
