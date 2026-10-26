package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
)

func TestPHP84ArrayAnyLoweringAndTargetBoundaries(t *testing.T) {
	const builtin = `<?php
function MaTcHeS($value, $key): bool { return $value === $key; }
echo ArRaY_AnY(['match'], 'mAtChEs');`

	php84, err := frontend.Parse(builtin, "8.4")
	if err != nil {
		t.Fatal(err)
	}
	output, err := Generate(php84)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"rt.ArrayAny(",
		"func(value, key any) any { return F_matches(value, key) }",
	} {
		if !strings.Contains(string(output), want) {
			t.Fatalf("PHP 8.4 array_any lowering missing %q:\n%s", want, output)
		}
	}

	php83, err := frontend.Parse(builtin, "8.3")
	if err != nil {
		t.Fatal(err)
	}
	if output, err := Generate(php83); err == nil || len(output) != 0 || !strings.Contains(err.Error(), "unsupported function array_any") {
		t.Fatalf("PHP 8.3 absent builtin generated %q: %v", output, err)
	}

	const userFunction = `<?php
function ArRaY_AnY($value) { return $value; }
echo array_any('user');`
	php83, err = frontend.Parse(userFunction, "8.3")
	if err != nil {
		t.Fatal(err)
	}
	output, err = Generate(php83)
	if err != nil || !strings.Contains(string(output), `F_array_any("user")`) || strings.Contains(string(output), "rt.ArrayAny(") {
		t.Fatalf("PHP 8.3 user function was not preserved: %v\n%s", err, output)
	}

	php84, err = frontend.Parse(userFunction, "8.4")
	if err != nil {
		t.Fatal(err)
	}
	if output, err = Generate(php84); err == nil || len(output) != 0 || !strings.Contains(err.Error(), "function name reserved in PHP 8.4: ArRaY_AnY") {
		t.Fatalf("PHP 8.4 user declaration generated %q: %v", output, err)
	}
}

func TestPHP84ArrayAnyStatefulCallbackUsesContextAdapter(t *testing.T) {
	program, err := frontend.Parse(`<?php
function token($value, $key): bool { return strtok($value, ',') === 'hit'; }
echo array_any(['hit,next'], 'token');`, "8.4")
	if err != nil {
		t.Fatal(err)
	}
	output, err := Generate(program)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"rt.ArrayAny(",
		"FContext_token(execution, value, key)",
	} {
		if !strings.Contains(string(output), want) {
			t.Fatalf("stateful array_any lowering missing %q:\n%s", want, output)
		}
	}
}

func TestPHP84ArrayAnyUnsupportedSourceFailsBeforeGeneration(t *testing.T) {
	for name, source := range map[string]string{
		"missing_array":      `function matches($value, $key): bool { return true; } echo array_any('not array', 'matches');`,
		"missing_callback":   `function matches($value, $key): bool { return true; } echo array_any([]);`,
		"extra_argument":     `function matches($value, $key): bool { return true; } echo array_any([], 'matches', null);`,
		"dynamic_callback":   `function matches($value, $key): bool { return true; } $callback = 'matches'; echo array_any([], $callback);`,
		"builtin_callback":   `echo array_any([], 'is_string');`,
		"method_callback":    `class Matcher { public function test($value, $key): bool { return true; } } echo array_any([], ['Matcher', 'test']);`,
		"arrow_callback":     `echo array_any([], fn($value, $key) => true);`,
		"one_parameter":      `function matches($value): bool { return true; } echo array_any([], 'matches');`,
		"three_parameters":   `function matches($value, $key, $extra): bool { return true; } echo array_any([], 'matches');`,
		"optional_parameter": `function matches($value, $key = null): bool { return true; } echo array_any([], 'matches');`,
		"typed_parameter":    `function matches(string $value, $key): bool { return true; } echo array_any([], 'matches');`,
		"no_bool_return":     `function matches($value, $key) { return true; } echo array_any([], 'matches');`,
		"wrong_bool_return":  `function matches($value, $key): bool { return 'yes'; } echo array_any([], 'matches');`,
		"dead_invalid":       `function matches($value, $key): bool { return true; } echo true || array_any([], 'matches', null);`,
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
