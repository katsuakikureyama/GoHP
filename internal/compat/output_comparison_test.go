package compat

import (
	"strings"
	"testing"
)

// Synthetic observations exercise classification, not native PHP behavior.
func comparableOutputRuns(t *testing.T) (Report, Report) {
	t.Helper()
	_, a := savedRun(t, "8.3", "pass", RunSchema)
	_, b := savedRun(t, "8.4", "pass", RunSchema)
	for _, r := range []*Report{&a, &b} {
		r.Metadata.Commit = strings.Repeat("c", 40)
		r.Metadata.Tree = strings.Repeat("d", 40)
	}
	return a, b
}

func TestCompareOutputsSeparatesNativeAndTranspilerObservations(t *testing.T) {
	for _, tt := range []struct {
		name   string
		change func(*Report, *Report)
		status string
	}{
		{"same", func(*Report, *Report) {}, "matching-stdout-both-runs"},
		{"native_difference", func(a, b *Report) { b.Cases[0].Outputs = testExecutionOutputs("else", "else") }, "native-stdout-difference"},
		{"transpiler_mismatch", func(a, b *Report) {
			b.Cases[0].Outputs = testExecutionOutputs("same", "else")
			b.Cases[0].Phase.Stdout = "fail"
			b.Cases[0].Phase.Differential = "fail"
			b.Cases[0].Status = "failed"
			b.Overall = "failed"
		}, "transpiler-stdout-mismatch"},
		{"different_commit", func(a, b *Report) { b.Metadata.Commit = strings.Repeat("e", 40) }, "not-comparable"},
		{"unknown_revision", func(a, b *Report) { a.Metadata.Commit = ""; b.Metadata.Commit = "" }, "not-comparable"},
		{"different_source", func(a, b *Report) { b.Metadata.SourceSHA256 = strings.Repeat("e", 64) }, "not-comparable"},
		{"different_catalog", func(a, b *Report) { b.CatalogSHA256 = strings.Repeat("e", 64) }, "not-comparable"},
		{"different_go", func(a, b *Report) { b.Metadata.GoVersion = "go1.other" }, "not-comparable"},
		{"different_flags", func(a, b *Report) { b.Metadata.BuildFlags = "other" }, "not-comparable"},
		{"dirty", func(a, b *Report) { b.Metadata.Dirty = "dirty" }, "not-comparable"},
		{"fixture_changed", func(a, b *Report) { b.Cases[0].FixtureSHA256 = strings.Repeat("e", 64) }, "not-comparable"},
		{"missing_case", func(a, b *Report) { b.Cases = nil }, "not-comparable"},
		{"missing_output", func(a, b *Report) { b.Cases[0].Outputs = nil }, "not-comparable"},
		{"outside_target", func(a, b *Report) {
			b.Cases[0].Status = "not-measured"
			b.Cases[0].Outputs, b.Cases[0].Phase, b.Cases[0].Diagnostics = nil, nil, nil
			b.Cases[0].GeneratedSHA256 = ""
		}, "not-comparable"},
		{"mixed_output_schemas", func(a, b *Report) { b.Schema = RunSchemaOutputs }, "matching-stdout-both-runs"},
		{"legacy", func(a, b *Report) { b.Schema = RunSchemaQuiet; b.Cases[0].Outputs = nil }, "not-comparable"},
		{"unrelated_failure", func(a, b *Report) { b.Overall = "failed"; b.Cases[0].Status = "incomplete" }, "not-comparable"},
		{"nonzero", func(a, b *Report) { b.Cases[0].Outputs.Go.ExitCode = 7; b.Cases[0].Phase.Exit = "fail" }, "not-comparable"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			a, b := comparableOutputRuns(t)
			tt.change(&a, &b)
			r := CompareOutputs(a, b)
			if len(r.Cases) != 1 || r.Cases[0].Status != tt.status {
				t.Fatalf("want %s got %+v", tt.status, r)
			}
			if r.Left.Overall != a.Overall || r.Right.Overall != b.Overall {
				t.Fatal("original verdict lost")
			}
			if len(r.OracleDifferences) == 0 {
				t.Fatal("oracle differences hidden")
			}
		})
	}
	a, b := comparableOutputRuns(t)
	r := CompareOutputs(a, b)
	r.Cases[0].ID = "[x](javascript:bad)<script>|\n"
	text := string(OutputComparisonMarkdown(r))
	if strings.Contains(text, "<script>") || strings.Contains(text, "[x](javascript:") || !strings.Contains(text, "do not establish") {
		t.Fatal("unsafe or misleading comparison Markdown")
	}
}
