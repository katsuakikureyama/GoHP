package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestUnknownObjectArrayLowering(t *testing.T) {
	const source = `<?php
class C {
  public $child = null;
  public function make($value = null) {
    if ($value) { return new C(); }
    return new C();
  }
  public function passthrough($value) {
    return $value;
  }
  public function property() {
    return $this->child;
  }
}
function identity($value) {
  return $value;
}
function direct($value) {
  $shape = (array)$value;
}
function from_property(C $box) {
  $shape = (array)$box->child;
}
function from_nested(C $box) {
  $shape = (array)$box->child->child;
}
function from_array(array $items) {
  $item = $items[0];
  $shape = (array)$item;
}
function from_call($value) {
  $shape = (array)identity($value);
}
function from_method_property(C $box) {
  $shape = (array)$box->property();
}
function foreign_receiver($box) {
  $shape = (array)$box->make(true);
}
function known_conditional_factory(C $box) {
  $shape = (array)$box->make(true);
}
function passthrough_known(C $box) {
  $shape = (array)$box->passthrough(new C());
}
function known_class() {
  $box = new C();
  $shape = (array)$box;
}
`
	p, err := frontend.Parse(source, testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	generated, mapping, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	code := string(generated)

	for _, tt := range []struct {
		function string
		line     int
	}{
		{"F_direct", 19},
		{"F_from_property", 22},
		{"F_from_nested", 25},
		{"F_from_array", 29},
		{"F_from_call", 32},
		{"F_from_method_property", 35},
		{"F_foreign_receiver", 38},
		{"F_known_conditional_factory", 41},
		{"F_passthrough_known", 44},
		{"F_known_class", 48},
	} {
		body := generatedFunction(t, code, tt.function)
		if got := strings.Count(body, "rt.ObjectToArray("); got != 1 {
			t.Fatalf("%s ObjectToArray calls = %d, want 1:\n%s", tt.function, got, body)
		}
		requireMappedUnknownObjectArrayCast(t, code, mapping, tt.function, tt.line)
	}

	known := generatedFunction(t, code, "F_known_class")
	if !strings.Contains(known, `rt.ObjectToArray(v["box"])`) {
		t.Fatalf("known generated-class cast changed its lowering:\n%s", known)
	}
}

func TestUnknownObjectArrayUnsupportedSource(t *testing.T) {
	for name, source := range map[string]string{
		"known_string":     `$value = 'text'; $shape = (array)$value;`,
		"known_integer":    `$value = 1; $shape = (array)$value;`,
		"known_boolean":    `$value = false; $shape = (array)$value;`,
		"known_null":       `$value = null; $shape = (array)$value;`,
		"known_array":      `$value = []; $shape = (array)$value;`,
		"known_closure":    `$value = fn() => null; $shape = (array)$value;`,
		"undefined":        `$shape = (array)$missing;`,
		"global":           `$shape = (array)$GLOBALS;`,
		"missing_property": `class C { public $value = null; } $box = new C(); $shape = (array)$box->missing;`,
		"dynamic_property": `class C { public $value = null; } $box = new C(); $name = 'value'; $shape = (array)$box->$name;`,
		"private_property": `class C { private $value = null; } $box = new C(); $shape = (array)$box->value;`,
		"inheritance":      `class ParentBox { public $value = null; } class ChildBox extends ParentBox {} $box = new ChildBox(); $shape = (array)$box;`,
	} {
		t.Run(name, func(t *testing.T) {
			p, err := frontend.Parse("<?php "+source, testTarget(t))
			var out []byte
			if err == nil {
				out, err = Generate(p)
			}
			if err == nil || len(out) != 0 || !strings.Contains(err.Error(), "PHP00") {
				t.Fatalf("unsupported object-array source generated output: %s; %v", out, err)
			}
		})
	}
}

func TestUnknownObjectArrayMalformedIR(t *testing.T) {
	variable := func() ir.Expr { return ir.Expr{Kind: "var", Text: "value", Line: 2} }
	program := func(expr ir.Expr) ir.Program {
		return ir.Program{
			Version:    ir.Version,
			PHPVersion: testTarget(t),
			Functions: []ir.Function{{
				Name: "convert", Params: []string{"value"}, Line: 1,
				Body: []ir.Stmt{{Kind: "assign", Name: "shape", Line: 2, Expr: expr}},
			}},
		}
	}
	valid := ir.Expr{Kind: "object_array_cast", Line: 2, Args: []ir.Expr{variable()}}
	if out, err := Generate(program(valid)); err != nil || len(out) == 0 {
		t.Fatalf("valid unknown-parameter object-array cast was rejected: %s; %v", out, err)
	}

	for name, expr := range map[string]ir.Expr{
		"text":             {Kind: "object_array_cast", Text: "array", Line: 2, Args: []ir.Expr{variable()}},
		"params":           {Kind: "object_array_cast", Line: 2, Params: []string{}, Args: []ir.Expr{variable()}},
		"missing_operand":  {Kind: "object_array_cast", Line: 2},
		"extra_operand":    {Kind: "object_array_cast", Line: 2, Args: []ir.Expr{variable(), variable()}},
		"operand_metadata": {Kind: "object_array_cast", Line: 2, Args: []ir.Expr{{Kind: "var", Text: "value", Line: 2, Params: []string{}}}},
	} {
		t.Run(name, func(t *testing.T) {
			out, err := Generate(program(expr))
			if err == nil || len(out) != 0 || !strings.Contains(err.Error(), "IR002") {
				t.Fatalf("malformed unknown object-array IR generated output: %s; %v", out, err)
			}
		})
	}
}

func requireMappedUnknownObjectArrayCast(t *testing.T, code string, mapping SourceMap, function string, wantLine int) {
	t.Helper()
	inFunction := false
	for generatedLine, text := range strings.Split(code, "\n") {
		if strings.HasPrefix(text, "func "+function+"(") {
			inFunction = true
			continue
		}
		if inFunction && strings.HasPrefix(text, "func ") {
			break
		}
		if !inFunction || !strings.Contains(text, "rt.ObjectToArray(") {
			continue
		}
		if got, ok := mapping.PHPLine(generatedLine + 1); !ok || got != wantLine {
			t.Fatalf("%s object-array cast source map = %d, %v; want %d", function, got, ok, wantLine)
		}
		return
	}
	t.Fatalf("%s mapped object-array cast missing:\n%s", function, code)
}
