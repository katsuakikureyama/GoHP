package compat

import "reflect"

// QuietSuccessProfile is deliberately narrower than diagnostic equivalence:
// positive fixtures must exit normally and emit no stderr on either side.
const QuietSuccessProfile = "quiet-success/v1"

type StderrObservation struct {
	Bytes  int    `json:"bytes"`
	SHA256 string `json:"sha256"`
}

type DiagnosticEvidence struct {
	Profile string             `json:"profile"`
	PHPArgs []string           `json:"php_args"`
	Status  string             `json:"status"`
	PHP     *StderrObservation `json:"php,omitempty"`
	Go      *StderrObservation `json:"go,omitempty"`
}

// QuietPHPArgs returns a fresh slice: callers append a file or -r program.
// These explicit per-process overrides are separate from the startup baseline.
func QuietPHPArgs() []string {
	return []string{"-d", "error_reporting=-1", "-d", "display_errors=stderr", "-d", "display_startup_errors=1", "-d", "log_errors=0", "-d", "html_errors=0"}
}

func diagnosticShape(d *DiagnosticEvidence, schema string) string {
	if schema != RunSchema && schema != RunSchemaQuiet {
		if d != nil {
			return "legacy payload cannot carry quiet stderr evidence"
		}
		return ""
	}
	if d == nil || d.Profile != QuietSuccessProfile || !reflect.DeepEqual(d.PHPArgs, QuietPHPArgs()) {
		return "differential payload requires the quiet-success fixture profile"
	}
	expected := "not-checked"
	if d.PHP != nil && d.Go != nil {
		expected = "pass"
	}
	for _, observation := range []*StderrObservation{d.PHP, d.Go} {
		if observation == nil {
			continue
		}
		if observation.Bytes < 0 || !validSHA256(observation.SHA256) || (observation.Bytes == 0) != (observation.SHA256 == hashBytes(nil)) {
			return "differential payload has invalid stderr length/hash evidence"
		}
		if observation.Bytes > 0 {
			expected = "fail"
		}
	}
	if d.Status != expected {
		return "differential payload contradicts quiet stderr observations"
	}
	return ""
}
