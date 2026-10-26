// Package corpus checks pinned source bytes without parsing or executing them.
package corpus

const ManifestSchema = "gohp.source-corpus/v0.1"
const ReportSchema = "gohp.source-corpus.inspection/v0.1"

type Origin struct {
	Repository string `json:"repository"`
	Revision   string `json:"revision"`
}

type Selection struct {
	Unit              string `json:"unit"`
	Description       string `json:"description"`
	DependencyClosure string `json:"dependency_closure"`
}

type File struct {
	Path   string `json:"path"`
	Role   string `json:"role"`
	Bytes  int64  `json:"bytes"`
	SHA256 string `json:"sha256"`
}

// Manifest describes a selected set of whole files, not all repository files.
// Source/version strings are declarations: Inspect does not authenticate them.
type Manifest struct {
	Schema     string    `json:"schema"`
	ID         string    `json:"id"`
	Project    string    `json:"project"`
	Version    string    `json:"version"`
	Source     Origin    `json:"source"`
	Selection  Selection `json:"selection"`
	PHPTargets []string  `json:"php_targets"`
	Files      []File    `json:"files"`
}

type Observation struct {
	Path   string `json:"path"`
	Role   string `json:"role"`
	Bytes  int64  `json:"bytes"`
	SHA256 string `json:"sha256"`
	Status string `json:"status"`
}

type Measurements struct {
	SubjectFiles    int   `json:"subject_files"`
	SubjectBytes    int64 `json:"subject_bytes"`
	ProvenanceFiles int   `json:"provenance_files"`
	ProvenanceBytes int64 `json:"provenance_bytes"`
	FunctionUnits   *int  `json:"function_units"`
	SyntaxNodes     *int  `json:"syntax_nodes"`
	CallSites       *int  `json:"call_sites"`
	CompiledUnits   *int  `json:"compiled_units"`
	ExecutedCases   *int  `json:"executed_cases"`
}

type Report struct {
	Schema         string        `json:"schema"`
	Status         string        `json:"status"`
	Compatibility  string        `json:"compatibility"`
	AnalysisStatus string        `json:"analysis_status"`
	ManifestSHA256 string        `json:"manifest_sha256"`
	Manifest       Manifest      `json:"manifest"`
	Files          []Observation `json:"files"`
	Measurements   Measurements  `json:"measurements"`
	Notices        []string      `json:"notices"`
}
