package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestNestedArrayProbesLoweringAndSourceMap(t *testing.T) {
	const source = `<?php
function dynamic_value($value) { return $value; }
function fallback($name, $value) { echo $name; return $value; }
$root = ['first' => ['second' => 'value']];
$key = 'second';
$unknown = dynamic_value($root);
$one_level = isset($root['first']);
$nested = isset($root['first'][$key]);
$multi = isset($root['first'][$key], $root['first']['missing']);
$known = $root['first'][$key] ?? fallback('known-fallback', 'wrong');
$missing = $root['first']['missing'] ?? fallback('missing-fallback', 'fallback');
$unknown_value = $unknown['first'][$key] ?? false;
$chain = $root['first']['missing'] ?? $root['first'][$key] ?? false;
$ordinary = $root['first']['second'];
`
	p, err := frontend.Parse(source, "8.3")
	if err != nil {
		t.Fatal(err)
	}
	path := p.Body[4].Expr
	if path.Kind != "isset_path" || path.Text != "" || len(path.Args) != 3 || path.Args[0].Kind != "var" || path.Args[1].Kind != "string" || path.Args[2].Kind != "var" {
		t.Fatalf("nested isset IR = %#v, want flat root/key/key isset_path", path)
	}
	multi := p.Body[5].Expr
	if multi.Kind != "isset_multi" || len(multi.Args) != 2 || multi.Args[0].Kind != "isset_path" || multi.Args[1].Kind != "isset_path" {
		t.Fatalf("multi nested isset IR = %#v, want two isset_path operands", multi)
	}
	chain := p.Body[9].Expr
	if chain.Kind != "coalesce" || len(chain.Args) != 2 || chain.Args[1].Kind != "coalesce" {
		t.Fatalf("coalesce chain IR = %#v, want right-associative coalesce", chain)
	}

	generated, sourceMap, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	text := string(generated)
	for _, want := range []string{
		`rt.ArrayIsset(v["root"], "first")`,
		`rt.ArrayPathIsset(v["root"], "first", v["key"])`,
		`rt.ArrayPathIsset(v["root"], "first", "missing")`,
		`value, present := rt.ArrayPathProbe(v["root"], "first", v["key"])`,
		`value, present := rt.ArrayPathProbe(v["unknown"], "first", v["key"])`,
		`return F_fallback("missing-fallback", "fallback")`,
		`rt.Get(rt.Get(v["root"], "first"), "second")`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("nested probe lowering missing %q:\n%s", want, generated)
		}
	}
	if strings.Contains(text, `rt.Get(v["root"], "first", v["key"])`) {
		t.Fatalf("nested paths must not lower to an invalid eager multi-key Get:\n%s", generated)
	}
	knownProbe := strings.Index(text, `rt.ArrayPathProbe(v["root"], "first", v["key"])`)
	knownFallback := strings.Index(text, `return F_fallback("known-fallback", "wrong")`)
	if knownProbe < 0 || knownFallback < knownProbe {
		t.Fatalf("coalesce must probe once before its lazy fallback:\n%s", generated)
	}
	knownStart := strings.Index(text, `v["known"] =`)
	if knownStart < 0 {
		t.Fatalf("missing generated known coalesce assignment:\n%s", generated)
	}
	knownEnd := strings.Index(text[knownStart:], "// gohp:end")
	if knownEnd < 0 || strings.Count(text[knownStart:knownStart+knownEnd], "rt.ArrayPathProbe(") != 1 || !strings.Contains(text[knownStart:knownStart+knownEnd], "if present {") {
		t.Fatalf("coalesce must emit one probe and a present-only fallback boundary:\n%s", generated)
	}
	for line, generatedLine := range strings.Split(text, "\n") {
		if strings.Contains(generatedLine, `v["nested"] = rt.Clone(rt.ArrayPathIsset(v["root"], "first", v["key"]))`) {
			if got, ok := sourceMap.PHPLine(line + 1); !ok || got != 8 {
				t.Fatalf("nested isset maps to PHP line %d, %v; want 8", got, ok)
			}
			return
		}
	}
	t.Fatalf("nested isset was not source-mapped:\n%s", generated)
}

func TestNestedArrayProbesBoundaries(t *testing.T) {
	for name, source := range map[string]string{
		"coalesce_literal_left":       "echo []['a'] ?? false;",
		"coalesce_undefined_root":     "echo $root['a'] ?? false;",
		"coalesce_undefined_key":      "$root = ['a' => []]; echo $root['a'][$key] ?? false;",
		"coalesce_property_left":      "class Box { public $value = null; } $box = new Box(); echo $box->value ?? false;",
		"coalesce_call_left":          "function root() { return []; } echo root()['a'] ?? false;",
		"coalesce_effectful_key":      "function key() { return 'a'; } $root = ['a' => []]; echo $root[key()]['b'] ?? false;",
		"coalesce_scalar_root":        "$root = 'text'; echo $root['a']['b'] ?? false;",
		"coalesce_scalar_key":         "$root = ['a' => []]; $key = false; echo $root['a'][$key] ?? false;",
		"nested_isset_undefined_root": "echo isset($root['a']['b']);",
		"nested_isset_property":       "class Box { public $value = null; } $box = new Box(); echo isset($box->value['a']);",
		"nested_isset_effectful_key":  "function key() { return 'a'; } $root = ['a' => []]; echo isset($root[key()]['b']);",
	} {
		t.Run(name, func(t *testing.T) {
			p, err := frontend.Parse("<?php "+source, "8.3")
			if err == nil {
				_, err = Generate(p)
			}
			if err == nil || (!strings.Contains(err.Error(), "PHP00") && !strings.Contains(err.Error(), "IR002")) {
				t.Fatalf("unsupported nested array probe was accepted: %v", err)
			}
		})
	}

	root := ir.Expr{Kind: "var", Text: "root", Line: 1}
	key := ir.Expr{Kind: "string", Text: "key", Line: 1}
	index := ir.Expr{Kind: "index", Line: 1, Args: []ir.Expr{root, key}}
	for name, expr := range map[string]ir.Expr{
		"isset_path_too_short":   {Kind: "isset_path", Line: 1, Args: []ir.Expr{root, key}},
		"isset_path_text":        {Kind: "isset_path", Text: "isset", Line: 1, Args: []ir.Expr{root, key, key}},
		"isset_path_nested_root": {Kind: "isset_path", Line: 1, Args: []ir.Expr{index, key, key}},
		"coalesce_bad_left":      {Kind: "coalesce", Line: 1, Args: []ir.Expr{ir.Expr{Kind: "array", Line: 1}, ir.Expr{Kind: "bool", Text: "false", Line: 1}}},
		"coalesce_bad_arity":     {Kind: "coalesce", Line: 1, Args: []ir.Expr{index}},
		"coalesce_text":          {Kind: "coalesce", Text: "??", Line: 1, Args: []ir.Expr{index, ir.Expr{Kind: "bool", Text: "false", Line: 1}}},
	} {
		t.Run(name, func(t *testing.T) {
			p := ir.Program{Version: ir.Version, PHPVersion: "8.3", Body: []ir.Stmt{
				{Kind: "assign", Name: "root", Line: 1, Expr: ir.Expr{Kind: "array", Line: 1}},
				{Kind: "echo", Line: 2, Expr: expr},
			}}
			out, err := Generate(p)
			if err == nil || len(out) != 0 || (!strings.Contains(err.Error(), "PHP00") && !strings.Contains(err.Error(), "IR002")) {
				t.Fatalf("malformed nested probe IR generated output %s: %v", out, err)
			}
		})
	}
}
