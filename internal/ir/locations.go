package ir

import (
	"errors"
	"fmt"
	"path"
	"strings"
	"unicode"
	"unicode/utf8"
)

// SourceFile describes original input bytes, not synthetic concatenated text.
// Manually supplied IR metadata is not proof that a source buffer was verified.
type SourceFile struct {
	ID     string `json:"source_id"`
	Bytes  int64  `json:"bytes"`
	SHA256 string `json:"sha256"`
}

type Location struct {
	SourceID string `json:"source_id"`
	Line     int    `json:"line"`
}

type SourceError struct {
	SourceID string
	Err      error
}

func (e *SourceError) Error() string { return e.SourceID + ": " + e.Err.Error() }
func (e *SourceError) Unwrap() error { return e.Err }

// WithSource preserves the deepest known source rather than relabeling an
// already-located dependency error as its caller's source.
func WithSource(err error, sourceID string) error {
	if err == nil || sourceID == "" {
		return err
	}
	var located *SourceError
	if errors.As(err, &located) {
		return err
	}
	return &SourceError{SourceID: sourceID, Err: err}
}

func ValidSourceID(id string) bool {
	if id == "" || !utf8.ValidString(id) || strings.HasPrefix(id, "/") || strings.ContainsAny(id, "\\:") || path.Clean(id) != id {
		return false
	}
	for _, r := range id {
		if r == 0 || unicode.IsControl(r) {
			return false
		}
	}
	for _, part := range strings.Split(id, "/") {
		if part == "" || part == "." || part == ".." {
			return false
		}
	}
	return true
}

// ValidateSources rejects ambiguous/mixed line-only and file-aware IR. Within
// a declaration or statement scope, children cannot claim another source file.
func ValidateSources(p Program) error {
	bad := func(message string) error { return fmt.Errorf("IR002: %s", message) }
	known := map[string]bool{}
	for _, source := range p.Sources {
		if !ValidSourceID(source.ID) || known[source.ID] || source.Bytes < 0 || len(source.SHA256) != 64 {
			return bad("invalid or duplicate source metadata")
		}
		for _, r := range source.SHA256 {
			if !(r >= '0' && r <= '9' || r >= 'a' && r <= 'f') {
				return bad("source hash must be canonical SHA256")
			}
		}
		known[source.ID] = true
	}
	resolve := func(id, parent string) (string, error) {
		if len(known) == 0 {
			if id != "" {
				return "", bad("source identity requires a source table")
			}
			return "", nil
		}
		if id == "" {
			id = parent
		}
		if !known[id] {
			return "", bad("missing or unlisted source identity")
		}
		if parent != "" && id != parent {
			return "", bad("nested source identity conflicts with containing source")
		}
		return id, nil
	}
	var statements func([]Stmt, string) error
	statements = func(body []Stmt, parent string) error {
		for _, statement := range body {
			id, err := resolve(statement.SourceID, parent)
			if err != nil {
				return err
			}
			if err := statements(statement.Body, id); err != nil {
				return err
			}
			if err := statements(statement.ElseBody, id); err != nil {
				return err
			}
		}
		return nil
	}
	function := func(f Function, parent string) error {
		id, err := resolve(f.SourceID, parent)
		if err != nil {
			return err
		}
		return statements(f.Body, id)
	}
	for _, f := range p.Functions {
		if err := function(f, ""); err != nil {
			return err
		}
	}
	for _, c := range p.Classes {
		id, err := resolve(c.SourceID, "")
		if err != nil {
			return err
		}
		for _, f := range c.Methods {
			if err := function(f, id); err != nil {
				return err
			}
		}
	}
	return statements(p.Body, "")
}
