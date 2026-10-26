// Package corpus verifies a selected, read-only application source corpus.
// It deliberately does not fetch upstreams, inspect dependencies, parse PHP or
// execute any source code.
package corpus

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	manifestSchema           = "gohp.corpus.manifest/v0.1"
	verificationSchema       = "gohp.corpus.verification/v0.1"
	maxManifestBytes   int64 = 1 << 20
	maxFileBytes       int64 = 16 << 20
	maxFiles                 = 64
	maxTotalBytes      int64 = 64 << 20
	maxJSONDepth             = 64
)

// Manifest is a fixed selection declaration. Metadata is not used to infer
// paths, download sources, or establish upstream authenticity.
type Manifest struct {
	Schema     string         `json:"schema"`
	ID         string         `json:"id"`
	Upstream   Upstream       `json:"upstream"`
	Selection  Selection      `json:"selection"`
	PHPTargets []string       `json:"php_targets"`
	Files      []ManifestFile `json:"files"`
}

type Upstream struct {
	Repository string `json:"repository"`
	Revision   string `json:"revision"`
	Version    string `json:"version"`
}

type Selection struct {
	Unit         string `json:"unit"`
	Exclusions   string `json:"exclusions"`
	Dependencies string `json:"dependencies"`
}

type ManifestFile struct {
	Path   string `json:"path"`
	Role   string `json:"role"`
	Bytes  int64  `json:"bytes"`
	SHA256 string `json:"sha256"`
}

// Report is emitted only after every selected file has been bound to its
// declared bytes and SHA-256 digest.
type Report struct {
	Schema          string              `json:"schema"`
	Status          string              `json:"status"`
	ManifestSHA256  string              `json:"manifest_sha256"`
	Manifest        Manifest            `json:"manifest"`
	SourceFileCount int                 `json:"source_file_count"`
	SourceBytes     int64               `json:"source_bytes"`
	Compatibility   string              `json:"compatibility"`
	Measurements    Measurements        `json:"measurements"`
	Targets         []TargetMeasurement `json:"targets"`
	Notice          string              `json:"notice"`
}

type Measurements struct {
	FunctionUnits  *float64 `json:"function_units"`
	Callsites      *float64 `json:"callsites"`
	CompileRatio   *float64 `json:"compile_ratio"`
	ExecutionRatio *float64 `json:"execution_ratio"`
}

type TargetMeasurement struct {
	PHPTarget string `json:"php_target"`
	Compile   string `json:"compile"`
	Execution string `json:"execution"`
}

// Verify reads a small strict JSON manifest and checks its listed whole files
// beneath sourceRoot. The caller is responsible for supplying a private root
// that is not mutated concurrently; Verify performs no writes.
func Verify(manifestPath, sourceRoot string) (Report, error) {
	report, _, err := load(manifestPath, sourceRoot, false)
	return report, err
}

// Source contains the exact bytes already checked against a manifest entry.
// Only source-role files are retained; notices/provenance are still verified.
type Source struct {
	Path  string
	Bytes []byte
}

// LoadSources returns all checked source buffers only after the entire manifest
// selection passes. Consumers must use these buffers, not reopen source paths.
func LoadSources(manifestPath, sourceRoot string) (Report, []Source, error) {
	return load(manifestPath, sourceRoot, true)
}

func load(manifestPath, sourceRoot string, retain bool) (Report, []Source, error) {
	manifestBytes, err := readRegular(manifestPath, maxManifestBytes)
	if err != nil {
		return Report{}, nil, fmt.Errorf("manifest: %w", err)
	}
	manifest, err := parseManifest(manifestBytes)
	if err != nil {
		return Report{}, nil, fmt.Errorf("manifest: %w", err)
	}
	root, err := resolvedDirectory(sourceRoot)
	if err != nil {
		return Report{}, nil, fmt.Errorf("source root: %w", err)
	}

	var sourceBytes int64
	var sourceCount int
	var sources []Source
	for _, entry := range manifest.Files {
		resolved, err := resolveSelected(root, entry.Path)
		if err != nil {
			return Report{}, nil, fmt.Errorf("%s: %w", entry.Path, err)
		}
		info, err := os.Stat(resolved)
		if err != nil {
			return Report{}, nil, fmt.Errorf("%s: %w", entry.Path, err)
		}
		if !info.Mode().IsRegular() {
			return Report{}, nil, fmt.Errorf("%s: selected file is not regular", entry.Path)
		}
		if info.Size() != entry.Bytes {
			return Report{}, nil, fmt.Errorf("%s: size mismatch: got %d, want %d", entry.Path, info.Size(), entry.Bytes)
		}
		contents, err := readRegular(resolved, maxFileBytes)
		if err != nil {
			return Report{}, nil, fmt.Errorf("%s: %w", entry.Path, err)
		}
		digest := sha256.Sum256(contents)
		if int64(len(contents)) != entry.Bytes || hex.EncodeToString(digest[:]) != entry.SHA256 {
			return Report{}, nil, fmt.Errorf("%s: size or sha256 mismatch", entry.Path)
		}
		if entry.Role == "source" {
			sourceCount++
			sourceBytes += entry.Bytes
			if retain {
				sources = append(sources, Source{Path: entry.Path, Bytes: contents})
			}
		}
	}
	manifestDigest := sha256.Sum256(manifestBytes)
	targets := make([]TargetMeasurement, len(manifest.PHPTargets))
	for i, target := range manifest.PHPTargets {
		targets[i] = TargetMeasurement{PHPTarget: target, Compile: "not-measured", Execution: "not-measured"}
	}
	return Report{
		Schema: verificationSchema, Status: "source-verified",
		ManifestSHA256: hex.EncodeToString(manifestDigest[:]), Manifest: manifest,
		SourceFileCount: sourceCount, SourceBytes: sourceBytes,
		Compatibility: "unverified", Measurements: Measurements{}, Targets: targets,
		Notice: "Source-byte equality only: this does not verify upstream authenticity, dependency closure, parser coverage, or PHP compatibility.",
	}, sources, nil
}

func readRegular(filename string, limit int64) ([]byte, error) {
	info, err := os.Lstat(filename)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("must be a regular file")
	}
	if info.Size() > limit {
		return nil, fmt.Errorf("exceeds %d byte limit", limit)
	}
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > limit {
		return nil, fmt.Errorf("exceeds %d byte limit", limit)
	}
	return b, nil
}

func resolvedDirectory(root string) (string, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("must be a directory")
	}
	return resolved, nil
}

func resolveSelected(root, selected string) (string, error) {
	if !canonicalPath(selected) {
		return "", fmt.Errorf("path must be canonical relative slash path")
	}
	joined := filepath.Join(root, filepath.FromSlash(selected))
	resolved, err := filepath.EvalSymlinks(joined)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(root, resolved)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", fmt.Errorf("path escapes source root")
	}
	return resolved, nil
}

func canonicalPath(value string) bool {
	platformPath := filepath.FromSlash(value)
	if value == "" || strings.Contains(value, "\\") || strings.HasPrefix(value, "/") || filepath.IsAbs(platformPath) || filepath.VolumeName(platformPath) != "" || path.Clean(value) != value {
		return false
	}
	for _, r := range value {
		if r == 0 || unicode.IsControl(r) {
			return false
		}
	}
	for _, part := range strings.Split(value, "/") {
		if part == "" || part == "." || part == ".." {
			return false
		}
	}
	return true
}

func parseManifest(raw []byte) (Manifest, error) {
	value, err := strictJSON(raw)
	if err != nil {
		return Manifest{}, err
	}
	obj, err := object(value, "root", "schema", "id", "upstream", "selection", "php_targets", "files")
	if err != nil {
		return Manifest{}, err
	}
	manifest := Manifest{}
	if manifest.Schema, err = requiredString(obj, "schema"); err != nil {
		return Manifest{}, err
	}
	if manifest.Schema != manifestSchema {
		return Manifest{}, fmt.Errorf("schema must be %q", manifestSchema)
	}
	if manifest.ID, err = requiredString(obj, "id"); err != nil {
		return Manifest{}, err
	}
	if strings.TrimSpace(manifest.ID) == "" {
		return Manifest{}, fmt.Errorf("id must be nonempty")
	}
	if manifest.Upstream, err = parseUpstream(obj["upstream"]); err != nil {
		return Manifest{}, err
	}
	if manifest.Selection, err = parseSelection(obj["selection"]); err != nil {
		return Manifest{}, err
	}
	if manifest.PHPTargets, err = requiredStrings(obj, "php_targets"); err != nil {
		return Manifest{}, err
	}
	if len(manifest.PHPTargets) == 0 {
		return Manifest{}, fmt.Errorf("php_targets must be nonempty")
	}
	seenTargets := map[string]bool{}
	for _, target := range manifest.PHPTargets {
		if target != "8.3" && target != "8.4" {
			return Manifest{}, fmt.Errorf("php_targets contains disabled target %q", target)
		}
		if seenTargets[target] {
			return Manifest{}, fmt.Errorf("php_targets contains duplicate target")
		}
		seenTargets[target] = true
	}
	files, ok := obj["files"].([]any)
	if !ok || len(files) == 0 || len(files) > maxFiles {
		return Manifest{}, fmt.Errorf("files must contain 1 to %d entries", maxFiles)
	}
	seenPaths := map[string]bool{}
	var total int64
	var sources, notices int
	for i, value := range files {
		file, err := parseFile(value)
		if err != nil {
			return Manifest{}, fmt.Errorf("files[%d]: %w", i, err)
		}
		if seenPaths[file.Path] {
			return Manifest{}, fmt.Errorf("files contains duplicate paths")
		}
		seenPaths[file.Path] = true
		total += file.Bytes
		if total > maxTotalBytes {
			return Manifest{}, fmt.Errorf("files exceed %d byte aggregate limit", maxTotalBytes)
		}
		switch file.Role {
		case "source":
			sources++
		case "notice":
			notices++
		}
		manifest.Files = append(manifest.Files, file)
	}
	if sources == 0 {
		return Manifest{}, fmt.Errorf("files must include a source")
	}
	if notices == 0 {
		return Manifest{}, fmt.Errorf("files must include a notice")
	}
	return manifest, nil
}

func parseUpstream(value any) (Upstream, error) {
	obj, err := object(value, "upstream", "repository", "revision", "version")
	if err != nil {
		return Upstream{}, err
	}
	u := Upstream{}
	if u.Repository, err = requiredString(obj, "repository"); err != nil {
		return u, err
	}
	if u.Revision, err = requiredString(obj, "revision"); err != nil {
		return u, err
	}
	if u.Version, err = requiredString(obj, "version"); err != nil {
		return u, err
	}
	if strings.TrimSpace(u.Repository) == "" || strings.TrimSpace(u.Version) == "" {
		return u, fmt.Errorf("upstream fields must be nonempty")
	}
	if !lowerHex(u.Revision, 40) {
		return u, fmt.Errorf("upstream revision must be 40 lowercase hexadecimal characters")
	}
	return u, nil
}

func parseSelection(value any) (Selection, error) {
	obj, err := object(value, "selection", "unit", "exclusions", "dependencies")
	if err != nil {
		return Selection{}, err
	}
	s := Selection{}
	if s.Unit, err = requiredString(obj, "unit"); err != nil {
		return s, err
	}
	if s.Dependencies, err = requiredString(obj, "dependencies"); err != nil {
		return s, err
	}
	if s.Exclusions, err = requiredString(obj, "exclusions"); err != nil {
		return s, err
	}
	if s.Unit != "whole-file" || s.Dependencies != "unresolved" || strings.TrimSpace(s.Exclusions) == "" {
		return s, fmt.Errorf("selection must use whole-file, nonempty exclusions, and unresolved dependencies")
	}
	return s, nil
}

func parseFile(value any) (ManifestFile, error) {
	obj, err := object(value, "file", "path", "role", "bytes", "sha256")
	if err != nil {
		return ManifestFile{}, err
	}
	f := ManifestFile{}
	if f.Path, err = requiredString(obj, "path"); err != nil {
		return f, err
	}
	if !canonicalPath(f.Path) {
		return f, fmt.Errorf("path must be canonical relative slash path")
	}
	if f.Role, err = requiredString(obj, "role"); err != nil {
		return f, err
	}
	if f.Role != "source" && f.Role != "notice" && f.Role != "provenance" {
		return f, fmt.Errorf("role must be source, notice, or provenance")
	}
	if f.Bytes, err = positiveInt64(obj["bytes"], "bytes"); err != nil {
		return f, err
	}
	if f.Bytes > maxFileBytes {
		return f, fmt.Errorf("bytes exceed %d byte limit", maxFileBytes)
	}
	if f.SHA256, err = requiredString(obj, "sha256"); err != nil {
		return f, err
	}
	if !lowerHex(f.SHA256, 64) {
		return f, fmt.Errorf("sha256 must be 64 lowercase hexadecimal characters")
	}
	return f, nil
}

func object(value any, label string, allowed ...string) (map[string]any, error) {
	obj, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s must be an object", label)
	}
	set := map[string]bool{}
	for _, key := range allowed {
		set[key] = true
	}
	if len(obj) != len(set) {
		return nil, fmt.Errorf("%s has missing or unknown fields", label)
	}
	for key := range obj {
		if !set[key] {
			return nil, fmt.Errorf("%s contains unknown field %q", label, key)
		}
	}
	for key := range set {
		if _, ok := obj[key]; !ok {
			return nil, fmt.Errorf("%s is missing field %q", label, key)
		}
	}
	return obj, nil
}

func requiredString(obj map[string]any, key string) (string, error) {
	value, ok := obj[key].(string)
	if !ok {
		return "", fmt.Errorf("%s must be a string", key)
	}
	return value, nil
}

func requiredStrings(obj map[string]any, key string) ([]string, error) {
	values, ok := obj[key].([]any)
	if !ok {
		return nil, fmt.Errorf("%s must be an array of strings", key)
	}
	result := make([]string, len(values))
	for i, value := range values {
		var ok bool
		result[i], ok = value.(string)
		if !ok {
			return nil, fmt.Errorf("%s must be an array of strings", key)
		}
	}
	return result, nil
}

func positiveInt64(value any, key string) (int64, error) {
	number, ok := value.(json.Number)
	if !ok {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}
	n, err := number.Int64()
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}
	return n, nil
}

func lowerHex(value string, length int) bool {
	if len(value) != length {
		return false
	}
	for _, c := range value {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}

func strictJSON(raw []byte) (any, error) {
	if !utf8.Valid(raw) {
		return nil, fmt.Errorf("JSON must be valid UTF-8")
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	value, err := strictValue(decoder, 0)
	if err != nil {
		return nil, err
	}
	if _, err := decoder.Token(); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("trailing JSON value")
		}
		return nil, fmt.Errorf("trailing data: %w", err)
	}
	return value, nil
}

func strictValue(decoder *json.Decoder, depth int) (any, error) {
	if depth > maxJSONDepth {
		return nil, fmt.Errorf("JSON exceeds %d nested levels", maxJSONDepth)
	}
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	switch token := token.(type) {
	case json.Delim:
		switch token {
		case '{':
			result := map[string]any{}
			for decoder.More() {
				keyToken, err := decoder.Token()
				if err != nil {
					return nil, err
				}
				key, ok := keyToken.(string)
				if !ok {
					return nil, fmt.Errorf("object key is not a string")
				}
				if _, exists := result[key]; exists {
					return nil, fmt.Errorf("duplicate JSON object key %q", key)
				}
				value, err := strictValue(decoder, depth+1)
				if err != nil {
					return nil, err
				}
				result[key] = value
			}
			end, err := decoder.Token()
			if err != nil || end != json.Delim('}') {
				return nil, fmt.Errorf("unterminated object")
			}
			return result, nil
		case '[':
			var result []any
			for decoder.More() {
				value, err := strictValue(decoder, depth+1)
				if err != nil {
					return nil, err
				}
				result = append(result, value)
			}
			end, err := decoder.Token()
			if err != nil || end != json.Delim(']') {
				return nil, fmt.Errorf("unterminated array")
			}
			return result, nil
		default:
			return nil, fmt.Errorf("unexpected JSON delimiter")
		}
	default:
		return token, nil
	}
}
