package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestStrCaseCmpLoweringAndSourceMap(t *testing.T) {
	p, err := frontend.Parse("<?php\n$result = StRcAsEcMp('UTF-8', 'utf-8');\n", "8.3")
	if err != nil {
		t.Fatal(err)
	}
	generated, sourceMap, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	for line, text := range strings.Split(string(generated), "\n") {
		if strings.Contains(text, `rt.StrCaseCmp("UTF-8", "utf-8")`) {
			if got, ok := sourceMap.PHPLine(line + 1); !ok || got != 2 {
				t.Fatalf("strcasecmp maps to PHP line %d, %v; want 2", got, ok)
			}
			return
		}
	}
	t.Fatalf("strcasecmp did not lower through runtime guard:\n%s", generated)
}

func TestStrCaseCmpMalformedArity(t *testing.T) {
	for _, count := range []int{0, 1, 3} {
		args := make([]ir.Expr, count)
		for i := range args {
			args[i] = ir.Expr{Kind: "string", Text: "x", Line: 1}
		}
		p := ir.Program{Version: ir.Version, PHPVersion: "8.3", Body: []ir.Stmt{
			{Kind: "echo", Line: 1, Expr: ir.Expr{Kind: "call", Text: "strcasecmp", Args: args, Line: 1}},
		}}
		if _, err := Generate(p); err == nil || !strings.Contains(err.Error(), "strcasecmp expects two arguments") {
			t.Fatalf("strcasecmp with %d arguments was not rejected: %v", count, err)
		}
	}
}
