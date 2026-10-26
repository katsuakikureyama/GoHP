// Package compat turns one bounded go test -json run into compatibility
// evidence. It deliberately does not infer PHP compatibility from a generated
// artifact alone.
package compat

// Metadata identifies the environment that produced one evidence run.
// Empty optional oracle/CI fields are preserved rather than invented.
type Metadata struct {
	Commit            string       `json:"commit"`
	Tree              string       `json:"tree"`
	Dirty             string       `json:"dirty"`
	GoVersion         string       `json:"go_version"`
	GOOS              string       `json:"goos"`
	GOARCH            string       `json:"goarch"`
	PHPTarget         string       `json:"php_target"`
	PHPVersion        string       `json:"php_version"`
	Image             string       `json:"php_image"`
	GoImage           string       `json:"go_image"`
	RunURL            string       `json:"run_url"`
	RunID             string       `json:"run_id"`
	Attempt           string       `json:"attempt"`
	Command           string       `json:"command"`
	StartedAt         string       `json:"started_at"`
	BuildFlags        string       `json:"build_flags"`
	Provenance        string       `json:"provenance"`
	SourceSHA256      string       `json:"source_sha256"`
	TestExit          int          `json:"test_exit"`
	PHPBaseline       *PHPBaseline `json:"php_baseline,omitempty"`
	PHPBaselineSHA256 string       `json:"php_baseline_sha256,omitempty"`
}

type Catalog struct {
	Schema   string  `json:"schema"`
	Maturity string  `json:"maturity"`
	Entries  []Entry `json:"entries"`
	SHA256   string  `json:"-"`
	root     string
	cases    map[string]Case
}

type Entry struct {
	ID             string   `json:"id"`
	Kind           string   `json:"kind"`
	Name           string   `json:"name"`
	Implementation string   `json:"implementation"`
	Contract       string   `json:"contract"`
	PHPTargets     []string `json:"php_targets"`
	Checks         []Check  `json:"checks"`
	Limitations    []string `json:"limitations"`
}

type Check struct {
	CaseID     string   `json:"case_id"`
	Test       string   `json:"test"`
	File       string   `json:"file"`
	Kind       string   `json:"kind"`
	Fixture    string   `json:"fixture,omitempty"`
	Assertions []string `json:"assertions"`
}

// Case is the normalized, reusable execution identity for a catalog check.
type Case struct {
	ID      string `json:"id"`
	Test    string `json:"test"`
	Package string `json:"package"`
	File    string `json:"file"`
	Kind    string `json:"kind"`
	Fixture string `json:"fixture,omitempty"`
}

type PhaseEvidence struct {
	QuietStderr  string `json:"quiet_stderr,omitempty"`
	Parse        string `json:"parse"`
	Codegen      string `json:"codegen"`
	GoBuild      string `json:"go_build"`
	GoExecute    string `json:"go_execute"`
	Differential string `json:"differential"`
	Stdout       string `json:"stdout"`
	Exit         string `json:"exit"`
	Stderr       string `json:"stderr"`
	Exceptions   string `json:"exceptions"`
	SideEffects  string `json:"side_effects"`
}

type CaseEvidence struct {
	Diagnostics     *DiagnosticEvidence `json:"diagnostics,omitempty"`
	ID              string              `json:"id"`
	Test            string              `json:"test"`
	Package         string              `json:"package"`
	Kind            string              `json:"kind"`
	Fixture         string              `json:"fixture,omitempty"`
	FixtureSHA256   string              `json:"fixture_sha256,omitempty"`
	GeneratedSHA256 string              `json:"generated_sha256,omitempty"`
	Status          string              `json:"status"`
	Phase           *PhaseEvidence      `json:"phase,omitempty"`
	Detail          string              `json:"detail,omitempty"`
}

type FeatureRow struct {
	ID             string   `json:"id"`
	Target         string   `json:"target"`
	Kind           string   `json:"kind"`
	Name           string   `json:"name"`
	Implementation string   `json:"implementation"`
	Contract       string   `json:"contract"`
	Limitations    []string `json:"limitations"`
	Checks         []string `json:"checks"`
	Status         string   `json:"status"`
}

type Report struct {
	Schema        string         `json:"schema"`
	Metadata      Metadata       `json:"metadata"`
	CatalogSHA256 string         `json:"catalog_sha256"`
	EventSHA256   string         `json:"event_sha256"`
	Overall       string         `json:"overall"`
	Cases         []CaseEvidence `json:"cases"`
	Features      []FeatureRow   `json:"features"`
}
