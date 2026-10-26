package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestArrayAppendLoweringAndSourceMap(t *testing.T) {
	p, err := frontend.Parse("<?php\n$items = [];\n$items[] = 'added';\n", testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	if target := p.Body[1]; target.Kind != "assign_append" || target.Target.Kind != "var" || target.Target.Text != "items" {
		t.Fatalf("append target is not explicit: %#v", target)
	}
	generated, sourceMap, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	for line, text := range strings.Split(string(generated), "\n") {
		if strings.Contains(text, `v["items"] = rt.ArrayAppend(v["items"], "added")`) {
			if got, ok := sourceMap.PHPLine(line + 1); !ok || got != 3 {
				t.Fatalf("append maps to PHP line %d, %v; want 3", got, ok)
			}
			return
		}
	}
	t.Fatalf("append was not lowered through ArrayAppend:\n%s", generated)
}

func TestArrayAppendMalformedTarget(t *testing.T) {
	variable := ir.Expr{Kind: "var", Text: "items", Line: 2}
	value := ir.Expr{Kind: "string", Text: "added", Line: 2}
	for _, tt := range []struct {
		name   string
		target ir.Expr
		legacy string
	}{
		{"missing", ir.Expr{}, ""},
		{"index", ir.Expr{Kind: "index", Args: []ir.Expr{variable, value}}, ""},
		{"call", ir.Expr{Kind: "call", Text: "items"}, ""},
		{"extra_args", ir.Expr{Kind: "var", Text: "items", Args: []ir.Expr{value}}, ""},
		{"bad_name", ir.Expr{Kind: "var", Text: "bad-name"}, ""},
		{"legacy_name", variable, "items"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := ir.Program{Version: ir.Version, PHPVersion: testTarget(t), Body: []ir.Stmt{
				{Kind: "assign", Name: "items", Line: 1, Expr: ir.Expr{Kind: "array"}},
				{Kind: "assign_append", Name: tt.legacy, Target: tt.target, Expr: value, Line: 2},
			}}
			if _, err := Generate(p); err == nil || !strings.Contains(err.Error(), "IR002") {
				t.Fatalf("malformed append target was not rejected: %v", err)
			}
		})
	}
}
