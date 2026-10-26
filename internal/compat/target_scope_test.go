package compat

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestCatalogTargetScopes(t *testing.T) {
	check := func(id, test string, targets []string) Check {
		return Check{
			CaseID: id, Test: test, File: "internal/compat/target_scope_test.go", Kind: "differential", Fixture: "tests/fixtures/target_scope.php",
			PHPTargets: targets, Assertions: []string{"target-scoped differential execution"},
		}
	}
	entry := func(id string, targets []string, checks ...Check) Entry {
		return Entry{ID: id, Kind: "tooling", Name: id, Implementation: "partial", Contract: "target scope", PHPTargets: targets, Checks: checks}
	}

	catalog := Catalog{Schema: "gohp.compatibility.catalog/v0.2", Maturity: "experimental", Entries: []Entry{
		entry("tool.default", []string{"8.3", "8.4"}, check("case.default", "TestDefault", nil)),
		entry("tool.only83", []string{"8.3"}, check("case.only83", "TestOnly83", []string{"8.3"})),
		entry("tool.shared83", []string{"8.3"}, check("case.shared", "TestShared", nil)),
		entry("tool.shared84", []string{"8.4"}, check("case.shared", "TestShared", nil)),
	}}
	if err := catalog.validateCatalog(false); err != nil {
		t.Fatalf("validate scoped catalog: %v", err)
	}
	if got := strings.Join(catalog.cases["case.default"].PHPTargets, ","); got != "8.3,8.4" {
		t.Fatalf("default case scope = %q, want inherited entry targets", got)
	}
	if got := strings.Join(catalog.cases["case.shared"].PHPTargets, ","); got != "8.3,8.4" {
		t.Fatalf("reused case scope = %q, want normalized union", got)
	}
	for _, target := range []string{"8.3", "8.4"} {
		cases, err := catalog.DifferentialCases(target)
		if err != nil {
			t.Fatalf("differential cases for %s: %v", target, err)
		}
		ids := make([]string, len(cases))
		for i := range cases {
			ids[i] = cases[i].ID
		}
		want := "case.default,case.shared"
		if target == "8.3" {
			want = "case.default,case.only83,case.shared"
		}
		if got := strings.Join(ids, ","); got != want {
			t.Fatalf("differential cases for %s = %q, want %q", target, got, want)
		}
	}
	if _, err := catalog.DifferentialCases("9.9"); err == nil {
		t.Fatal("unsupported execution target accepted")
	}

	for _, invalid := range []struct {
		name   string
		mutate func(*Catalog)
	}{
		{"explicit empty", func(c *Catalog) { c.Entries[0].Checks[0].PHPTargets = []string{} }},
		{"duplicate", func(c *Catalog) { c.Entries[0].Checks[0].PHPTargets = []string{"8.3", "8.3"} }},
		{"outside entry", func(c *Catalog) { c.Entries[1].Checks[0].PHPTargets = []string{"8.4"} }},
		{"legacy field", func(c *Catalog) {
			c.Schema, c.Entries[0].Checks[0].PHPTargets = "gohp.compatibility.catalog/v0.1", []string{"8.3"}
		}},
	} {
		t.Run(invalid.name, func(t *testing.T) {
			copy := catalog
			encoded, err := json.Marshal(catalog)
			if err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(encoded, &copy); err != nil {
				t.Fatal(err)
			}
			invalid.mutate(&copy)
			if err := copy.validateCatalog(false); err == nil {
				t.Fatal("invalid target scope accepted")
			}
		})
	}
}

func TestEvaluateTargetScopedCases(t *testing.T) {
	catalog := targetScopedUnitCatalog(t, false)
	meta := targetScopeMetadata("8.4")
	events := targetScopeEvents(t, "TestShared", "pass")
	report, err := evaluate(catalog, strings.NewReader(events), meta, nil, RunSchema)
	if err != nil {
		t.Fatal(err)
	}
	if report.Overall != "pass" {
		t.Fatalf("out-of-scope omitted case made complete 8.4 run fail: %#v", report)
	}
	outside := targetScopeCase(t, report, "case.only83")
	if outside.Status != "not-measured" || outside.Detail != "case is not applicable to selected PHP target" || outside.Phase != nil || outside.Outputs != nil || outside.Diagnostics != nil || outside.GeneratedSHA256 != "" {
		t.Fatalf("out-of-scope case acquired execution evidence: %#v", outside)
	}
	if row := targetScopeRow(t, report, "tool.shared", "8.4"); row.Status != "compiler-check-pass" || strings.Join(row.Checks, ",") != "case.shared" {
		t.Fatalf("active target row did not use its applicable check: %#v", row)
	}
	if row := targetScopeRow(t, report, "tool.shared", "8.3"); row.Status != "not-measured" {
		t.Fatalf("non-selected target row promoted: %#v", row)
	}

	for _, terminal := range []struct {
		name, test, action string
		wantOverall        string
	}{
		{"missing active", "", "", "incomplete"},
		{"skipped active", "TestShared", "skip", "incomplete"},
	} {
		t.Run(terminal.name, func(t *testing.T) {
			var raw string
			if terminal.test == "" {
				raw = targetScopeEvents(t, "", "")
			} else {
				raw = targetScopeEvents(t, terminal.test, terminal.action)
			}
			r, err := evaluate(catalog, strings.NewReader(raw), meta, nil, RunSchema)
			if err != nil {
				t.Fatal(err)
			}
			if r.Overall != terminal.wantOverall || targetScopeCase(t, r, "case.shared").Status == "pass" {
				t.Fatalf("active reused case was not required: %#v", r)
			}
		})
	}

	emptyChecks := targetScopedUnitCatalog(t, true)
	r, err := evaluate(emptyChecks, strings.NewReader(events), meta, nil, RunSchema)
	if err != nil {
		t.Fatal(err)
	}
	row := targetScopeRow(t, r, "tool.only83-on-both", "8.4")
	if len(row.Checks) != 0 || row.Status != "unverified" || r.Overall != "incomplete" {
		t.Fatalf("empty target-specific check set promoted to pass: row=%#v report=%#v", row, r)
	}
	emptyChecks.Entries[len(emptyChecks.Entries)-1].Implementation = "unsupported"
	r, err = evaluate(emptyChecks, strings.NewReader(events), meta, nil, RunSchema)
	if err != nil || targetScopeRow(t, r, "tool.only83-on-both", "8.4").Status != "unsupported" || r.Overall != "pass" {
		t.Fatalf("explicit unsupported scope must stay unsupported, not a pass or an incomplete claim: %#v, %v", r, err)
	}
}

func TestTargetScopedArchiveReplay(t *testing.T) {
	catalog := targetScopedUnitCatalog(t, false)
	catalogBytes, err := json.Marshal(catalog)
	if err != nil {
		t.Fatal(err)
	}
	catalog.SHA256 = hashBytes(catalogBytes)
	events := targetScopeEvents(t, "TestShared", "pass")
	report, err := evaluate(catalog, strings.NewReader(events), targetScopeMetadata("8.4"), nil, RunSchema)
	if err != nil {
		t.Fatal(err)
	}
	if targetScopeCase(t, report, "case.only83").Status != "not-measured" {
		t.Fatalf("test setup did not produce scoped evidence: %#v", report)
	}

	writeArchive := func(t *testing.T, r Report) string {
		t.Helper()
		dir := t.TempDir()
		writeTestFile(t, filepath.Join(dir, "catalog.json"), catalogBytes)
		writeTestFile(t, filepath.Join(dir, "go-test.jsonl"), []byte(events))
		writeTestJSON(t, filepath.Join(dir, "run.json"), r)
		return dir
	}
	if replayed, err := LoadArchive(writeArchive(t, report)); err != nil || replayed.Overall != "pass" || targetScopeCase(t, replayed, "case.only83").Status != "not-measured" {
		t.Fatalf("target-scoped v0.5 archive did not replay: %#v, %v", replayed, err)
	}

	fabricated := report
	fabricated.Cases = append([]CaseEvidence(nil), report.Cases...)
	for i := range fabricated.Cases {
		if fabricated.Cases[i].ID == "case.only83" {
			fabricated.Cases[i].Status = "pass"
			fabricated.Cases[i].Outputs = testExecutionOutputs("fabricated", "fabricated")
		}
	}
	if _, err := LoadArchive(writeArchive(t, fabricated)); err == nil {
		t.Fatal("archive accepted fabricated evidence for an out-of-scope case")
	}

	legacy := report
	legacy.Schema = RunSchemaOutputs
	if _, err := LoadArchive(writeArchive(t, legacy)); err == nil {
		t.Fatal("pre-v0.5 run accepted a v0.2 target-scoped catalog")
	}
}

func targetScopedUnitCatalog(t *testing.T, includeEmptyTargetRow bool) Catalog {
	t.Helper()
	check := func(id, test string, targets []string) Check {
		return Check{CaseID: id, Test: test, File: "internal/compat/target_scope_test.go", Kind: "unit", PHPTargets: targets, Assertions: []string{"target-scoped execution"}}
	}
	entries := []Entry{
		{ID: "tool.shared", Kind: "tooling", Name: "shared", Implementation: "implemented", Contract: "shared target case", PHPTargets: []string{"8.3", "8.4"}, Checks: []Check{check("case.shared", "TestShared", nil)}},
		{ID: "tool.only83", Kind: "tooling", Name: "only83", Implementation: "implemented", Contract: "8.3-only case", PHPTargets: []string{"8.3"}, Checks: []Check{check("case.only83", "TestOnly83", nil)}},
	}
	if includeEmptyTargetRow {
		entries = append(entries, Entry{ID: "tool.only83-on-both", Kind: "tooling", Name: "only83 on both", Implementation: "implemented", Contract: "target-filtered row", PHPTargets: []string{"8.3", "8.4"}, Checks: []Check{check("case.only83", "TestOnly83", []string{"8.3"})}})
	}
	catalog := Catalog{Schema: "gohp.compatibility.catalog/v0.2", Maturity: "experimental", Entries: entries}
	if err := catalog.validateCatalog(false); err != nil {
		t.Fatalf("validate test catalog: %v", err)
	}
	return catalog
}

func targetScopeMetadata(target string) Metadata {
	meta := testMetadata()
	meta.SourceSHA256 = strings.Repeat("a", 64)
	meta.PHPTarget, meta.PHPVersion = target, target+".30"
	meta.PHPBaseline = testPHPBaseline(meta.PHPVersion)
	meta.PHPBaselineSHA256 = PHPBaselineHash(meta.PHPBaseline)
	return meta
}

func targetScopeEvents(t *testing.T, test, terminal string) string {
	t.Helper()
	pkg := modulePath + "/internal/compat"
	events := []map[string]string{{"Action": "start", "Package": pkg}}
	if test != "" {
		events = append(events, map[string]string{"Action": "run", "Package": pkg, "Test": test}, map[string]string{"Action": terminal, "Package": pkg, "Test": test})
	}
	events = append(events, map[string]string{"Action": "pass", "Package": pkg})
	return eventLines(t, events)
}

func targetScopeCase(t *testing.T, report Report, id string) CaseEvidence {
	t.Helper()
	for _, evidence := range report.Cases {
		if evidence.ID == id {
			return evidence
		}
	}
	t.Fatalf("case %q was not recorded", id)
	return CaseEvidence{}
}

func targetScopeRow(t *testing.T, report Report, id, target string) FeatureRow {
	t.Helper()
	for _, row := range report.Features {
		if row.ID == id && row.Target == target {
			return row
		}
	}
	t.Fatalf("feature %q target %q was not recorded", id, target)
	return FeatureRow{}
}
