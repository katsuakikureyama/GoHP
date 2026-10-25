package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestLogicalOrLoweringAndSourceMap(t *testing.T) {
	p, err := frontend.Parse("<?php\n$left = '0';\n$result = $left || false;\n", "8.3")
	if err != nil {
		t.Fatal(err)
	}
	generated, sourceMap, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	for line, text := range strings.Split(string(generated), "\n") {
		if strings.Contains(text, `rt.Bool(v["left"]) || rt.Bool(false)`) {
			if got, ok := sourceMap.PHPLine(line + 1); !ok || got != 3 {
				t.Fatalf("logical or maps to PHP line %d, %v; want 3", got, ok)
			}
			return
		}
	}
	t.Fatalf("logical or did not lower to a guarded short circuit:\n%s", generated)
}

func TestLogicalOrMalformedArity(t *testing.T) {
	for _, count := range []int{0, 1, 3} {
		args := make([]ir.Expr, count)
		for i := range args {
			args[i] = ir.Expr{Kind: "null", Line: 1}
		}
		p := ir.Program{Version: ir.Version, PHPVersion: "8.3", Body: []ir.Stmt{
			{Kind: "echo", Line: 1, Expr: ir.Expr{Kind: "binary", Text: "||", Args: args, Line: 1}},
		}}
		if _, err := Generate(p); err == nil || !strings.Contains(err.Error(), "IR002") {
			t.Fatalf("logical or with %d operands was not rejected: %v", count, err)
		}
	}
}
