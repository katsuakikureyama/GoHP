package compat

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

type testEvent struct {
	Action  string `json:"Action"`
	Package string `json:"Package"`
	Test    string `json:"Test"`
	Output  string `json:"Output"`
}

type lifecycle struct {
	terminal string
	invalid  bool
	ran      bool
	count    int
}

type loggedPayload struct {
	casePayload
	packageName string
	test        string
}

type casePayload struct {
	Outputs         *ExecutionOutputs   `json:"outputs,omitempty"`
	Diagnostics     *DiagnosticEvidence `json:"diagnostics,omitempty"`
	CaseID          string              `json:"case_id"`
	PHPTarget       string              `json:"php_target"`
	PHPVersion      string              `json:"php_version"`
	FixtureSHA256   string              `json:"fixture_sha256"`
	GeneratedSHA256 string              `json:"generated_sha256"`
	Phases          struct {
		Parse        string `json:"parse"`
		Codegen      string `json:"codegen"`
		GoBuild      string `json:"go-build"`
		GoExecute    string `json:"go-execute"`
		Differential string `json:"differential"`
	} `json:"phases"`
	Comparisons struct {
		Stdout      string `json:"stdout"`
		Exit        string `json:"exit"`
		Stderr      string `json:"stderr"`
		Exceptions  string `json:"exceptions"`
		SideEffects string `json:"side_effects"`
	} `json:"comparisons"`
}

// Evaluate turns exactly one go test -json stream into a run-scoped report.
// It deliberately makes absent or contradictory evidence incomplete rather
// than promoting it to a compatibility pass.
func Evaluate(c Catalog, events io.Reader, meta Metadata) (Report, error) {
	return evaluate(c, events, meta, c.fixtureHash, RunSchema)
}

type fixtureLookup func(Case) (hash, detail string)

func evaluate(c Catalog, events io.Reader, meta Metadata, fixture fixtureLookup, schema string) (Report, error) {
	if err := validateBaseline(schema, meta); err != nil {
		return Report{}, err
	}
	if err := validateMetadata(meta); err != nil {
		return Report{}, err
	}
	raw, err := io.ReadAll(events)
	if err != nil {
		return Report{}, fmt.Errorf("read go test events: %w", err)
	}
	parsed, err := parseEvents(raw)
	if err != nil {
		return Report{}, err
	}
	lifecycles := make(map[string]*lifecycle)
	payloads := make(map[string][]loggedPayload)
	// test2json may split one log line across several output events. Keep
	// fragments isolated by package/test, including when parallel tests interleave.
	// Older archives retain their original event-level interpretation on replay.
	pending := make(map[string]string)
	logPayload := func(output, pkg, test string) {
		if payload, ok := payloadFromOutput(output); ok {
			payloads[payload.CaseID] = append(payloads[payload.CaseID], loggedPayload{casePayload: payload, packageName: pkg, test: test})
		}
	}
	for _, event := range parsed {
		packageKey := event.Package + "\x00"
		if lifecycles[packageKey] == nil {
			lifecycles[packageKey] = &lifecycle{}
		}
		key := event.Package + "\x00" + event.Test
		if event.Action == "start" || event.Action == "run" || event.Action == "pass" || event.Action == "fail" || event.Action == "skip" {
			state := lifecycles[key]
			if state == nil {
				state = &lifecycle{}
				lifecycles[key] = state
			}
			begin := event.Action == "run" || event.Action == "start"
			if begin {
				if state.ran || state.terminal != "" || (event.Action == "start") != (event.Test == "") {
					state.invalid = true
				}
				state.ran = true
			}
			if state.terminal != "" && state.terminal != event.Action {
				state.invalid = true
			}
			if !begin {
				state.count++
				if state.count > 1 || !state.ran {
					state.invalid = true
				}
				state.terminal = event.Action
			}
		}
		if event.Action == "output" {
			if schema != RunSchema {
				logPayload(event.Output, event.Package, event.Test)
				continue
			}
			output := pending[key] + event.Output
			for {
				i := strings.IndexByte(output, '\n')
				if i < 0 {
					break
				}
				if i > 4*1024*1024 {
					return Report{}, fmt.Errorf("go test output line exceeds 4 MiB")
				}
				logPayload(output[:i], event.Package, event.Test)
				output = output[i+1:]
			}
			if len(output) > 4*1024*1024 {
				return Report{}, fmt.Errorf("go test output line exceeds 4 MiB")
			}
			pending[key] = output
		} else if event.Action == "pass" || event.Action == "fail" || event.Action == "skip" {
			logPayload(pending[key], event.Package, event.Test)
			delete(pending, key)
		}
	}
	for key, output := range pending {
		pkg, test, _ := strings.Cut(key, "\x00")
		logPayload(output, pkg, test)
	}
	report := Report{
		Schema:        schema,
		Metadata:      meta,
		CatalogSHA256: c.SHA256,
		EventSHA256:   hashBytes(raw),
	}
	for _, testCase := range sortedCases(c) {
		report.Cases = append(report.Cases, evaluateCase(testCase, lifecycles, payloads, meta, fixture, schema))
	}
	if overall(nil, lifecycles, meta.TestExit) != "pass" {
		for i := range report.Cases {
			if report.Cases[i].Status == "pass" {
				report.Cases[i].Status = "incomplete"
				report.Cases[i].Detail = "test run has failed, skipped or incomplete execution"
			}
		}
	}
	report.Features = evaluateFeatures(c, report.Cases, meta)
	report.Overall = overall(report.Features, lifecycles, meta.TestExit)
	return report, nil
}

func validateMetadata(meta Metadata) error {
	// Replay retains old oracle versions as historical evidence. The live runner
	// separately applies the current minimum patch policy before executing tests.
	if !knownEvidenceTarget(meta.PHPTarget) {
		return fmt.Errorf("evidence has unsupported PHP target %q", meta.PHPTarget)
	}
	if meta.PHPVersion != "" && !strings.HasPrefix(meta.PHPVersion, meta.PHPTarget+".") {
		return fmt.Errorf("oracle PHP version does not match target")
	}
	for name, value := range map[string]string{
		"go_version": meta.GoVersion, "goos": meta.GOOS, "goarch": meta.GOARCH,
		"command": meta.Command, "started_at": meta.StartedAt, "dirty": meta.Dirty,
		"provenance": meta.Provenance,
	} {
		if blank(value) {
			return fmt.Errorf("evidence metadata requires %s", name)
		}
	}
	if !oneOf(meta.Provenance, "local-git", "ci-checkout-env", "unbound-checkout") {
		return fmt.Errorf("unknown evidence provenance %q", meta.Provenance)
	}
	return nil
}

func parseEvents(raw []byte) ([]testEvent, error) {
	var result []testEvent
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 4*1024*1024)
	line := 0
	for scanner.Scan() {
		line++
		if len(bytes.TrimSpace(scanner.Bytes())) == 0 {
			continue
		}
		var event testEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return nil, fmt.Errorf("invalid go test JSON event at line %d: %w", line, err)
		}
		if blank(event.Action) || blank(event.Package) {
			return nil, fmt.Errorf("incomplete go test event at line %d", line)
		}
		if !oneOf(event.Action, "start", "run", "pause", "cont", "pass", "fail", "skip", "output", "bench") {
			return nil, fmt.Errorf("unknown go test action at line %d", line)
		}
		result = append(result, event)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read go test events: %w", err)
	}
	return result, nil
}

func (c Catalog) fixtureHash(testCase Case) (string, string) {
	fixture, err := c.checkedPath(testCase.Fixture)
	if err != nil {
		return "", "fixture no longer resolves beneath repository root"
	}
	b, err := os.ReadFile(fixture)
	if err != nil {
		return "", "fixture is no longer readable"
	}
	return hashBytes(b), ""
}

func evaluateCase(testCase Case, states map[string]*lifecycle, payloads map[string][]loggedPayload, meta Metadata, fixture fixtureLookup, schema string) CaseEvidence {
	evidence := CaseEvidence{ID: testCase.ID, Test: testCase.Test, Package: testCase.Package, Kind: testCase.Kind, Fixture: testCase.Fixture}
	if testCase.Fixture != "" {
		evidence.FixtureSHA256, evidence.Detail = fixture(testCase)
		if evidence.Detail != "" {
			evidence.Status = "incomplete"
			return evidence
		}
	}
	var payload casePayload
	payloadDetail := ""
	if testCase.Kind == "differential" {
		payload, payloadDetail = identityPayload(testCase, evidence.FixtureSHA256, payloads[testCase.ID], meta, schema)
		if payloadDetail == "" {
			evidence.GeneratedSHA256 = payload.GeneratedSHA256
			evidence.Phase = payloadPhase(payload)
			evidence.Diagnostics = payload.Diagnostics
			evidence.Outputs = payload.Outputs
			if payload.Diagnostics != nil {
				evidence.Phase.QuietStderr = payload.Diagnostics.Status
			}
		}
	}
	status, detail := lifecycleStatus(testCase, states)
	if status != "pass" {
		evidence.Status, evidence.Detail = status, detail
		return evidence
	}
	if meta.TestExit != 0 {
		evidence.Status, evidence.Detail = "incomplete", "go test process exited non-zero"
		return evidence
	}
	if testCase.Kind != "differential" {
		evidence.Status = "pass"
		return evidence
	}
	if payloadDetail != "" {
		evidence.Status, evidence.Detail = "incomplete", payloadDetail
		return evidence
	}
	if detail := positiveDifferentialDetail(payload, schema); detail != "" {
		evidence.Detail = detail
		if strings.Contains(detail, "failed") {
			evidence.Status = "failed"
		} else {
			evidence.Status = "incomplete"
		}
		return evidence
	}
	evidence.Status = "pass"
	return evidence
}

func lifecycleStatus(testCase Case, states map[string]*lifecycle) (string, string) {
	packageState := states[testCase.Package+"\x00"]
	if packageState == nil || !packageState.ran || packageState.count != 1 || packageState.terminal == "" {
		return "incomplete", "package has no terminal event"
	}
	if packageState.invalid || packageState.terminal == "fail" {
		return "failed", "package failed or has contradictory terminal events"
	}
	if packageState.terminal != "pass" {
		return "incomplete", "package did not pass"
	}
	parts := strings.Split(testCase.Test, "/")
	for i := 1; i <= len(parts); i++ {
		name := strings.Join(parts[:i], "/")
		state := states[testCase.Package+"\x00"+name]
		if state == nil || !state.ran || state.count != 1 || state.terminal == "" {
			return "incomplete", "test or parent has no terminal event"
		}
		if state.invalid || state.terminal == "fail" {
			return "failed", "test or parent failed or has contradictory terminal events"
		}
		if state.terminal != "pass" {
			return "incomplete", "test or parent did not pass"
		}
	}
	return "pass", ""
}

func payloadFromOutput(output string) (casePayload, bool) {
	const marker = "GOHP_CASE "
	i := strings.Index(output, marker)
	if i < 0 {
		return casePayload{}, false
	}
	var payload casePayload
	if err := strictJSON([]byte(strings.TrimSpace(output[i+len(marker):])), &payload); err != nil {
		return casePayload{}, false
	}
	return payload, true
}

func identityPayload(testCase Case, fixtureHash string, values []loggedPayload, meta Metadata, schema string) (casePayload, string) {
	if len(values) != 1 {
		return casePayload{}, "differential evidence must contain exactly one GOHP_CASE payload"
	}
	value := values[0]
	p := value.casePayload
	if value.packageName != testCase.Package || value.test != testCase.Test || p.CaseID != testCase.ID || p.PHPTarget != meta.PHPTarget || meta.PHPVersion == "" || p.PHPVersion != meta.PHPVersion {
		return casePayload{}, "differential payload does not match case, target or PHP version"
	}
	if p.FixtureSHA256 != fixtureHash {
		return casePayload{}, "differential payload has an invalid fixture hash"
	}
	if detail := payloadShape(p, schema); detail != "" {
		return casePayload{}, detail
	}
	return p, ""
}

func payloadPhase(p casePayload) *PhaseEvidence {
	return &PhaseEvidence{Parse: p.Phases.Parse, Codegen: p.Phases.Codegen, GoBuild: p.Phases.GoBuild, GoExecute: p.Phases.GoExecute, Differential: p.Phases.Differential, Stdout: p.Comparisons.Stdout, Exit: p.Comparisons.Exit, Stderr: p.Comparisons.Stderr, Exceptions: p.Comparisons.Exceptions, SideEffects: p.Comparisons.SideEffects}
}

func payloadShape(p casePayload, schema string) string {
	if detail := diagnosticShape(p.Diagnostics, schema); detail != "" {
		return detail
	}
	if detail := outputShape(p, schema); detail != "" {
		return detail
	}
	if (p.GeneratedSHA256 != "" || p.Phases.Codegen == "pass") && !validSHA256(p.GeneratedSHA256) {
		return "differential payload has an invalid generated source hash"
	}
	for _, phase := range []string{p.Phases.Parse, p.Phases.Codegen, p.Phases.GoBuild, p.Phases.GoExecute, p.Phases.Differential} {
		if !oneOf(phase, "pass", "fail", "not-run") {
			return "differential payload has an invalid phase status"
		}
	}
	for _, comparison := range []string{p.Comparisons.Stdout, p.Comparisons.Exit} {
		if !oneOf(comparison, "pass", "fail", "not-checked") {
			return "differential payload has an invalid comparison status"
		}
	}
	for _, comparison := range []string{p.Comparisons.Stderr, p.Comparisons.Exceptions, p.Comparisons.SideEffects} {
		if comparison != "not-checked" {
			return "differential payload records unsupported comparison evidence"
		}
	}
	return ""
}

func positiveDifferentialDetail(p casePayload, schema string) string {
	if schema == RunSchema || schema == RunSchemaQuiet {
		if p.Diagnostics != nil && p.Diagnostics.Status == "fail" {
			return "quiet stderr requirement failed"
		}
		if p.Diagnostics == nil || p.Diagnostics.Status != "pass" {
			return "differential payload does not prove quiet stderr"
		}
	}
	for _, status := range []string{p.Phases.Parse, p.Phases.Codegen, p.Phases.GoBuild, p.Phases.GoExecute, p.Phases.Differential, p.Comparisons.Stdout, p.Comparisons.Exit} {
		if status == "fail" {
			return "differential execution or comparison failed"
		}
	}
	for _, phase := range []string{p.Phases.Parse, p.Phases.Codegen, p.Phases.GoBuild, p.Phases.GoExecute, p.Phases.Differential} {
		if phase != "pass" {
			return "differential payload does not prove every execution phase"
		}
	}
	if p.Comparisons.Stdout != "pass" || p.Comparisons.Exit != "pass" {
		return "differential payload does not prove stdout and exit comparisons"
	}
	return ""
}

func validSHA256(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func evaluateFeatures(c Catalog, cases []CaseEvidence, meta Metadata) []FeatureRow {
	byID := make(map[string]CaseEvidence, len(cases))
	for _, evidence := range cases {
		byID[evidence.ID] = evidence
	}
	var rows []FeatureRow
	for _, entry := range c.Entries {
		for _, target := range entry.PHPTargets {
			row := FeatureRow{ID: entry.ID, Target: target, Kind: entry.Kind, Name: entry.Name, Implementation: entry.Implementation, Contract: entry.Contract, Limitations: entry.Limitations}
			for _, check := range entry.Checks {
				row.Checks = append(row.Checks, check.CaseID)
			}
			sort.Strings(row.Checks)
			if target != meta.PHPTarget {
				row.Status = "not-measured"
			} else {
				row.Status = featureStatus(entry, row.Checks, byID)
			}
			rows = append(rows, row)
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].ID == rows[j].ID {
			return rows[i].Target < rows[j].Target
		}
		return rows[i].ID < rows[j].ID
	})
	return rows
}

func featureStatus(entry Entry, ids []string, cases map[string]CaseEvidence) string {
	allPass := true
	for _, id := range ids {
		status := cases[id].Status
		if status == "failed" {
			return "failed"
		}
		if status != "pass" {
			allPass = false
		}
	}
	if !allPass {
		return "incomplete"
	}
	if entry.Implementation == "unsupported" {
		return "unsupported"
	}
	if entry.Kind == "tooling" && len(ids) != 0 {
		return "compiler-check-pass"
	}
	for _, check := range entry.Checks {
		if check.Kind == "differential" {
			return "differential-pass (case-limited)"
		}
	}
	return "unverified"
}

func overall(features []FeatureRow, states map[string]*lifecycle, testExit int) string {
	status := "pass"
	if testExit != 0 || len(states) == 0 {
		status = "incomplete"
	}
	for key, state := range states {
		if state.invalid || state.terminal == "fail" {
			return "failed"
		}
		// Go emits a package-level skip for packages with no test files. Required
		// catalog cases in such a package still fail their own lifecycle check.
		noTests := strings.HasSuffix(key, "\x00") && state.terminal == "skip"
		if !state.ran || state.count != 1 || (state.terminal != "pass" && !noTests) {
			status = "incomplete"
		}
	}
	for _, feature := range features {
		if feature.Status == "failed" {
			return "failed"
		}
		if feature.Status == "incomplete" || feature.Status == "unverified" {
			status = "incomplete"
		}
	}
	return status
}
