package codegen

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestSourceAwareLocationsAndMap(t *testing.T) {
	helper := parseIdentifiedSource(t, "lib/helper.php", `<?php
function helper() {
  return 7;
}
`)
	box := parseIdentifiedSource(t, "lib/box.php", `<?php
class Box {
  public $value = 9;
  public function get() {
    return $this->value;
  }
}
`)
	main := parseIdentifiedSource(t, `app/main "quoted".php`, `<?php
$result = helper();
echo $result;
$box = new Box();
echo $box->get();
`)

	if got := helper.Functions[0].SourceID; got != "lib/helper.php" {
		t.Fatalf("function source ID = %q, want lib/helper.php", got)
	}
	if got := box.Classes[0].SourceID; got != "lib/box.php" || box.Classes[0].Line != 2 {
		t.Fatalf("class location = %#v, want lib/box.php:2", box.Classes[0])
	}
	if got := main.Body[0].SourceID; got != `app/main "quoted".php` {
		t.Fatalf("top-level statement source ID = %q", got)
	}

	combined := ir.Program{
		Version:    ir.Version,
		PHPVersion: testTarget(t),
		Sources:    append(append(append([]ir.SourceFile{}, helper.Sources...), box.Sources...), main.Sources...),
		Functions:  helper.Functions,
		Classes:    box.Classes,
		Body:       main.Body,
	}
	generated, sourceMap, err := GenerateWithMap(combined)
	if err != nil {
		t.Fatal(err)
	}
	if sourceMap.Schema != "gohp.sourcemap/v0.2" || sourceMap.Precision != "statement-start-file-line" {
		t.Fatalf("source-aware map header = %#v", sourceMap)
	}
	if len(sourceMap.Sources) != 3 || sourceMap.Sources[0] != helper.Sources[0] || sourceMap.Sources[1] != box.Sources[0] || sourceMap.Sources[2] != main.Sources[0] {
		t.Fatalf("source-aware map sources = %#v", sourceMap.Sources)
	}
	if strings.Contains(string(generated), `app/main "quoted".php`) || strings.Contains(string(generated), "lib/helper.php") {
		t.Fatalf("source IDs must not be interpolated into generated marker comments:\n%s", generated)
	}
	if _, ok := sourceMap.PHPLine(1); ok {
		t.Fatal("v0.2 PHPLine must fail closed because it would discard the source ID")
	}

	// helper's return and main's first echo both originate at numeric PHP line 3.
	// The file identity must remain distinguishable in the v0.2 map.
	locations := map[string]ir.Location{
		`return rt.Clone(int64(7))`: {SourceID: "lib/helper.php", Line: 3},
		`rt.Echo(v["result"])`:      {SourceID: `app/main "quoted".php`, Line: 3},
		`return rt.Clone((v["this"]).(interface{ Get(string) any }).Get("value"))`: {SourceID: "lib/box.php", Line: 5},
	}
	for fragment, want := range locations {
		got, ok := sourceLocationForFragment(generated, sourceMap, fragment)
		if !ok || got != want {
			t.Errorf("%q location = %#v, %v; want %#v", fragment, got, ok, want)
		}
	}
	if got, ok := sourceLocationForFragment(generated, sourceMap, `rt.Echo(v["result"])`); !ok || got.Line != helper.Functions[0].Body[0].Line || got.SourceID == helper.Functions[0].SourceID {
		t.Fatalf("same numeric PHP line must retain distinct source identity: %#v, %v", got, ok)
	}
	for line, text := range strings.Split(string(generated), "\n") {
		if strings.TrimSpace(text) == "" || strings.HasPrefix(strings.TrimSpace(text), "//") {
			if _, ok := sourceMap.PHPLocation(line + 1); ok {
				t.Fatalf("scaffolding/comment line %d was source-mapped", line+1)
			}
		}
	}
}

func TestSourceAwareDiagnostics(t *testing.T) {
	t.Run("semantic dependency", func(t *testing.T) {
		dependency := parseIdentifiedSource(t, "lib/dependency.php", "<?php\nfunction accept(Missing $value) {}\n")
		p := parseIdentifiedSource(t, "app/entry.php", "<?php\necho 1;\n")
		p.Sources = append(p.Sources, dependency.Sources...)
		p.Functions = dependency.Functions
		out, sourceMap, err := GenerateWithMap(p)
		if err == nil || out != nil || sourceMap.Schema != "" || !strings.Contains(err.Error(), "lib/dependency.php") || !strings.Contains(err.Error(), "line 2") {
			t.Fatalf("semantic diagnostic/artifacts = %q, %#v, %v", out, sourceMap, err)
		}
	})
	t.Run("codegen undefined variable", func(t *testing.T) {
		dependency := parseIdentifiedSource(t, "lib/undefined.php", "<?php\nfunction unresolved(){return $missing;}\n")
		p := parseIdentifiedSource(t, "app/entry.php", "<?php\necho 1;\n")
		p.Sources = append(p.Sources, dependency.Sources...)
		p.Functions = dependency.Functions
		out, sourceMap, err := GenerateWithMap(p)
		if err == nil || out != nil || sourceMap.Schema != "" || !strings.Contains(err.Error(), "lib/undefined.php") || !strings.Contains(err.Error(), "line 2") {
			t.Fatalf("codegen diagnostic/artifacts = %q, %#v, %v", out, sourceMap, err)
		}
	})
}

func TestSourceAwareParseFailures(t *testing.T) {
	for _, source := range []string{"<?php\necho \"unterminated;", "<?php\nrequire_once __DIR__ . '/dependency.php';"} {
		p, err := frontend.ParseSource("lib/broken.php", source, testTarget(t), frontend.Options{})
		if err == nil || !strings.Contains(err.Error(), "lib/broken.php") || !strings.Contains(err.Error(), "line 2") || p.Version != "" || p.Sources != nil || p.Body != nil {
			t.Fatalf("parse failure leaked program/location: %+v %v", p, err)
		}
	}
	if p, err := frontend.ParseSource("../outside.php", "<?php echo 1;", testTarget(t), frontend.Options{}); err == nil || p.Version != "" {
		t.Fatal("invalid input source ID accepted")
	}
}

func TestLegacySourceMapRemainsV1(t *testing.T) {
	const source = "<?php\necho 1;\n"
	p, err := frontend.Parse(source, testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	generated, sourceMap, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	again, err := Generate(p)
	if err != nil || string(generated) != string(again) {
		t.Fatalf("legacy Generate bytes changed: %v\n%s\n%s", err, generated, again)
	}
	if sourceMap.Schema != "gohp.sourcemap/v0.1" || sourceMap.Sources != nil {
		t.Fatalf("legacy map = %#v", sourceMap)
	}
	encoded, err := json.Marshal(sourceMap)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "sources") || strings.Contains(string(encoded), "source_id") {
		t.Fatalf("legacy map JSON unexpectedly carries v0.2 identity fields: %s", encoded)
	}
	for line, text := range strings.Split(string(generated), "\n") {
		if strings.Contains(text, "rt.Echo(int64(1))") {
			if got, ok := sourceMap.PHPLine(line + 1); !ok || got != 2 {
				t.Fatalf("legacy PHPLine = %d, %v; want 2", got, ok)
			}
			return
		}
	}
	t.Fatal("legacy generated echo missing")
}

func TestSourceMetadataRejectsMalformedPrograms(t *testing.T) {
	valid := ir.SourceFile{ID: "lib/one.php", Bytes: 5, SHA256: strings.Repeat("a", 64)}
	base := ir.Program{Version: ir.Version, PHPVersion: testTarget(t)}
	cases := map[string]ir.Program{
		"empty ID":       withSources(base, []ir.SourceFile{{Bytes: 5, SHA256: valid.SHA256}}),
		"absolute ID":    withSources(base, []ir.SourceFile{{ID: "/lib/one.php", Bytes: 5, SHA256: valid.SHA256}}),
		"escaping ID":    withSources(base, []ir.SourceFile{{ID: "../one.php", Bytes: 5, SHA256: valid.SHA256}}),
		"control ID":     withSources(base, []ir.SourceFile{{ID: "lib/\none.php", Bytes: 5, SHA256: valid.SHA256}}),
		"duplicate ID":   withSources(base, []ir.SourceFile{valid, valid}),
		"bad hash":       withSources(base, []ir.SourceFile{{ID: valid.ID, Bytes: valid.Bytes, SHA256: "not-a-sha256"}}),
		"negative bytes": withSources(base, []ir.SourceFile{{ID: valid.ID, Bytes: -1, SHA256: valid.SHA256}}),
	}
	for name, p := range cases {
		t.Run(name, func(t *testing.T) { assertNoSourceAwareArtifacts(t, p) })
	}

	t.Run("unlisted top-level statement", func(t *testing.T) {
		p := withSources(base, []ir.SourceFile{valid})
		p.Body = []ir.Stmt{{Kind: "echo", Line: 2, SourceID: "missing.php", Expr: ir.Expr{Kind: "int", Text: "1", Line: 2}}}
		assertNoSourceAwareArtifacts(t, p)
	})
	t.Run("metadata on legacy program", func(t *testing.T) {
		p := base
		p.Body = []ir.Stmt{{Kind: "echo", Line: 2, SourceID: valid.ID, Expr: ir.Expr{Kind: "int", Text: "1", Line: 2}}}
		assertNoSourceAwareArtifacts(t, p)
	})
	t.Run("missing declaration identity", func(t *testing.T) {
		p := parseIdentifiedSource(t, "lib/one.php", "<?php function one(){return 1;}")
		p.Functions[0].SourceID = ""
		assertNoSourceAwareArtifacts(t, p)
	})
	t.Run("conflicting nested source ID", func(t *testing.T) {
		one := parseIdentifiedSource(t, "lib/one.php", "<?php\nfunction one() { return 1; }\n")
		two := parseIdentifiedSource(t, "lib/two.php", "<?php\necho 2;\n")
		one.Sources = append(one.Sources, two.Sources[0])
		one.Functions[0].Body[0].SourceID = "lib/two.php"
		assertNoSourceAwareArtifacts(t, one)
	})
}

func parseIdentifiedSource(t *testing.T, id, source string) ir.Program {
	t.Helper()
	p, err := frontend.ParseSource(id, source, testTarget(t), frontend.Options{})
	if err != nil {
		t.Fatalf("ParseSource(%q): %v", id, err)
	}
	if len(p.Sources) != 1 {
		t.Fatalf("ParseSource(%q) sources = %#v", id, p.Sources)
	}
	sum := sha256.Sum256([]byte(source))
	if got, want := p.Sources[0], (ir.SourceFile{ID: id, Bytes: int64(len(source)), SHA256: hex.EncodeToString(sum[:])}); got != want {
		t.Fatalf("ParseSource(%q) metadata = %#v, want %#v", id, got, want)
	}
	return p
}

func sourceLocationForFragment(generated []byte, sourceMap SourceMap, fragment string) (ir.Location, bool) {
	for line, text := range strings.Split(string(generated), "\n") {
		if strings.Contains(text, fragment) {
			return sourceMap.PHPLocation(line + 1)
		}
	}
	return ir.Location{}, false
}

func withSources(p ir.Program, sources []ir.SourceFile) ir.Program {
	p.Sources = sources
	return p
}

func assertNoSourceAwareArtifacts(t *testing.T, p ir.Program) {
	t.Helper()
	out, sourceMap, err := GenerateWithMap(p)
	if err == nil || out != nil || sourceMap.Schema != "" || sourceMap.Mappings != nil || sourceMap.Sources != nil {
		t.Fatalf("malformed source metadata generated artifacts: %q, %#v, %v", out, sourceMap, err)
	}
}
