package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestWhileLoweringAndSourceMap(t *testing.T) {
	p, err := frontend.Parse("<?php $i = 0;\nWhIlE ($i !== 2) {\n $i = $i + 1;\n if ($i === 1) { continue; echo 'dead-continue'; }\n break;\n echo 'dead-break';\n}\necho 'after';", "8.3")
	if err != nil {
		t.Fatal(err)
	}
	if s := p.Body[1]; s.Kind != "while" || s.Name != "" || s.Expr.Kind != "binary" {
		t.Fatalf("while IR: %#v", s)
	}
	out, mapping, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(out), "dead-") {
		t.Fatal("dead tail emitted")
	}
	for fragment, want := range map[string]int{"for rt.Bool(": 2, "continue": 4, "break": 5, `rt.Echo("after")`: 8} {
		found := false
		for line, text := range strings.Split(string(out), "\n") {
			if strings.TrimSpace(text) == fragment || (fragment != "break" && fragment != "continue" && strings.Contains(text, fragment)) {
				found = true
				if got, ok := mapping.PHPLine(line + 1); !ok || got != want {
					t.Fatalf("%s mapping %d %v", fragment, got, ok)
				}
			}
		}
		if !found {
			t.Fatalf("missing %s:\n%s", fragment, out)
		}
	}
}

func TestWhileTypeState(t *testing.T) {
	for _, source := range []string{
		"$a = []; while (false) { $a = false; break; $a = []; } echo $a;",
		"$a = []; while (false) { foreach ([false] as $a) { break; } } echo $a;",
		"$a = []; while (false) { if (false) { $a = false; continue; } $a = []; } echo $a;",
		"class X { public $v = null; } $o = new X(); $i = 0; while ($i !== 1) { $o->v = $i; $i = $i + 1; } echo $o->v;",
	} {
		p, err := frontend.Parse("<?php "+source, "8.3")
		if err != nil {
			t.Fatal(err)
		}
		if _, err = Generate(p); err != nil {
			t.Fatal(err)
		}
	}
}

func TestWhileUnsupportedSource(t *testing.T) {
	for name, source := range map[string]string{
		"unbraced": "while (false) echo 1;", "empty_statement": "while (false);",
		"alternative": "while (false): endwhile;", "empty_condition": "while () {}",
		"assignment_condition": "$a = true; while ($a = false) {}",
		"for":                  "for (;;) { break; }", "do": "do {} while (false);", "switch": "switch (1) { default: break; }",
		"level_break": "while (false) { break 1; }", "level_continue": "while (false) { continue 2; }",
		"outside_function":    "function f() { break; } while (false) { $r = f(); }",
		"outside_method":      "class X { public function f() { continue; } }",
		"undefined_condition": "while ($missing) {}", "body_condition_binding": "while ($x) { $x = false; }",
		"escaped_binding":           "while (false) { $x = 1; } echo $x;",
		"object_condition":          "class X {} while (new X()) { break; }",
		"callable_condition":        "while (fn() => false) { break; }",
		"known_bad_first_condition": "$a = false; while (count($a)) { $a = []; }",
		"dead_bad":                  "while (false) { break; echo count(false); }",
		"dead_undefined":            "while (false) { continue; echo $missing; }",
		"redeclare":                 "function WhIlE() { return 1; }",
		"loop_carried_class":        "class X { public $v = null; } $o = new X(); while (false) { $o->v = 1; $o = false; }",
	} {
		t.Run(name, func(t *testing.T) {
			p, err := frontend.Parse("<?php "+source, "8.3")
			var out []byte
			if err == nil {
				out, err = Generate(p)
			}
			if err == nil || len(out) != 0 || !strings.Contains(err.Error(), "PHP00") {
				t.Fatalf("unsupported output %s: %v", out, err)
			}
		})
	}
}

func TestWhileMalformedIR(t *testing.T) {
	for name, stmt := range map[string]ir.Stmt{
		"missing_condition": {Kind: "while"},
		"name":              {Kind: "while", Name: "x"}, "key": {Kind: "while", Key: "x"}, "value": {Kind: "while", Value: "x"},
		"target_metadata":   {Kind: "while", Target: ir.Expr{Args: []ir.Expr{}}},
		"else":              {Kind: "while", ElseBody: []ir.Stmt{}},
		"condition_payload": {Kind: "while", Expr: ir.Expr{Kind: "bool", Text: "false", Args: []ir.Expr{{Kind: "null", Text: "null"}}}},
		"bad_nested_write":  {Kind: "while", Expr: ir.Expr{Kind: "bool", Text: "false"}, Body: []ir.Stmt{{Kind: "assign_index", Target: ir.Expr{Kind: "index"}}}},
	} {
		t.Run(name, func(t *testing.T) {
			p := ir.Program{Version: ir.Version, PHPVersion: "8.3", Body: []ir.Stmt{stmt}}
			out, err := Generate(p)
			if err == nil || len(out) != 0 || !strings.Contains(err.Error(), "IR002") {
				t.Fatalf("malformed output %s: %v", out, err)
			}
		})
	}
}
