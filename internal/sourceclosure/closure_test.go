package sourceclosure

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestInspectGraph(t *testing.T) {
	files := map[string]string{
		"app/entry.php":                  "<?php\nrequire_once __DIR__ . '/lib/first.php';\nrequire_once __DIR__ . '/lib/second.php';\nrequire_once __DIR__ . '/lib/first.php';\n",
		"app/lib/first.php":              "<?php\nclass First {}\nrequire_once __DIR__ . '/nested/shared.php';\n",
		"app/lib/second.php":             "<?php\nfunction second() {}\nrequire_once __DIR__ . '/nested/shared.php';\n",
		"app/lib/nested/shared.php":      "<?php\nclass Shared {}\nrequire_once __DIR__ . '/deeper/tail.php';\n",
		"app/lib/nested/deeper/tail.php": "<?php\nfunction tail() {}\n",
		"unreached.php":                  "<?php\nfunction untouched() {}\n",
	}
	root, manifest := writeFixture(t, files, []string{"8.3", "8.4"})

	for _, version := range []string{"8.3", "8.4"} {
		report, err := Inspect(manifest, root, "app/entry.php", version, frontend.Options{})
		if err != nil {
			t.Fatalf("Inspect(%s): %v", version, err)
		}
		again, err := Inspect(manifest, root, "app/entry.php", version, frontend.Options{})
		if err != nil {
			t.Fatalf("second Inspect(%s): %v", version, err)
		}
		if !reflect.DeepEqual(again, report) {
			t.Fatalf("Inspect(%s) was not deterministic:\nfirst: %#v\nagain: %#v", version, report, again)
		}
		if report.Schema != "gohp.source-closure/v0.1" || report.Status != "graph-checked" || report.PHPTarget != version || report.RegexProfile != "" || report.Entry != "app/entry.php" || report.Compatibility != "unverified" {
			t.Fatalf("unexpected report for %s: %#v", version, report)
		}
		manifestBytes, err := os.ReadFile(manifest)
		if err != nil {
			t.Fatal(err)
		}
		wantDigest := sha256.Sum256(manifestBytes)
		if report.ManifestSHA256 != hex.EncodeToString(wantDigest[:]) {
			t.Fatalf("manifest identity = %q, want exact manifest digest", report.ManifestSHA256)
		}
		wantModules := []string{"app/entry.php", "app/lib/first.php", "app/lib/nested/shared.php", "app/lib/nested/deeper/tail.php", "app/lib/second.php"}
		if got := moduleIDs(report.Modules); !reflect.DeepEqual(got, wantModules) {
			t.Fatalf("module DFS order = %v, want %v", got, wantModules)
		}
		for _, module := range report.Modules {
			contents := files[module.SourceID]
			wantHash := sha256.Sum256([]byte(contents))
			if module.Bytes != int64(len(contents)) || module.SHA256 != hex.EncodeToString(wantHash[:]) {
				t.Fatalf("module identity was not retained for %s: %#v", module.SourceID, module)
			}
		}
		wantDeclarations := map[string][]frontend.ModuleDeclaration{
			"app/lib/first.php":              {{Kind: "class", Name: "First", Line: 2}},
			"app/lib/nested/shared.php":      {{Kind: "class", Name: "Shared", Line: 2}},
			"app/lib/nested/deeper/tail.php": {{Kind: "function", Name: "tail", Line: 2}},
			"app/lib/second.php":             {{Kind: "function", Name: "second", Line: 2}},
		}
		for sourceID, want := range wantDeclarations {
			if got := moduleFor(t, report.Modules, sourceID).Declarations; !reflect.DeepEqual(got, want) {
				t.Fatalf("declarations for %s = %#v, want %#v", sourceID, got, want)
			}
		}
		wantEdges := []Edge{
			{From: "app/entry.php", Line: 2, Suffix: "/lib/first.php", To: "app/lib/first.php"},
			{From: "app/lib/first.php", Line: 3, Suffix: "/nested/shared.php", To: "app/lib/nested/shared.php"},
			{From: "app/lib/nested/shared.php", Line: 3, Suffix: "/deeper/tail.php", To: "app/lib/nested/deeper/tail.php"},
			{From: "app/entry.php", Line: 3, Suffix: "/lib/second.php", To: "app/lib/second.php"},
			{From: "app/lib/second.php", Line: 3, Suffix: "/nested/shared.php", To: "app/lib/nested/shared.php", Reused: true},
			{From: "app/entry.php", Line: 4, Suffix: "/lib/first.php", To: "app/lib/first.php", Reused: true},
		}
		if !reflect.DeepEqual(report.Edges, wantEdges) {
			t.Fatalf("edge DFS order = %#v, want %#v", report.Edges, wantEdges)
		}
		if !reflect.DeepEqual(report.Unreached, []string{"unreached.php"}) {
			t.Fatalf("unreached source selection = %v", report.Unreached)
		}
	}
}

func TestInspectGraphRejectsEdges(t *testing.T) {
	cases := []struct {
		name  string
		files map[string]string
		entry string
		want  string
	}{
		{
			name: "unlisted source", entry: "entry.php", want: "entry.php",
			files: map[string]string{"entry.php": "<?php\nrequire_once __DIR__ . '/missing.php';\n"},
		},
		{
			name: "cycle", entry: "entry.php", want: "two.php",
			files: map[string]string{
				"entry.php": "<?php\nrequire_once __DIR__ . '/two.php';\n",
				"two.php":   "<?php\nrequire_once __DIR__ . '/entry.php';\n",
			},
		},
		{
			name: "dynamic", entry: "entry.php", want: "entry.php",
			files: map[string]string{"entry.php": "<?php\nrequire_once $path;\n"},
		},
		{
			name: "conditional", entry: "entry.php", want: "entry.php",
			files: map[string]string{"entry.php": "<?php\nif (true) { require_once __DIR__ . '/child.php'; }\n", "child.php": "<?php\n"},
		},
		{
			name: "local", entry: "entry.php", want: "entry.php",
			files: map[string]string{"entry.php": "<?php\nfunction local_load() { require_once __DIR__ . '/child.php'; }\n", "child.php": "<?php\n"},
		},
		{
			name: "top level code", entry: "entry.php", want: "entry.php",
			files: map[string]string{"entry.php": "<?php\n$x = 1;\n"},
		},
		{
			name: "declaration after load", entry: "entry.php", want: "entry.php",
			files: map[string]string{"entry.php": "<?php\nrequire_once __DIR__ . '/child.php';\nfunction after() {}\n", "child.php": "<?php\n"},
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			root, manifest := writeFixture(t, test.files, []string{"8.3"})
			report, err := Inspect(manifest, root, test.entry, "8.3", frontend.Options{})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Inspect error = %v, want offending source %q", err, test.want)
			}
			assertZeroReport(t, report)
		})
	}
}

func TestInspectGraphRejectsCollisions(t *testing.T) {
	t.Run("same kind is case insensitive", func(t *testing.T) {
		files := map[string]string{
			"entry.php": "<?php\nrequire_once __DIR__ . '/one.php';\nrequire_once __DIR__ . '/two.php';\n",
			"one.php":   "<?php\nfunction Shared() {}\n",
			"two.php":   "<?php\nfunction shared() {}\n",
		}
		root, manifest := writeFixture(t, files, []string{"8.3"})
		report, err := Inspect(manifest, root, "entry.php", "8.3", frontend.Options{})
		if err == nil || !strings.Contains(err.Error(), "two.php") {
			t.Fatalf("same-kind duplicate error = %v", err)
		}
		assertZeroReport(t, report)
	})

	t.Run("same bytes retain distinct source identities", func(t *testing.T) {
		files := map[string]string{
			"entry.php": "<?php\nrequire_once __DIR__ . '/one.php';\nrequire_once __DIR__ . '/two.php';\n",
			"one.php":   "<?php\n",
			"two.php":   "<?php\n",
		}
		root, manifest := writeFixture(t, files, []string{"8.3"})
		report, err := Inspect(manifest, root, "entry.php", "8.3", frontend.Options{})
		if err != nil {
			t.Fatal(err)
		}
		if got, want := moduleIDs(report.Modules), []string{"entry.php", "one.php", "two.php"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("identical files were deduplicated or reordered: %v", got)
		}
		if report.Modules[1].SHA256 != report.Modules[2].SHA256 || report.Modules[1].SourceID == report.Modules[2].SourceID {
			t.Fatalf("content hash replaced source identity: %#v", report.Modules)
		}
	})

	t.Run("class and function names have separate namespaces", func(t *testing.T) {
		files := map[string]string{
			"entry.php":    "<?php\nrequire_once __DIR__ . '/class.php';\nrequire_once __DIR__ . '/function.php';\n",
			"class.php":    "<?php\nclass Same {}\n",
			"function.php": "<?php\nfunction same() {}\n",
		}
		root, manifest := writeFixture(t, files, []string{"8.3"})
		if _, err := Inspect(manifest, root, "entry.php", "8.3", frontend.Options{}); err != nil {
			t.Fatalf("class/function namespace collision: %v", err)
		}
	})
}

func TestInspectGraphRejectsManifestDrift(t *testing.T) {
	files := map[string]string{
		"entry.php": "<?php\n",
		"late.php":  "<?php\nfunction later() {}\n",
	}
	for _, test := range []struct {
		name string
		path string
	}{
		{name: "late source", path: "late.php"},
		{name: "notice", path: "NOTICE"},
		{name: "provenance", path: "PROVENANCE"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root, manifest := writeFixture(t, files, []string{"8.3"})
			if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(test.path)), []byte("changed\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			report, err := Inspect(manifest, root, "entry.php", "8.3", frontend.Options{})
			if err == nil || !strings.Contains(err.Error(), test.path) {
				t.Fatalf("Inspect error = %v, want drift path %q", err, test.path)
			}
			assertZeroReport(t, report)
		})
	}
}

func TestInspectGraphTargetProfile(t *testing.T) {
	files := map[string]string{"entry.php": "<?php\nfunction usable() {}\n"}
	t.Run("version is enabled but not selected", func(t *testing.T) {
		root, manifest := writeFixture(t, files, []string{"8.3"})
		report, err := Inspect(manifest, root, "entry.php", "8.4", frontend.Options{})
		if err == nil || !strings.Contains(err.Error(), "not selected") {
			t.Fatalf("Inspect error = %v", err)
		}
		assertZeroReport(t, report)
	})
	t.Run("version is not enabled", func(t *testing.T) {
		root, manifest := writeFixture(t, files, []string{"8.3"})
		report, err := Inspect(manifest, root, "entry.php", "8.2", frontend.Options{})
		if err == nil || !strings.Contains(err.Error(), "not enabled") {
			t.Fatalf("Inspect error = %v", err)
		}
		assertZeroReport(t, report)
	})
	t.Run("experimental pcre2 is php83 only", func(t *testing.T) {
		root, manifest := writeFixture(t, files, []string{"8.3", "8.4"})
		report, err := Inspect(manifest, root, "entry.php", "8.4", frontend.Options{ExperimentalPCRE2: true})
		if err == nil || !strings.Contains(err.Error(), "requires PHP 8.3") {
			t.Fatalf("Inspect error = %v", err)
		}
		assertZeroReport(t, report)
	})
	t.Run("experimental pcre2 is recorded for php83", func(t *testing.T) {
		root, manifest := writeFixture(t, files, []string{"8.3"})
		report, err := Inspect(manifest, root, "entry.php", "8.3", frontend.Options{ExperimentalPCRE2: true})
		if err != nil {
			t.Fatal(err)
		}
		if report.RegexProfile != ir.Preg83Profile {
			t.Fatalf("regex profile = %q, want %q", report.RegexProfile, ir.Preg83Profile)
		}
	})
}

func moduleIDs(modules []Module) []string {
	ids := make([]string, len(modules))
	for i, module := range modules {
		ids[i] = module.SourceID
	}
	return ids
}

func moduleFor(t *testing.T, modules []Module, sourceID string) Module {
	t.Helper()
	for _, module := range modules {
		if module.SourceID == sourceID {
			return module
		}
	}
	t.Fatalf("module %q was not found in %#v", sourceID, modules)
	return Module{}
}

func assertZeroReport(t *testing.T, report Report) {
	t.Helper()
	if !reflect.DeepEqual(report, Report{}) {
		t.Fatalf("failed inspection returned partial report: %#v", report)
	}
}

func writeFixture(t *testing.T, sources map[string]string, targets []string) (string, string) {
	t.Helper()
	root := t.TempDir()
	files := make([]map[string]any, 0, len(sources)+2)
	for _, name := range sortedNames(sources) {
		if err := writeFixtureFile(root, name, sources[name]); err != nil {
			t.Fatal(err)
		}
		files = append(files, fixtureFile(name, "source", sources[name]))
	}
	for _, extra := range []struct{ name, role, contents string }{
		{"NOTICE", "notice", "static closure test notice\n"},
		{"PROVENANCE", "provenance", "static closure test provenance\n"},
	} {
		if err := writeFixtureFile(root, extra.name, extra.contents); err != nil {
			t.Fatal(err)
		}
		files = append(files, fixtureFile(extra.name, extra.role, extra.contents))
	}
	manifest := map[string]any{
		"schema":      "gohp.corpus.manifest/v0.1",
		"id":          "source-closure-test",
		"upstream":    map[string]any{"repository": "test/source-closure", "revision": strings.Repeat("a", 40), "version": "0"},
		"selection":   map[string]any{"unit": "whole-file", "exclusions": "fixture only", "dependencies": "unresolved"},
		"php_targets": targets,
		"files":       files,
	}
	b, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "manifest.json")
	if err := os.WriteFile(path, b, 0o600); err != nil {
		t.Fatal(err)
	}
	return root, path
}

func writeFixtureFile(root, name, contents string) error {
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(contents), 0o600)
}

func fixtureFile(name, role, contents string) map[string]any {
	digest := sha256.Sum256([]byte(contents))
	return map[string]any{"path": name, "role": role, "bytes": int64(len(contents)), "sha256": hex.EncodeToString(digest[:])}
}

func sortedNames(values map[string]string) []string {
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
