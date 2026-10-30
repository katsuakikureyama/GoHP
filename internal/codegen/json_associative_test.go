package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestJSONAssociativeLoweringAndSourceMap(t *testing.T) {
	const source = `<?php
function decode($text) { return JsOn_DeCoDe($text, true); }
$decoded = json_decode('{"flag":false}', true);
$code = json_last_error();
$message = json_last_error_msg();
$callback = fn($value) => json_decode($value, true);
$results = array_map('decode', ['null', 'false']);
class Box {
    public function decode($method_text) { return json_decode($method_text, true); }
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
	text := string(generated)
	for fragment, wantLine := range map[string]int{
		`return rt.Clone(execution.JSONDecodeAssociative(v["text"]))`: 2,
		`v["decoded"] = rt.Clone(execution.JSONDecodeAssociative(`:    3,
		`v["code"] = rt.Clone(execution.JSONLastError())`:             4,
		`v["message"] = rt.Clone(execution.JSONLastErrorMessage())`:   5,
		`return execution.JSONDecodeAssociative(v["value"])`:          6,
		`return FContext_decode(execution, value)`:                    7,
	} {
		found := false
		for line, generatedLine := range strings.Split(text, "\n") {
			if !strings.Contains(generatedLine, fragment) {
				continue
			}
			found = true
			if got, ok := mapping.PHPLine(line + 1); !ok || got != wantLine {
				t.Fatalf("%q maps to %d/%v, want %d", fragment, got, ok, wantLine)
			}
		}
		if !found {
			t.Fatalf("missing %q:\n%s", fragment, generated)
		}
	}
	for _, fragment := range []string{
		`return FContext_decode(&rt.Execution{}, args...)`,
		`rt.NewCallableWithExecution`,
		`MContext_decode(execution *rt.Execution`,
	} {
		if !strings.Contains(text, fragment) {
			t.Fatalf("missing shared/isolated execution routing %q:\n%s", fragment, generated)
		}
	}
	if strings.Contains(generatedFunction(t, text, "FContext_decode"), "NewScalarFrame") {
		t.Fatal("mixed JSON values must retain legacy storage")
	}
}

func TestJSONAssociativeUnsupportedSource(t *testing.T) {
	for name, source := range map[string]string{
		"no_args":           "echo json_decode();",
		"one_arg":           "echo json_decode('null');",
		"object_mode":       "echo json_decode('{}', false);",
		"nullable_mode":     "echo json_decode('{}', null);",
		"dynamic_mode":      "$mode = true; echo json_decode('{}', $mode);",
		"depth":             "echo json_decode('{}', true, 32);",
		"flags":             "echo json_decode('{}', true, 512, 0);",
		"input_null":        "echo json_decode(null, true);",
		"input_bool":        "echo json_decode(false, true);",
		"input_int":         "echo json_decode(1, true);",
		"input_array":       "echo json_decode([], true);",
		"input_object":      "class Box {} echo json_decode(new Box(), true);",
		"input_callable":    "echo json_decode(fn() => '{}', true);",
		"undefined":         "echo json_decode($missing, true);",
		"dead_invalid":      "echo true || json_decode([], true);",
		"error_arity":       "echo json_last_error(1);",
		"message_arity":     "echo json_last_error_msg(1);",
		"named":             "echo json_decode(json: '{}', associative: true);",
		"unpack":            "echo json_decode(...['{}', true]);",
		"redeclare":         "function JsOn_DeCoDe($a,$b) { return $a; }",
		"redeclare_error":   "function JsOn_LaSt_ErRoR() { return 0; }",
		"redeclare_message": "function JsOn_LaSt_ErRoR_MsG() { return ''; }",
		"builtin_callback":  "$values = array_map('json_decode', ['{}']);",
	} {
		t.Run(name, func(t *testing.T) {
			p, err := frontend.Parse("<?php "+source, testTarget(t))
			var generated []byte
			if err == nil {
				generated, err = Generate(p)
			}
			if err == nil || len(generated) != 0 || !strings.Contains(err.Error(), "PHP00") {
				t.Fatalf("unsupported JSON source generated %s: %v", generated, err)
			}
		})
	}
}

func TestJSONAssociativeMalformedIR(t *testing.T) {
	valid := ir.Expr{Kind: "call", Text: "json_decode", Line: 2, Args: []ir.Expr{
		{Kind: "string", Text: "null", Line: 2}, {Kind: "bool", Text: "true", Line: 2},
	}}
	for name, mutate := range map[string]func(*ir.Expr){
		"hidden_params": func(e *ir.Expr) { e.Params = []string{} },
		"hidden_input":  func(e *ir.Expr) { e.Args[0].Args = []ir.Expr{{Kind: "int", Text: "0"}} },
		"hidden_mode":   func(e *ir.Expr) { e.Args[1].Params = []string{} },
		"invalid_bool":  func(e *ir.Expr) { e.Args[1].Text = "truth" },
	} {
		t.Run(name, func(t *testing.T) {
			e := valid
			e.Args = append([]ir.Expr{}, valid.Args...)
			mutate(&e)
			p := ir.Program{Version: ir.Version, PHPVersion: testTarget(t), Body: []ir.Stmt{{Kind: "assign", Name: "decoded", Line: 2, Expr: e}}}
			generated, err := Generate(p)
			if err == nil || len(generated) != 0 || !strings.Contains(err.Error(), "IR002") {
				t.Fatalf("malformed JSON IR generated %s: %v", generated, err)
			}
		})
	}
}
