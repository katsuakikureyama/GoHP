package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestPHP84ArrayAllLoweringAndTargetBoundaries(t *testing.T) {
	const builtin = `<?php
function MaTcHeS($value, $key): bool { return $value === $key; }
echo ArRaY_AlL(['match'], 'mAtChEs');`

	php84, err := frontend.Parse(builtin, "8.4")
	if err != nil {
		t.Fatal(err)
	}
	output, err := Generate(php84)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"rt.ArrayAll(",
		"func(value, key any) any { return F_matches(value, key) }",
	} {
		if !strings.Contains(string(output), want) {
			t.Fatalf("PHP 8.4 array_all lowering missing %q:\n%s", want, output)
		}
	}

	php83, err := frontend.Parse(builtin, "8.3")
	if err != nil {
		t.Fatal(err)
	}
	if output, err := Generate(php83); err == nil || len(output) != 0 || !strings.Contains(err.Error(), "unsupported function array_all") {
		t.Fatalf("PHP 8.3 absent builtin generated %q: %v", output, err)
	}

	const userFunction = `<?php
function ArRaY_AlL($value) { return $value; }
echo array_all('user');`
	php83, err = frontend.Parse(userFunction, "8.3")
	if err != nil {
		t.Fatal(err)
	}
	output, err = Generate(php83)
	if err != nil || !strings.Contains(string(output), `F_array_all("user")`) || strings.Contains(string(output), "rt.ArrayAll(") {
		t.Fatalf("PHP 8.3 user function was not preserved: %v\n%s", err, output)
	}

	php84, err = frontend.Parse(userFunction, "8.4")
	if err != nil {
		t.Fatal(err)
	}
	if output, err = Generate(php84); err == nil || len(output) != 0 || !strings.Contains(err.Error(), "function name reserved in PHP 8.4: ArRaY_AlL") {
		t.Fatalf("PHP 8.4 user declaration generated %q: %v", output, err)
	}
}

func TestPHP84ArrayAllStatefulCallbackUsesContextAdapter(t *testing.T) {
	program, err := frontend.Parse(`<?php
function token($value, $key): bool { return strtok($value, ',') === 'hit'; }
echo array_all(['hit,next'], 'token');`, "8.4")
	if err != nil {
		t.Fatal(err)
	}
	output, err := Generate(program)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"rt.ArrayAll(",
		"FContext_token(execution, value, key)",
	} {
		if !strings.Contains(string(output), want) {
			t.Fatalf("stateful array_all lowering missing %q:\n%s", want, output)
		}
	}
}

func TestPHP84ArrayAllUnsupportedSourceFailsBeforeGeneration(t *testing.T) {
	for name, source := range map[string]string{
		"missing_array":      `function matches($value, $key): bool { return true; } echo array_all('not array', 'matches');`,
		"missing_callback":   `function matches($value, $key): bool { return true; } echo array_all([]);`,
		"extra_argument":     `function matches($value, $key): bool { return true; } echo array_all([], 'matches', null);`,
		"dynamic_callback":   `function matches($value, $key): bool { return true; } $callback = 'matches'; echo array_all([], $callback);`,
		"builtin_callback":   `echo array_all([], 'is_string');`,
		"method_callback":    `class Matcher { public function test($value, $key): bool { return true; } } echo array_all([], ['Matcher', 'test']);`,
		"arrow_callback":     `echo array_all([], fn($value, $key) => true);`,
		"one_parameter":      `function matches($value): bool { return true; } echo array_all([], 'matches');`,
		"three_parameters":   `function matches($value, $key, $extra): bool { return true; } echo array_all([], 'matches');`,
		"optional_parameter": `function matches($value, $key = null): bool { return true; } echo array_all([], 'matches');`,
		"typed_parameter":    `function matches(string $value, $key): bool { return true; } echo array_all([], 'matches');`,
		"no_bool_return":     `function matches($value, $key) { return true; } echo array_all([], 'matches');`,
		"wrong_bool_return":  `function matches($value, $key): bool { return 'yes'; } echo array_all([], 'matches');`,
		"dead_invalid":       `function matches($value, $key): bool { return true; } echo true || array_all([], 'matches', null);`,
	} {
		t.Run(name, func(t *testing.T) {
			program, err := frontend.Parse("<?php "+source, "8.4")
			var output []byte
			if err == nil {
				output, err = Generate(program)
			}
			if err == nil || len(output) != 0 || !strings.Contains(err.Error(), "PHP00") {
				t.Fatalf("unsupported source generated %q: %v", output, err)
			}
		})
	}
}

func TestPHP84ArrayAllWrongCallbackIRIsRejectedBeforeGeneration(t *testing.T) {
	p := ir.Program{
		Version:    ir.Version,
		PHPVersion: "8.4",
		Functions: []ir.Function{{
			Name:       "matches",
			Params:     []string{"value", "key"},
			ReturnType: "bool",
			Line:       1,
			Body:       []ir.Stmt{{Kind: "return", Line: 1, Expr: ir.Expr{Kind: "bool", Text: "true", Line: 1}}},
		}},
		Body: []ir.Stmt{{
			Kind: "echo",
			Line: 1,
			Expr: ir.Expr{Kind: "call", Text: "array_all", Line: 1, Args: []ir.Expr{
				{Kind: "array", Line: 1},
				{Kind: "string", Text: "matches", Line: 1},
				{Kind: "null", Text: "null", Line: 1},
			}},
		}},
	}
	output, err := Generate(p)
	if err == nil || len(output) != 0 || !strings.Contains(err.Error(), "PHP004") {
		t.Fatalf("wrong array_all callback IR was not rejected safely: %q %v", output, err)
	}
}
