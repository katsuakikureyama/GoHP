package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestParserOutputLowering(t *testing.T) {
	const source = `<?php
class Datum { public $html; public function __construct($html) { $this->html = $html; } }
class Parser {
  public $output;
  public function freeform($html) { return new Datum($html); }
  public function add($html, $start = null, $length = null) {
    if (0 === $length) { return; }
    $this->output[] = (array)$this->freeform(substr($html, $start ?? 0, $length));
  }
}
$p = new Parser();
$p->output = [];
$p->add('abcd');
$p->add('abcd', 1, 2);
$short = substr('abcd', 1);
`
	p, err := frontend.Parse(source, "8.3")
	if err != nil {
		t.Fatal(err)
	}
	method := p.Classes[1].Methods[1]
	if method.OptionalNullParams != 2 {
		t.Fatalf("optional count = %d", method.OptionalNullParams)
	}
	ret := method.Body[0].Body[0]
	if ret.Kind != "return" || ret.Expr.Kind != "null" || ret.Expr.Text != "null" {
		t.Fatalf("bare return not canonical null: %#v", ret)
	}
	appendStmt := method.Body[1]
	if appendStmt.Kind != "assign_property_append" || appendStmt.Target.Kind != "property" || appendStmt.Target.Text != "output" || appendStmt.Target.Args[0].Text != "this" {
		t.Fatalf("append IR = %#v", appendStmt)
	}
	out, mapping, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	text := string(out)
	for _, want := range []string{"args = rt.NullDefaultArgs(args, 1, 3)", `rt.Substr("abcd", int64(1), nil)`, `rt.ArrayAppend(target.Get("output"), value)`, "rt.ObjectToArray("} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q:\n%s", want, out)
		}
	}
	found := false
	for line, generated := range strings.Split(text, "\n") {
		if strings.Contains(generated, `target.Set("output",`) {
			found = true
			if got, ok := mapping.PHPLine(line + 1); !ok || got != 8 {
				t.Fatalf("append map = %d,%v; want8", got, ok)
			}
		}
	}
	if !found {
		t.Fatal("no mapped append statement")
	}
}

func TestParserOutputUnsupportedSource(t *testing.T) {
	for name, source := range map[string]string{
		"method_below_min":        `class C { public function f($x,$y=null) {} } $c=new C(); $c->f();`,
		"method_above_max":        `class C { public function f($x=null) {} } $c=new C(); $c->f(1,2);`,
		"method_nonnull_default":  `class C { public function f($x=0) {} }`,
		"method_typed_default":    `class C { public function f(array $x=null) {} }`,
		"bare_top_return":         `return;`,
		"bare_typed_return":       `function f(): array { return; }`,
		"bare_constructor_return": `class C { public function __construct() { return; } }`,
		"append_unknown_receiver": `function f($o) { $o->items[] = 1; }`,
		"append_missing_property": `class C {} $c=new C(); $c->items[]=1;`,
		"append_nested":           `class C { public $items; } $c=new C(); $c->items[0][]=1;`,
		"append_chained":          `class C { public $items; } $c=new C(); $c->items->next[]=1;`,
		"cast_nonsimple_factory":  `class C { public function make($x) { if ($x) { return new C(); } return new C(); } } $c=new C(); $a=(array)$c->make(true);`,
		"cast_foreign_receiver":   `class C { public function make() { return new C(); } } function f($o) { return (array)$o->make(); }`,
		"cast_passthrough":        `class C { public function make($x) { return $x; } } $c=new C(); $a=(array)$c->make(new C());`,
		"substr_arity":            `$x=substr('a');`,
		"substr_extra":            `$x=substr('a',0,null,1);`,
		"substr_value":            `$x=substr(null,0);`,
		"substr_offset":           `$x=substr('a','0');`,
		"substr_length":           `$x=substr('a',0,false);`,
		"substr_redeclare":        `function SuBsTr($x) { return $x; }`,
	} {
		t.Run(name, func(t *testing.T) {
			p, err := frontend.Parse("<?php "+source, "8.3")
			var out []byte
			if err == nil {
				out, err = Generate(p)
			}
			if err == nil || len(out) != 0 {
				t.Fatalf("unsupported source generated %s: %v", out, err)
			}
		})
	}
}

func TestParserOutputMalformedIR(t *testing.T) {
	null := ir.Expr{Kind: "null", Text: "null", Line: 1}
	base := ir.Expr{Kind: "var", Text: "box", Line: 2}
	target := ir.Expr{Kind: "property", Text: "items", Line: 2, Args: []ir.Expr{base}}
	for name, stmt := range map[string]ir.Stmt{
		"wrong_target":   {Kind: "assign_property_append", Target: base, Expr: null, Line: 2},
		"wrong_receiver": {Kind: "assign_property_append", Target: ir.Expr{Kind: "property", Text: "items", Args: []ir.Expr{null}, Line: 2}, Expr: null, Line: 2},
		"extra_metadata": {Kind: "assign_property_append", Target: target, Name: "items", Expr: null, Line: 2},
		"body_payload":   {Kind: "assign_property_append", Target: target, Body: []ir.Stmt{}, Expr: null, Line: 2},
	} {
		t.Run(name, func(t *testing.T) {
			p := ir.Program{Version: ir.Version, PHPVersion: "8.3", Classes: []ir.Class{{Name: "C", Properties: []ir.Property{{Name: "items", Value: null}}}}, Body: []ir.Stmt{{Kind: "assign", Name: "box", Expr: ir.Expr{Kind: "new", Text: "C", Line: 1}, Line: 1}, stmt}}
			out, err := Generate(p)
			if err == nil || len(out) != 0 || !strings.Contains(err.Error(), "IR002") {
				t.Fatalf("malformed IR generated %s: %v", out, err)
			}
		})
	}
}
