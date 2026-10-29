package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestUnknownPropertyWriteLowering(t *testing.T) {
	const parameterSource = `<?php
function tick() { return 'next'; }
function mutate($receiver) {
    $receiver->value = tick();
    $receiver->items[] = 'next';
}
`
	p, err := frontend.Parse(parameterSource, testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	out, mapping, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	code := string(out)
	body := generatedFunction(t, code, "F_mutate")
	for _, want := range []string{
		`value := any(F_tick())`,
		`target := rt.ResolveProperty(v["receiver"], "value")`,
		`target.Assign(value)`,
		`value := any("next")`,
		`target := rt.ResolveProperty(v["receiver"], "items")`,
		`target.Append(value)`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("unknown property write lowering missing %q:\n%s", want, body)
		}
	}
	if got := strings.Count(body, "F_tick()"); got != 1 {
		t.Fatalf("property-write RHS call count = %d, want 1:\n%s", got, body)
	}
	if strings.Index(body, "value := any(F_tick())") > strings.Index(body, `target := rt.ResolveProperty(v["receiver"], "value")`) {
		t.Fatalf("property-write RHS must precede writable-slot resolution:\n%s", body)
	}
	requireMappedPropertyWrite(t, code, mapping, `rt.ResolveProperty(v["receiver"], "value")`, 4)
	requireMappedPropertyWrite(t, code, mapping, `rt.ResolveProperty(v["receiver"], "items")`, 5)

	const valueSources = `<?php
class Box { public $value = null; }
class Factory {
    public function uncertain($flag) {
        if ($flag) { return new Box(); }
        return new Box();
    }
}
function uncertain_return($flag) {
    if ($flag) { return new Box(); }
    return new Box();
}
function from_array() {
    $items = [new Box()];
    $item = $items[0];
    $item->value = 'array';
}
function from_results($factory) {
    $from_method = $factory->uncertain(true);
    $from_return = uncertain_return(true);
    $from_method->value = 'method';
    $from_return->value = 'return';
}
function widened($flag) {
    $box = new Box();
    if ($flag) {
        $box = [];
    } else {
        $box = new Box();
    }
    $box->value = 'widened';
}
`
	p, err = frontend.Parse(valueSources, testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	out, mapping, err = GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	code = string(out)
	for _, tc := range []struct {
		function, fragment string
		line               int
	}{
		{"F_from_array", `rt.ResolveProperty(v["item"], "value")`, 16},
		{"F_from_results", `rt.ResolveProperty(v["from_method"], "value")`, 21},
		{"F_from_results", `rt.ResolveProperty(v["from_return"], "value")`, 22},
		{"F_widened", `rt.ResolveProperty(v["box"], "value")`, 31},
	} {
		body := generatedFunction(t, code, tc.function)
		if !strings.Contains(body, tc.fragment) {
			t.Fatalf("%s did not lower an unknown receiver through the runtime property protocol:\n%s", tc.function, body)
		}
		requireMappedPropertyWrite(t, code, mapping, tc.fragment, tc.line)
	}
	widened := generatedFunction(t, code, "F_widened")
	if !strings.Contains(widened, "v := map[string]any{}") || strings.Contains(widened, "NewScalarFrame") || strings.Contains(widened, "rt.ValueSlot") {
		t.Fatalf("control-flow-widened receiver must use guarded legacy storage:\n%s", widened)
	}
}

func TestUnknownPropertyWriteUnsupportedSource(t *testing.T) {
	for name, source := range map[string]string{
		"known_scalar":                 `$box = 1; $box->value = 'after';`,
		"known_array":                  `$box = []; $box->value = 'after';`,
		"known_null":                   `$box = null; $box->value = 'after';`,
		"known_callable":               `$box = fn() => null; $box->value = 'after';`,
		"undefined":                    `$box->value = 'after';`,
		"global":                       `$GLOBALS->value = 'after';`,
		"dynamic_property":             `class Box { public $value = null; } $box = new Box(); $name = 'value'; $box->$name = 'after';`,
		"index_receiver":               `class Box { public $value = null; } $boxes = [new Box()]; $boxes[0]->value = 'after';`,
		"call_receiver":                `function build() { return null; } build()->value = 'after';`,
		"missing_known_property":       `class Box { public $value = null; } $box = new Box(); $box->missing = 'after';`,
		"conditional_only_unknown":     `if (true) { $box = null; } $box->value = 'after';`,
		"direct_concat_remains_closed": `function f($box) { $box->value .= 'after'; }`,
	} {
		t.Run(name, func(t *testing.T) {
			p, err := frontend.Parse("<?php "+source, testTarget(t))
			var out []byte
			if err == nil {
				out, err = Generate(p)
			}
			if err == nil || len(out) != 0 {
				t.Fatalf("unsupported unknown property write was accepted: %s: %v", out, err)
			}
		})
	}
}

func TestUnknownPropertyWriteMalformedIR(t *testing.T) {
	variable := func(name string) ir.Expr { return ir.Expr{Kind: "var", Text: name, Line: 2} }
	validTarget := ir.Expr{Kind: "property", Text: "value", Line: 2, Args: []ir.Expr{variable("receiver")}}
	valid := ir.Stmt{Kind: "assign_property", Line: 2, Target: validTarget, Expr: ir.Expr{Kind: "string", Text: "after", Line: 2}}
	program := func(stmt ir.Stmt) ir.Program {
		return ir.Program{
			Version:    ir.Version,
			PHPVersion: testTarget(t),
			Functions: []ir.Function{{
				Name: "mutate", Line: 1, Params: []string{"receiver"}, Body: []ir.Stmt{stmt},
			}},
		}
	}

	for _, tt := range []struct {
		name string
		stmt ir.Stmt
	}{
		{"target_kind", ir.Stmt{Kind: "assign_property", Line: 2, Target: variable("receiver"), Expr: valid.Expr}},
		{"target_missing_name", ir.Stmt{Kind: "assign_property", Line: 2, Target: ir.Expr{Kind: "property", Line: 2, Args: []ir.Expr{variable("receiver")}}, Expr: valid.Expr}},
		{"target_hidden_params", ir.Stmt{Kind: "assign_property", Line: 2, Target: ir.Expr{Kind: "property", Text: "value", Line: 2, Params: []string{}, Args: []ir.Expr{variable("receiver")}}, Expr: valid.Expr}},
		{"target_extra_receiver", ir.Stmt{Kind: "assign_property", Line: 2, Target: ir.Expr{Kind: "property", Text: "value", Line: 2, Args: []ir.Expr{variable("receiver"), variable("other")}}, Expr: valid.Expr}},
		{"target_bad_name", ir.Stmt{Kind: "assign_property", Line: 2, Target: ir.Expr{Kind: "property", Text: "bad-name", Line: 2, Args: []ir.Expr{variable("receiver")}}, Expr: valid.Expr}},
		{"base_hidden_params", ir.Stmt{Kind: "assign_property", Line: 2, Target: ir.Expr{Kind: "property", Text: "value", Line: 2, Args: []ir.Expr{{Kind: "var", Text: "receiver", Line: 2, Params: []string{}}}}, Expr: valid.Expr}},
		{"base_hidden_args", ir.Stmt{Kind: "assign_property", Line: 2, Target: ir.Expr{Kind: "property", Text: "value", Line: 2, Args: []ir.Expr{{Kind: "var", Text: "receiver", Line: 2, Args: []ir.Expr{}}}}, Expr: valid.Expr}},
		{"base_call", ir.Stmt{Kind: "assign_property", Line: 2, Target: ir.Expr{Kind: "property", Text: "value", Line: 2, Args: []ir.Expr{{Kind: "call", Text: "build", Line: 2}}}, Expr: valid.Expr}},
		{"hidden_name", ir.Stmt{Kind: "assign_property", Name: "hidden", Line: 2, Target: validTarget, Expr: valid.Expr}},
		{"hidden_key", ir.Stmt{Kind: "assign_property", Key: "hidden", Line: 2, Target: validTarget, Expr: valid.Expr}},
		{"hidden_value", ir.Stmt{Kind: "assign_property", Value: "hidden", Line: 2, Target: validTarget, Expr: valid.Expr}},
		{"hidden_body", ir.Stmt{Kind: "assign_property", Line: 2, Target: validTarget, Expr: valid.Expr, Body: []ir.Stmt{}}},
		{"hidden_else_body", ir.Stmt{Kind: "assign_property", Line: 2, Target: validTarget, Expr: valid.Expr, ElseBody: []ir.Stmt{}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			out, err := Generate(program(tt.stmt))
			if err == nil || len(out) != 0 || !strings.Contains(err.Error(), "IR002") {
				t.Fatalf("malformed unknown-receiver property write generated %s: %v", out, err)
			}
		})
	}

	if out, err := Generate(program(valid)); err != nil || len(out) == 0 {
		t.Fatalf("valid unknown-parameter property write was rejected: %s %v", out, err)
	}
}

func requireMappedPropertyWrite(t *testing.T, code string, mapping SourceMap, fragment string, wantLine int) {
	t.Helper()
	for generatedLine, text := range strings.Split(code, "\n") {
		if !strings.Contains(text, fragment) {
			continue
		}
		if got, ok := mapping.PHPLine(generatedLine + 1); !ok || got != wantLine {
			t.Fatalf("property write source map for %q = %d,%v; want %d", fragment, got, ok, wantLine)
		}
		return
	}
	t.Fatalf("mapped property write %q missing:\n%s", fragment, code)
}
