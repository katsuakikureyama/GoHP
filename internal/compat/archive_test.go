package compat

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A saved differential run is deliberately independent of the consumer's
// filesystem: there is no tests/fixture.php or tests/archive_test.go here.
func savedRun(t *testing.T, target, terminal string, schemas ...string) (string, Report) {
	t.Helper()
	entry := Entry{ID: "php.example", Kind: "builtin", Name: "example", Implementation: "partial", Contract: "fixture only", PHPTargets: []string{"7.4", "8.3"}, Limitations: []string{"no side effects compared"}, Checks: []Check{{CaseID: "case.example", Test: "TestExample", File: "tests/archive_test.go", Kind: "differential", Fixture: "tests/fixture.php", Assertions: []string{"stdout and successful exits"}}}}
	c := Catalog{Schema: "gohp.compatibility.catalog/v0.1", Maturity: "experimental", Entries: []Entry{entry}}
	if err := c.validateCatalog(false); err != nil {
		t.Fatal(err)
	}
	catalogBytes, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	c.SHA256 = hashBytes(catalogBytes)
	meta := testMetadata()
	meta.SourceSHA256 = strings.Repeat("a", 64)
	meta.PHPTarget, meta.PHPVersion = target, target+".30"
	schema := RunSchemaLegacy
	if len(schemas) != 0 {
		schema = schemas[0]
	}
	meta.PHPBaseline, meta.PHPBaselineSHA256 = nil, ""
	if schema != RunSchemaLegacy {
		meta.PHPBaseline = testPHPBaseline(meta.PHPVersion)
		meta.PHPBaselineSHA256 = PHPBaselineHash(meta.PHPBaseline)
	}
	fixtureHash := strings.Repeat("b", 64)
	p := casePayload{CaseID: "case.example", PHPTarget: target, PHPVersion: meta.PHPVersion, FixtureSHA256: fixtureHash, GeneratedSHA256: strings.Repeat("c", 64)}
	p.Phases.Parse, p.Phases.Codegen, p.Phases.GoBuild, p.Phases.GoExecute, p.Phases.Differential = "pass", "pass", "pass", "pass", "pass"
	p.Comparisons.Stdout, p.Comparisons.Exit = "pass", "pass"
	p.Comparisons.Stderr, p.Comparisons.Exceptions, p.Comparisons.SideEffects = "not-checked", "not-checked", "not-checked"
	if schema == RunSchema {
		p.Diagnostics = testQuietEvidence()
	}
	payload, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	pkg := modulePath + "/tests"
	events := eventLines(t, []map[string]string{
		{"Action": "start", "Package": pkg},
		{"Action": "run", "Package": pkg, "Test": "TestExample"},
		{"Action": "output", "Package": pkg, "Test": "TestExample", "Output": "GOHP_CASE " + string(payload)},
		{"Action": terminal, "Package": pkg, "Test": "TestExample"},
		{"Action": terminal, "Package": pkg},
	})
	if terminal == "fail" {
		meta.TestExit = 1
	}
	r, err := evaluate(c, strings.NewReader(events), meta, func(Case) (string, string) { return fixtureHash, "" }, schema)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "catalog.json"), catalogBytes)
	writeTestFile(t, filepath.Join(dir, "go-test.jsonl"), []byte(events))
	writeTestJSON(t, filepath.Join(dir, "run.json"), r)
	return dir, r
}

func writeTestFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
}

func writeTestJSON(t *testing.T, path string, value any) {
	t.Helper()
	b, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, path, b)
}

func TestLoadArchiveReplaysWithoutSourceAndRejectsChangedEvidence(t *testing.T) {
	for _, terminal := range []string{"pass", "fail", "skip"} {
		t.Run(terminal, func(t *testing.T) {
			dir, expected := savedRun(t, "8.3", terminal)
			r, err := LoadArchive(dir)
			if err != nil || r.Overall != expected.Overall || r.Features[1].Status != expected.Features[1].Status {
				t.Fatalf("read stored %s run: %+v, %v", terminal, r, err)
			}
		})
	}
	for _, change := range []string{"events", "catalog", "verdict", "phase", "duplicate case", "fixture hash", "missing report", "extra field", "escaping symlink"} {
		t.Run(change, func(t *testing.T) {
			dir, r := savedRun(t, "8.3", "pass")
			switch change {
			case "events":
				writeTestFile(t, filepath.Join(dir, "go-test.jsonl"), []byte("{}\n"))
			case "catalog":
				writeTestFile(t, filepath.Join(dir, "catalog.json"), []byte("{}"))
			case "verdict":
				r.Features[0].Status = "differential-pass (case-limited)"
				writeTestJSON(t, filepath.Join(dir, "run.json"), r)
			case "phase":
				r.Cases[0].Phase.Stderr = "pass"
				writeTestJSON(t, filepath.Join(dir, "run.json"), r)
			case "duplicate case":
				r.Cases = append(r.Cases, r.Cases[0])
				writeTestJSON(t, filepath.Join(dir, "run.json"), r)
			case "fixture hash":
				r.Cases[0].FixtureSHA256 = strings.Repeat("d", 64)
				writeTestJSON(t, filepath.Join(dir, "run.json"), r)
			case "missing report":
				if err := os.Remove(filepath.Join(dir, "run.json")); err != nil {
					t.Fatal(err)
				}
			case "extra field":
				b, err := json.Marshal(r)
				if err != nil {
					t.Fatal(err)
				}
				writeTestFile(t, filepath.Join(dir, "run.json"), append([]byte(`{"unexpected":true,`), b[1:]...))
			case "escaping symlink":
				path := filepath.Join(dir, "run.json")
				outside := filepath.Join(t.TempDir(), "outside.json")
				writeTestJSON(t, outside, r)
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(outside, path); err != nil {
					t.Skipf("symlink unavailable: %v", err)
				}
			}
			if _, err := LoadArchive(dir); err == nil {
				t.Fatal("changed/incomplete archive accepted")
			}
		})
	}
}

func TestViewKeepsRunsAndFailureVisibleThroughFilters(t *testing.T) {
	_, older := savedRun(t, "7.4", "pass")
	_, newer := savedRun(t, "8.3", "fail")
	newer.Metadata.Commit = strings.Repeat("d", 40)
	newer.Metadata.Image = "different-image"
	v, err := BuildView([]Report{older, newer}, ViewFilter{Target: "8.3", Kind: "builtin"})
	if err != nil {
		t.Fatal(err)
	}
	if len(v.Runs) != 2 || v.Runs[0].Features[0].Status != "not-measured" || v.Runs[1].Features[0].Status != "failed" || v.Runs[1].Metadata.Image != "different-image" {
		t.Fatalf("runs merged or metadata lost: %+v", v)
	}
	v, err = BuildView([]Report{older, newer}, ViewFilter{Status: "differential-pass (case-limited)"})
	if err != nil {
		t.Fatal(err)
	}
	if len(v.Runs[0].Features) != 1 || len(v.Runs[1].Features) != 0 || v.Runs[1].Overall != "failed" || len(v.Runs[1].Cases) != 1 {
		t.Fatalf("filter hid failed run: %+v", v)
	}
	md := string(ViewMarkdown(v))
	if !strings.Contains(md, "No matching feature rows") || !strings.Contains(md, "**failed**") || !strings.Contains(md, newer.Metadata.Commit) {
		t.Fatalf("filtered Markdown lost failure/provenance: %s", md)
	}
	v.Runs[0].Features[0].Name = "[click](javascript:bad)<script>|\nrow"
	md = string(ViewMarkdown(v))
	if strings.Contains(md, "<script>") || strings.Contains(md, "[click](javascript:") || strings.Contains(md, "|\nrow") {
		t.Fatalf("unescaped archive text: %s", md)
	}
	for _, f := range []ViewFilter{{Target: "8.5"}, {Kind: "unknown"}, {Status: "supported"}} {
		if _, err := BuildView([]Report{older}, f); err == nil {
			t.Fatalf("invalid filter accepted: %+v", f)
		}
	}
}
