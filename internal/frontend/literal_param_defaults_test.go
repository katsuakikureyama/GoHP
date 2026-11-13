package frontend

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestLiteralParameterDefaults(t *testing.T) {
	target := testTarget(t)

	for name, tc := range map[string]struct {
		source       string
		wantDefaults []ir.Expr
		wantOptional int
	}{
		"null_default": {
			source:       `function f($x = null) {}`,
			wantDefaults: []ir.Expr{{Kind: "null", Text: "null"}},
			wantOptional: 1,
		},
		"true_default": {
			source:       `function f($x = true) {}`,
			wantDefaults: []ir.Expr{{Kind: "bool", Text: "true"}},
			wantOptional: 1,
		},
		"false_default": {
			source:       `function f($x = false) {}`,
			wantDefaults: []ir.Expr{{Kind: "bool", Text: "false"}},
			wantOptional: 1,
		},
		"TRUE_uppercase": {
			source:       `function f($x = TRUE) {}`,
			wantDefaults: []ir.Expr{{Kind: "bool", Text: "true"}},
			wantOptional: 1,
		},
		"FALSE_uppercase": {
			source:       `function f($x = FALSE) {}`,
			wantDefaults: []ir.Expr{{Kind: "bool", Text: "false"}},
			wantOptional: 1,
		},
		"NULL_uppercase": {
			source:       `function f($x = NULL) {}`,
			wantDefaults: []ir.Expr{{Kind: "null", Text: "null"}},
			wantOptional: 1,
		},
		"int_default_zero": {
			source:       `function f($x = 0) {}`,
			wantDefaults: []ir.Expr{{Kind: "int", Text: "0"}},
			wantOptional: 1,
		},
		"int_default_positive": {
			source:       `function f($x = 10) {}`,
			wantDefaults: []ir.Expr{{Kind: "int", Text: "10"}},
			wantOptional: 1,
		},
		"string_single_quoted": {
			source:       `function f($x = 'hello') {}`,
			wantDefaults: []ir.Expr{{Kind: "string", Text: "hello"}},
			wantOptional: 1,
		},
		"string_double_quoted": {
			source:       `function f($x = "hello") {}`,
			wantDefaults: []ir.Expr{{Kind: "string", Text: "hello"}},
			wantOptional: 1,
		},
		"string_empty": {
			source:       `function f($x = '') {}`,
			wantDefaults: []ir.Expr{{Kind: "string", Text: ""}},
			wantOptional: 1,
		},
		"empty_array_brackets": {
			source:       `function f($x = []) {}`,
			wantDefaults: []ir.Expr{{Kind: "array"}},
			wantOptional: 1,
		},
		"empty_array_func": {
			source:       `function f($x = array()) {}`,
			wantDefaults: []ir.Expr{{Kind: "array"}},
			wantOptional: 1,
		},
		"mixed_defaults": {
			source: `function f($a, $b = 1, $c = 'text', $d = false, $e = null) {}`,
			wantDefaults: []ir.Expr{
				{},
				{Kind: "int", Text: "1"},
				{Kind: "string", Text: "text"},
				{Kind: "bool", Text: "false"},
				{Kind: "null", Text: "null"},
			},
			wantOptional: 4,
		},
		"constructor_defaults": {
			source: `class C { public function __construct($x = 1, $y = 'a') {} }`,
			wantDefaults: []ir.Expr{
				{Kind: "int", Text: "1"},
				{Kind: "string", Text: "a"},
			},
			wantOptional: 2,
		},
		"method_defaults": {
			source:       `class C { public function m($x = true) {} }`,
			wantDefaults: []ir.Expr{{Kind: "bool", Text: "true"}},
			wantOptional: 1,
		},
		"static_method_defaults": {
			source:       `class C { public static function m($x = 0) {} }`,
			wantDefaults: []ir.Expr{{Kind: "int", Text: "0"}},
			wantOptional: 1,
		},
	} {
		t.Run(name, func(t *testing.T) {
			p, err := Parse("<?php "+tc.source, target)
			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}
			var fn ir.Function
			if len(p.Functions) > 0 {
				fn = p.Functions[0]
			} else if len(p.Classes) > 0 {
				fn = p.Classes[0].Methods[0]
			}
			if fn.OptionalNullParams != tc.wantOptional {
				t.Fatalf("OptionalNullParams = %d, want %d", fn.OptionalNullParams, tc.wantOptional)
			}
			if len(fn.ParamDefaults) != len(tc.wantDefaults) {
				t.Fatalf("ParamDefaults len = %d, want %d", len(fn.ParamDefaults), len(tc.wantDefaults))
			}
			for i, want := range tc.wantDefaults {
				got := fn.ParamDefaults[i]
				if got.Kind != want.Kind {
					t.Fatalf("ParamDefaults[%d].Kind = %q, want %q", i, got.Kind, want.Kind)
				}
				if got.Text != want.Text {
					t.Fatalf("ParamDefaults[%d].Text = %q, want %q", i, got.Text, want.Text)
				}
			}
		})
	}
}

func TestLiteralParameterDefaultsRejections(t *testing.T) {
	target := testTarget(t)

	for name, source := range map[string]string{
		"negative_int":            `function f($x = -1) {}`,
		"float":                   `function f($x = 1.5) {}`,
		"constant":                `function f($x = E_ERROR) {}`,
		"expression":              `function f($x = 1 + 2) {}`,
		"nonempty_array":          `function f($x = [1]) {}`,
		"nonempty_array_func":     `function f($x = array(1)) {}`,
		"new":                     `function f($x = new C()) {}`,
		"required_after_optional": `function f($x = null, $y) {}`,
		"var_default":             `function f($x = $y) {}`,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := Parse("<?php "+source, target)
			if err == nil {
				t.Fatal("expected rejection, got nil")
			}
			if !strings.Contains(err.Error(), "PHP00") {
				t.Fatalf("expected PHP00 error, got: %v", err)
			}
		})
	}
}
