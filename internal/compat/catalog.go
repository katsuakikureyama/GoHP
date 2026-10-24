package compat

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const modulePath = "github.com/katsuakikureyama/GoHP"

// LoadCatalog strictly reads compatibility/catalog.json beneath root and
// validates every check against the checked-out test and fixture files.
func LoadCatalog(root string) (Catalog, error) {
	root, err := filepath.EvalSymlinks(root)
	if err != nil {
		return Catalog{}, fmt.Errorf("resolve catalog root: %w", err)
	}
	probe := Catalog{root: root}
	catalogFile, err := probe.checkedPath(filepath.Join("compatibility", "catalog.json"))
	if err != nil {
		return Catalog{}, fmt.Errorf("resolve compatibility catalog: %w", err)
	}
	b, err := os.ReadFile(catalogFile)
	if err != nil {
		return Catalog{}, fmt.Errorf("read compatibility catalog: %w", err)
	}
	var c Catalog
	if err := strictJSON(b, &c); err != nil {
		return Catalog{}, fmt.Errorf("decode compatibility catalog: %w", err)
	}
	c.SHA256 = hashBytes(b)
	c.root = root
	if err := c.validate(); err != nil {
		return Catalog{}, err
	}
	return c, nil
}

func strictJSON(b []byte, dst any) error {
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err := d.Decode(dst); err != nil {
		return err
	}
	var extra any
	if err := d.Decode(&extra); err == nil {
		return fmt.Errorf("unexpected JSON value after document")
	} else if err != io.EOF {
		return err
	}
	return nil
}

func (c *Catalog) validate() error {
	return c.validateCatalog(true)
}

// Archive validation checks identities and contracts without opening source
// paths from a historical catalog in the consumer's current checkout.
func (c *Catalog) validateCatalog(sourceAvailable bool) error {
	if c.Schema != "gohp.compatibility.catalog/v0.1" {
		return fmt.Errorf("unsupported catalog schema %q", c.Schema)
	}
	if c.Maturity != "draft" && c.Maturity != "experimental" {
		return fmt.Errorf("unsupported catalog maturity %q", c.Maturity)
	}
	if len(c.Entries) == 0 {
		return fmt.Errorf("catalog has no entries")
	}
	c.cases = make(map[string]Case)
	entryIDs := make(map[string]struct{})
	for i := range c.Entries {
		e := &c.Entries[i]
		if blank(e.ID) || blank(e.Name) || blank(e.Contract) {
			return fmt.Errorf("entry %d requires id, name and non-empty contract", i)
		}
		if _, ok := entryIDs[e.ID]; ok {
			return fmt.Errorf("duplicate entry id %q", e.ID)
		}
		entryIDs[e.ID] = struct{}{}
		if !oneOf(e.Kind, "syntax", "operator", "type", "builtin", "extension", "tooling", "application") {
			return fmt.Errorf("entry %q has unsupported kind %q", e.ID, e.Kind)
		}
		if !oneOf(e.Implementation, "implemented", "partial", "unsupported") {
			return fmt.Errorf("entry %q has unsupported implementation %q", e.ID, e.Implementation)
		}
		if err := targets(e.PHPTargets); err != nil {
			return fmt.Errorf("entry %q: %w", e.ID, err)
		}
		for j := range e.Checks {
			if err := c.validateCheck(e.ID, &e.Checks[j], sourceAvailable); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Catalog) validateCheck(entryID string, check *Check, sourceAvailable bool) error {
	if blank(check.CaseID) || blank(check.Test) || blank(check.File) || len(check.Assertions) == 0 {
		return fmt.Errorf("entry %q has an incomplete check", entryID)
	}
	if !oneOf(check.Kind, "differential", "negative", "unit", "compiler") {
		return fmt.Errorf("entry %q check %q has unsupported kind %q", entryID, check.CaseID, check.Kind)
	}
	for _, assertion := range check.Assertions {
		if blank(assertion) {
			return fmt.Errorf("entry %q check %q has an empty assertion", entryID, check.CaseID)
		}
	}
	file, err := relativePath(check.File)
	if err != nil {
		return fmt.Errorf("entry %q check %q file: %w", entryID, check.CaseID, err)
	}
	if !strings.HasSuffix(file, "_test.go") {
		return fmt.Errorf("entry %q check %q file is not Go source", entryID, check.CaseID)
	}
	if check.Kind == "differential" && blank(check.Fixture) {
		return fmt.Errorf("entry %q check %q requires a fixture", entryID, check.CaseID)
	}
	if !blank(check.Fixture) {
		if _, err := relativePath(check.Fixture); err != nil {
			return fmt.Errorf("entry %q check %q fixture: %w", entryID, check.CaseID, err)
		}
	}
	if sourceAvailable {
		file, err = c.checkedPath(check.File)
		if err != nil {
			return fmt.Errorf("entry %q check %q file: %w", entryID, check.CaseID, err)
		}
		if !blank(check.Fixture) {
			if _, err := c.checkedPath(check.Fixture); err != nil {
				return fmt.Errorf("entry %q check %q fixture: %w", entryID, check.CaseID, err)
			}
		}
		if err := hasTopLevelTest(file, topLevelTest(check.Test)); err != nil {
			return fmt.Errorf("entry %q check %q: %w", entryID, check.CaseID, err)
		}
	}
	dir := filepath.ToSlash(filepath.Dir(check.File))
	pkg := modulePath
	if dir != "." {
		pkg += "/" + dir
	}
	identity := Case{ID: check.CaseID, Test: check.Test, Package: pkg, File: check.File, Kind: check.Kind, Fixture: check.Fixture}
	if old, exists := c.cases[check.CaseID]; exists {
		if old.Test != identity.Test || old.File != identity.File || old.Kind != identity.Kind || old.Fixture != identity.Fixture {
			return fmt.Errorf("case %q is reused with a different test, file, kind or fixture", check.CaseID)
		}
	} else {
		c.cases[check.CaseID] = identity
	}
	return nil
}

func relativePath(name string) (string, error) {
	if filepath.IsAbs(name) || name == "" || filepath.Clean(name) == "." {
		return "", fmt.Errorf("must be a non-empty relative file")
	}
	clean := filepath.Clean(name)
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path traversal is not allowed")
	}
	return clean, nil
}

func (c Catalog) checkedPath(name string) (string, error) {
	clean, err := relativePath(name)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(filepath.Join(c.root, clean))
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(c.root, resolved)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("symlink escapes repository root")
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("must be a regular file")
	}
	return resolved, nil
}

func hasTopLevelTest(filename, name string) error {
	if !strings.HasPrefix(name, "Test") || name == "Test" {
		return fmt.Errorf("invalid Go test name %q", name)
	}
	f, err := parser.ParseFile(token.NewFileSet(), filename, nil, 0)
	if err != nil {
		return err
	}
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if ok && fn.Recv == nil && fn.Name.Name == name && isTestingTSignature(fn) {
			return nil
		}
	}
	return fmt.Errorf("top-level test %q with func(*testing.T) signature not found", name)
}

func isTestingTSignature(fn *ast.FuncDecl) bool {
	if fn.Type.Results != nil && len(fn.Type.Results.List) != 0 {
		return false
	}
	if fn.Type.Params == nil || len(fn.Type.Params.List) != 1 {
		return false
	}
	star, ok := fn.Type.Params.List[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	selector, ok := star.X.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "T" {
		return false
	}
	ident, ok := selector.X.(*ast.Ident)
	return ok && ident.Name == "testing"
}

func topLevelTest(test string) string { return strings.SplitN(test, "/", 2)[0] }
func blank(s string) bool             { return strings.TrimSpace(s) == "" }
func oneOf(value string, values ...string) bool {
	for _, candidate := range values {
		if value == candidate {
			return true
		}
	}
	return false
}
func targets(values []string) error {
	if len(values) == 0 {
		return fmt.Errorf("requires php_targets")
	}
	seen := make(map[string]bool)
	for _, value := range values {
		if !oneOf(value, "7.4", "8.3") || seen[value] {
			return fmt.Errorf("invalid or duplicate PHP target %q", value)
		}
		seen[value] = true
	}
	return nil
}
func hashBytes(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func sortedCases(c Catalog) []Case {
	result := make([]Case, 0, len(c.cases))
	for _, testCase := range c.cases {
		result = append(result, testCase)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}
