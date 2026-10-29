package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestArrayEmptyLoweringAndSourceMap(t *testing.T) {
	p, err := frontend.Parse("<?php\nfunction probe(array $a): bool { return EmPtY($a); }\necho probe([]);", testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	expr := p.Functions[0].Body[0].Expr
	if expr.Kind != "empty_value" || expr.Text != "" || len(expr.Args) != 1 || expr.Args[0].Kind != "var" {
		t.Fatalf("empty was not preserved as a variable-only construct: %#v", expr)
	}
	out, mapping, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	for line, text := range strings.Split(string(out), "\n") {
		if strings.Contains(text, `return rt.BoolReturn(rt.EmptyValue(v["a"]))`) {
			if got, ok := mapping.PHPLine(line + 1); !ok || got != 2 {
				t.Fatalf("empty return mapping = %d, %v; want 2", got, ok)
			}
			return
		}
	}
	t.Fatalf("missing guarded empty lowering:\n%s", out)
}

func TestArrayEmptyUnsupportedSource(t *testing.T) {
	for name, source := range map[string]string{
		"missing":       "echo empty();",
		"extra":         "$a = []; echo empty($a, $a);",
		"literal":       "echo empty([]);",
		"index":         "$a = [[]]; echo empty($a[0]);",
		"property":      "class X { public $a = null; } $x = new X(); echo empty($x->a);",
		"call":          "function source() { return []; } echo empty(source());",
		"parentheses":   "$a = []; echo empty(($a));",
		"undefined":     "echo empty($missing);",
		"conditional":   "if (true) { $a = []; } echo empty($a);",
		"null":          "$a = null; echo empty($a);",
		"bool":          "$a = false; echo empty($a);",
		"integer":       "$a = 0; echo empty($a);",
		"object":        "class X {} $a = new X(); echo empty($a);",
		"callable":      "$a = fn() => []; echo empty($a);",
		"redeclare":     "function EmPtY($a) { return $a; }",
		"bool_as_array": "$a = []; echo count(empty($a));",
		"dead_invalid":  "$a = null; echo true || empty($a);",
	} {
		t.Run(name, func(t *testing.T) {
			p, err := frontend.Parse("<?php "+source, testTarget(t))
			var out []byte
			if err == nil {
				out, err = Generate(p)
			}
			if err == nil || len(out) != 0 || !strings.Contains(err.Error(), "PHP00") {
				t.Fatalf("unsupported source generated output: %s; %v", out, err)
			}
		})
	}
}

func TestArrayEmptyMalformedIR(t *testing.T) {
	variable := ir.Expr{Kind: "var", Text: "a", Line: 2}
	for name, expr := range map[string]ir.Expr{
		"generic_call": {Kind: "call", Text: "EmPtY", Args: []ir.Expr{variable}},
		"missing":      {Kind: "empty_array"},
		"extra":        {Kind: "empty_array", Args: []ir.Expr{variable, variable}},
		"text":         {Kind: "empty_array", Text: "empty", Args: []ir.Expr{variable}},
		"params":       {Kind: "empty_array", Params: []string{}, Args: []ir.Expr{variable}},
		"literal":      {Kind: "empty_array", Args: []ir.Expr{{Kind: "array"}}},
		"var_operands": {Kind: "empty_array", Args: []ir.Expr{{Kind: "var", Text: "a", Args: []ir.Expr{variable}}}},
	} {
		t.Run(name, func(t *testing.T) {
			p := ir.Program{Version: ir.Version, PHPVersion: testTarget(t), Body: []ir.Stmt{
				{Kind: "assign", Name: "a", Line: 1, Expr: ir.Expr{Kind: "array"}},
				{Kind: "echo", Line: 2, Expr: expr},
			}}
			out, err := Generate(p)
			if err == nil || len(out) != 0 || !strings.Contains(err.Error(), "IR002") {
				t.Fatalf("malformed IR generated output: %s; %v", out, err)
			}
		})
	}
}
