package compat

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func testQuietEvidence() *DiagnosticEvidence {
	return &DiagnosticEvidence{Profile: QuietSuccessProfile, PHPArgs: QuietPHPArgs(), Status: "pass", PHP: &StderrObservation{Bytes: 0, SHA256: hashBytes(nil)}, Go: &StderrObservation{Bytes: 0, SHA256: hashBytes(nil)}}
}

func TestQuietEvidenceCannotPromoteUnobservedOrNoisyRuns(t *testing.T) {
	for _, tt := range []struct {
		name   string
		change func(*casePayload)
		want   string
	}{
		{"quiet", func(*casePayload) {}, "pass"},
		{"absent", func(p *casePayload) { p.Diagnostics = nil }, "incomplete"},
		{"unobserved_php", func(p *casePayload) { p.Diagnostics.PHP = nil }, "incomplete"},
		{"unobserved_go", func(p *casePayload) { p.Diagnostics.Go = nil }, "incomplete"},
		{"not_run", func(p *casePayload) {
			p.Diagnostics.PHP = nil
			p.Diagnostics.Go = nil
			p.Diagnostics.Status = "not-checked"
		}, "incomplete"},
		{"unknown_profile", func(p *casePayload) { p.Diagnostics.Profile = "anything" }, "incomplete"},
		{"disabled_warnings", func(p *casePayload) { p.Diagnostics.PHPArgs[1] = "error_reporting=0" }, "incomplete"},
		{"missing_hash", func(p *casePayload) { p.Diagnostics.Go.SHA256 = "" }, "incomplete"},
		{"contradictory_hash", func(p *casePayload) { p.Diagnostics.PHP.SHA256 = hashBytes([]byte("warning")) }, "incomplete"},
		{"negative_length", func(p *casePayload) { p.Diagnostics.PHP.Bytes = -1 }, "incomplete"},
		{"noise_claimed_quiet", func(p *casePayload) {
			p.Diagnostics.PHP = &StderrObservation{Bytes: 7, SHA256: hashBytes([]byte("warning"))}
		}, "incomplete"},
		{"matching_noise", func(p *casePayload) {
			p.Diagnostics.PHP = &StderrObservation{Bytes: 7, SHA256: hashBytes([]byte("warning"))}
			p.Diagnostics.Go = &StderrObservation{Bytes: 7, SHA256: hashBytes([]byte("warning"))}
			p.Diagnostics.Status = "fail"
		}, "failed"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir, original := savedRun(t, "8.3", "pass", RunSchema)
			catalogBytes, err := os.ReadFile(filepath.Join(dir, "catalog.json"))
			if err != nil {
				t.Fatal(err)
			}
			var c Catalog
			if err := strictJSON(catalogBytes, &c); err != nil {
				t.Fatal(err)
			}
			if err := c.validateCatalog(false); err != nil {
				t.Fatal(err)
			}
			c.SHA256 = hashBytes(catalogBytes)
			raw, err := os.ReadFile(filepath.Join(dir, "go-test.jsonl"))
			if err != nil {
				t.Fatal(err)
			}
			events, err := parseEvents(raw)
			if err != nil {
				t.Fatal(err)
			}
			var changed bytes.Buffer
			for _, event := range events {
				if p, ok := payloadFromOutput(event.Output); ok {
					tt.change(&p)
					encoded, err := json.Marshal(p)
					if err != nil {
						t.Fatal(err)
					}
					event.Output = "GOHP_CASE " + string(encoded)
				}
				if err := json.NewEncoder(&changed).Encode(event); err != nil {
					t.Fatal(err)
				}
			}
			r, err := evaluate(c, &changed, original.Metadata, func(Case) (string, string) { return original.Cases[0].FixtureSHA256, "" }, RunSchema)
			if err != nil {
				t.Fatal(err)
			}
			if r.Cases[0].Status != tt.want {
				t.Fatalf("got %s want %s: %+v", r.Cases[0].Status, tt.want, r.Cases[0])
			}
			if tt.want != "pass" && r.Overall == "pass" {
				t.Fatal("invalid quiet evidence promoted run")
			}
		})
	}
}

func TestQuietArchivePreservesHistoricalScope(t *testing.T) {
	var reports []Report
	for _, schema := range []string{RunSchemaLegacy, RunSchemaBaseline, RunSchemaQuiet, RunSchema} {
		dir, expected := savedRun(t, "8.3", "pass", schema)
		r, err := LoadArchive(dir)
		if err != nil || !reflect.DeepEqual(r, expected) {
			t.Fatalf("load %s: %v", schema, err)
		}
		if (r.Cases[0].Diagnostics != nil) != (schema == RunSchema || schema == RunSchemaQuiet) {
			t.Fatal("legacy diagnostic evidence synthesized")
		}
		reports = append(reports, r)
		if schema == RunSchemaBaseline {
			r.Schema = RunSchema
			writeTestJSON(t, filepath.Join(dir, "run.json"), r)
			if _, err := LoadArchive(dir); err == nil {
				t.Fatal("old evidence acquired quiet guarantee")
			}
		}
	}
	v, err := BuildView(reports, ViewFilter{})
	if err != nil {
		t.Fatal(err)
	}
	for i, r := range reports {
		if v.Runs[i].RunSchema != r.Schema {
			t.Fatal("view lost original schema")
		}
	}
	for _, change := range []string{"observation", "phase", "downgrade"} {
		dir, r := savedRun(t, "8.3", "pass", RunSchema)
		switch change {
		case "observation":
			r.Cases[0].Diagnostics.Go.Bytes = 1
		case "phase":
			r.Cases[0].Phase.QuietStderr = "not-checked"
		case "downgrade":
			r.Schema = RunSchemaBaseline
		}
		writeTestJSON(t, filepath.Join(dir, "run.json"), r)
		if _, err := LoadArchive(dir); err == nil {
			t.Fatalf("accepted changed %s", change)
		}
	}
}
