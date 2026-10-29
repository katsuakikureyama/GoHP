package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestStringEmptyLoweringAndSourceMap(t *testing.T) {
	const source = `<?php
function has_html(string $html): bool {
    return !empty($html);
}
$html = substr('hello', 1);
if (!empty($html)) {
    echo 'text:';
}
$html = '';
$blank = empty($html);
$html = [];
$array_blank = empty($html);
$callback = fn($value) => !empty($value);
echo $callback('0') ? 'callback-empty' : 'callback-text';
`
	p, err := frontend.Parse(source, testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	generated, sourceMap, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	text := string(generated)
	for _, want := range []string{
		`return rt.BoolReturn((!rt.Bool(rt.VariableEmpty(v["html"]))))`,
		`if rt.Bool((!rt.Bool(rt.VariableEmpty(v["html"])))) {`,
		`v["blank"] = rt.Clone(rt.VariableEmpty(v["html"]))`,
		`v["array_blank"] = rt.Clone(rt.VariableEmpty(v["html"]))`,
		`rt.NewCallable(1, func(args ...any) any {`,
		`return (!rt.Bool(rt.VariableEmpty(v["value"])))`,
		`rt.Invoke(v["callback"], "0")`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("string empty lowering omitted %q:\n%s", want, generated)
		}
	}
	if got := strings.Count(generatedFunction(t, text, "F_has_html"), `rt.VariableEmpty(v["html"])`); got != 1 {
		t.Fatalf("typed string helper has %d empty reads, want one:\n%s", got, generated)
	}
	for fragment, wantLine := range map[string]int{
		`return rt.BoolReturn((!rt.Bool(rt.VariableEmpty(v["html"]))))`: 3,
		`if rt.Bool((!rt.Bool(rt.VariableEmpty(v["html"])))) {`:         6,
		`v["blank"] = rt.Clone(rt.VariableEmpty(v["html"]))`:            10,
		`v["array_blank"] = rt.Clone(rt.VariableEmpty(v["html"]))`:      12,
		`rt.NewCallable(1, func(args ...any) any {`:                     13,
		`rt.Invoke(v["callback"], "0")`:                                 14,
	} {
		found := false
		for line, generatedLine := range strings.Split(text, "\n") {
			if !strings.Contains(generatedLine, fragment) {
				continue
			}
			found = true
			if got, ok := sourceMap.PHPLine(line + 1); !ok || got != wantLine {
				t.Fatalf("%q maps to PHP line %d (%v), want %d", fragment, got, ok, wantLine)
			}
			break
		}
		if !found {
			t.Fatalf("source-mapped lowering missing %q:\n%s", fragment, generated)
		}
	}
}

func TestStringEmptyUnsupportedSource(t *testing.T) {
	for name, source := range map[string]string{
		"null":             "$value = null; echo empty($value);",
		"bool":             "$value = false; echo empty($value);",
		"integer":          "$value = 0; echo empty($value);",
		"object":           "class Box {} $value = new Box(); echo empty($value);",
		"closure":          "$value = fn() => ''; echo empty($value);",
		"undefined":        "echo empty($value);",
		"conditional_only": "if (true) { $value = ''; } echo empty($value);",
		"property":         "class Box { public $value = ''; } $box = new Box(); echo empty($box->value);",
		"index":            "$values = ['value']; echo empty($values[0]);",
		"literal":          "echo empty('');",
		"call":             "function source() { return ''; } echo empty(source());",
		"extra_arguments":  "$value = ''; echo empty($value, $value);",
		"global":           "echo empty($GLOBALS);",
		"redeclare":        "function EmPtY($value) { return $value; }",
		"dead_invalid":     "$value = null; echo true || empty($value);",
	} {
		t.Run(name, func(t *testing.T) {
			p, err := frontend.Parse("<?php "+source, testTarget(t))
			var output []byte
			if err == nil {
				output, err = Generate(p)
			}
			if err == nil || len(output) != 0 || !strings.Contains(err.Error(), "PHP00") {
				t.Fatalf("unsupported string empty source generated output: %s; %v", output, err)
			}
		})
	}
}

func TestStringEmptyMalformedIR(t *testing.T) {
	variable := ir.Expr{Kind: "var", Text: "value", Line: 2}
	for name, expr := range map[string]ir.Expr{
		"missing_operand":      {Kind: "empty_var", Line: 2},
		"extra_operand":        {Kind: "empty_var", Line: 2, Args: []ir.Expr{variable, variable}},
		"unexpected_text":      {Kind: "empty_var", Text: "empty", Line: 2, Args: []ir.Expr{variable}},
		"unexpected_params":    {Kind: "empty_var", Params: []string{}, Line: 2, Args: []ir.Expr{variable}},
		"var_operands":         {Kind: "empty_var", Line: 2, Args: []ir.Expr{{Kind: "var", Text: "value", Line: 2, Args: []ir.Expr{variable}}}},
		"var_params":           {Kind: "empty_var", Line: 2, Args: []ir.Expr{{Kind: "var", Text: "value", Line: 2, Params: []string{}}}},
		"invalid_var_name":     {Kind: "empty_var", Line: 2, Args: []ir.Expr{{Kind: "var", Text: "not-valid", Line: 2}}},
		"non_variable_literal": {Kind: "empty_var", Line: 2, Args: []ir.Expr{{Kind: "string", Text: "", Line: 2}}},
		"non_variable_call":    {Kind: "empty_var", Line: 2, Args: []ir.Expr{{Kind: "call", Text: "source", Line: 2}}},
	} {
		t.Run(name, func(t *testing.T) {
			output, err := Generate(stringEmptyProgram(testTarget(t), expr))
			if err == nil || len(output) != 0 || !strings.Contains(err.Error(), "IR002") {
				t.Fatalf("malformed empty_var generated output: %s; %v", output, err)
			}
		})
	}
	for name, expr := range map[string]ir.Expr{
		"generic_empty_call":  {Kind: "call", Text: "EmPtY", Line: 2, Args: []ir.Expr{variable}},
		"generic_empty_extra": {Kind: "call", Text: "empty", Line: 2, Args: []ir.Expr{variable, variable}},
	} {
		t.Run(name, func(t *testing.T) {
			output, err := Generate(stringEmptyProgram(testTarget(t), expr))
			if err == nil || len(output) != 0 || !strings.Contains(err.Error(), "IR002") {
				t.Fatalf("generic empty call generated output: %s; %v", output, err)
			}
		})
	}
}

func TestLegacyArrayEmptyIR(t *testing.T) {
	array := ir.Expr{Kind: "var", Text: "items", Line: 2}
	legacyArray := ir.Program{Version: ir.Version, PHPVersion: testTarget(t), Body: []ir.Stmt{
		{Kind: "assign", Name: "items", Line: 1, Expr: ir.Expr{Kind: "array", Line: 1}},
		{Kind: "echo", Line: 2, Expr: ir.Expr{Kind: "empty_array", Line: 2, Args: []ir.Expr{array}}},
	}}
	output, err := Generate(legacyArray)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(output), `rt.ArrayEmpty(v["items"])`) || strings.Contains(string(output), "rt.VariableEmpty") {
		t.Fatalf("legacy array empty changed lowering:\n%s", output)
	}

	legacyString := ir.Program{Version: ir.Version, PHPVersion: testTarget(t), Body: []ir.Stmt{
		{Kind: "assign", Name: "html", Line: 1, Expr: ir.Expr{Kind: "string", Text: "", Line: 1}},
		{Kind: "echo", Line: 2, Expr: ir.Expr{Kind: "empty_array", Line: 2, Args: []ir.Expr{{Kind: "var", Text: "html", Line: 2}}}},
	}}
	if output, err := Generate(legacyString); err == nil || len(output) != 0 || !strings.Contains(err.Error(), "PHP004") {
		t.Fatalf("legacy string empty was accepted: %s; %v", output, err)
	}

	parsed, err := frontend.Parse("<?php $html = ''; echo empty($html);", testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	if got := parsed.Body[1].Expr.Kind; got != "empty_var" {
		t.Fatalf("frontend string empty IR kind = %q, want empty_var", got)
	}
	output, err = Generate(parsed)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(output), `rt.VariableEmpty(v["html"])`) || strings.Contains(string(output), "rt.ArrayEmpty") {
		t.Fatalf("frontend string empty did not use the new lowering:\n%s", output)
	}
}

func stringEmptyProgram(target string, expr ir.Expr) ir.Program {
	return ir.Program{Version: ir.Version, PHPVersion: target, Body: []ir.Stmt{
		{Kind: "assign", Name: "value", Line: 1, Expr: ir.Expr{Kind: "string", Text: "", Line: 1}},
		{Kind: "echo", Line: 2, Expr: expr},
	}}
}
