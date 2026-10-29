package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestPropertyIssetLowering(t *testing.T) {
	const source = `<?php
class Node { public $child; public $value; }
function probe($node, $array, $plain) { return isset($node->child->value, $array['key'], $plain); }
$node = new Node();
echo IsSeT($node->missing);
`
	p, err := frontend.Parse(source, testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	address := p.Functions[0].Body[0].Expr.Args[0]
	if address.Kind != "isset_property" || len(address.Args) != 3 || address.Args[0].Kind != "var" || address.Args[1].Kind != "string" || address.Args[1].Text != "child" || address.Args[2].Text != "value" {
		t.Fatalf("property address IR: %+v", address)
	}
	out, mapping, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	text := string(out)
	for _, want := range []string{`PHPPropertyLookup(k string) (any, bool)`, `return self.P_child, true`, `rt.PropertyIsset(v["node"], "child", "value") && rt.ArrayIsset(v["array"], "key") && rt.IssetValue(v["plain"])`, `rt.PropertyIsset(v["node"], "missing")`} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(text, `.Get("missing")`) || strings.Contains(text, "NewScalarFrame") || strings.Contains(text, "rt.ValueSlot") {
		t.Fatal("presence must use quiet lookup and remain legacy storage")
	}
	for i, line := range strings.Split(text, "\n") {
		if strings.Contains(line, `rt.PropertyIsset(v["node"], "missing")`) {
			if got, ok := mapping.PHPLine(i + 1); !ok || got != 5 {
				t.Fatalf("source map: %d %v", got, ok)
			}
			return
		}
	}
	t.Fatal("missing mapped property probe")
}

func TestPropertyIssetUnsupportedSource(t *testing.T) {
	for _, source := range []string{
		`echo isset($missing->value);`,
		`if(true){$x=null;} echo isset($x->value);`,
		`$x=null; $name='value'; echo isset($x->$name);`,
		`$x=null; echo isset($x->{'value'});`,
		`$x=null; echo isset($x->value[0]);`,
		`$x=[]; echo isset($x[0]->value);`,
		`$x=null; echo isset(($x)->value);`,
		`function source(){return null;} echo isset(source()->value);`,
		`$x=null; echo isset($x->method()->value);`,
		`$x=fn()=>true; echo isset($x->value);`,
		`class Node{} $x=new Node(); echo $x->missing;`,
	} {
		p, err := frontend.Parse("<?php "+source, testTarget(t))
		var out []byte
		if err == nil {
			out, err = Generate(p)
		}
		if err == nil || len(out) != 0 {
			t.Fatalf("unsupported %s: %s %v", source, out, err)
		}
	}
}

func TestPropertyIssetMalformedIR(t *testing.T) {
	root := ir.Expr{Kind: "var", Text: "node", Line: 2}
	name := ir.Expr{Kind: "string", Text: "value", Line: 2}
	for _, expr := range []ir.Expr{
		{Kind: "isset_property", Args: []ir.Expr{root}},
		{Kind: "isset_property", Text: "hidden", Args: []ir.Expr{root, name}},
		{Kind: "isset_property", Params: []string{}, Args: []ir.Expr{root, name}},
		{Kind: "isset_property", Args: []ir.Expr{{Kind: "var", Text: "node", Args: []ir.Expr{}}, name}},
		{Kind: "isset_property", Args: []ir.Expr{{Kind: "var", Text: "node", Params: []string{}}, name}},
		{Kind: "isset_property", Args: []ir.Expr{{Kind: "call", Text: "source"}, name}},
		{Kind: "isset_property", Args: []ir.Expr{root, {Kind: "var", Text: "name"}}},
		{Kind: "isset_property", Args: []ir.Expr{root, {Kind: "string", Text: "bad-name"}}},
		{Kind: "isset_property", Args: []ir.Expr{root, {Kind: "string", Text: ""}}},
		{Kind: "isset_property", Args: []ir.Expr{root, {Kind: "string", Text: "value", Args: []ir.Expr{}}}},
		{Kind: "isset_property", Args: []ir.Expr{root, {Kind: "string", Text: "value", Params: []string{}}}},
	} {
		p := ir.Program{Version: ir.Version, PHPVersion: testTarget(t), Body: []ir.Stmt{
			{Kind: "assign", Name: "node", Line: 1, Expr: ir.Expr{Kind: "null", Text: "null", Line: 1}},
			{Kind: "echo", Line: 2, Expr: expr},
		}}
		if out, err := Generate(p); err == nil || len(out) != 0 {
			t.Fatalf("malformed %+v: %s %v", expr, out, err)
		}
	}
}
