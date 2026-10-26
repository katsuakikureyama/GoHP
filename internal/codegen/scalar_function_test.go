package codegen

import (
	"reflect"
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestScalarFunctionParameterLowering(t *testing.T) {
	p, err := frontend.Parse(`<?php
function typed(InT $n, STRING $s, Bool $b): bool {
    echo 'body';
    $n = 'changed';
    return $b;
}
function mixed(int $n, array $a, callable $f, $tail = null) { return $a; }
function non_scalar(int $n) { return [$n]; }
function reassign(int $n) { $n = 'changed'; return strlen($n); }
`, testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(p.Functions[0].ParamTypes, []string{"int", "string", "bool"}) ||
		!reflect.DeepEqual(p.Functions[1].ParamTypes, []string{"int", "array", "callable", ""}) || p.Functions[1].OptionalNullParams != 1 {
		t.Fatalf("parameter metadata: %#v", p.Functions)
	}
	out, mapping, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	body := strings.Index(s, `rt.Echo("body")`)
	for _, guard := range []string{
		`v.Write("n", rt.ScalarParameter(args[0], "int"))`,
		`v.Write("s", rt.ScalarParameter(args[1], "string"))`,
		`v.Write("b", rt.ScalarParameter(args[2], "bool"))`,
	} {
		if at := strings.Index(s, guard); at < 0 || at > body {
			t.Fatalf("guard missing/late %s:\n%s", guard, s)
		}
	}
	if strings.Count(s, "rt.NewScalarFrame()") != 2 || !strings.Contains(s, `rt.BoolReturn(v.Read("b"))`) ||
		strings.Index(s, "defer v.Close()") > strings.Index(s, "rt.ScalarParameter") {
		t.Fatalf("frame selection/lifetime/return: %s", s)
	}
	for _, guard := range []string{`v["n"] = rt.ScalarParameter(args[0], "int")`, `v["a"] = rt.ArrayParameter(args[1])`, `v["f"] = rt.CallableParameter(args[2])`, `rt.NullDefaultArgs(args, 3, 4)`} {
		if !strings.Contains(s, guard) {
			t.Fatalf("legacy mixed guard missing %s: %s", guard, s)
		}
	}
	for i, line := range strings.Split(s, "\n") {
		if strings.Contains(line, `rt.Echo("body")`) {
			if got, ok := mapping.PHPLine(i + 1); !ok || got != 3 {
				t.Fatalf("statement mapping: %d %v", got, ok)
			}
		}
	}
}

func TestScalarFunctionParameterBoundaries(t *testing.T) {
	for _, source := range []string{
		`function f(int $v) {} f('1');`,
		`function f(string $v) {} f(1);`,
		`function f(bool $v) {} f(0);`,
		`function f(int $v) {} f(null);`,
		`function f(int $v) {} f([]);`,
		`class X {} function f(int $v) {} f(new X());`,
		`function caller() { later(false); } function later(int $v) {}`,
		`function f(string $v) {} if(false) { f(1); }`,
		`function f(int $v) {} f();`,
		`function f(int $v) {} f(1, 2);`,
		`function f(int $v) { return strlen($v); }`,
		`function f(bool $v): bool { $v = 0; return $v; }`,
		`function f(int $v) { $v = false; later($v); } function later(int $v) {}`,
		`function f(int $v = null) {}`,
		`function f(string $v = 'x') {}`,
		`function f(?int $v) {}`,
		`function f(int|string $v) {}`,
		`function f(int&X $v) {}`,
		`function f(float $v) {}`,
		`function f(mixed $v) {}`,
		`function f(int &$v) {}`,
		`function f(int ...$v) {}`,
		`function f(int $v) {} f(v: 1);`,
		`class X { public function f(int $v) {} }`,
		`$f = fn(int $v) => $v;`,
	} {
		p, err := frontend.Parse("<?php "+source, testTarget(t))
		var out []byte
		if err == nil {
			out, err = Generate(p)
		}
		if err == nil || len(out) != 0 {
			t.Fatalf("accepted %s: %v", source, err)
		}
	}
}

func TestScalarFunctionParameterMalformedIR(t *testing.T) {
	for _, f := range []ir.Function{
		{Name: "f", Params: []string{"v"}, ParamTypes: []string{"INT"}},
		{Name: "f", Params: []string{"v"}, ParamTypes: []string{"int", "string"}},
		{Name: "f", Params: []string{"v"}, ParamTypes: []string{"bool"}, OptionalNullParams: 1},
		{Name: "f", Params: []string{"v"}, ParamTypes: []string{"?int"}},
		{Name: "f", Params: []string{"v"}, ParamTypes: []string{"float"}},
	} {
		p := ir.Program{Version: ir.Version, PHPVersion: testTarget(t), Functions: []ir.Function{f}}
		if out, err := Generate(p); err == nil || len(out) != 0 {
			t.Fatalf("accepted malformed %#v: %v", f, err)
		}
	}
}
