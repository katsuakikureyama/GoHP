package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestCallStatementLoweringAndSourceMap(t *testing.T) {
	const source = `<?php
function marker($value) { echo $value; return ['ignored']; }
class Box { public function touch($value) { return $value; } }
marker('once');
$box = new Box();
$box->touch('method');
$cb = fn($value) => marker($value);
$cb('arrow');
strlen('ignored');
while (true) {
  break;
  marker('dead');
}
marker('after');`
	p, err := frontend.Parse(source, "8.3")
	if err != nil {
		t.Fatal(err)
	}
	if p.Body[0].Kind != "call_stmt" || p.Body[0].Expr.Kind != "call" || p.Body[2].Expr.Kind != "method" || p.Body[4].Expr.Kind != "invoke" {
		t.Fatalf("call IR: %#v", p.Body)
	}
	out, mapping, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	for fragment, want := range map[string]int{`_ = F_marker("once")`: 4, `.M_touch("method")`: 6, `_ = rt.Invoke(v["cb"], "arrow")`: 8, `_ = rt.Strlen("ignored")`: 9, `_ = F_marker("after")`: 14} {
		found := 0
		for line, text := range strings.Split(string(out), "\n") {
			if strings.Contains(text, fragment) {
				found++
				if got, ok := mapping.PHPLine(line + 1); !ok || got != want {
					t.Fatalf("%s mapping %d %v", fragment, got, ok)
				}
			}
		}
		if found != 1 {
			t.Fatalf("%s occurs %d times:\n%s", fragment, found, out)
		}
	}
	if strings.Contains(string(out), `F_marker("dead")`) {
		t.Fatal("dead call emitted")
	}
}

func TestCallStatementUnsupportedSource(t *testing.T) {
	for name, source := range map[string]string{
		"unknown": "missing();", "arity": "strlen();", "type": "count(false);",
		"literal": "42;", "variable": "$x = 1; $x;", "array": "[1];", "language_construct": "isset($x);",
		"binary": "strlen('x') + 1;", "conditional": "strlen('x') ? strlen('a') : strlen('b');",
		"parentheses": "(strlen('x'));", "new": "class X {} new X();",
		"undefined_receiver": "$missing->f();", "undefined_arrow": "$missing();",
		"scalar_receiver": "$x = false; $x->f();", "scalar_callee": "$x = 'strlen'; $x('x');",
		"method_arity":        "class X { public function f($x) {return $x;} } $x = new X(); $x->f();",
		"chained":             "class X { public function f() {return $this;} } $x = new X(); $x->f()->f();",
		"dynamic_name":        "class X {} $x = new X(); $n = 'f'; $x->$n();",
		"constructor_reentry": "class X { public function __construct() {} } $x = new X(); $x->__construct();",
		"dead_unknown":        "while (false) { break; missing(); }",
		"dead_type":           "while (false) { continue; count(false); }",
		"dead_undefined":      "while (false) { break; strlen($missing); }",
		"no_binding":          "function f() {return 1;} f(); echo $x;",
	} {
		t.Run(name, func(t *testing.T) {
			p, err := frontend.Parse("<?php "+source, "8.3")
			var out []byte
			if err == nil {
				out, err = Generate(p)
			}
			if err == nil || len(out) != 0 || !strings.Contains(err.Error(), "PHP00") {
				t.Fatalf("unsupported %s: %v", out, err)
			}
		})
	}
}

func TestCallStatementMalformedIR(t *testing.T) {
	valid := ir.Expr{Kind: "call", Text: "strlen", Args: []ir.Expr{{Kind: "string", Text: "x"}}}
	for name, stmt := range map[string]ir.Stmt{
		"name": {Kind: "call_stmt", Expr: valid, Name: "x"}, "key": {Kind: "call_stmt", Expr: valid, Key: "x"}, "value": {Kind: "call_stmt", Expr: valid, Value: "x"},
		"body": {Kind: "call_stmt", Expr: valid, Body: []ir.Stmt{}}, "else": {Kind: "call_stmt", Expr: valid, ElseBody: []ir.Stmt{}},
		"target":  {Kind: "call_stmt", Expr: valid, Target: ir.Expr{Args: []ir.Expr{}}},
		"missing": {Kind: "call_stmt"}, "literal": {Kind: "call_stmt", Expr: ir.Expr{Kind: "int", Text: "1"}},
		"method_receiver": {Kind: "call_stmt", Expr: ir.Expr{Kind: "method", Text: "f"}},
		"invoke_receiver": {Kind: "call_stmt", Expr: ir.Expr{Kind: "invoke"}},
		"call_payload":    {Kind: "call_stmt", Expr: ir.Expr{Kind: "call", Text: "strlen", Args: valid.Args, Params: []string{"hidden"}}},
	} {
		t.Run(name, func(t *testing.T) {
			p := ir.Program{Version: ir.Version, PHPVersion: "8.3", Body: []ir.Stmt{stmt}}
			out, err := Generate(p)
			if err == nil || len(out) != 0 {
				t.Fatalf("malformed output %s: %v", out, err)
			}
		})
	}
}
