package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestStrictNotEqualLoweringAndSourceMap(t *testing.T) {
	p, err := frontend.Parse("<?php\n$value = 1;\n$result = $value !== '1';\n", testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	generated, sourceMap, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	for line, text := range strings.Split(string(generated), "\n") {
		if strings.Contains(text, `!rt.Identical(v["value"], "1")`) {
			if got, ok := sourceMap.PHPLine(line + 1); !ok || got != 3 {
				t.Fatalf("!== maps to PHP line %d, %v; want 3", got, ok)
			}
			return
		}
	}
	t.Fatalf("!== did not lower through the scalar comparison guard:\n%s", generated)
}

func TestStrictNotEqualMalformedArity(t *testing.T) {
	for _, count := range []int{0, 1, 3} {
		args := make([]ir.Expr, count)
		for i := range args {
			args[i] = ir.Expr{Kind: "null", Line: 1}
		}
		p := ir.Program{Version: ir.Version, PHPVersion: testTarget(t), Body: []ir.Stmt{
			{Kind: "echo", Line: 1, Expr: ir.Expr{Kind: "binary", Text: "!==", Args: args, Line: 1}},
		}}
		if _, err := Generate(p); err == nil || !strings.Contains(err.Error(), "IR002") {
			t.Fatalf("!== with %d operands was not rejected: %v", count, err)
		}
	}
}
