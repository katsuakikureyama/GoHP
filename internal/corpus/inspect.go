package corpus

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/katsuakikureyama/GoHP/internal/phpversion"
)

const (
	maxManifestBytes = 1 << 20
	maxFileBytes     = 16 << 20
	maxTotalBytes    = 64 << 20
	maxFiles         = 64
)

var (
	idPattern   = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,95}$`)
	partPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)
	gitPattern  = regexp.MustCompile(`^[0-9a-f]{40}$`)
	hashPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

// Inspect validates the entire selection before returning a report. It reads
// source bytes only: no PHP parsing/execution, Git subprocess, network or writes.
// Use a private, stable directory; path checks are not a hostile-filesystem
// sandbox or a snapshot transaction against concurrent mutation.
func Inspect(manifestPath, sourceRoot string) (Report, error) {
	m, manifestHash, err := readManifest(manifestPath)
	if err != nil {
		return Report{}, err
	}
	root, err := filepath.Abs(sourceRoot)
	if err != nil {
		return Report{}, err
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return Report{}, err
	}
	info, err := os.Stat(root)
	if err != nil {
		return Report{}, err
	}
	if !info.IsDir() {
		return Report{}, fmt.Errorf("source root must be a directory")
	}
	files := make([]Observation, 0, len(m.Files))
	var measured Measurements
	for _, file := range m.Files {
		observed, err := inspectFile(root, file)
		if err != nil {
			return Report{}, fmt.Errorf("source %s: %w", file.Path, err)
		}
		files = append(files, observed)
		if file.Role == "subject" {
			measured.SubjectFiles++
			measured.SubjectBytes += observed.Bytes
		} else {
			measured.ProvenanceFiles++
			measured.ProvenanceBytes += observed.Bytes
		}
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return Report{
		Schema: ReportSchema, Status: "source-identity-match", Compatibility: "unverified",
		AnalysisStatus: "not-measured", ManifestSHA256: manifestHash, Manifest: m,
		Files: files, Measurements: measured,
		Notices: []string{
			"Only the listed whole-file paths and bytes were checked; unlisted files were not read.",
			"Repository, revision, project and version are manifest declarations, not authenticated origin or license approval.",
			"PHP targets are intended measurement scopes; no PHP oracle, compiler or application was executed.",
			"Functions, syntax, callsites, dependencies, compilation and execution remain unmeasured; no compatibility percentage is available.",
			"Provenance files are checked but do not increase the selected source denominator.",
		},
	}, nil
}

func readManifest(name string) (Manifest, string, error) {
	info, err := os.Lstat(name)
	if err != nil {
		return Manifest{}, "", err
	}
	if !info.Mode().IsRegular() || info.Size() > maxManifestBytes {
		return Manifest{}, "", fmt.Errorf("manifest must be a regular file no larger than 1 MiB")
	}
	f, err := os.Open(name)
	if err != nil {
		return Manifest{}, "", err
	}
	defer f.Close()
	info, err = f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return Manifest{}, "", fmt.Errorf("manifest must be a regular file")
	}
	b, err := io.ReadAll(io.LimitReader(f, maxManifestBytes+1))
	if err != nil {
		return Manifest{}, "", err
	}
	if len(b) > maxManifestBytes || !utf8.Valid(b) {
		return Manifest{}, "", fmt.Errorf("manifest must be UTF-8 JSON no larger than 1 MiB")
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	if err := checkJSON(d, 0); err != nil {
		return Manifest{}, "", fmt.Errorf("manifest JSON: %w", err)
	}
	if _, err := d.Token(); err != io.EOF {
		return Manifest{}, "", fmt.Errorf("manifest contains trailing JSON data")
	}
	d = json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	var m Manifest
	if err := d.Decode(&m); err != nil {
		return Manifest{}, "", fmt.Errorf("manifest: %w", err)
	}
	if err := validateManifest(m); err != nil {
		return Manifest{}, "", err
	}
	return m, fmt.Sprintf("%x", sha256.Sum256(b)), nil
}

// Disallow ambiguous duplicate keys and nulls before decoding typed fields.
// A null cannot stand in for a missing/zero observation in the manifest.
func checkJSON(d *json.Decoder, depth int) error {
	if depth > 64 {
		return fmt.Errorf("JSON exceeds 64 nesting levels")
	}
	t, err := d.Token()
	if err != nil {
		return err
	}
	if t == nil {
		return fmt.Errorf("null is not allowed in a manifest")
	}
	switch t {
	case json.Delim('{'):
		seen := make(map[string]bool)
		for d.More() {
			key, err := d.Token()
			if err != nil {
				return err
			}
			s, ok := key.(string)
			if !ok || seen[s] || strings.ToLower(s) != s {
				return fmt.Errorf("invalid, noncanonical or duplicate JSON object key")
			}
			seen[s] = true
			if err := checkJSON(d, depth+1); err != nil {
				return err
			}
		}
		end, err := d.Token()
		if err != nil || end != json.Delim('}') {
			return fmt.Errorf("invalid JSON object end")
		}
	case json.Delim('['):
		for d.More() {
			if err := checkJSON(d, depth+1); err != nil {
				return err
			}
		}
		end, err := d.Token()
		if err != nil || end != json.Delim(']') {
			return fmt.Errorf("invalid JSON array end")
		}
	}
	return nil
}

func validateManifest(m Manifest) error {
	if m.Schema != ManifestSchema || !idPattern.MatchString(m.ID) {
		return fmt.Errorf("invalid manifest schema or corpus id")
	}
	if !text(m.Project, 128) || !text(m.Version, 128) || !text(m.Selection.Description, 1024) {
		return fmt.Errorf("project, version and selection description must be bounded nonempty printable text")
	}
	u, err := url.Parse(m.Source.Repository)
	if err != nil || len(m.Source.Repository) > 1024 || u.Scheme != "https" || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Opaque != "" {
		return fmt.Errorf("source repository must be an HTTPS URL without credentials, query or fragment")
	}
	if !gitPattern.MatchString(m.Source.Revision) {
		return fmt.Errorf("source revision must be a full lowercase 40-hex Git commit")
	}
	if m.Selection.Unit != "whole-file" || m.Selection.DependencyClosure != "unresolved" {
		return fmt.Errorf("selection must be whole-file with unresolved dependency closure")
	}
	targets := make(map[string]bool)
	if len(m.PHPTargets) == 0 {
		return fmt.Errorf("manifest needs intended PHP targets")
	}
	for _, target := range m.PHPTargets {
		if err := phpversion.ValidateTarget(target); err != nil {
			return err
		}
		if targets[target] {
			return fmt.Errorf("duplicate PHP target")
		}
		targets[target] = true
	}
	if len(m.Files) == 0 || len(m.Files) > maxFiles {
		return fmt.Errorf("manifest must list 1 to 64 files")
	}
	paths := make(map[string]bool)
	var total int64
	subjects := 0
	for _, file := range m.Files {
		if !relative(file.Path) || paths[file.Path] {
			return fmt.Errorf("source paths must be unique canonical relative slash paths")
		}
		paths[file.Path] = true
		if file.Role != "subject" && file.Role != "provenance" {
			return fmt.Errorf("file role must be subject or provenance")
		}
		if file.Role == "subject" {
			subjects++
		}
		if file.Bytes <= 0 || file.Bytes > maxFileBytes || !hashPattern.MatchString(file.SHA256) {
			return fmt.Errorf("each file needs 1 byte to 16 MiB of declared bytes and a lowercase SHA-256")
		}
		total += file.Bytes
		if total > maxTotalBytes {
			return fmt.Errorf("selected files exceed the 64 MiB total limit")
		}
	}
	if subjects == 0 {
		return fmt.Errorf("manifest needs at least one subject file")
	}
	return nil
}

func text(s string, limit int) bool {
	return len(s) > 0 && len(s) <= limit && strings.TrimSpace(s) == s &&
		strings.IndexFunc(s, unicode.IsControl) < 0
}

func relative(s string) bool {
	if len(s) == 0 || len(s) > 512 || path.IsAbs(s) || path.Clean(s) != s {
		return false
	}
	for _, part := range strings.Split(s, "/") {
		if part == "." || part == ".." || !partPattern.MatchString(part) {
			return false
		}
	}
	return true
}

func inspectFile(root string, expected File) (Observation, error) {
	name := root
	parts := strings.Split(expected.Path, "/")
	for i, part := range parts {
		name = filepath.Join(name, part)
		info, err := os.Lstat(name)
		if err != nil {
			return Observation{}, err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return Observation{}, fmt.Errorf("symlink path components are not allowed")
		}
		if i < len(parts)-1 {
			if !info.IsDir() {
				return Observation{}, fmt.Errorf("parent must be a directory")
			}
		} else if !info.Mode().IsRegular() || info.Size() != expected.Bytes {
			return Observation{}, fmt.Errorf("expected regular file with %d bytes", expected.Bytes)
		}
	}
	f, err := os.Open(name)
	if err != nil {
		return Observation{}, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() != expected.Bytes {
		return Observation{}, fmt.Errorf("opened file does not match declared type/size")
	}
	h := sha256.New()
	n, err := io.Copy(h, io.LimitReader(f, expected.Bytes+1))
	if err != nil {
		return Observation{}, err
	}
	hash := fmt.Sprintf("%x", h.Sum(nil))
	if n != expected.Bytes || hash != expected.SHA256 {
		return Observation{}, fmt.Errorf("source size or SHA-256 mismatch")
	}
	return Observation{Path: expected.Path, Role: expected.Role, Bytes: n, SHA256: hash, Status: "match"}, nil
}
