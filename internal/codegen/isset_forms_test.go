package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestIssetFormsLowering(t *testing.T) {
	const source = `<?php
function capture($value) { return $value; }
function probe($key, $items, $third) { return capture(IsSeT($key, $items[$key], $third)); }
$key = 'present';
$items = ['present' => false];
$third = '';
if (isset($key, $items[$key])) { echo 'if:'; }
$logical = false && isset($key, $items[$key]);
$ternary = isset($key, $items[$key]) ? 'yes' : 'no';
echo probe($key, $items, $third);
`
	p, err := frontend.Parse(source, "8.3")
	if err != nil {
		t.Fatal(err)
	}

	probe := p.Functions[1].Body[0].Expr
	if probe.Kind != "call" || len(probe.Args) != 1 {
		t.Fatalf("probe return did not retain capture call: %#v", probe)
	}
	multi := probe.Args[0]
	if multi.Kind != "isset_multi" || multi.Text != "" || len(multi.Args) != 3 {
		t.Fatalf("multi isset IR = %#v, want three address operands", multi)
	}
	if multi.Args[0].Kind != "isset_var" || multi.Args[1].Kind != "isset_index" || multi.Args[2].Kind != "isset_var" {
		t.Fatalf("multi isset operand forms = %#v", multi.Args)
	}
	if got := p.Body[3].Expr.Kind; got != "isset_multi" {
		t.Fatalf("two-address isset IR kind = %q, want isset_multi", got)
	}
	if got := p.Body[4].Expr.Kind; got != "binary" {
		t.Fatalf("logical isset expression IR kind = %q, want binary", got)
	}

	generated, sourceMap, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	text := string(generated)
	for _, want := range []string{
		`rt.IssetValue(v["key"]) && rt.ArrayIsset(v["items"], v["key"]) && rt.IssetValue(v["third"])`,
		`if rt.Bool((rt.IssetValue(v["key"]) && rt.ArrayIsset(v["items"], v["key"]))) {`,
		`rt.Bool(false) && rt.Bool((rt.IssetValue(v["key"]) && rt.ArrayIsset(v["items"], v["key"])))`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("generated Go omitted canonical isset lowering %q:\n%s", want, generated)
		}
	}
	if strings.Contains(text, `rt.Get(v["items"], v["key"])`) {
		t.Fatalf("isset indexed address lowered through an eager read boundary:\n%s", generated)
	}
	for line, generatedLine := range strings.Split(text, "\n") {
		if strings.Contains(generatedLine, `if rt.Bool((rt.IssetValue(v["key"]) && rt.ArrayIsset(v["items"], v["key"]))) {`) {
			if got, ok := sourceMap.PHPLine(line + 1); !ok || got != 7 {
				t.Fatalf("multi isset if maps to PHP line %d (%v), want 7", got, ok)
			}
			return
		}
	}
	t.Fatalf("multi isset if lowering was not source-mapped:\n%s", generated)
}

func TestIssetFormsBoundary(t *testing.T) {
	for _, tt := range []struct {
		name string
		src  string
	}{
		{"no_operands", "<?php echo isset();"},
		{"literal_operand", "<?php echo isset(null);"},
		{"call_operand", "<?php function source() { return null; } echo isset(source());"},
		{"nested_index", "<?php $items = [[]]; echo isset($items[0]['child']);"},
		{"property", "<?php class Item { public $value = null; } $item = new Item(); echo isset($item->value);"},
		{"parenthesized_operand", "<?php $value = 1; echo isset(($value));"},
		{"parenthesized_key", "<?php $items = []; $key = 'key'; echo isset($items[($key)]);"},
		{"undefined_bare", "<?php echo isset($missing);"},
		{"undefined_multi_operand", "<?php $value = 1; echo isset($value, $missing);"},
		{"dead_undefined_multi_operand", "<?php $value = 1; $items = []; if (false) { echo isset($value, $items[$missing]); }"},
		{"known_scalar_root", "<?php $value = 'scalar'; echo isset($value['key']);"},
		{"known_scalar_key", "<?php $items = []; $key = false; echo isset($items[$key]);"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p, err := frontend.Parse(tt.src, "8.3")
			if err == nil {
				_, err = Generate(p)
			}
			if err == nil || !strings.Contains(err.Error(), "PHP00") {
				t.Fatalf("unsupported isset form was accepted: %v", err)
			}
		})
	}

	unknown := "<?php function probe($key, $items) { return isset($key, $items[$key]); } echo probe(null, 'scalar');"
	p, err := frontend.Parse(unknown, "8.3")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Generate(p); err != nil {
		t.Fatalf("defined unknown operands must reach their runtime guards: %v", err)
	}

	value := ir.Expr{Kind: "var", Text: "value", Line: 2}
	items := ir.Expr{Kind: "var", Text: "items", Line: 2}
	key := ir.Expr{Kind: "string", Text: "key", Line: 2}
	for _, tt := range []struct {
		name string
		expr ir.Expr
	}{
		{"isset_var_no_operand", ir.Expr{Kind: "isset_var", Line: 2}},
		{"isset_var_extra_operand", ir.Expr{Kind: "isset_var", Line: 2, Args: []ir.Expr{value, value}}},
		{"isset_var_non_variable", ir.Expr{Kind: "isset_var", Line: 2, Args: []ir.Expr{{Kind: "null", Line: 2}}}},
		{"isset_var_unexpected_text", ir.Expr{Kind: "isset_var", Text: "isset", Line: 2, Args: []ir.Expr{value}}},
		{"isset_multi_one_operand", ir.Expr{Kind: "isset_multi", Line: 2, Args: []ir.Expr{{Kind: "isset_var", Line: 2, Args: []ir.Expr{value}}}}},
		{"isset_multi_text", ir.Expr{Kind: "isset_multi", Text: "isset", Line: 2, Args: []ir.Expr{{Kind: "isset_var", Line: 2, Args: []ir.Expr{value}}, {Kind: "isset_var", Line: 2, Args: []ir.Expr{value}}}}},
		{"isset_multi_non_address", ir.Expr{Kind: "isset_multi", Line: 2, Args: []ir.Expr{{Kind: "isset_var", Line: 2, Args: []ir.Expr{value}}, value}}},
		{"isset_multi_undefined_operand", ir.Expr{Kind: "isset_multi", Line: 2, Args: []ir.Expr{{Kind: "isset_var", Line: 2, Args: []ir.Expr{value}}, {Kind: "isset_index", Line: 2, Args: []ir.Expr{items, key}}}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := ir.Program{Version: ir.Version, PHPVersion: "8.3", Body: []ir.Stmt{
				{Kind: "assign", Name: "value", Line: 1, Expr: ir.Expr{Kind: "int", Text: "1", Line: 1}},
				{Kind: "echo", Line: 2, Expr: tt.expr},
			}}
			if _, err := Generate(p); err == nil || !strings.Contains(err.Error(), "IR002") && !strings.Contains(err.Error(), "PHP004") {
				t.Fatalf("malformed isset IR was accepted: %v", err)
			}
		})
	}
}
