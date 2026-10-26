package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestArrayWriteLoweringAndSourceMap(t *testing.T) {
	const source = `<?php
$items = ['old' => 'before'];
$key = 'old';
$items[$key] = 'after';
echo $items['old'];
`
	p, err := frontend.Parse(source, testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	generated, sourceMap, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for line, text := range strings.Split(string(generated), "\n") {
		if !strings.Contains(text, `v["items"] = rt.ArraySet(v["items"], v["key"], "after")`) {
			continue
		}
		found = true
		if got, ok := sourceMap.PHPLine(line + 1); !ok || got != 4 {
			t.Fatalf("array write maps to PHP line %d, %v; want 4", got, ok)
		}
	}
	if !found {
		t.Fatalf("generated Go did not lower the write through ArraySet:\n%s", generated)
	}
}

func TestArrayWriteMalformedTarget(t *testing.T) {
	arrayLiteral := ir.Expr{Kind: "array", Line: 1}
	variable := func(name string) ir.Expr { return ir.Expr{Kind: "var", Text: name, Line: 2} }
	validKey := ir.Expr{Kind: "string", Text: "key", Line: 2}
	validTarget := ir.Expr{Kind: "index", Line: 2, Args: []ir.Expr{variable("items"), validKey}}
	for _, tt := range []struct {
		name string
		stmt ir.Stmt
		code string
	}{
		{
			name: "not_index", code: "IR002",
			stmt: ir.Stmt{Kind: "assign_index", Line: 2, Target: variable("items"), Expr: validKey},
		},
		{
			name: "wrong_arity", code: "IR002",
			stmt: ir.Stmt{Kind: "assign_index", Line: 2, Target: ir.Expr{Kind: "index", Line: 2, Args: []ir.Expr{variable("items")}}, Expr: validKey},
		},
		{
			name: "nested_target", code: "IR002",
			stmt: ir.Stmt{Kind: "assign_index", Line: 2, Target: ir.Expr{Kind: "index", Line: 2, Args: []ir.Expr{{Kind: "index", Line: 2, Args: []ir.Expr{variable("items"), validKey}}, validKey}}, Expr: validKey},
		},
		{
			name: "non_variable_base", code: "IR002",
			stmt: ir.Stmt{Kind: "assign_index", Line: 2, Target: ir.Expr{Kind: "index", Line: 2, Args: []ir.Expr{arrayLiteral, validKey}}, Expr: validKey},
		},
		{
			name: "call_key", code: "PHP004",
			stmt: ir.Stmt{Kind: "assign_index", Line: 2, Target: ir.Expr{Kind: "index", Line: 2, Args: []ir.Expr{variable("items"), {Kind: "call", Text: "marker", Line: 2}}}, Expr: validKey},
		},
		{
			name: "legacy_name_populated", code: "IR002",
			stmt: ir.Stmt{Kind: "assign_index", Name: "items", Line: 2, Target: validTarget, Expr: validKey},
		},
		{
			name: "undefined_root", code: "PHP004",
			stmt: ir.Stmt{Kind: "assign_index", Line: 2, Target: validTarget, Expr: validKey},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			body := []ir.Stmt{tt.stmt}
			if tt.name != "undefined_root" {
				body = append([]ir.Stmt{{Kind: "assign", Name: "items", Line: 1, Expr: arrayLiteral}}, body...)
			}
			p := ir.Program{Version: ir.Version, PHPVersion: testTarget(t), Body: body}
			if _, err := Generate(p); err == nil || !strings.Contains(err.Error(), tt.code) {
				t.Fatalf("malformed array target was not rejected safely: %v", err)
			}
		})
	}
}
