package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestBlockDataConstructorLoweringAndSourceMap(t *testing.T) {
	const source = `<?php
class Bucket {
  public $implicit;
  public $label = 'default';
  public function __CONSTRUCT($label, $items) {
    $this->label = $label;
    $this->implicit = $items;
  }
}
$first = new Bucket('first', ['a', 'b']);
$second = new Bucket('second', ['c']);
$shape = (array)$first;
$count = count($shape);
foreach ($shape as $key => $value) { echo $key; }
echo $first->label;
`
	p, err := frontend.Parse(source, testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	property := p.Classes[0].Properties[0]
	if property.Name != "implicit" || property.Value.Kind != "null" || property.Value.Text != "null" {
		t.Fatalf("implicit public property = %#v, want canonical null default", property)
	}
	constructor := p.Classes[0].Methods[0]
	if !strings.EqualFold(constructor.Name, "__construct") || len(constructor.Params) != 2 || constructor.OptionalNullParams != 0 || constructor.ReturnType != "" {
		t.Fatalf("constructor IR = %#v, want fixed untyped __construct", constructor)
	}
	if first := p.Body[0].Expr; first.Kind != "new" || first.Text != "Bucket" || len(first.Args) != 2 {
		t.Fatalf("first construction IR = %#v, want new Bucket(args...)", first)
	}
	if cast := p.Body[2].Expr; cast.Kind != "object_array_cast" || cast.Text != "" || len(cast.Args) != 1 || cast.Args[0].Kind != "var" || cast.Args[0].Text != "first" {
		t.Fatalf("object cast IR = %#v, want object_array_cast($first)", cast)
	}

	generated, sourceMap, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	text := string(generated)
	for _, want := range []string{
		`func New_bucket(args ...any) any`,
		`rt.Args(args, 2)`,
		`self.M___construct(args...)`,
		`func (self *C_bucket) M___construct(args ...any) any`,
		`v["label"] = rt.Clone(args[0])`,
		`v["items"] = rt.Clone(args[1])`,
		`value := any(v["label"])`,
		`target := rt.ResolveProperty(v["this"], "label")`,
		`value := any(v["items"])`,
		`target := rt.ResolveProperty(v["this"], "implicit")`,
		`target.Assign(value)`,
		`func (self *C_bucket) PHPProperties() rt.Array`,
		`rt.ObjectToArray(v["first"])`,
		`rt.Count(v["shape"])`,
		`rt.Entries(v["shape"])`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("constructor/cast lowering missing %q:\n%s", want, generated)
		}
	}
	if strings.Count(text, `rt.ObjectToArray(v["first"])`) != 1 {
		t.Fatalf("object cast receiver must be evaluated once:\n%s", generated)
	}
	first := strings.Index(text, `New_bucket("first", rt.NewArray(`)
	second := strings.Index(text, `New_bucket("second", rt.NewArray(`)
	if first < 0 || second < first {
		t.Fatalf("fresh instances must each call the generated constructor:\n%s", generated)
	}
	implicit := strings.Index(text, `Key: "implicit"`)
	label := strings.Index(text, `Key: "label"`)
	if implicit < 0 || label < implicit || !strings.Contains(text, `Value: self.P_implicit`) || !strings.Contains(text, `Value: self.P_label`) {
		t.Fatalf("object-to-array properties must be declaration ordered values passed through NewArray cloning:\n%s", generated)
	}
	for line, generatedLine := range strings.Split(text, "\n") {
		if strings.Contains(generatedLine, `v["shape"] = rt.Clone(rt.ObjectToArray(v["first"]))`) {
			if got, ok := sourceMap.PHPLine(line + 1); !ok || got != 12 {
				t.Fatalf("object cast maps to PHP line %d, %v; want 12", got, ok)
			}
			return
		}
	}
	t.Fatalf("object cast was not source-mapped:\n%s", generated)
}

func TestBlockDataConstructorBoundaries(t *testing.T) {
	for name, source := range map[string]string{
		"no_constructor_arguments":     `class Box { public $value; } $box = new Box('value');`,
		"constructor_missing_arg":      `class Box { public function __construct($value) {} } $box = new Box();`,
		"constructor_extra_arg":        `class Box { public function __construct($value) {} } $box = new Box('one', 'two');`,
		"unknown_class":                `$box = new Missing();`,
		"typed_constructor":            `class Box { public function __construct(array $value) {} }`,
		"non_null_default_constructor": `class Box { public function __construct($value = false) {} }`,
		"constructor_return_type":      `class Box { public function __construct(): bool {} }`,
		"constructor_return_value":     `class Box { public function __construct() { return 1; } }`,
		"other_magic_method":           `class Box { public function __clone() {} }`,
		"direct_constructor_reentry":   `class Box { public function __construct() {} } $box = new Box(); $box->__construct();`,
		"inheritance":                  `class Child extends ParentBox {}`,
		"cast_unknown_parameter":       `function copy($value) { $array = (array)$value; }`,
		"cast_array":                   `$value = []; $array = (array)$value;`,
		"cast_scalar":                  `$value = 'text'; $array = (array)$value;`,
		"cast_callable":                `$value = fn($item) => $item; $array = (array)$value;`,
	} {
		t.Run(name, func(t *testing.T) {
			p, err := frontend.Parse("<?php "+source, testTarget(t))
			var out []byte
			if err == nil {
				out, err = Generate(p)
			}
			if err == nil || !strings.Contains(err.Error(), "PHP00") {
				t.Fatalf("unsupported constructor/cast source was accepted: %v", err)
			}
			if len(out) != 0 {
				t.Fatalf("rejected constructor/cast source unexpectedly generated output:\n%s", out)
			}
		})
	}
}

func TestBlockDataConstructorMalformedIR(t *testing.T) {
	variable := func(name string) ir.Expr { return ir.Expr{Kind: "var", Text: name, Line: 3} }
	canonicalNull := ir.Expr{Kind: "null", Text: "null", Line: 1}
	constructor := ir.Function{
		Name:   "__construct",
		Params: []string{"value"},
		Line:   2,
		Body: []ir.Stmt{{
			Kind: "assign_property", Line: 3,
			Target: ir.Expr{Kind: "property", Text: "slot", Line: 3, Args: []ir.Expr{variable("this")}},
			Expr:   variable("value"),
		}},
	}
	program := func(methods []ir.Function, body []ir.Stmt) ir.Program {
		return ir.Program{
			Version: ir.Version, PHPVersion: testTarget(t),
			Classes: []ir.Class{{Name: "Box", Properties: []ir.Property{{Name: "slot", Value: canonicalNull}}, Methods: methods}},
			Body:    body,
		}
	}
	newBox := func(args ...ir.Expr) ir.Expr { return ir.Expr{Kind: "new", Text: "Box", Line: 3, Args: args} }
	assignment := func(name string, line int, expr ir.Expr) ir.Stmt {
		return ir.Stmt{Kind: "assign", Name: name, Line: line, Expr: expr}
	}
	cast := func(arg ir.Expr) ir.Stmt {
		return assignment("shape", 4, ir.Expr{Kind: "object_array_cast", Line: 4, Args: []ir.Expr{arg}})
	}
	construct := assignment("box", 3, newBox(ir.Expr{Kind: "string", Text: "value", Line: 3}))
	validCast := cast(variable("box"))
	if out, err := Generate(program([]ir.Function{constructor}, []ir.Stmt{construct, validCast})); err != nil || len(out) == 0 {
		t.Fatalf("canonical constructor/cast IR was rejected: %v\n%s", err, out)
	}

	for _, tt := range []struct {
		name    string
		methods []ir.Function
		body    []ir.Stmt
		code    string
	}{
		{"new_missing_args", []ir.Function{constructor}, []ir.Stmt{assignment("box", 3, newBox())}, "PHP004"},
		{"new_extra_args", []ir.Function{constructor}, []ir.Stmt{assignment("box", 3, newBox(variable("box"), variable("box")))}, "PHP004"},
		{"new_empty_class", []ir.Function{constructor}, []ir.Stmt{assignment("box", 3, ir.Expr{Kind: "new", Line: 3, Args: []ir.Expr{{Kind: "string", Text: "value", Line: 3}}})}, "IR002"},
		{"new_params", []ir.Function{constructor}, []ir.Stmt{assignment("box", 3, ir.Expr{Kind: "new", Text: "Box", Line: 3, Params: []string{}, Args: []ir.Expr{{Kind: "string", Text: "value", Line: 3}}})}, "IR002"},
		{"constructor_invalid_default_count", []ir.Function{{Name: "__construct", Params: []string{"value"}, OptionalNullParams: 2, Line: 2}}, []ir.Stmt{}, "IR002"},
		{"constructor_typed", []ir.Function{{Name: "__construct", Params: []string{"value"}, ParamTypes: []string{"array"}, Line: 2}}, []ir.Stmt{}, "PHP004"},
		{"constructor_return_type", []ir.Function{{Name: "__construct", ReturnType: "bool", Line: 2}}, []ir.Stmt{}, "PHP004"},
		{"constructor_return_value", []ir.Function{{Name: "__construct", Line: 2, Body: []ir.Stmt{{Kind: "return", Line: 3, Expr: ir.Expr{Kind: "int", Text: "1", Line: 3}}}}}, []ir.Stmt{}, "PHP004"},
		{"duplicate_constructor", []ir.Function{constructor, {Name: "__CONSTRUCT", Line: 4}}, []ir.Stmt{}, "PHP004"},
		{"cast_text", []ir.Function{constructor}, []ir.Stmt{construct, assignment("shape", 4, ir.Expr{Kind: "object_array_cast", Text: "array", Line: 4, Args: []ir.Expr{variable("box")}})}, "IR002"},
		{"cast_wrong_arity", []ir.Function{constructor}, []ir.Stmt{construct, assignment("shape", 4, ir.Expr{Kind: "object_array_cast", Line: 4, Args: []ir.Expr{variable("box"), variable("box")}})}, "IR002"},
		{"cast_params", []ir.Function{constructor}, []ir.Stmt{construct, assignment("shape", 4, ir.Expr{Kind: "object_array_cast", Line: 4, Params: []string{}, Args: []ir.Expr{variable("box")}})}, "IR002"},
		{"cast_unknown", []ir.Function{constructor}, []ir.Stmt{cast(variable("missing"))}, "PHP004"},
		{"cast_scalar", []ir.Function{constructor}, []ir.Stmt{cast(ir.Expr{Kind: "string", Text: "text", Line: 4})}, "PHP004"},
		{"cast_array", []ir.Function{constructor}, []ir.Stmt{cast(ir.Expr{Kind: "array", Line: 4})}, "PHP004"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			out, err := Generate(program(tt.methods, tt.body))
			if err == nil || len(out) != 0 || !strings.Contains(err.Error(), tt.code) {
				t.Fatalf("malformed constructor/cast IR generated output %s: %v", out, err)
			}
		})
	}

}
