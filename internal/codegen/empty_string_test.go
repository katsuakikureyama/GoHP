package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestEmptyStringLowering(t *testing.T) {
	p, err := frontend.ParseSource("lib/html.php", "<?php\nfunction probe(string $html): bool { return EmPtY($html); }\n", testTarget(t), frontend.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got := p.Functions[0].Body[0].Expr; got.Kind != "empty_value" || got.Text != "" || len(got.Args) != 1 || got.Args[0].Kind != "var" {
		t.Fatalf("empty IR: %+v", got)
	}
	out, m, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(out), "NewScalarFrame") || strings.Contains(string(out), "ValueSlot") {
		t.Fatal("empty_value scopes must remain legacy")
	}
	for i, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, `return rt.BoolReturn(rt.EmptyValue(v["html"]))`) {
			if got, ok := m.PHPLocation(i + 1); !ok || got.SourceID != "lib/html.php" || got.Line != 2 {
				t.Fatalf("file map: %+v %v", got, ok)
			}
			return
		}
	}
	t.Fatalf("missing string empty lowering: %s", out)
}

func TestEmptyStringMalformedIR(t *testing.T) {
	root := ir.Expr{Kind: "var", Text: "html", Line: 2}
	for _, expression := range []ir.Expr{
		{Kind: "empty_value"},
		{Kind: "empty_value", Args: []ir.Expr{root, root}},
		{Kind: "empty_value", Text: "empty", Args: []ir.Expr{root}},
		{Kind: "empty_value", Params: []string{}, Args: []ir.Expr{root}},
		{Kind: "empty_value", Args: []ir.Expr{{Kind: "string", Text: ""}}},
		{Kind: "empty_value", Args: []ir.Expr{{Kind: "var", Text: "html", Args: []ir.Expr{root}}}},
		{Kind: "empty_value", Args: []ir.Expr{{Kind: "var", Text: "html", Params: []string{}}}},
	} {
		p := ir.Program{Version: ir.Version, PHPVersion: testTarget(t), Body: []ir.Stmt{
			{Kind: "assign", Name: "html", Line: 1, Expr: ir.Expr{Kind: "string", Text: "0", Line: 1}},
			{Kind: "echo", Line: 2, Expr: expression},
		}}
		if out, err := Generate(p); err == nil || len(out) != 0 {
			t.Fatalf("malformed %+v: %s %v", expression, out, err)
		}
	}
}

func TestEmptyStringPreservesLegacyRawArrayIR(t *testing.T) {
	p, err := frontend.Parse("<?php $a=[];echo empty($a);", testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	p.Body[1].Expr.Kind = "empty_array"
	out, err := Generate(p)
	if err != nil || !strings.Contains(string(out), "rt.ArrayEmpty(") {
		t.Fatalf("legacy raw boundary: %s %v", out, err)
	}
	p.Body[0].Expr = ir.Expr{Kind: "string", Text: "", Line: 1}
	if out, err := Generate(p); err == nil || len(out) != 0 {
		t.Fatal("legacy array-only IR silently broadened")
	}
}
