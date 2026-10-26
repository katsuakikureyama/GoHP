package compat

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

type ComparisonRun struct {
	Schema        string   `json:"run_schema"`
	Metadata      Metadata `json:"metadata"`
	Overall       string   `json:"overall"`
	CatalogSHA256 string   `json:"catalog_sha256"`
	EventSHA256   string   `json:"event_sha256"`
}

type OutputComparison struct {
	Schema            string                 `json:"schema"`
	Notice            string                 `json:"notice"`
	Left              ComparisonRun          `json:"left"`
	Right             ComparisonRun          `json:"right"`
	ContextIssues     []string               `json:"context_issues"`
	OracleDifferences []string               `json:"oracle_differences"`
	Cases             []OutputComparisonCase `json:"cases"`
}

type OutputComparisonCase struct {
	ID             string        `json:"id"`
	Status         string        `json:"status"`
	Detail         string        `json:"detail"`
	LeftMatch      string        `json:"left_php_go_stdout"`
	RightMatch     string        `json:"right_php_go_stdout"`
	NativeRelation string        `json:"native_stdout_relation"`
	Left           *CaseEvidence `json:"left,omitempty"`
	Right          *CaseEvidence `json:"right,omitempty"`
}

const outputComparisonNotice = "Observation-only comparison of two validated archives, not a support or release gate. Native stdout differences are across recorded oracle environments: they do not establish that PHP version caused a difference or that it is intentional. Hashes bind recorded bytes, not authenticity. No tests are rerun. Exceptions, nonzero-exit equivalence, arbitrary diagnostics and side effects are not compared. Historical runs never gain missing observations."

// CompareOutputs consumes reports already validated by LoadArchive. It never
// combines their verdicts or chooses a preferred run. Target-local observations
// stay visible even when cross-run provenance prevents comparison.
func CompareOutputs(left, right Report) OutputComparison {
	r := OutputComparison{Schema: "gohp.compatibility.output-comparison/v0.1", Notice: outputComparisonNotice,
		Left: comparisonRun(left), Right: comparisonRun(right), ContextIssues: []string{}, OracleDifferences: []string{}, Cases: []OutputComparisonCase{}}
	if !hasOutputObservations(left.Schema) || !hasOutputObservations(right.Schema) {
		r.ContextIssues = append(r.ContextIssues, "both runs require v0.4 or later output observations")
	}
	lm, rm := left.Metadata, right.Metadata
	if lm.Dirty != "clean" || rm.Dirty != "clean" {
		r.ContextIssues = append(r.ContextIssues, "both checkouts must be clean")
	}
	for _, field := range []struct{ name, left, right string }{
		{"commit", lm.Commit, rm.Commit}, {"tree", lm.Tree, rm.Tree},
		{"source_sha256", lm.SourceSHA256, rm.SourceSHA256}, {"catalog_sha256", left.CatalogSHA256, right.CatalogSHA256},
		{"go_version", lm.GoVersion, rm.GoVersion}, {"goos", lm.GOOS, rm.GOOS}, {"goarch", lm.GOARCH, rm.GOARCH},
		{"command", lm.Command, rm.Command},
	} {
		valid := field.left != "" && field.right != "" && field.left == field.right
		if field.name == "commit" || field.name == "tree" {
			decoded, err := hex.DecodeString(field.left)
			valid = valid && err == nil && len(decoded) == 20
		}
		if field.name == "source_sha256" || field.name == "catalog_sha256" {
			valid = valid && validSHA256(field.left)
		}
		if !valid {
			r.ContextIssues = append(r.ContextIssues, field.name+" is missing, invalid or differs")
		}
	}
	if lm.BuildFlags != rm.BuildFlags {
		r.ContextIssues = append(r.ContextIssues, "build_flags differ")
	}
	if lm.GoImage != rm.GoImage {
		r.ContextIssues = append(r.ContextIssues, "go_image differs")
	}
	for _, field := range []struct{ name, left, right string }{
		{"php_target", lm.PHPTarget, rm.PHPTarget}, {"php_version", lm.PHPVersion, rm.PHPVersion},
		{"php_image", lm.Image, rm.Image}, {"php_baseline_sha256", lm.PHPBaselineSHA256, rm.PHPBaselineSHA256},
	} {
		if field.left != field.right {
			r.OracleDifferences = append(r.OracleDifferences, field.name)
		}
	}
	leftCases, rightCases := differentialCases(left), differentialCases(right)
	ids := map[string]bool{}
	for id := range leftCases {
		ids[id] = true
	}
	for id := range rightCases {
		ids[id] = true
	}
	ordered := make([]string, 0, len(ids))
	for id := range ids {
		ordered = append(ordered, id)
	}
	sort.Strings(ordered)
	for _, id := range ordered {
		row := OutputComparisonCase{ID: id, Status: "not-comparable", NativeRelation: "not-compared", Left: leftCases[id], Right: rightCases[id]}
		row.LeftMatch, row.RightMatch = localStdoutRelation(row.Left), localStdoutRelation(row.Right)
		switch {
		case len(r.ContextIssues) != 0:
			row.Detail = "cross-run context does not match; target-local observations only"
		case row.Left == nil || row.Right == nil:
			row.Detail = "case is absent in one run"
		case row.Left.Fixture != row.Right.Fixture || !validSHA256(row.Left.FixtureSHA256) || row.Left.FixtureSHA256 != row.Right.FixtureSHA256 || row.Left.Test != row.Right.Test || row.Left.Package != row.Right.Package:
			row.Detail = "case source or execution identity differs"
		case row.LeftMatch == "mismatch" || row.RightMatch == "mismatch":
			row.Status, row.Detail = "transpiler-stdout-mismatch", "at least one run observed PHP and generated Go stdout differ; no combined success verdict"
		case left.Overall != "pass" || right.Overall != "pass" || row.Left.Status != "pass" || row.Right.Status != "pass":
			row.Detail = "a run or case is not passing; no cross-run native classification"
		case row.LeftMatch != "match" || row.RightMatch != "match":
			row.Detail = "successful quiet process outputs are not fully observed"
		case sameStdout(row.Left.Outputs.PHP, row.Right.Outputs.PHP):
			row.Status, row.NativeRelation, row.Detail = "matching-stdout-both-runs", "same", "both runs match their PHP oracle and recorded native stdout agrees"
		default:
			row.Status, row.NativeRelation, row.Detail = "native-stdout-difference", "different", "each generated result matches its own PHP oracle; difference cause and intent are not established"
		}
		r.Cases = append(r.Cases, row)
	}
	return r
}

func hasOutputObservations(schema string) bool {
	return schema == RunSchemaOutputs || schema == RunSchema
}

func comparisonRun(r Report) ComparisonRun {
	return ComparisonRun{Schema: r.Schema, Metadata: r.Metadata, Overall: r.Overall, CatalogSHA256: r.CatalogSHA256, EventSHA256: r.EventSHA256}
}

func differentialCases(r Report) map[string]*CaseEvidence {
	result := map[string]*CaseEvidence{}
	for i := range r.Cases {
		if r.Cases[i].Kind == "differential" {
			result[r.Cases[i].ID] = &r.Cases[i]
		}
	}
	return result
}

func localStdoutRelation(c *CaseEvidence) string {
	if c == nil || c.Outputs == nil || c.Outputs.PHP == nil || c.Outputs.Go == nil || c.Phase == nil || c.Diagnostics == nil || c.Diagnostics.Profile != QuietSuccessProfile || c.Diagnostics.Status != "pass" || c.Phase.QuietStderr != "pass" || c.Phase.GoExecute != "pass" || c.Phase.Exit != "pass" || c.Outputs.PHP.ExitCode != 0 || c.Outputs.Go.ExitCode != 0 {
		return "not-measured"
	}
	if c.Phase.Stdout == "pass" && sameStdout(c.Outputs.PHP, c.Outputs.Go) {
		return "match"
	}
	if c.Phase.Stdout == "fail" && !sameStdout(c.Outputs.PHP, c.Outputs.Go) {
		return "mismatch"
	}
	return "not-measured"
}

func OutputComparisonMarkdown(r OutputComparison) []byte {
	var b bytes.Buffer
	b.WriteString("# GoHP recorded output comparison\n\n")
	fmt.Fprintf(&b, "%s\n\n", viewText(r.Notice))
	for _, side := range []struct {
		name string
		run  ComparisonRun
	}{{"Left", r.Left}, {"Right", r.Right}} {
		m := side.run.Metadata
		fmt.Fprintf(&b, "%s: schema %s; overall %s; commit %s; target %s; PHP %s; baseline %s; events %s.\n\n", side.name, viewText(side.run.Schema), viewText(side.run.Overall), viewText(m.Commit), viewText(m.PHPTarget), viewText(m.PHPVersion), viewText(m.PHPBaselineSHA256), viewText(side.run.EventSHA256))
	}
	fmt.Fprintf(&b, "Context issues: %s.\n\nOracle metadata differences: %s. These do not establish causality.\n\n", viewText(joinOrNone(r.ContextIssues)), viewText(joinOrNone(r.OracleDifferences)))
	b.WriteString("| Case | Observation | Left PHP/Go stdout | Right PHP/Go stdout | Native stdout | Detail |\n|---|---|---|---|---|---|\n")
	for _, c := range r.Cases {
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s |\n", viewText(c.ID), viewText(c.Status), viewText(c.LeftMatch), viewText(c.RightMatch), viewText(c.NativeRelation), viewText(c.Detail))
	}
	return b.Bytes()
}

func joinOrNone(values []string) string {
	if len(values) == 0 {
		return "none recorded"
	}
	return strings.Join(values, "; ")
}
