// Package sourceinventory inventories native PHP lexical observations. It does
// not resolve calls, validate PHP grammar or establish GoHP compatibility.
package sourceinventory

import "github.com/katsuakikureyama/GoHP/internal/corpus"

const Schema = "gohp.corpus.lexical-inventory/v0.2"

// Token is a native tokenizer's byte-length observation, never source code to
// execute. Punctuation uses CHAR_ plus two uppercase hexadecimal digits.
type Token struct {
	Kind  string `json:"kind"`
	Bytes int    `json:"bytes"`
}

type Frequency struct {
	Kind  string `json:"kind"`
	Count int    `json:"count"`
	Bytes int    `json:"bytes"`
}

type Marker struct {
	Kind       string `json:"kind"`
	Name       string `json:"name,omitempty"`
	NameBase64 string `json:"name_base64,omitempty"`
	StartByte  int    `json:"start_byte"`
	EndByte    int    `json:"end_byte"`
	Line       int    `json:"line"`
}

type NameFrequency struct {
	Name       string `json:"name,omitempty"`
	NameBase64 string `json:"name_base64"`
	Count      int    `json:"count"`
}

type FileInventory struct {
	Path                    string          `json:"path"`
	SHA256                  string          `json:"sha256"`
	Bytes                   int             `json:"bytes"`
	Tokens                  int             `json:"tokens"`
	GrammarStatus           string          `json:"grammar_status"`
	UnparsedBytes           int             `json:"unparsed_bytes"`
	TokenKinds              []Frequency     `json:"token_kinds"`
	Declarations            []Marker        `json:"declaration_markers"`
	Parentheses             []Marker        `json:"open_parentheses"`
	CandidateNames          []NameFrequency `json:"name_parenthesis_frequencies"`
	NameEvidence            []NameEvidence  `json:"candidate_name_evidence"`
	NameEvidenceStatus      string          `json:"name_evidence_status"`
	UnclassifiedParentheses int             `json:"unclassified_parentheses"`
	SemanticUnits           *int            `json:"semantic_units"`
	ResolvedCalls           *int            `json:"resolved_calls"`
}

type Report struct {
	Schema            string            `json:"schema"`
	Status            string            `json:"status"`
	PHPTarget         string            `json:"php_target"`
	PHPVersion        string            `json:"php_version"`
	TokenizerSHA256   string            `json:"tokenizer_sha256"`
	NativeEnvironment NativeEnvironment `json:"native_environment"`
	Verification      corpus.Report     `json:"verification"`
	Files             []FileInventory   `json:"files"`
	Compatibility     string            `json:"compatibility"`
	Notice            string            `json:"notice"`
}
