package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestArraySetParsesAndLowersToRuntimeBoundary(t *testing.T) {
	p, err := frontend.Parse("<?php $items = ['first' => 'one']; $key = 'first'; $items[$key] = 'updated'; echo $items['first'];", "8.3")
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Body) != 4 || p.Body[2].Kind != "array_set" || p.Body[2].Name != "items" || p.Body[2].Index == nil || p.Body[2].Index.Kind != "var" || p.Body[2].Index.Text != "key" {
		t.Fatalf("array-element assignment was not represented as the narrow array_set IR: %#v", p.Body)
	}
	generated, err := Generate(p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(generated), "rt.ArraySet(") {
		t.Fatalf("generated Go did not use the runtime array element write boundary:\n%s", generated)
	}
}

func TestArraySetKnownUnsupportedSourceFailsBeforeGeneration(t *testing.T) {
	for _, tt := range []struct {
		name string
		src  string
	}{
		{"undefined_base", "<?php $items['key'] = 'value';"},
		{"scalar_base", "<?php $items = 1; $items['key'] = 'value';"},
		{"boolean_key", "<?php $items = []; $items[true] = 'value';"},
		{"null_key", "<?php $items = []; $items[null] = 'value';"},
		{"array_key", "<?php $items = []; $items[[]] = 'value';"},
		{"append", "<?php $items = []; $items[] = 'value';"},
		{"nested_target", "<?php $items = ['nested' => []]; $items['nested']['key'] = 'value';"},
		{"property_target", "<?php class Box { public $items = null; } $box = new Box(); $box->items['key'] = 'value';"},
		{"compound_assignment", "<?php $items = ['key' => 1]; $items['key'] += 1;"},
		{"reference_assignment", "<?php $items = []; $value = 'value'; $items['key'] = &$value;"},
		{"call_target", "<?php function items() { return []; } items()['key'] = 'value';"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p, err := frontend.Parse(tt.src, "8.3")
			if err == nil {
				_, err = Generate(p)
			}
			if err == nil || !strings.Contains(err.Error(), "PHP00") {
				t.Fatalf("unsupported array-element write was not rejected with a diagnostic: %v", err)
			}
		})
	}
}

func TestArraySetMalformedIRIsRejectedBeforeGeneration(t *testing.T) {
	emptyArray := ir.Expr{Kind: "array", Line: 1}
	validIndex := ir.Expr{Kind: "string", Text: "key", Line: 1}
	validValue := ir.Expr{Kind: "string", Text: "value", Line: 1}
	initialized := ir.Stmt{Kind: "assign", Name: "items", Line: 1, Expr: emptyArray}
	for _, tt := range []struct {
		name string
		body []ir.Stmt
		code string
	}{
		{
			name: "missing_index",
			body: []ir.Stmt{{Kind: "array_set", Name: "items", Line: 1, Expr: validValue}},
			code: "IR002",
		},
		{
			name: "unexpected_body",
			body: []ir.Stmt{initialized, {Kind: "array_set", Name: "items", Line: 1, Index: &validIndex, Expr: validValue, Body: []ir.Stmt{{Kind: "echo", Line: 1, Expr: validValue}}}},
			code: "IR002",
		},
		{
			name: "unexpected_key_field",
			body: []ir.Stmt{initialized, {Kind: "array_set", Name: "items", Line: 1, Index: &validIndex, Expr: validValue, Key: "key"}},
			code: "IR002",
		},
		{
			name: "unexpected_value_field",
			body: []ir.Stmt{initialized, {Kind: "array_set", Name: "items", Line: 1, Index: &validIndex, Expr: validValue, Value: "value"}},
			code: "IR002",
		},
		{
			name: "reserved_base",
			body: []ir.Stmt{{Kind: "array_set", Name: "GLOBALS", Line: 1, Index: &validIndex, Expr: validValue}},
			code: "PHP004",
		},
		{
			name: "undefined_base",
			body: []ir.Stmt{{Kind: "array_set", Name: "items", Line: 1, Index: &validIndex, Expr: validValue}},
			code: "PHP004",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := ir.Program{Version: ir.Version, PHPVersion: "8.3", Body: tt.body}
			if _, err := Generate(p); err == nil || !strings.Contains(err.Error(), tt.code) {
				t.Fatalf("malformed array_set IR was not rejected safely: %v", err)
			}
		})
	}
}
