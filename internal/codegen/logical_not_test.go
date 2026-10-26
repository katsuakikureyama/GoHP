package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestLogicalNotLoweringAndSourceMap(t *testing.T) {
	p, err := frontend.Parse("<?php\n$value = '0';\n$negated = !$value;\n", testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	e := p.Body[1].Expr
	if e.Kind != "not" || len(e.Args) != 1 || e.Args[0].Kind != "var" {
		t.Fatalf("not is not a unary IR operation: %#v", e)
	}
	generated, sourceMap, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	for line, text := range strings.Split(string(generated), "\n") {
		if strings.Contains(text, `!rt.Bool(v["value"])`) {
			if got, ok := sourceMap.PHPLine(line + 1); !ok || got != 3 {
				t.Fatalf("not maps to PHP line %d, %v; want 3", got, ok)
			}
			return
		}
	}
	t.Fatalf("not did not lower through the Bool guard:\n%s", generated)
}

func TestLogicalNotMalformedArity(t *testing.T) {
	for _, count := range []int{0, 2} {
		args := make([]ir.Expr, count)
		for i := range args {
			args[i] = ir.Expr{Kind: "null", Line: 1}
		}
		p := ir.Program{Version: ir.Version, PHPVersion: testTarget(t), Body: []ir.Stmt{
			{Kind: "echo", Line: 1, Expr: ir.Expr{Kind: "not", Args: args, Line: 1}},
		}}
		if _, err := Generate(p); err == nil || !strings.Contains(err.Error(), "IR002") {
			t.Fatalf("not with %d operands was not rejected: %v", count, err)
		}
	}
}
