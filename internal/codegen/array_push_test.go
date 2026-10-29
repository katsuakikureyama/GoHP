package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestArrayPushLoweringAndSourceMap(t *testing.T) {
	const source = `<?php
function tick($value) { return $value; }
class Box { public $items; public function __construct() { $this->items = []; } }
$items = [];
array_push($items);
$count = ARRAY_PUSH($items, tick('first'), tick('last'));
$box = new Box();
$propertyCount = array_push($box->items, tick('property'));
if (true) {
  $ifCount = array_push($items, 'conditional');
} else {
  $elseCount = array_push($items, 'otherwise');
}
while (false) {
  $loopCount = array_push($items, 'loop');
  break;
}
$alias = [];
$alias = array_push($alias, 'alias');
function local_lane() { $items = []; $count = array_push($items, 'local'); return $count; }
`
	p, err := frontend.Parse(source, testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	var pushes []ir.Stmt
	var visit func([]ir.Stmt)
	visit = func(stmts []ir.Stmt) {
		for _, stmt := range stmts {
			if stmt.Kind == "array_push" {
				pushes = append(pushes, stmt)
			}
			visit(stmt.Body)
			visit(stmt.ElseBody)
		}
	}
	visit(p.Body)
	if len(pushes) != 7 {
		t.Fatalf("array_push statements = %d, want 7: %#v", len(pushes), p.Body)
	}
	for _, stmt := range pushes {
		if stmt.Expr.Kind != "call" || stmt.Expr.Text != "array_push" || stmt.Expr.Params != nil {
			t.Fatalf("array_push must be normalized to its canonical call payload: %#v", stmt)
		}
	}
	if pushes[0].Name != "" || pushes[0].Target.Kind != "var" || pushes[0].Target.Text != "items" || len(pushes[0].Expr.Args) != 0 {
		t.Fatalf("zero-argument array_push IR = %#v", pushes[0])
	}
	if pushes[1].Name != "count" || pushes[1].Target.Kind != "var" || len(pushes[1].Expr.Args) != 2 {
		t.Fatalf("case-insensitive local array_push IR = %#v", pushes[1])
	}
	if pushes[2].Name != "propertyCount" || pushes[2].Target.Kind != "property" || pushes[2].Target.Text != "items" || len(pushes[2].Target.Args) != 1 || pushes[2].Target.Args[0].Kind != "var" || pushes[2].Target.Args[0].Text != "box" {
		t.Fatalf("property array_push IR = %#v", pushes[2])
	}

	generated, mapping, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	code := string(generated)
	for fragment, wantLine := range map[string]int{
		`_ = rt.ArrayPushLocal(v, "items")`:                                  5,
		`v["count"] = rt.ArrayPushLocal(v, "items", pushValue0, pushValue1)`: 6,
		`v["propertyCount"] = target.Push(pushValue0)`:                       8,
		`v["ifCount"] = rt.ArrayPushLocal(v, "items", pushValue0)`:           10,
		`v["elseCount"] = rt.ArrayPushLocal(v, "items", pushValue0)`:         12,
		`v["loopCount"] = rt.ArrayPushLocal(v, "items", pushValue0)`:         15,
		`v["alias"] = rt.ArrayPushLocal(v, "alias", pushValue0)`:             19,
	} {
		arrayPushRequireMapped(t, code, mapping, fragment, wantLine)
	}
	if got := strings.Count(code, `rt.Clone(F_tick(`); got != 3 {
		t.Fatalf("array_push payload effects = %d, want 3 evaluated exactly once:\n%s", got, code)
	}
	resolve := strings.Index(code, `target := rt.ResolveProperty(v["box"], "items")`)
	propertyTick := strings.Index(code, `pushValue0 := rt.Clone(F_tick("property"))`)
	propertyPush := strings.Index(code, `v["propertyCount"] = target.Push(pushValue0)`)
	if resolve < 0 || propertyTick < resolve || propertyPush < propertyTick {
		t.Fatalf("property target must resolve before payloads and push after payloads:\n%s", code)
	}
	if strings.Count(code, `v["alias"] = rt.ArrayPushLocal(v, "alias", pushValue0)`) != 1 {
		t.Fatalf("result/target alias must publish the count last:\n%s", code)
	}
	local := generatedFunction(t, code, "F_local_lane")
	if !strings.Contains(local, "v := map[string]any{}") || strings.Contains(local, "NewScalarFrame") || strings.Contains(local, "rt.ValueSlot") {
		t.Fatalf("array_push must retain the guarded legacy local scope:\n%s", local)
	}
}

func TestArrayPushUnsupportedSource(t *testing.T) {
	for name, source := range map[string]string{
		"missing_target":     `array_push();`,
		"known_nonarray":     `$items = false; array_push($items);`,
		"undefined":          `array_push($items);`,
		"reserved_target":    `$GLOBALS = []; array_push($GLOBALS);`,
		"literal_target":     `array_push([], 'value');`,
		"call_target":        `function make() { return []; } array_push(make(), 'value');`,
		"index_target":       `$items = [[]]; array_push($items[0], 'value');`,
		"nested_property":    `class Box { public $child = null; } $box = new Box(); array_push($box->child->items, 'value');`,
		"dynamic_property":   `class Box { public $items; } $box = new Box(); $name = 'items'; array_push($box->$name, 'value');`,
		"named_argument":     `$items = []; array_push(array: $items, value: 'value');`,
		"unpacked_argument":  `$items = []; array_push(...$items);`,
		"reference_argument": `$items = []; $value = 'value'; array_push($items, &$value);`,
		"builtin_redeclare":  `function ARRAY_PUSH($items) { return 0; }`,
		"embedded_echo":      `$items = []; echo array_push($items, 'value');`,
		"embedded_return":    `function f($items) { return array_push($items, 'value'); }`,
		"embedded_condition": `$items = []; if (array_push($items, 'value')) { echo 'bad'; }`,
		"embedded_argument":  `function f($value) {} $items = []; f(array_push($items, 'value'));`,
		"property_result":    `class Box { public $count; } $box = new Box(); $items = []; $box->count = array_push($items, 'value');`,
		"bool_result_guard":  `function f(): bool { $items = []; $count = array_push($items); return $count; }`,
		"dead_invalid":       `while (false) { array_push(false); break; }`,
		"array_pop_remains":  `$items = []; array_pop($items);`,
	} {
		t.Run(name, func(t *testing.T) {
			p, err := frontend.Parse("<?php "+source, testTarget(t))
			var out []byte
			if err == nil {
				out, err = Generate(p)
			}
			if err == nil || len(out) != 0 || !strings.Contains(err.Error(), "PHP00") {
				t.Fatalf("unsupported array_push source generated output %s: %v", out, err)
			}
		})
	}
}

func TestArrayPushMalformedIR(t *testing.T) {
	variable := func(name string) ir.Expr { return ir.Expr{Kind: "var", Text: name, Line: 2} }
	value := ir.Expr{Kind: "string", Text: "value", Line: 2}
	valid := ir.Stmt{
		Kind: "array_push", Line: 2, Target: variable("items"),
		Expr: ir.Expr{Kind: "call", Text: "array_push", Line: 2, Args: []ir.Expr{value}},
	}
	program := func(stmt ir.Stmt, initial bool) ir.Program {
		body := []ir.Stmt{stmt}
		if initial {
			body = append([]ir.Stmt{{Kind: "assign", Name: "items", Line: 1, Expr: ir.Expr{Kind: "array", Line: 1}}}, body...)
		}
		return ir.Program{Version: ir.Version, PHPVersion: testTarget(t), Body: body}
	}
	if out, err := Generate(program(valid, true)); err != nil || len(out) == 0 {
		t.Fatalf("valid canonical array_push IR was rejected: %s; %v", out, err)
	}

	for _, tt := range []struct {
		name    string
		stmt    ir.Stmt
		initial bool
		code    string
	}{
		{"text", ir.Stmt{Kind: "array_push", Line: 2, Target: variable("items"), Expr: ir.Expr{Kind: "call", Text: "Array_Push", Line: 2}}, true, "IR002"},
		{"params", ir.Stmt{Kind: "array_push", Line: 2, Target: variable("items"), Expr: ir.Expr{Kind: "call", Text: "array_push", Line: 2, Params: []string{}}}, true, "IR002"},
		{"target_kind", ir.Stmt{Kind: "array_push", Line: 2, Target: ir.Expr{Kind: "array", Line: 2}, Expr: valid.Expr}, true, "PHP004"},
		{"target_args", ir.Stmt{Kind: "array_push", Line: 2, Target: ir.Expr{Kind: "var", Text: "items", Line: 2, Args: []ir.Expr{value}}, Expr: valid.Expr}, true, "IR002"},
		{"stmt_name", ir.Stmt{Kind: "array_push", Name: "GLOBALS", Line: 2, Target: variable("items"), Expr: valid.Expr}, true, "PHP004"},
		{"key", ir.Stmt{Kind: "array_push", Key: "hidden", Line: 2, Target: variable("items"), Expr: valid.Expr}, true, "IR002"},
		{"value", ir.Stmt{Kind: "array_push", Value: "hidden", Line: 2, Target: variable("items"), Expr: valid.Expr}, true, "IR002"},
		{"body", ir.Stmt{Kind: "array_push", Line: 2, Target: variable("items"), Expr: valid.Expr, Body: []ir.Stmt{}}, true, "IR002"},
		{"else_body", ir.Stmt{Kind: "array_push", Line: 2, Target: variable("items"), Expr: valid.Expr, ElseBody: []ir.Stmt{}}, true, "IR002"},
		{"value_metadata", ir.Stmt{Kind: "array_push", Line: 2, Target: variable("items"), Expr: ir.Expr{Kind: "call", Text: "array_push", Line: 2, Args: []ir.Expr{{Kind: "string", Text: "value", Line: 2, Params: []string{}}}}}, true, "IR002"},
		{"undefined_target", valid, false, "PHP004"},
		{"name_before_use", ir.Stmt{Kind: "array_push", Name: "result", Line: 2, Target: variable("items"), Expr: ir.Expr{Kind: "call", Text: "array_push", Line: 2, Args: []ir.Expr{variable("result")}}}, true, "PHP004"},
		{"raw_generic_call", ir.Stmt{Kind: "assign", Name: "count", Line: 2, Expr: ir.Expr{Kind: "call", Text: "array_push", Line: 2, Args: []ir.Expr{variable("items")}}}, true, "PHP004"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			out, err := Generate(program(tt.stmt, tt.initial))
			if err == nil || len(out) != 0 || !strings.Contains(err.Error(), tt.code) {
				t.Fatalf("malformed array_push IR generated output %s: %v", out, err)
			}
		})
	}
}

func arrayPushRequireMapped(t *testing.T, code string, mapping SourceMap, fragment string, wantLine int) {
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
