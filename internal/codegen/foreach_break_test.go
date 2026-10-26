package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestForeachBreakLoweringAndSourceMap(t *testing.T) {
	p, err := frontend.Parse("<?php\nforeach ([1, 2] as $v) {\n if ($v === 1) {\n  BrEaK;\n  echo 'dead-tail';\n }\n echo $v;\n}\necho 'after';", testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	if got := p.Body[0].Body[0].Body[0]; got.Kind != "break" || got.Expr.Kind != "" {
		t.Fatalf("break IR: %#v", got)
	}
	out, mapping, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(out), "dead-tail") {
		t.Fatal("unreachable tail emitted")
	}
	for fragment, want := range map[string]int{"break": 4, `rt.Echo(v["v"])`: 7, `rt.Echo("after")`: 9} {
		found := false
		for line, text := range strings.Split(string(out), "\n") {
			if strings.TrimSpace(text) == fragment || strings.Contains(text, fragment) && fragment != "break" {
				found = true
				if got, ok := mapping.PHPLine(line + 1); !ok || got != want {
					t.Fatalf("%s mapping = %d, %v, want %d", fragment, got, ok, want)
				}
			}
		}
		if !found {
			t.Fatalf("missing %s:\n%s", fragment, out)
		}
	}
}

func TestForeachBreakUnsupportedSource(t *testing.T) {
	for name, source := range map[string]string{
		"top_level": "break;", "if_outside": "if (false) { break; }",
		"function_outside":        "function f() { break; }",
		"method_outside":          "class X { public function f() { break; } }",
		"called_function_outside": "function f() { break; } foreach ([1] as $v) { $r = f(); }",
		"explicit_one":            "foreach ([1] as $v) { break 1; }",
		"outer_level":             "foreach ([1] as $v) { foreach ([2] as $w) { break 2; } }",
		"dynamic_level":           "foreach ([1] as $v) { break $v; }",
		"parenthesized_level":     "foreach ([1] as $v) { break(1); }",
		"while_level":             "while (true) { break 1; }",
		"redeclare":               "function BrEaK() { return 1; }",
		"dead_semantic":           "foreach ([1] as $v) { break; echo count(false); }",
		"dead_codegen":            "foreach ([1] as $v) { break; echo $undefined; }",
		"branch_binding":          "foreach ([1] as $v) { if (false) { break; $x = 1; } echo $x; }",
		"loop_binding":            "foreach ([1] as $v) { break; $x = 1; } echo $x;",
	} {
		t.Run(name, func(t *testing.T) {
			p, err := frontend.Parse("<?php "+source, testTarget(t))
			var out []byte
			if err == nil {
				out, err = Generate(p)
			}
			if err == nil || len(out) != 0 || !strings.Contains(err.Error(), "PHP00") {
				t.Fatalf("unsupported source produced %s: %v", out, err)
			}
		})
	}
}

func TestForeachBreakMalformedIR(t *testing.T) {
	for name, stmt := range map[string]ir.Stmt{
		"expression":          {Kind: "break", Expr: ir.Expr{Kind: "int", Text: "1"}},
		"expression_metadata": {Kind: "break", Expr: ir.Expr{Args: []ir.Expr{}}},
		"target":              {Kind: "break", Target: ir.Expr{Kind: "var", Text: "v"}},
		"name":                {Kind: "break", Name: "label"}, "key": {Kind: "break", Key: "k"},
		"value": {Kind: "break", Value: "v"}, "body": {Kind: "break", Body: []ir.Stmt{}},
		"else": {Kind: "break", ElseBody: []ir.Stmt{}},
	} {
		t.Run(name, func(t *testing.T) {
			p := ir.Program{Version: ir.Version, PHPVersion: testTarget(t), Body: []ir.Stmt{{Kind: "foreach", Value: "v", Expr: ir.Expr{Kind: "array"}, Body: []ir.Stmt{stmt}}}}
			out, err := Generate(p)
			if err == nil || len(out) != 0 || !strings.Contains(err.Error(), "IR002") {
				t.Fatalf("malformed break produced %s: %v", out, err)
			}
		})
	}
	for _, body := range [][]ir.Stmt{{{Kind: "break"}}, {{Kind: "if", Expr: ir.Expr{Kind: "bool", Text: "false"}, Body: []ir.Stmt{{Kind: "break"}}}}} {
		p := ir.Program{Version: ir.Version, PHPVersion: testTarget(t), Body: body}
		if out, err := Generate(p); err == nil || len(out) != 0 || !strings.Contains(err.Error(), "enclosing foreach") {
			t.Fatalf("outside-loop IR: %s, %v", out, err)
		}
	}
}

func TestForeachBreakDeadTailTypeState(t *testing.T) {
	// The reachable false assignment must not be hidden by the dead array
	// assignment, or scalar echo after the loop is incorrectly rejected.
	p, err := frontend.Parse("<?php $a = []; foreach ([1] as $v) { $a = false; break; $a = []; } echo $a;", testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Generate(p); err != nil {
		t.Fatal(err)
	}
}
