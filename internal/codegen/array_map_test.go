package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestArrayMapAndAddSlashesLowerToRuntimeBoundaries(t *testing.T) {
	p, err := frontend.Parse("<?php $items = ['first' => \"O'Reilly\"]; $mapped = ArRaY_MaP('MaPpEr', $items); foreach ($mapped as $key => $value) { echo $key . ':' . $value; } function mapper($value) { return AdDsLaShEs($value); }", "8.3")
	if err != nil {
		t.Fatal(err)
	}
	generated, err := Generate(p)
	if err != nil {
		t.Fatal(err)
	}
	text := string(generated)
	for _, want := range []string{"rt.ArrayMap(func(value any) any { return F_mapper(value) }", "rt.AddSlashes("} {
		if !strings.Contains(text, want) {
			t.Fatalf("generated Go did not use runtime boundary %q:\n%s", want, generated)
		}
	}
}

func TestArrayMapAndAddSlashesKnownUnsupportedFormsFailBeforeGeneration(t *testing.T) {
	for _, tt := range []struct {
		name string
		src  string
	}{
		{"missing_array", "<?php function mapper($value) { return $value; } echo array_map('mapper');"},
		{"multiple_arrays", "<?php function mapper($value) { return $value; } echo array_map('mapper', [], []);"},
		{"dynamic_callback", "<?php function mapper($value) { return $value; } $callback = 'mapper'; echo array_map($callback, []);"},
		{"builtin_callback", "<?php echo array_map('strtolower', []);"},
		{"null_callback", "<?php function mapper($value) { return $value; } echo array_map(null, []);"},
		{"method_callback", "<?php class Mapper { public function map($value) { return $value; } } echo array_map(['Mapper', 'map'], []);"},
		{"closure_callback", "<?php echo array_map(function($value) { return $value; }, []);"},
		{"reference_array_argument", "<?php function mapper($value) { return $value; } $items = []; echo array_map('mapper', &$items);"},
		{"unknown_callback", "<?php echo array_map('missing', []);"},
		{"zero_argument_callback", "<?php function zero() { return 'x'; } echo array_map('zero', []);"},
		{"two_argument_callback", "<?php function pair($left, $right) { return $left; } echo array_map('pair', []);"},
		{"scalar_array_argument", "<?php function mapper($value) { return $value; } echo array_map('mapper', 'not array');"},
		{"addslashes_missing", "<?php echo addslashes();"},
		{"addslashes_extra", "<?php echo addslashes('x', 'y');"},
		{"addslashes_static_non_string", "<?php echo addslashes(1);"},
		{"addslashes_redeclare", "<?php function AdDsLaShEs($value) { return $value; }"},
		{"array_map_redeclare", "<?php function ArRaY_mAp($value) { return $value; }"},
		{"array_result_not_string_for_concat", "<?php function mapper($value) { return $value; } echo array_map('mapper', []) . 'x';"},
		{"array_result_not_string_for_lowercase", "<?php function mapper($value) { return $value; } echo strtolower(array_map('mapper', []));"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p, err := frontend.Parse(tt.src, "8.3")
			if err == nil {
				_, err = Generate(p)
			}
			if err == nil || !strings.Contains(err.Error(), "PHP00") {
				t.Fatalf("unsupported source was not rejected with a diagnostic: %v", err)
			}
		})
	}
}

func TestArrayMapWrongCallbackIRIsRejectedBeforeGeneration(t *testing.T) {
	p := ir.Program{
		Version: ir.Version, PHPVersion: "8.3",
		Functions: []ir.Function{{Name: "wrapper", Params: []string{"value"}, Line: 1, Body: []ir.Stmt{{
			Kind: "return", Line: 1, Expr: ir.Expr{Kind: "call", Text: "array_map", Line: 1, Args: []ir.Expr{
				{Kind: "int", Text: "1", Line: 1}, {Kind: "var", Text: "value", Line: 1},
			}},
		}}}},
	}
	if _, err := Generate(p); err == nil || !strings.Contains(err.Error(), "PHP004") {
		t.Fatalf("wrong array_map callback IR was not rejected safely: %v", err)
	}
}
