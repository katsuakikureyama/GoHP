package semantic

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func testSemanticTarget(t *testing.T) string {
	t.Helper()
	return "8.3"
}

func TestLiteralParamDefaultsValidation(t *testing.T) {
	target := testSemanticTarget(t)

	// Valid: all admitted literal kinds
	for name, fn := range map[string]ir.Function{
		"null_default": {
			Name: "f", Params: []string{"x"}, OptionalNullParams: 1, Line: 1,
			ParamDefaults: []ir.Expr{{Kind: "null", Text: "null", Line: 1}},
		},
		"bool_true": {
			Name: "f", Params: []string{"x"}, OptionalNullParams: 1, Line: 1,
			ParamDefaults: []ir.Expr{{Kind: "bool", Text: "true", Line: 1}},
		},
		"bool_false": {
			Name: "f", Params: []string{"x"}, OptionalNullParams: 1, Line: 1,
			ParamDefaults: []ir.Expr{{Kind: "bool", Text: "false", Line: 1}},
		},
		"int_zero": {
			Name: "f", Params: []string{"x"}, OptionalNullParams: 1, Line: 1,
			ParamDefaults: []ir.Expr{{Kind: "int", Text: "0", Line: 1}},
		},
		"int_positive": {
			Name: "f", Params: []string{"x"}, OptionalNullParams: 1, Line: 1,
			ParamDefaults: []ir.Expr{{Kind: "int", Text: "10", Line: 1}},
		},
		"string_default": {
			Name: "f", Params: []string{"x"}, OptionalNullParams: 1, Line: 1,
			ParamDefaults: []ir.Expr{{Kind: "string", Text: "hello", Line: 1}},
		},
		"empty_array": {
			Name: "f", Params: []string{"x"}, OptionalNullParams: 1, Line: 1,
			ParamDefaults: []ir.Expr{{Kind: "array", Line: 1}},
		},
		"mixed_with_required": {
			Name: "f", Params: []string{"a", "b", "c"}, OptionalNullParams: 2, Line: 1,
			ParamDefaults: []ir.Expr{
				{},
				{Kind: "int", Text: "1", Line: 1},
				{Kind: "bool", Text: "false", Line: 1},
			},
		},
		"no_defaults": {
			Name: "f", Params: []string{"x"}, OptionalNullParams: 0, Line: 1,
		},
	} {
		t.Run(name, func(t *testing.T) {
			p := ir.Program{Version: ir.Version, PHPVersion: target, Functions: []ir.Function{fn}}
			if err := Validate(p); err != nil {
				t.Fatalf("valid IR rejected: %v", err)
			}
		})
	}
}

func TestLiteralParamDefaultsMalformedIR(t *testing.T) {
	target := testSemanticTarget(t)

	for name, fn := range map[string]ir.Function{
		"defaults_misaligned": {
			Name: "f", Params: []string{"x", "y"}, OptionalNullParams: 1, Line: 1,
			ParamDefaults: []ir.Expr{{Kind: "int", Text: "1", Line: 1}},
		},
		"required_has_default": {
			Name: "f", Params: []string{"x", "y"}, OptionalNullParams: 1, Line: 1,
			ParamDefaults: []ir.Expr{
				{Kind: "int", Text: "1", Line: 1},
				{Kind: "null", Text: "null", Line: 1},
			},
		},
		"negative_int_default": {
			Name: "f", Params: []string{"x"}, OptionalNullParams: 1, Line: 1,
			ParamDefaults: []ir.Expr{{Kind: "int", Text: "-1", Line: 1}},
		},
		"nonempty_array_default": {
			Name: "f", Params: []string{"x"}, OptionalNullParams: 1, Line: 1,
			ParamDefaults: []ir.Expr{{Kind: "array", Line: 1, Args: []ir.Expr{{Kind: "pair", Line: 1}}}},
		},
		"unsupported_default_kind": {
			Name: "f", Params: []string{"x"}, OptionalNullParams: 1, Line: 1,
			ParamDefaults: []ir.Expr{{Kind: "call", Text: "foo", Line: 1}},
		},
		"typed_param_with_default": {
			Name: "f", Params: []string{"x"}, ParamTypes: []string{"array"}, OptionalNullParams: 1, Line: 1,
			ParamDefaults: []ir.Expr{{Kind: "null", Text: "null", Line: 1}},
		},
		"negative_optional_count": {
			Name: "f", Params: []string{"x"}, OptionalNullParams: -1, Line: 1,
		},
		"excessive_optional_count": {
			Name: "f", Params: []string{"x"}, OptionalNullParams: 2, Line: 1,
		},
	} {
		t.Run(name, func(t *testing.T) {
			p := ir.Program{Version: ir.Version, PHPVersion: target, Functions: []ir.Function{fn}}
			err := Validate(p)
			if err == nil {
				t.Fatal("malformed IR accepted")
			}
			if !strings.Contains(err.Error(), "IR002") && !strings.Contains(err.Error(), "PHP00") {
				t.Fatalf("expected IR002/PHP00 error, got: %v", err)
			}
		})
	}
}
