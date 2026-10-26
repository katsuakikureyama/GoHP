package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestPublicPropertyWriteLoweringAndSourceMap(t *testing.T) {
	const source = `<?php
function marker($value) { return $value; }
class Box {
  public $items = null;
  public $label = 'before';
  public function replace($value) {
    $this->items = $value;
    return count($this->items);
  }
}
$box = new Box();
$alias = $box;
$alias->label = marker('after');
$box->items = ['one', 'two'];
$count = count($box->items);
foreach ($box->items as $item) { echo $item; }
echo $box->replace(['three']);
foreach ($box->items as $item) { echo $item; }
`
	p, err := frontend.Parse(source, "8.3")
	if err != nil {
		t.Fatal(err)
	}
	aliasWrite := p.Body[2]
	if aliasWrite.Kind != "assign_property" || aliasWrite.Name != "" || aliasWrite.Target.Kind != "property" ||
		aliasWrite.Target.Text != "label" || len(aliasWrite.Target.Args) != 1 || aliasWrite.Target.Args[0].Kind != "var" || aliasWrite.Target.Args[0].Text != "alias" {
		t.Fatalf("property write IR = %#v, want assign_property(property($alias, label))", aliasWrite)
	}
	methodWrite := p.Classes[0].Methods[0].Body[0]
	if methodWrite.Kind != "assign_property" || len(methodWrite.Target.Args) != 1 || methodWrite.Target.Args[0].Text != "this" {
		t.Fatalf("method property write IR = %#v, want $this receiver", methodWrite)
	}

	generated, sourceMap, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	text := string(generated)
	for _, want := range []string{
		`func (self *C_box) Set(k string, value any)`,
		`case "items":`,
		`self.P_items = rt.Clone(value)`,
		`case "label":`,
		`self.P_label = rt.Clone(value)`,
		`panic("undefined property assignment")`,
		`(v["this"]).(interface{ Set(string, any) }).Set("items", v["value"])`,
		`(v["alias"]).(interface{ Set(string, any) }).Set("label", F_marker("after"))`,
		`rt.Count((v["box"]).(interface{ Get(string) any }).Get("items"))`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("property write lowering missing %q:\n%s", want, generated)
		}
	}
	if strings.Count(text, `.Set("label", F_marker("after"))`) != 1 {
		t.Fatalf("property-write RHS must be evaluated once:\n%s", generated)
	}
	if got := strings.Count(text, `rt.Entries((v["box"]).(interface{ Get(string) any }).Get("items"))`); got != 2 {
		t.Fatalf("property value must remain usable as an array after alias/method writes; got %d foreach reads:\n%s", got, generated)
	}
	methodCall := strings.Index(text, `.M_replace(rt.NewArray(`)
	secondRead := strings.LastIndex(text, `rt.Entries((v["box"]).(interface{ Get(string) any }).Get("items"))`)
	if methodCall < 0 || secondRead < methodCall {
		t.Fatalf("method write must precede the post-method property read:\n%s", generated)
	}
	for line, generatedLine := range strings.Split(text, "\n") {
		if strings.Contains(generatedLine, `.Set("items", rt.NewArray(`) {
			if got, ok := sourceMap.PHPLine(line + 1); !ok || got != 14 {
				t.Fatalf("property write maps to PHP line %d, %v; want 14", got, ok)
			}
			return
		}
	}
	t.Fatalf("property write was not source-mapped:\n%s", generated)
}

func TestPublicPropertyWriteBoundaries(t *testing.T) {
	for name, source := range map[string]string{
		"undefined_receiver":        `$box->value = 'after';`,
		"unknown_parameter":         `class Box { public $value = null; } function replace($box) { $box->value = 'after'; }`,
		"foreign_call_result":       `class Box { public $value = null; } function build() { return new Box(); } $box = build(); $box->value = 'after';`,
		"scalar_receiver":           `$box = 1; $box->value = 'after';`,
		"callable_receiver":         `$box = fn() => null; $box->value = 'after';`,
		"undeclared_property":       `class Box { public $value = null; } $box = new Box(); $box->missing = 'after';`,
		"dynamic_property":          `class Box { public $value = null; } $box = new Box(); $name = 'value'; $box->$name = 'after';`,
		"nested_property":           `class Box { public $value = null; } $box = new Box(); $box->value->next = 'after';`,
		"property_index":            `class Box { public $value = null; } $box = new Box(); $box->value[0] = 'after';`,
		"property_append":           `class Box { public $value = null; } $box = new Box(); $box->value[] = 'after';`,
		"compound_assignment":       `class Box { public $value = null; } $box = new Box(); $box->value += 1;`,
		"reference_assignment":      `class Box { public $value = null; } $box = new Box(); $box->value =& $other;`,
		"assignment_expression_rhs": `class Box { public $value = null; } $box = new Box(); $result = $box->value = 'after';`,
	} {
		t.Run(name, func(t *testing.T) {
			p, err := frontend.Parse("<?php "+source, "8.3")
			var out []byte
			if err == nil {
				out, err = Generate(p)
			}
			if err == nil || !strings.Contains(err.Error(), "PHP00") {
				t.Fatalf("unsupported property write was accepted: %v", err)
			}
			if len(out) != 0 {
				t.Fatalf("rejected property write unexpectedly generated output:\n%s", out)
			}
		})
	}
}

func TestPublicPropertyWriteMalformedIR(t *testing.T) {
	variable := func(name string) ir.Expr { return ir.Expr{Kind: "var", Text: name, Line: 2} }
	validTarget := ir.Expr{Kind: "property", Text: "items", Line: 2, Args: []ir.Expr{variable("box")}}
	validValue := ir.Expr{Kind: "array", Line: 2}
	valid := ir.Stmt{Kind: "assign_property", Line: 2, Target: validTarget, Expr: validValue}
	program := func(stmt ir.Stmt, withBox bool) ir.Program {
		body := []ir.Stmt{stmt}
		if withBox {
			body = append([]ir.Stmt{{Kind: "assign", Name: "box", Line: 1, Expr: ir.Expr{Kind: "new", Text: "Box", Line: 1}}}, body...)
		}
		return ir.Program{
			Version: ir.Version, PHPVersion: "8.3",
			Classes: []ir.Class{{Name: "Box", Properties: []ir.Property{{Name: "items", Value: ir.Expr{Kind: "null", Text: "null", Line: 1}}}}},
			Body:    body,
		}
	}

	for _, tt := range []struct {
		name    string
		stmt    ir.Stmt
		withBox bool
		code    string
	}{
		{"target_kind", ir.Stmt{Kind: "assign_property", Line: 2, Target: variable("box"), Expr: validValue}, true, "IR002"},
		{"target_missing_property", ir.Stmt{Kind: "assign_property", Line: 2, Target: ir.Expr{Kind: "property", Line: 2, Args: []ir.Expr{variable("box")}}, Expr: validValue}, true, "IR002"},
		{"target_params", ir.Stmt{Kind: "assign_property", Line: 2, Target: ir.Expr{Kind: "property", Text: "items", Line: 2, Params: []string{}, Args: []ir.Expr{variable("box")}}, Expr: validValue}, true, "IR002"},
		{"target_wrong_arity", ir.Stmt{Kind: "assign_property", Line: 2, Target: ir.Expr{Kind: "property", Text: "items", Line: 2, Args: []ir.Expr{variable("box"), variable("other")}}, Expr: validValue}, true, "IR002"},
		{"nested_receiver", ir.Stmt{Kind: "assign_property", Line: 2, Target: ir.Expr{Kind: "property", Text: "items", Line: 2, Args: []ir.Expr{{Kind: "property", Text: "items", Line: 2, Args: []ir.Expr{variable("box")}}}}, Expr: validValue}, true, "IR002"},
		{"receiver_extra_args", ir.Stmt{Kind: "assign_property", Line: 2, Target: ir.Expr{Kind: "property", Text: "items", Line: 2, Args: []ir.Expr{{Kind: "var", Text: "box", Line: 2, Args: []ir.Expr{}}}}, Expr: validValue}, true, "IR002"},
		{"legacy_name", ir.Stmt{Kind: "assign_property", Name: "box", Line: 2, Target: validTarget, Expr: validValue}, true, "IR002"},
		{"key_payload", ir.Stmt{Kind: "assign_property", Key: "key", Line: 2, Target: validTarget, Expr: validValue}, true, "IR002"},
		{"value_payload", ir.Stmt{Kind: "assign_property", Value: "value", Line: 2, Target: validTarget, Expr: validValue}, true, "IR002"},
		{"body_payload", ir.Stmt{Kind: "assign_property", Line: 2, Target: validTarget, Expr: validValue, Body: []ir.Stmt{}}, true, "IR002"},
		{"else_payload", ir.Stmt{Kind: "assign_property", Line: 2, Target: validTarget, Expr: validValue, ElseBody: []ir.Stmt{}}, true, "IR002"},
		{"unknown_receiver", valid, false, "PHP004"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			out, err := Generate(program(tt.stmt, tt.withBox))
			if err == nil || len(out) != 0 || !strings.Contains(err.Error(), tt.code) {
				t.Fatalf("malformed property write generated output %s: %v", out, err)
			}
		})
	}
}
