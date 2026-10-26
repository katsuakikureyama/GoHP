package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestArrayIssetLowersWithoutIndexReadAndPreservesExpressionContexts(t *testing.T) {
	const source = `<?php
function capture($value) { return $value; }
function probe($items, $key) { return capture(IsSeT($items[$key])); }
$items = ['present' => false];
$key = 'present';
if (isset($items[$key])) { echo 'if:'; }
$guard = false && isset($items[$key]);
echo probe($items, $key);
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
	if strings.Contains(text, "rt.Get(") {
		t.Fatalf("isset must not lower through the missing-key read boundary:\n%s", generated)
	}
	if got := strings.Count(text, "rt.ArrayIsset("); got != 3 {
		t.Fatalf("ArrayIsset lowering count = %d, want 3:\n%s", got, generated)
	}
	for _, want := range []string{
		`return rt.Clone(F_capture(rt.ArrayIsset(v["items"], v["key"])))`,
		`if rt.Bool(rt.ArrayIsset(v["items"], v["key"])) {`,
		`rt.Bool(false) && rt.Bool(rt.ArrayIsset(v["items"], v["key"]))`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("generated Go did not preserve isset expression context %q:\n%s", want, generated)
		}
	}
	for line, generatedLine := range strings.Split(text, "\n") {
		if strings.Contains(generatedLine, `if rt.Bool(rt.ArrayIsset(v["items"], v["key"])) {`) {
			if got, ok := sourceMap.PHPLine(line + 1); !ok || got != 6 {
				t.Fatalf("isset if maps to PHP line %d, %v; want 6", got, ok)
			}
			return
		}
	}
	t.Fatalf("isset if lowering was not source-mapped:\n%s", generated)
}

func TestArrayIssetStrictSourceFormsFailClosed(t *testing.T) {
	for _, tt := range []struct {
		name string
		src  string
	}{
		{"no_arguments", "<?php $items = []; echo isset();"},
		{"property", "<?php class Item { public $value = null; } $item = new Item(); echo isset($item->value);"},
		{"string_offset", "<?php echo isset('text'[0]);"},
		{"parenthesized_operand", "<?php $items = []; echo isset(($items)['key']);"},
		{"parenthesized_key", "<?php $items = []; $key = 'key'; echo isset($items[($key)]);"},
		{"undefined_root", "<?php echo isset($missing['key']);"},
		{"undefined_key", "<?php $items = []; echo isset($items[$key]);"},
		{"conditional_root", "<?php if (true) { $items = []; } echo isset($items['key']);"},
		{"conditional_key", "<?php $items = []; if (true) { $key = 'key'; } echo isset($items[$key]);"},
		{"effectful_key", "<?php function marker() { return 'key'; } $items = []; echo isset($items[marker()]);"},
		{"effectful_root", "<?php function source() { return []; } echo isset(source()['key']);"},
		{"known_non_array", "<?php $value = 'scalar'; echo isset($value['key']);"},
		{"known_non_key", "<?php $items = []; $key = true; echo isset($items[$key]);"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p, err := frontend.Parse(tt.src, testTarget(t))
			if err == nil {
				_, err = Generate(p)
			}
			if err == nil || !strings.Contains(err.Error(), "PHP00") {
				t.Fatalf("unsupported isset source was not rejected with a diagnostic: %v", err)
			}
		})
	}
}

func TestArrayIssetMalformedIRAndRedeclarationFailBeforeGeneration(t *testing.T) {
	items := ir.Expr{Kind: "var", Text: "items", Line: 2}
	key := ir.Expr{Kind: "string", Text: "key", Line: 2}
	array := ir.Expr{Kind: "array", Line: 1}
	for _, tt := range []struct {
		name string
		expr ir.Expr
		code string
	}{
		{"generic_call", ir.Expr{Kind: "call", Text: "isset", Line: 2, Args: []ir.Expr{items}}, "IR002"},
		{"missing_operand", ir.Expr{Kind: "isset_index", Line: 2, Args: []ir.Expr{items}}, "IR002"},
		{"extra_operand", ir.Expr{Kind: "isset_index", Line: 2, Args: []ir.Expr{items, key, key}}, "IR002"},
		{"non_variable_root", ir.Expr{Kind: "isset_index", Line: 2, Args: []ir.Expr{array, key}}, "IR002"},
		{"nested_root", ir.Expr{Kind: "isset_index", Line: 2, Args: []ir.Expr{{Kind: "index", Line: 2, Args: []ir.Expr{items, key}}, key}}, "IR002"},
		{"invalid_key", ir.Expr{Kind: "isset_index", Line: 2, Args: []ir.Expr{items, {Kind: "bool", Text: "true", Line: 2}}}, "PHP004"},
		{"unexpected_text", ir.Expr{Kind: "isset_index", Text: "isset", Line: 2, Args: []ir.Expr{items, key}}, "IR002"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := ir.Program{Version: ir.Version, PHPVersion: testTarget(t), Body: []ir.Stmt{
				{Kind: "assign", Name: "items", Line: 1, Expr: array},
				{Kind: "echo", Line: 2, Expr: tt.expr},
			}}
			if _, err := Generate(p); err == nil || !strings.Contains(err.Error(), tt.code) {
				t.Fatalf("malformed isset IR was not rejected safely: %v", err)
			}
		})
	}

	p, err := frontend.Parse("<?php function IsSeT($value) { return $value; }", testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Generate(p); err == nil || !strings.Contains(err.Error(), "PHP004") {
		t.Fatalf("isset redeclaration was not rejected safely: %v", err)
	}
}
