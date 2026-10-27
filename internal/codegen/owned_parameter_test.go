package codegen

import (
	"strconv"
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestOwnedParameterSelection(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		owned        bool
		scalar       bool
	}{
		{
			"int_parameter",
			`function selected(int $number) { $items = [$number]; return $items; }`,
			true, false,
		},
		{
			"string_parameter",
			`function selected(string $word) { $items = [$word]; return $items; }`,
			true, false,
		},
		{
			"bool_parameter",
			`function selected(bool $flag) { $items = [$flag]; return $items; }`,
			true, false,
		},
		{
			"multiple_parameters_with_control_flow",
			`function selected(int $number, string $word, bool $flag) { $items = ['word' => $word]; if ($flag) { $items[] = $number; } else { $items[] = 0; } return $items; }`,
			true, false,
		},
		{
			"read_only_unused_parameter_has_a_slot",
			`function selected(int $unused) { $items = []; return $items; }`,
			true, false,
		},
		{
			"scalar_parameter_replaced_by_flat_array",
			`function selected(int $value) { $value = []; $value[] = 1; return $value; }`,
			true, false,
		},
		{
			"pure_scalar_parameters_keep_scalar_frame",
			`function selected(int $number, string $word, bool $flag) { return $flag ? $word : $number; }`,
			false, true,
		},
		{
			"array_parameter_even_if_overwritten",
			`function selected(array $input) { $input = []; return $input; }`,
			false, false,
		},
		{
			"untyped_parameter_even_if_overwritten",
			`function selected($input) { $input = []; return $input; }`,
			false, false,
		},
		{
			"callable_parameter_even_if_overwritten",
			`function selected(callable $input) { $input = []; return $input; }`,
			false, false,
		},
		{
			"mixed_scalar_and_untyped_parameters",
			`function selected(int $number, $input) { $items = [$number]; return $items; }`,
			false, false,
		},
		{
			"untyped_null_default_stays_legacy",
			`function selected(int $number, $input = null) { $items = [$number]; return $items; }`,
			false, false,
		},
		{
			"parameter_domain_bounds_loop_writes",
			`function selected(int $number) { $items = []; while (false) { $number = []; } return $items; }`,
			false, false,
		},
		{
			"dead_nested_value_stays_unsafe",
			`function selected(int $number) { $items = []; return $items; $bad = [[$number]]; }`,
			false, false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, err := frontend.Parse("<?php "+tc.source, testTarget(t))
			if err != nil {
				t.Fatal(err)
			}
			out, err := Generate(p)
			if err != nil {
				t.Fatal(err)
			}
			body := generatedFunction(t, string(out), "F_selected")
			if got := strings.Contains(body, "map[string]*rt.ValueSlot{}"); got != tc.owned {
				t.Fatalf("owned parameter selection = %v, want %v:\n%s", got, tc.owned, body)
			}
			if got := strings.Contains(body, "rt.NewScalarFrame()"); got != tc.scalar {
				t.Fatalf("scalar-frame selection = %v, want %v:\n%s", got, tc.scalar, body)
			}
			if !tc.owned && !tc.scalar && !strings.Contains(body, "v := map[string]any{}") {
				t.Fatalf("legacy fallback missing:\n%s", body)
			}
		})
	}
}

func TestOwnedParameterLoweringAndMap(t *testing.T) {
	p, err := frontend.Parse(`<?php
function selected(int $number, string $word, bool $flag, int $unused): array {
    $items = ['word' => $word, 'number' => $number];
    if ($flag) {
        $items['yes'] = $number;
    }
    return $items;
}`, testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	out, mapping, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	body := generatedFunction(t, string(out), "F_selected")

	firstGuard := strings.Index(body, "rt.ScalarParameter(")
	for _, local := range []string{"flag", "items", "number", "unused", "word"} {
		declare := `v["` + local + `"] = &rt.ValueSlot{}`
		close := `defer v["` + local + `"].Close()`
		if strings.Count(body, declare) != 1 || strings.Count(body, close) != 1 {
			t.Fatalf("owned local %q is not allocated and closed exactly once:\n%s", local, body)
		}
		if strings.Index(body, close) > firstGuard {
			t.Fatalf("cleanup for %q was not registered before the first parameter guard:\n%s", local, body)
		}
	}
	args := strings.Index(body, "rt.Args(args, 4)")
	if args < 0 {
		t.Fatalf("arity guard missing:\n%s", body)
	}
	previous := args
	for i, check := range []struct {
		name, annotation string
	}{
		{"number", "int"},
		{"word", "string"},
		{"flag", "bool"},
		{"unused", "int"},
	} {
		guard := `v["` + check.name + `"].Write(rt.ScalarParameter(args[` + strconv.Itoa(i) + `], "` + check.annotation + `"))`
		at := strings.Index(body, guard)
		if at < 0 || at < previous {
			t.Fatalf("parameter guard missing or out of order %q:\n%s", guard, body)
		}
		previous = at
	}
	if strings.Index(body, `v["number"] = &rt.ValueSlot{}`) > strings.Index(body, `v["number"].Write(rt.ScalarParameter`) {
		t.Fatalf("parameter storage was written before it was allocated:\n%s", body)
	}
	for index, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "map[string]*rt.ValueSlot{}") || strings.Contains(line, "defer v[") || strings.Contains(line, "rt.ScalarParameter(") {
			if got, ok := mapping.PHPLine(index + 1); ok {
				t.Fatalf("synthetic parameter/lifetime line maps to PHP line %d: %s", got, line)
			}
		}
	}
	for _, check := range []struct {
		fragment string
		phpLine  int
	}{
		{`temporary.MoveTo(v["items"])`, 3},
		{"if rt.Bool(", 4},
		{`v["items"].ArrayWrite(key, value)`, 5},
		{"return rt.ArrayReturn((func() any {", 7},
	} {
		found := false
		for index, line := range strings.Split(string(out), "\n") {
			if !strings.Contains(line, check.fragment) {
				continue
			}
			found = true
			if got, ok := mapping.PHPLine(index + 1); !ok || got != check.phpLine {
				t.Fatalf("%q maps to %d (%v), want %d", check.fragment, got, ok, check.phpLine)
			}
		}
		if !found {
			t.Fatalf("generated source missing %q:\n%s", check.fragment, out)
		}
	}
}

func TestOwnedParameterPreservesRejection(t *testing.T) {
	for _, source := range []string{
		`<?php function invalid(int &$value) { $items = []; return $items; }`,
		`<?php function invalid(int $value) { $items = []; return $items; } invalid('1');`,
		`<?php function invalid(bool $value) { $items = []; return $items; } invalid(0);`,
		`<?php function invalid(int $value) { $items = []; return $missing; }`,
	} {
		p, err := frontend.Parse(source, testTarget(t))
		if err == nil {
			var out []byte
			out, err = Generate(p)
			if len(out) != 0 {
				t.Fatal("invalid source emitted output")
			}
		}
		if err == nil {
			t.Fatalf("invalid source accepted: %s", source)
		}
	}
	for _, f := range []ir.Function{
		{Name: "invalid", Params: []string{"value"}, ParamTypes: []string{"INT"}, Line: 1},
		{Name: "invalid", Params: []string{"value"}, ParamTypes: []string{"int", "string"}, Line: 1},
		{Name: "invalid", Params: []string{"value"}, ParamTypes: []string{"bool"}, OptionalNullParams: 1, Line: 1},
	} {
		p := ir.Program{Version: ir.Version, PHPVersion: testTarget(t), Functions: []ir.Function{f}}
		if out, err := Generate(p); err == nil || len(out) != 0 {
			t.Fatalf("malformed parameter metadata accepted: %#v: %v", f, err)
		}
	}
}
