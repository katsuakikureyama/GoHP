package compat

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func testExecutionOutputs(php, goOutput string) *ExecutionOutputs {
	return &ExecutionOutputs{PHP: &ProcessOutput{StdoutBytes: len(php), StdoutSHA256: hashBytes([]byte(php)), ExitCode: 0}, Go: &ProcessOutput{StdoutBytes: len(goOutput), StdoutSHA256: hashBytes([]byte(goOutput)), ExitCode: 0}}
}

func TestOutputObservationsValidateVerdicts(t *testing.T) {
	for _, tt := range []struct {
		name   string
		change func(*casePayload)
		valid  bool
	}{
		{"matching", func(*casePayload) {}, true},
		{"absent", func(p *casePayload) { p.Outputs = nil }, false},
		{"missing_php", func(p *casePayload) { p.Outputs.PHP = nil }, false},
		{"launch_failure", func(p *casePayload) {
			p.Outputs.PHP = nil
			p.Diagnostics.PHP = nil
			p.Diagnostics.Status = "not-checked"
			p.Comparisons.Stdout = "not-checked"
			p.Comparisons.Exit = "fail"
		}, true},
		{"mismatch_claimed_pass", func(p *casePayload) { p.Outputs.Go.StdoutSHA256 = hashBytes([]byte("else")) }, false},
		{"mismatch", func(p *casePayload) {
			p.Outputs.Go.StdoutSHA256 = hashBytes([]byte("else"))
			p.Comparisons.Stdout = "fail"
		}, true},
		{"same_claimed_failure", func(p *casePayload) { p.Comparisons.Stdout = "fail" }, false},
		{"nonzero_claimed_success", func(p *casePayload) { p.Outputs.Go.ExitCode = 7 }, false},
		{"nonzero", func(p *casePayload) {
			p.Outputs.Go.ExitCode = 7
			p.Comparisons.Exit = "fail"
			p.Comparisons.Stdout = "not-checked"
		}, true},
		{"success_claimed_exit_failure", func(p *casePayload) { p.Comparisons.Exit = "fail" }, false},
		{"empty_hash_nonempty_length", func(p *casePayload) { p.Outputs.Go.StdoutSHA256 = hashBytes(nil) }, false},
		{"negative_length", func(p *casePayload) { p.Outputs.Go.StdoutBytes = -1 }, false},
		{"invalid_exit", func(p *casePayload) { p.Outputs.Go.ExitCode = -2 }, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := casePayload{Outputs: testExecutionOutputs("same", "same"), Diagnostics: testQuietEvidence()}
			p.Comparisons.Stdout, p.Comparisons.Exit = "pass", "pass"
			tt.change(&p)
			for _, schema := range []string{RunSchemaOutputs, RunSchema} {
				if detail := outputShape(p, schema); (detail == "") != tt.valid {
					t.Fatalf("schema=%s valid=%v: %s", schema, tt.valid, detail)
				}
			}
		})
	}
}

func TestProcessOutputRequiresExplicitZeroObservations(t *testing.T) {
	valid := `{"stdout_bytes":0,"stdout_sha256":"` + hashBytes(nil) + `","exit_code":0}`
	for _, input := range []string{`{}`, strings.Replace(valid, `"exit_code":0`, `"exit_code":null`, 1), strings.Replace(valid, `,"exit_code":0`, "", 1), strings.Replace(valid, `"stdout_bytes":0,`, "", 1), strings.Replace(valid, `"exit_code":0`, `"exit_code":0,"extra":1`, 1)} {
		var p ProcessOutput
		if err := json.Unmarshal([]byte(input), &p); err == nil {
			t.Fatalf("incomplete output accepted: %s", input)
		}
	}
	var p ProcessOutput
	if err := json.Unmarshal([]byte(valid), &p); err != nil {
		t.Fatal(err)
	}
}

func TestOutputArchiveKeepsHistoricalObservationsAbsent(t *testing.T) {
	for _, schema := range []string{RunSchemaLegacy, RunSchemaBaseline, RunSchemaQuiet, RunSchemaOutputs, RunSchema} {
		dir, _ := savedRun(t, "8.3", "pass", schema)
		r, err := LoadArchive(dir)
		if err != nil {
			t.Fatal(err)
		}
		if (r.Cases[0].Outputs != nil) != (schema == RunSchemaOutputs || schema == RunSchema) {
			t.Fatal("historical output invented")
		}
		if schema == RunSchemaQuiet {
			r.Schema = RunSchema
			writeTestJSON(t, filepath.Join(dir, "run.json"), r)
			if _, err := LoadArchive(dir); err == nil {
				t.Fatal("historical evidence acquired output guarantee")
			}
		}
	}
	for _, change := range []string{"hash", "exit", "missing", "downgrade"} {
		dir, r := savedRun(t, "8.3", "pass", RunSchema)
		switch change {
		case "hash":
			r.Cases[0].Outputs.Go.StdoutSHA256 = hashBytes([]byte("else"))
		case "exit":
			r.Cases[0].Outputs.Go.ExitCode = 7
		case "missing":
			r.Cases[0].Outputs = nil
		case "downgrade":
			r.Schema = RunSchemaQuiet
		}
		writeTestJSON(t, filepath.Join(dir, "run.json"), r)
		if _, err := LoadArchive(dir); err == nil {
			t.Fatalf("changed %s accepted", change)
		}
	}
}
