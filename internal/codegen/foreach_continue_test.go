package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestForeachContinueLoweringAndSourceMap(t *testing.T) {
	p, err := frontend.Parse("<?php\nforeach ([1, 2] as $v) {\n if ($v === 1) {\n  CoNtInUe;\n  echo 'dead-tail';\n }\n echo $v;\n}\necho 'after';", "8.3")
	if err != nil {
		t.Fatal(err)
	}
	if got := p.Body[0].Body[0].Body[0]; got.Kind != "continue" || got.Expr.Kind != "" {
		t.Fatalf("continue IR: %#v", got)
	}
	out, mapping, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(out), "dead-tail") {
		t.Fatal("unreachable tail emitted")
	}
	for fragment, want := range map[string]int{"continue": 4, `rt.Echo(v["v"])`: 7, `rt.Echo("after")`: 9} {
		found := false
		for line, text := range strings.Split(string(out), "\n") {
			if strings.TrimSpace(text) == fragment || strings.Contains(text, fragment) && fragment != "continue" {
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

func TestForeachContinueUnsupportedSource(t *testing.T) {
	for name, source := range map[string]string{
		"top_level": "continue;", "if_outside": "if (false) { continue; }",
		"function_outside":        "function f() { continue; }",
		"method_outside":          "class X { public function f() { continue; } }",
		"called_function_outside": "function f() { continue; } foreach ([1] as $v) { $r = f(); }",
		"explicit_one":            "foreach ([1] as $v) { continue 1; }",
		"outer_level":             "foreach ([1] as $v) { foreach ([2] as $w) { continue 2; } }",
		"dynamic_level":           "foreach ([1] as $v) { continue $v; }",
		"parenthesized_level":     "foreach ([1] as $v) { continue(1); }",
		"break":                   "foreach ([1] as $v) { break; }",
		"while":                   "while (true) { continue; }",
		"redeclare":               "function CoNtInUe() { return 1; }",
		"dead_semantic":           "foreach ([1] as $v) { continue; echo count(false); }",
		"dead_codegen":            "foreach ([1] as $v) { continue; echo $undefined; }",
		"branch_binding":          "foreach ([1] as $v) { if (false) { continue; $x = 1; } echo $x; }",
		"loop_binding":            "foreach ([1] as $v) { continue; $x = 1; } echo $x;",
	} {
		t.Run(name, func(t *testing.T) {
			p, err := frontend.Parse("<?php "+source, "8.3")
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

func TestForeachContinueMalformedIR(t *testing.T) {
	for name, stmt := range map[string]ir.Stmt{
		"expression":          {Kind: "continue", Expr: ir.Expr{Kind: "int", Text: "1"}},
		"expression_metadata": {Kind: "continue", Expr: ir.Expr{Args: []ir.Expr{}}},
		"target":              {Kind: "continue", Target: ir.Expr{Kind: "var", Text: "v"}},
		"name":                {Kind: "continue", Name: "label"}, "key": {Kind: "continue", Key: "k"},
		"value": {Kind: "continue", Value: "v"}, "body": {Kind: "continue", Body: []ir.Stmt{}},
		"else": {Kind: "continue", ElseBody: []ir.Stmt{}},
	} {
		t.Run(name, func(t *testing.T) {
			p := ir.Program{Version: ir.Version, PHPVersion: "8.3", Body: []ir.Stmt{{Kind: "foreach", Value: "v", Expr: ir.Expr{Kind: "array"}, Body: []ir.Stmt{stmt}}}}
			out, err := Generate(p)
			if err == nil || len(out) != 0 || !strings.Contains(err.Error(), "IR002") {
				t.Fatalf("malformed continue produced %s: %v", out, err)
			}
		})
	}
	for _, body := range [][]ir.Stmt{{{Kind: "continue"}}, {{Kind: "if", Expr: ir.Expr{Kind: "bool", Text: "false"}, Body: []ir.Stmt{{Kind: "continue"}}}}} {
		p := ir.Program{Version: ir.Version, PHPVersion: "8.3", Body: body}
		if out, err := Generate(p); err == nil || len(out) != 0 || !strings.Contains(err.Error(), "enclosing foreach") {
			t.Fatalf("outside-loop IR: %s, %v", out, err)
		}
	}
}

func TestForeachContinueDeadTailTypeState(t *testing.T) {
	// The reachable false assignment must not be hidden by the dead array
	// assignment, or scalar echo after the loop is incorrectly rejected.
	p, err := frontend.Parse("<?php $a = []; foreach ([1] as $v) { $a = false; continue; $a = []; } echo $a;", "8.3")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Generate(p); err != nil {
		t.Fatal(err)
	}
}
