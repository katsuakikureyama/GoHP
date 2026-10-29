package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestArrayPopLoweringAndSourceMap(t *testing.T) {
	const source = `<?php
class Box { public $items; public function __construct() { $this->items = []; } }
$items = ['first', 'last'];
array_pop($items);
$value = ARRAY_POP($items);
$box = new Box();
$box->items = ['property'];
$propertyValue = array_pop($box->items);
if (true) {
  $ifValue = array_pop($items);
} else {
  $elseValue = array_pop($items);
}
while (false) {
  $loopValue = array_pop($items);
  break;
}
$alias = ['alias'];
$alias = array_pop($alias);
function local_lane() { $items = ['local']; $laneValue = array_pop($items); return $laneValue; }
function frame_read($items) { $frame = array_pop($items); return $frame->name; }
function bool_result($items): bool { $boolValue = array_pop($items); return $boolValue; }
`
	p, err := frontend.Parse(source, testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	var pops []ir.Stmt
	var visit func([]ir.Stmt)
	visit = func(stmts []ir.Stmt) {
		for _, stmt := range stmts {
			if stmt.Kind == "array_pop" {
				pops = append(pops, stmt)
			}
			visit(stmt.Body)
			visit(stmt.ElseBody)
		}
	}
	visit(p.Body)
	if len(pops) != 7 {
		t.Fatalf("array_pop statements = %d, want 7: %#v", len(pops), p.Body)
	}
	for _, stmt := range pops {
		if stmt.Expr.Kind != "call" || stmt.Expr.Text != "array_pop" || stmt.Expr.Args != nil || stmt.Expr.Params != nil {
			t.Fatalf("array_pop must use canonical zero-payload call IR: %#v", stmt)
		}
	}
	if pops[0].Name != "" || pops[0].Target.Kind != "var" || pops[0].Target.Text != "items" {
		t.Fatalf("standalone array_pop IR = %#v", pops[0])
	}
	if pops[1].Name != "value" || pops[1].Target.Kind != "var" || pops[1].Target.Text != "items" {
		t.Fatalf("case-insensitive local array_pop IR = %#v", pops[1])
	}
	if pops[2].Name != "propertyValue" || pops[2].Target.Kind != "property" || pops[2].Target.Text != "items" || len(pops[2].Target.Args) != 1 || pops[2].Target.Args[0].Kind != "var" || pops[2].Target.Args[0].Text != "box" {
		t.Fatalf("property array_pop IR = %#v", pops[2])
	}

	generated, mapping, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	code := string(generated)
	for fragment, wantLine := range map[string]int{
		`_ = rt.ArrayPopLocal(v, "items")`:              4,
		`v["value"] = rt.ArrayPopLocal(v, "items")`:     5,
		`v["propertyValue"] = target.Pop()`:             8,
		`v["ifValue"] = rt.ArrayPopLocal(v, "items")`:   10,
		`v["elseValue"] = rt.ArrayPopLocal(v, "items")`: 12,
		`v["loopValue"] = rt.ArrayPopLocal(v, "items")`: 15,
		`v["alias"] = rt.ArrayPopLocal(v, "alias")`:     19,
	} {
		arrayPopRequireMapped(t, code, mapping, fragment, wantLine)
	}
	if strings.Count(code, `v["alias"] = rt.ArrayPopLocal(v, "alias")`) != 1 {
		t.Fatalf("result/target alias must publish the popped value last:\n%s", code)
	}
	resolve := strings.Index(code, `target := rt.ResolveProperty(v["box"], "items")`)
	pop := strings.Index(code, `v["propertyValue"] = target.Pop()`)
	if resolve < 0 || pop < resolve {
		t.Fatalf("property target must be resolved before pop and result publication:\n%s", code)
	}
	local := generatedFunction(t, code, "F_local_lane")
	if !strings.Contains(local, "v := map[string]any{}") || strings.Contains(local, "NewScalarFrame") || strings.Contains(local, "rt.ValueSlot") {
		t.Fatalf("array_pop must retain the guarded legacy local scope:\n%s", local)
	}
}

func TestArrayPopUnsupportedSource(t *testing.T) {
	for name, source := range map[string]string{
		"missing_target":        `array_pop();`,
		"extra_argument":        `$items = []; array_pop($items, 'value');`,
		"known_nonarray":        `$items = false; array_pop($items);`,
		"undefined":             `array_pop($items);`,
		"reserved_target":       `$GLOBALS = []; array_pop($GLOBALS);`,
		"literal_target":        `array_pop([]);`,
		"call_target":           `function make() { return []; } array_pop(make());`,
		"index_target":          `$items = [[]]; array_pop($items[0]);`,
		"nested_property":       `class Box { public $child = null; } $box = new Box(); array_pop($box->child->items);`,
		"dynamic_property":      `class Box { public $items; } $box = new Box(); $name = 'items'; array_pop($box->$name);`,
		"named_argument":        `$items = []; array_pop(array: $items);`,
		"unpacked_argument":     `$items = []; array_pop(...$items);`,
		"reference_argument":    `$items = []; array_pop(&$items);`,
		"builtin_redeclare":     `function ARRAY_POP($items) { return null; }`,
		"embedded_echo":         `$items = []; echo array_pop($items);`,
		"embedded_return":       `function f($items) { return array_pop($items); }`,
		"embedded_condition":    `$items = []; if (array_pop($items)) { echo 'bad'; }`,
		"embedded_argument":     `function f($value) {} $items = []; f(array_pop($items));`,
		"arrow_embedded":        `$items = []; $f = fn() => array_pop($items);`,
		"property_result":       `class Box { public $value; } $box = new Box(); $items = []; $box->value = array_pop($items);`,
		"index_result":          `$items = []; $out = []; $out[0] = array_pop($items);`,
		"reserved_result":       `$items = []; $_value = array_pop($items);`,
		"dead_invalid":          `while (false) { array_pop(false); break; }`,
		"embedded_push_remains": `$items = []; echo array_push($items, 'value');`,
	} {
		t.Run(name, func(t *testing.T) {
			p, err := frontend.Parse("<?php "+source, testTarget(t))
			var out []byte
			if err == nil {
				out, err = Generate(p)
			}
			if err == nil || len(out) != 0 || !strings.Contains(err.Error(), "PHP00") {
				t.Fatalf("unsupported array_pop source generated output %s: %v", out, err)
			}
		})
	}
}

func TestArrayPopMalformedIR(t *testing.T) {
	variable := func(name string) ir.Expr { return ir.Expr{Kind: "var", Text: name, Line: 2} }
	valid := ir.Stmt{
		Kind: "array_pop", Line: 2, Target: variable("items"),
		Expr: ir.Expr{Kind: "call", Text: "array_pop", Line: 2},
	}
	program := func(stmt ir.Stmt, initial bool) ir.Program {
		body := []ir.Stmt{stmt}
		if initial {
			body = append([]ir.Stmt{{Kind: "assign", Name: "items", Line: 1, Expr: ir.Expr{Kind: "array", Line: 1}}}, body...)
		}
		return ir.Program{Version: ir.Version, PHPVersion: testTarget(t), Body: body}
	}
	if out, err := Generate(program(valid, true)); err != nil || len(out) == 0 {
		t.Fatalf("valid canonical array_pop IR was rejected: %s; %v", out, err)
	}

	value := ir.Expr{Kind: "string", Text: "hidden", Line: 2}
	for _, tt := range []struct {
		name    string
		stmt    ir.Stmt
		initial bool
		code    string
	}{
		{"text", ir.Stmt{Kind: "array_pop", Line: 2, Target: variable("items"), Expr: ir.Expr{Kind: "call", Text: "Array_Pop", Line: 2}}, true, "IR002"},
		{"args", ir.Stmt{Kind: "array_pop", Line: 2, Target: variable("items"), Expr: ir.Expr{Kind: "call", Text: "array_pop", Line: 2, Args: []ir.Expr{value}}}, true, "IR002"},
		{"empty_args_metadata", ir.Stmt{Kind: "array_pop", Line: 2, Target: variable("items"), Expr: ir.Expr{Kind: "call", Text: "array_pop", Line: 2, Args: []ir.Expr{}}}, true, "IR002"},
		{"params", ir.Stmt{Kind: "array_pop", Line: 2, Target: variable("items"), Expr: ir.Expr{Kind: "call", Text: "array_pop", Line: 2, Params: []string{}}}, true, "IR002"},
		{"target_kind", ir.Stmt{Kind: "array_pop", Line: 2, Target: ir.Expr{Kind: "array", Line: 2}, Expr: valid.Expr}, true, "PHP004"},
		{"target_args", ir.Stmt{Kind: "array_pop", Line: 2, Target: ir.Expr{Kind: "var", Text: "items", Line: 2, Args: []ir.Expr{value}}, Expr: valid.Expr}, true, "IR002"},
		{"property_root", ir.Stmt{Kind: "array_pop", Line: 2, Target: ir.Expr{Kind: "property", Text: "items", Line: 2, Args: []ir.Expr{{Kind: "call", Text: "root", Line: 2}}}, Expr: valid.Expr}, true, "PHP004"},
		{"stmt_name", ir.Stmt{Kind: "array_pop", Name: "GLOBALS", Line: 2, Target: variable("items"), Expr: valid.Expr}, true, "PHP004"},
		{"key", ir.Stmt{Kind: "array_pop", Key: "hidden", Line: 2, Target: variable("items"), Expr: valid.Expr}, true, "IR002"},
		{"value", ir.Stmt{Kind: "array_pop", Value: "hidden", Line: 2, Target: variable("items"), Expr: valid.Expr}, true, "IR002"},
		{"body", ir.Stmt{Kind: "array_pop", Line: 2, Target: variable("items"), Expr: valid.Expr, Body: []ir.Stmt{}}, true, "IR002"},
		{"else_body", ir.Stmt{Kind: "array_pop", Line: 2, Target: variable("items"), Expr: valid.Expr, ElseBody: []ir.Stmt{}}, true, "IR002"},
		{"target_metadata", ir.Stmt{Kind: "array_pop", Line: 2, Target: ir.Expr{Kind: "var", Text: "items", Line: 2, Params: []string{}}, Expr: valid.Expr}, true, "IR002"},
		{"undefined_target", valid, false, "PHP004"},
		{"raw_generic_call", ir.Stmt{Kind: "assign", Name: "value", Line: 2, Expr: ir.Expr{Kind: "call", Text: "array_pop", Line: 2, Args: []ir.Expr{variable("items")}}}, true, "PHP004"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			out, err := Generate(program(tt.stmt, tt.initial))
			if err == nil || len(out) != 0 || !strings.Contains(err.Error(), tt.code) {
				t.Fatalf("malformed array_pop IR generated output %s: %v", out, err)
			}
		})
	}
}

func arrayPopRequireMapped(t *testing.T, code string, mapping SourceMap, fragment string, wantLine int) {
	t.Helper()
	found := 0
	for line, text := range strings.Split(code, "\n") {
		if !strings.Contains(text, fragment) {
			continue
		}
		found++
		if got, ok := mapping.PHPLine(line + 1); !ok || got != wantLine {
			t.Fatalf("%q maps to PHP line %d, %v; want %d", fragment, got, ok, wantLine)
		}
	}
	if found != 1 {
		t.Fatalf("%q occurs %d times:\n%s", fragment, found, code)
	}
}
