package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestIfElseSourceMapAndEmptyBranch(t *testing.T) {
	const php = "<?php\nif (false) {\n echo 'then';\n} else {\n if (true) {\n  echo 'nested';\n } else {}\n echo 'otherwise';\n}\n"
	p, err := frontend.Parse(php, testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	if p.Body[0].ElseBody[0].ElseBody == nil {
		t.Fatal("empty else lost its explicit IR presence")
	}
	generated, sourceMap, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	for fragment, want := range map[string]int{
		"if rt.Bool(false)": 2, `rt.Echo("then")`: 3,
		"if rt.Bool(true)": 5, `rt.Echo("nested")`: 6, `rt.Echo("otherwise")`: 8,
	} {
		found := false
		for line, text := range strings.Split(string(generated), "\n") {
			if strings.Contains(text, fragment) {
				found = true
				if got, ok := sourceMap.PHPLine(line + 1); !ok || got != want {
					t.Errorf("%s maps to %d (%v), want %d", fragment, got, ok, want)
				}
			}
		}
		if !found {
			t.Errorf("missing generated statement %s", fragment)
		}
	}
}

func TestElseBodyRequiresIfIR(t *testing.T) {
	for _, otherwise := range [][]ir.Stmt{{}, {{Kind: "echo", Expr: ir.Expr{Kind: "null"}}}} {
		p := ir.Program{Version: ir.Version, PHPVersion: testTarget(t), Body: []ir.Stmt{
			{Kind: "echo", Expr: ir.Expr{Kind: "null"}, ElseBody: otherwise},
		}}
		if _, err := Generate(p); err == nil || !strings.Contains(err.Error(), "IR002") {
			t.Fatalf("non-if else body was ignored: %v", err)
		}
	}
}
