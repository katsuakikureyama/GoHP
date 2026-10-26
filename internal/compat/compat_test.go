package compat

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEvaluateRequiresExactDifferentialLifecycleAndPayload(t *testing.T) {
	testCase := Case{ID: "case.core", Test: "TestDifferential/core", Package: modulePath + "/tests", File: "tests/differential_test.go", Kind: "differential", Fixture: "tests/fixtures/core.php"}
	catalog := Catalog{
		Schema: "gohp.compatibility.catalog/v0.1", SHA256: strings.Repeat("a", 64), root: t.TempDir(),
		cases:   map[string]Case{"case.core": testCase},
		Entries: []Entry{{ID: "php.core", Kind: "syntax", Name: "Core", Implementation: "partial", Contract: "closed", PHPTargets: []string{"8.3"}, Checks: []Check{{CaseID: "case.core", Kind: "differential"}}}},
	}
	// Match LoadCatalog's canonical root, including macOS /var -> /private/var.
	canonical, err := filepath.EvalSymlinks(catalog.root)
	if err != nil {
		t.Fatal(err)
	}
	catalog.root = canonical
	if err := os.MkdirAll(filepath.Join(catalog.root, "tests", "fixtures"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(catalog.root, testCase.Fixture), []byte("<?php echo 1;"), 0600); err != nil {
		t.Fatal(err)
	}
	fixture, err := os.ReadFile(filepath.Join(catalog.root, testCase.Fixture))
	if err != nil {
		t.Fatal(err)
	}
	payload := map[string]any{
		"case_id": "case.core", "php_target": "8.3", "php_version": "8.3.30", "fixture_sha256": hashBytes(fixture), "generated_sha256": strings.Repeat("b", 64),
		"phases":      map[string]string{"parse": "pass", "codegen": "pass", "go-build": "pass", "go-execute": "pass", "differential": "pass"},
		"comparisons": map[string]string{"stdout": "pass", "exit": "pass", "stderr": "not-checked", "exceptions": "not-checked", "side_effects": "not-checked"},
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	base := []map[string]string{
		{"Action": "start", "Package": testCase.Package},
		{"Action": "run", "Package": testCase.Package, "Test": "TestDifferential"},
		{"Action": "run", "Package": testCase.Package, "Test": testCase.Test},
		{"Action": "output", "Package": testCase.Package, "Test": testCase.Test, "Output": "    GOHP_CASE " + string(encoded) + "\n"},
		{"Action": "pass", "Package": testCase.Package, "Test": testCase.Test},
		{"Action": "pass", "Package": testCase.Package, "Test": "TestDifferential"},
		{"Action": "pass", "Package": testCase.Package},
	}
	raw := eventLines(t, base)
	report, err := Evaluate(catalog, strings.NewReader(raw), testMetadata())
	if err != nil {
		t.Fatal(err)
	}
	if report.Overall != "pass" || report.Cases[0].Status != "pass" || report.Features[0].Status != "differential-pass (case-limited)" {
		t.Fatalf("unexpected passing report: %#v", report)
	}
	if report.Cases[0].Phase == nil || report.Cases[0].GeneratedSHA256 != strings.Repeat("b", 64) {
		t.Fatalf("missing measured phase evidence: %#v", report.Cases[0])
	}

	// A passed child cannot hide a failed parent; the phase data remains useful.
	for _, tt := range []struct {
		name   string
		change func([]map[string]string) []map[string]string
	}{
		{"missing package start", func(e []map[string]string) []map[string]string { return e[1:] }},
		{"missing test run", func(e []map[string]string) []map[string]string { return append(e[:2], e[3:]...) }},
		{"duplicate terminal", func(e []map[string]string) []map[string]string { return append(e, e[4]) }},
		{"wrong emitting test", func(e []map[string]string) []map[string]string { e[3]["Test"] = "TestOther"; return e }},
		{"wrong target", func(e []map[string]string) []map[string]string {
			e[3]["Output"] = strings.ReplaceAll(e[3]["Output"], `"php_target":"8.3"`, `"php_target":"7.4"`)
			return e
		}},
		{"stale fixture", func(e []map[string]string) []map[string]string {
			e[3]["Output"] = strings.ReplaceAll(e[3]["Output"], hashBytes(fixture), strings.Repeat("c", 64))
			return e
		}},
		{"unmeasured stderr claimed pass", func(e []map[string]string) []map[string]string {
			e[3]["Output"] = strings.ReplaceAll(e[3]["Output"], `"stderr":"not-checked"`, `"stderr":"pass"`)
			return e
		}},
		{"truncated other package", func(e []map[string]string) []map[string]string {
			return append(e, map[string]string{"Action": "start", "Package": modulePath + "/other"})
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			b, err := json.Marshal(base)
			if err != nil {
				t.Fatal(err)
			}
			var altered []map[string]string
			if err := json.Unmarshal(b, &altered); err != nil {
				t.Fatal(err)
			}
			r, err := Evaluate(catalog, strings.NewReader(eventLines(t, tt.change(altered))), testMetadata())
			if err != nil {
				t.Fatal(err)
			}
			if r.Overall == "pass" || r.Features[0].Status == "differential-pass (case-limited)" {
				t.Fatalf("invalid evidence promoted to PASS: %+v", r)
			}
		})
	}
	// Empty packages are emitted as skip by Go, not as a skipped required test.
	emptyPackage := append(append([]map[string]string{}, base...), map[string]string{"Action": "start", "Package": modulePath + "/ir"}, map[string]string{"Action": "skip", "Package": modulePath + "/ir"})
	if r, err := Evaluate(catalog, strings.NewReader(eventLines(t, emptyPackage)), testMetadata()); err != nil || r.Overall != "pass" {
		t.Fatalf("package without tests broke complete evidence: %+v, %v", r, err)
	}
	base[5]["Action"] = "fail"
	report, err = Evaluate(catalog, strings.NewReader(eventLines(t, base)), testMetadata())
	if err != nil {
		t.Fatal(err)
	}
	if report.Overall != "failed" || report.Cases[0].Status != "failed" || report.Cases[0].Phase == nil {
		t.Fatalf("parent failure was promoted or lost evidence: %#v", report)
	}
}

func TestEvaluateRejectsVacuousOrTruncatedEvidence(t *testing.T) {
	catalog := Catalog{
		Schema: "gohp.compatibility.catalog/v0.1", root: t.TempDir(), cases: map[string]Case{"case.compiler": {ID: "case.compiler", Test: "TestCompiler", Package: modulePath + "/cmd/gohp", File: "cmd/gohp/main_test.go", Kind: "compiler"}},
		Entries: []Entry{
			{ID: "tool.empty", Kind: "tooling", Name: "Empty", Implementation: "implemented", Contract: "closed", PHPTargets: []string{"8.3"}},
			{ID: "tool.checked", Kind: "tooling", Name: "Checked", Implementation: "implemented", Contract: "closed", PHPTargets: []string{"8.3"}, Checks: []Check{{CaseID: "case.compiler", Kind: "compiler"}}},
		},
	}
	raw := eventLines(t, []map[string]string{
		{"Action": "run", "Package": modulePath + "/cmd/gohp", "Test": "TestCompiler"},
		{"Action": "pass", "Package": modulePath + "/cmd/gohp", "Test": "TestCompiler"},
		// The omitted package terminal makes a truncated stream incomplete.
	})
	report, err := Evaluate(catalog, strings.NewReader(raw), testMetadata())
	if err != nil {
		t.Fatal(err)
	}
	features := make(map[string]FeatureRow, len(report.Features))
	for _, feature := range report.Features {
		features[feature.ID] = feature
	}
	if features["tool.empty"].Status != "unverified" || features["tool.checked"].Status != "incomplete" || report.Overall != "incomplete" {
		t.Fatalf("vacuous or truncated evidence passed: %#v", report)
	}
}

func TestLoadCatalogRejectsUnsafeAndIncompleteChecks(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "compatibility"), 0700); err != nil {
		t.Fatal(err)
	}
	base := `{"schema":"gohp.compatibility.catalog/v0.1","maturity":"draft","entries":[{"id":"x","kind":"syntax","name":"x","implementation":"partial","contract":"x","php_targets":["8.3"],"checks":[{"case_id":"x","test":"TestX","file":"../outside_test.go","kind":"unit","assertions":["x"]}],"limitations":[]}]}`
	if err := os.WriteFile(filepath.Join(root, "compatibility", "catalog.json"), []byte(base), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCatalog(root); err == nil || !strings.Contains(err.Error(), "path traversal") {
		t.Fatalf("unsafe check path accepted: %v", err)
	}
	unknown := strings.Replace(base, `"maturity":"draft"`, `"maturity":"draft","unexpected":true`, 1)
	if err := os.WriteFile(filepath.Join(root, "compatibility", "catalog.json"), []byte(unknown), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCatalog(root); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("unknown catalog field accepted: %v", err)
	}
}

func TestMarkdownEscapesCellsAndStatesUnmeasuredScope(t *testing.T) {
	b := string(Markdown(Report{Metadata: Metadata{PHPTarget: "8.3", Dirty: "dirty", Provenance: "unbound-checkout"}, Features: []FeatureRow{{ID: "x|y", Target: "7.4", Implementation: "partial", Status: "not-measured", Contract: "one\ntwo", Limitations: []string{"a|b"}}}}))
	if !strings.Contains(b, "x\\|y") || strings.Contains(b, "one\ntwo") || !strings.Contains(b, "not measured in this run") || !strings.Contains(b, "Stderr, exception and side-effect comparisons are not measured") {
		t.Fatalf("markdown is unsafe or omits scope: %q", b)
	}
}

func eventLines(t *testing.T, events []map[string]string) string {
	t.Helper()
	var b strings.Builder
	for _, event := range events {
		line, err := json.Marshal(event)
		if err != nil {
			t.Fatal(err)
		}
		b.Write(line)
		b.WriteByte('\n')
	}
	return b.String()
}

func testMetadata() Metadata {
	b := testPHPBaseline("8.3.30")
	return Metadata{GoVersion: "go1.test", GOOS: "test", GOARCH: "test", PHPTarget: "8.3", PHPVersion: "8.3.30", Command: "go test -json", StartedAt: "2026-10-25T00:00:00Z", Dirty: "clean", Provenance: "local-git", PHPBaseline: b, PHPBaselineSHA256: PHPBaselineHash(b)}
}
