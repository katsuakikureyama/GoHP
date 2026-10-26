package sourceinventory

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/corpus"
	"github.com/katsuakikureyama/GoHP/internal/phpversion"
)

func TestInventoryEnvironmentBinding(t *testing.T) {
	manifest, root := inventoryFixture(t, "first")
	raw, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatal(err)
	}
	var m corpus.Manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	second := []byte("second")
	if err := os.WriteFile(filepath.Join(root, "second.php"), second, 0600); err != nil {
		t.Fatal(err)
	}
	m.Files = append(m.Files, corpus.ManifestFile{Path: "second.php", Role: "source", Bytes: int64(len(second)), SHA256: fmt.Sprintf("%x", sha256.Sum256(second))})
	raw, err = json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifest, raw, 0600); err != nil {
		t.Fatal(err)
	}
	base := tokenization{PHPVersion: "8.3.36", InternalFunctions: []string{"strlen"}, Extensions: []string{"Core"}, ZendExtensions: []string{}}
	for _, tc := range []struct {
		name   string
		change func(*tokenization)
	}{
		{"same environment", func(*tokenization) {}},
		{"internal functions changed", func(r *tokenization) { r.InternalFunctions = []string{"strcmp", "strlen"} }},
		{"extensions changed", func(r *tokenization) { r.Extensions = []string{"Core", "tokenizer"} }},
		{"zend extensions changed", func(r *tokenization) { r.ZendExtensions = []string{"Zend OPcache"} }},
		{"version changed", func(r *tokenization) { r.PHPVersion = "8.3.37" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			scan := func(_ context.Context, source []byte) (tokenization, error) {
				r := base
				r.Tokens = []Token{{Kind: "T_INLINE_HTML", Bytes: len(source)}}
				if string(source) == "second" {
					tc.change(&r)
				}
				return r, nil
			}
			r, err := inventory(context.Background(), manifest, root, "8.3", scan)
			if tc.name == "same environment" {
				want := fmt.Sprintf("%x", sha256.Sum256([]byte(`[["strlen"],["Core"],[]]`)))
				if err != nil || r.NativeEnvironment.SHA256 != want || len(r.Files) != 2 {
					t.Fatalf("environment binding: %+v %v", r, err)
				}
			} else if err == nil || !reflect.DeepEqual(r, Report{}) {
				t.Fatal("environment/version drift exposed a successful or partial report")
			}
		})
	}
}

func TestNativeCandidateNameEvidence(t *testing.T) {
	target := os.Getenv("GOHP_PHP_TARGET")
	if target == "" {
		target = phpversion.DefaultTarget
	}
	// Namespaces, imports, attributes, conditional declarations and methods
	// deliberately prevent treating these matches as resolved global calls.
	source := `<?php
namespace App;
use function Vendor\work as imported;
function localName() {}
if (false) { function conditional() {} }
class C { function methodName() {} function strlen() {} }
#[strlen(1)] function decorated() {}
strlen('x'); STRLEN('x'); localName(); conditional(); methodName();
imported(); absent(); \strlen('x'); Vendor\strlen('x'); namespace\strlen('x');
$first = strlen(...);
`
	manifest, root := inventoryFixture(t, source)
	r, err := Inventory(context.Background(), manifest, root, target)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.NativeEnvironment.InternalFunctions) == 0 || len(r.NativeEnvironment.Extensions) == 0 || r.NativeEnvironment.ZendExtensions == nil || len(r.NativeEnvironment.SHA256) != 64 {
		t.Fatal("missing native environment evidence")
	}
	f := r.Files[0]
	if f.NameEvidenceStatus != "name-only-not-resolved" || f.ResolvedCalls != nil || f.SemanticUnits != nil || f.UnparsedBytes != len(source) || f.GrammarStatus != "not-validated" {
		t.Fatal("name matches became semantic evidence")
	}
	want := map[string]string{
		"strlen": "both-name-sources", "STRLEN": "both-name-sources",
		"localName": "selected-declaration-name", "conditional": "selected-declaration-name", "methodName": "selected-declaration-name",
		"imported": "unmatched-name", "absent": "unmatched-name",
		`\strlen`: "qualified-name-unresolved", `Vendor\strlen`: "qualified-name-unresolved", `namespace\strlen`: "qualified-name-unresolved",
	}
	if len(f.NameEvidence) != len(want) {
		t.Fatalf("evidence: %+v", f.NameEvidence)
	}
	for _, evidence := range f.NameEvidence {
		if evidence.Category != want[evidence.Name] {
			t.Fatalf("unexpected evidence: %+v", evidence)
		}
		if evidence.Name == "strlen" && evidence.Count != 3 {
			t.Fatal("attribute/first-class candidate disappeared")
		}
	}
}
