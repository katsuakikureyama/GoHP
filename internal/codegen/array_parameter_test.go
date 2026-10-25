package codegen

import (
	"reflect"
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestArrayParameterRetentionAndEntryGuards(t *testing.T) {
	p, err := frontend.Parse(`<?php
function mixed(array $left, $flag, ArRaY $right): bool {
    echo 'body';
    $left = false;
    return $left;
}
function plain($value) { return $value; }
`, "8.3")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(p.Functions[0].ParamTypes, []string{"array", "", "array"}) || len(p.Functions[1].ParamTypes) != 0 {
		t.Fatalf("parameter annotations lost/misaligned: %#v", p.Functions)
	}
	output, mapping, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	text := string(output)
	body := strings.Index(text, `rt.Echo("body")`)
	for _, guard := range []string{`v["left"] = rt.ArrayParameter(args[0])`, `v["right"] = rt.ArrayParameter(args[2])`, `v["flag"] = rt.Clone(args[1])`} {
		if at := strings.Index(text, guard); at < 0 || body < at {
			t.Fatalf("missing/misplaced entry guard %s:\n%s", guard, text)
		}
	}
	if strings.Count(text, "rt.ArrayParameter(") != 2 {
		t.Fatal("untyped arguments gained a type constraint")
	}
	for line, content := range strings.Split(text, "\n") {
		if strings.Contains(content, `rt.Echo("body")`) {
			if got, ok := mapping.PHPLine(line + 1); !ok || got != 3 {
				t.Fatalf("body mapping = %d %v", got, ok)
			}
			return
		}
	}
	t.Fatal("body marker missing")
}

func TestArrayParameterUnsupportedSource(t *testing.T) {
	for _, tc := range []struct{ name, source string }{
		{"null", "function take(array $a) { return $a; } take(null);"},
		{"bool", "function take(array $a) { return $a; } take(false);"},
		{"int", "function take(array $a) { return $a; } take(1);"},
		{"string", "function take(array $a) { return $a; } take('x');"},
		{"object", "class X {} function take(array $a) { return $a; } take(new X());"},
		{"forward_second_arg", "function first() { return later([], false); } function later($a, array $b) { return $b; }"},
		{"dead_call", "function take(array $a) { return $a; } if(false) { take(null); }"},
		{"missing_arg", "function take(array $a) { return $a; } take();"},
		{"extra_arg", "function take(array $a) { return $a; } take([], []);"},
		{"entry_type_known", "function take(array $a) { return strtolower($a); }"},
		{"bool_return_entry_type", "function take(array $a): bool { if (false) { return true; } else { return $a; } }"},
		{"reassigned_not_array", "function take(array $a) { $a = false; return later($a); } function later(array $b) { return $b; }"},
		{"nullable", "function take(?array $a) { return $a; }"},
		{"union", "function take(array|bool $a) { return $a; }"},
		{"intersection", "function take(array&X $a) { return $a; }"},
		{"callable", "function take(callable $a) { return $a; }"},
		{"bool_type", "function take(bool $a) { return $a; }"},
		{"default", "function take(array $a = []) { return $a; }"},
		{"reference", "function take(array &$a) { return $a; }"},
		{"variadic", "function take(array ...$a) { return $a; }"},
		{"method", "class X { public function take(array $a) { return $a; } }"},
		{"named_arg", "function take(array $a) { return $a; } take(a: []);"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, err := frontend.Parse("<?php "+tc.source, "8.3")
			if err == nil {
				var output []byte
				output, err = Generate(p)
				if len(output) != 0 {
					t.Fatal("unsupported source produced output")
				}
			}
			if err == nil || !strings.Contains(err.Error(), "PHP00") {
				t.Fatalf("unsupported source accepted: %v", err)
			}
		})
	}
}

func TestArrayParameterMalformedIR(t *testing.T) {
	for _, annotations := range [][]string{{"array", ""}, {"ARRAY"}, {"bool"}, {"callable"}, {"?array"}} {
		p := ir.Program{Version: ir.Version, PHPVersion: "8.3", Functions: []ir.Function{{
			Name: "take", Params: []string{"a"}, ParamTypes: annotations, Line: 1,
			Body: []ir.Stmt{{Kind: "return", Line: 1, Expr: ir.Expr{Kind: "var", Text: "a", Line: 1}}},
		}}}
		if output, err := Generate(p); err == nil || !strings.Contains(err.Error(), "IR002") || len(output) != 0 {
			t.Fatalf("malformed parameter types accepted: %#v: %v", annotations, err)
		}
	}
	// Validate all signatures before an earlier caller can index later metadata.
	p := ir.Program{Version: ir.Version, PHPVersion: "8.3", Functions: []ir.Function{
		{Name: "caller", Line: 1, Body: []ir.Stmt{{Kind: "return", Line: 1, Expr: ir.Expr{Kind: "call", Text: "later", Line: 1}}}},
		{Name: "later", ParamTypes: []string{"array"}, Line: 2},
	}}
	if output, err := Generate(p); err == nil || !strings.Contains(err.Error(), "IR002") || len(output) != 0 {
		t.Fatalf("malformed forward signature accepted: %v", err)
	}
}
