package sourceinventory

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/corpus"
	"github.com/katsuakikureyama/GoHP/internal/phpversion"
)

func inventoryFixture(t *testing.T, source string) (string, string) {
	t.Helper()
	root := t.TempDir()
	m := corpus.Manifest{Schema: "gohp.corpus.manifest/v0.1", ID: "synthetic", Upstream: corpus.Upstream{Repository: "example/project", Revision: strings.Repeat("a", 40), Version: "test"}, Selection: corpus.Selection{Unit: "whole-file", Exclusions: "Other files excluded", Dependencies: "unresolved"}, PHPTargets: []string{"8.3", "8.4"}}
	for _, file := range []struct{ path, role, text string }{{"source.php", "source", source}, {"NOTICE", "notice", "notice"}} {
		if err := os.WriteFile(filepath.Join(root, file.path), []byte(file.text), 0600); err != nil {
			t.Fatal(err)
		}
		m.Files = append(m.Files, corpus.ManifestFile{Path: file.path, Role: file.role, Bytes: int64(len(file.text)), SHA256: fmt.Sprintf("%x", sha256.Sum256([]byte(file.text)))})
	}
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(t.TempDir(), "manifest.json")
	if err := os.WriteFile(manifest, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return manifest, root
}

func TestInventorySelectionAndFailure(t *testing.T) {
	manifest, root := inventoryFixture(t, "plain text")
	calls := 0
	scan := func(ctx context.Context, source []byte) (tokenization, error) {
		calls++
		if string(source) != "plain text" {
			t.Fatal("scan did not receive checked source bytes")
		}
		return tokenization{PHPVersion: "8.3.36", Tokens: []Token{{Kind: "T_INLINE_HTML", Bytes: len(source)}}}, nil
	}
	r, err := inventory(context.Background(), manifest, root, "8.3", scan)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 || r.Status != "lexical-only" || r.Compatibility != "unverified" || r.Verification.Measurements.FunctionUnits != nil || len(r.Files) != 1 {
		t.Fatalf("unexpected inventory: %+v", r)
	}
	for _, target := range []string{"8.4", "8.2"} {
		r, err := inventory(context.Background(), manifest, root, target, scan)
		if err == nil || !reflect.DeepEqual(r, Report{}) {
			t.Fatalf("target %s accepted wrong native version or partial report", target)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "NOTICE"), []byte("changed notice"), 0600); err != nil {
		t.Fatal(err)
	}
	before := calls
	r, err = inventory(context.Background(), manifest, root, "8.3", scan)
	if err == nil || calls != before || !reflect.DeepEqual(r, Report{}) {
		t.Fatal("late notice failure started tokenizer or returned partial inventory")
	}
}

// Native PHP is mandatory in the configured CI matrix. This executes only the
// trusted tokenizer, not the deliberately invalid, side-effectful source data.
func TestNativeLexicalInventory(t *testing.T) {
	target := os.Getenv("GOHP_PHP_TARGET")
	if target == "" {
		target = phpversion.DefaultTarget
	}
	sentinel := filepath.Join(t.TempDir(), "SOURCE_EXECUTED")
	source := "<?php\nfile_put_contents('" + sentinel + "', 'executed');\n" + `
function Decl($x) { return strlen($x); }
function &refDecl() { return $x; }
$ref = function & () {};
$cb = fn($x) => $x;
$o->method(); Type::staticMethod(); new ClassName(); $cb();
$first = strlen(...);
if (true) { other(); }
($cb)();
$msg = "ghost()"; /* function ghost(){} */
$broken = ;
`
	manifest, root := inventoryFixture(t, source)
	r, err := Inventory(context.Background(), manifest, root, target)
	if err != nil {
		t.Fatal(err)
	}
	if err := phpversion.ValidateOracle(target, r.PHPVersion); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(sentinel); !os.IsNotExist(err) {
		t.Fatalf("input source may have executed: %v", err)
	}
	f := r.Files[0]
	if f.Bytes != len(source) || f.UnparsedBytes != len(source) || f.GrammarStatus != "not-validated" || len(f.Declarations) != 4 || f.SemanticUnits != nil || f.ResolvedCalls != nil || f.UnclassifiedParentheses < 2 {
		t.Fatalf("wrong lexical boundaries: %+v", f)
	}
	if f.Declarations[1].Kind != "named-function-marker" || f.Declarations[1].Name != "refDecl" || f.Declarations[2].Kind != "anonymous-function-marker" {
		t.Fatal("native by-reference markers misclassified")
	}
	names := map[string]int{}
	for _, name := range f.CandidateNames {
		names[name.Name] = name.Count
	}
	want := map[string]int{"file_put_contents": 1, "strlen": 2, "other": 1}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("name candidates=%v want %v (not resolved calls)", names, want)
	}
	var bytes, tokens int
	for _, kind := range f.TokenKinds {
		bytes += kind.Bytes
		tokens += kind.Count
	}
	if bytes != len(source) || tokens != f.Tokens {
		t.Fatal("not every original byte/token was counted")
	}
}
