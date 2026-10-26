package compat

import "fmt"

// ProcessOutput records observations, not raw stdout. ExitCode -1 means the
// process ended without a portable exit code (for example by signal). A nil
// process observation means execution was not observed, never successful zero.
type ProcessOutput struct {
	StdoutBytes  int    `json:"stdout_bytes"`
	StdoutSHA256 string `json:"stdout_sha256"`
	ExitCode     int    `json:"exit_code"`
}

func (p *ProcessOutput) UnmarshalJSON(data []byte) error {
	var wire struct {
		StdoutBytes  *int    `json:"stdout_bytes"`
		StdoutSHA256 *string `json:"stdout_sha256"`
		ExitCode     *int    `json:"exit_code"`
	}
	if err := strictJSON(data, &wire); err != nil {
		return err
	}
	if wire.StdoutBytes == nil || wire.StdoutSHA256 == nil || wire.ExitCode == nil {
		return fmt.Errorf("process output requires explicit byte count, hash and exit code")
	}
	p.StdoutBytes, p.StdoutSHA256, p.ExitCode = *wire.StdoutBytes, *wire.StdoutSHA256, *wire.ExitCode
	return nil
}

type ExecutionOutputs struct {
	PHP *ProcessOutput `json:"php,omitempty"`
	Go  *ProcessOutput `json:"go,omitempty"`
}

func sameStdout(a, b *ProcessOutput) bool {
	return a != nil && b != nil && a.StdoutBytes == b.StdoutBytes && a.StdoutSHA256 == b.StdoutSHA256
}

func outputShape(p casePayload, schema string) string {
	if schema != RunSchema && schema != RunSchemaOutputs {
		if p.Outputs != nil {
			return "legacy payload cannot carry process output evidence"
		}
		return ""
	}
	if p.Outputs == nil {
		return "differential payload requires process output evidence"
	}
	for _, output := range []*ProcessOutput{p.Outputs.PHP, p.Outputs.Go} {
		if output == nil {
			continue
		}
		if output.StdoutBytes < 0 || !validSHA256(output.StdoutSHA256) || (output.StdoutBytes == 0) != (output.StdoutSHA256 == hashBytes(nil)) || output.ExitCode < -1 {
			return "differential payload has invalid stdout or exit observation"
		}
	}
	if p.Diagnostics == nil || (p.Outputs.PHP == nil) != (p.Diagnostics.PHP == nil) || (p.Outputs.Go == nil) != (p.Diagnostics.Go == nil) {
		return "differential payload has inconsistent process observations"
	}
	both := p.Outputs.PHP != nil && p.Outputs.Go != nil
	if p.Comparisons.Stdout != "not-checked" {
		if !both || (p.Comparisons.Stdout == "pass") != sameStdout(p.Outputs.PHP, p.Outputs.Go) {
			return "differential stdout verdict contradicts output observations"
		}
	}
	if p.Comparisons.Exit == "pass" && (!both || p.Outputs.PHP.ExitCode != 0 || p.Outputs.Go.ExitCode != 0) {
		return "differential exit verdict contradicts process observations"
	}
	if p.Comparisons.Exit == "fail" && both && p.Outputs.PHP.ExitCode == 0 && p.Outputs.Go.ExitCode == 0 {
		return "differential exit failure contradicts successful observations"
	}
	return ""
}
