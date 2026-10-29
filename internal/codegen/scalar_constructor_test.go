package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestScalarConstructorLowering(t *testing.T) {
	p, err := frontend.Parse(`<?php
class Record {
 public $value;
 public function __CONSTRUCT(INT $n, STRING $s, BOOL $b, $extra = null) {
  $this->value = $n;
  $n = 'changed';
  echo $n;
 }
}
$r = new Record(1, 'x', false);`, testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	f := p.Classes[0].Methods[0]
	if strings.Join(f.ParamTypes, ",") != "int,string,bool," || f.OptionalNullParams != 1 {
		t.Fatalf("types: %#v", f)
	}
	out, mapping, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	for _, annotation := range []string{"int", "string", "bool"} {
		if strings.Count(s, `, "`+annotation+`")`) != 1 {
			t.Fatalf("guard for %s: %s", annotation, s)
		}
	}
	if !strings.Contains(s, "rt.NullDefaultArgs(args, 3, 4)") || strings.Index(s, "rt.ScalarParameter") > strings.Index(s, `rt.ResolveProperty(v["this"], "value")`) {
		t.Fatalf("entry guards missing or late: %s", s)
	}
	for i, line := range strings.Split(s, "\n") {
		if strings.Contains(line, `rt.ResolveProperty(v["this"], "value")`) {
			if got, ok := mapping.PHPLine(i + 1); !ok || got != 5 {
				t.Fatalf("constructor map: %d %v", got, ok)
			}
		}
	}
}

func TestScalarConstructorBoundaries(t *testing.T) {
	for _, source := range []string{
		`class X { public function __construct(int $x) {} } $x = new X('1');`,
		`class X { public function __construct(string $x) {} } $x = new X(1);`,
		`class X { public function __construct(bool $x) {} } $x = new X(0);`,
		`class X { public function __construct(int $x) {} } $x = new X(null);`,
		`class X { public function __construct(int $x) {} } $x = new X([]);`,
		`class X { public function __construct(int $x) {} } $x = new X();`,
		`class X { public function __construct(int $x) {} } $x = new X(1, 2);`,
		`class X { public function __construct(int $x = null) {} }`,
		`class X { public function __construct(string $x = 'x') {} }`,
		`class X { public function __construct(?int $x) {} }`,
		`class X { public function __construct(float $x) {} }`,
		`class X { public function __construct(array $x) {} }`,
		`class X { public function __construct(callable $x) {} }`,
		`class X { public function f(int $x) {} }`,
		`function __construct(float $x) {}`, `function f(string $x = null) {}`,
		`class X { private function __construct(int $x) {} }`,
		`class X { public function __construct(int $x) {return 1;} }`,
		`class X { public function __construct(int $x) {} } $x = new X(1); $x->__construct(2);`,
		`class X { public function __construct(int $x) { echo strlen($x); } }`,
	} {
		p, err := frontend.Parse("<?php "+source, testTarget(t))
		var out []byte
		if err == nil {
			out, err = Generate(p)
		}
		if err == nil || len(out) != 0 {
			t.Fatalf("accepted: %s: %v", source, err)
		}
	}
}

func TestScalarConstructorMalformedIR(t *testing.T) {
	for _, f := range []ir.Function{
		{Name: "__construct", Params: []string{"x"}, ParamTypes: []string{"INT"}},
		{Name: "__construct", Params: []string{"x"}, ParamTypes: []string{"string", "bool"}},
		{Name: "__construct", Params: []string{"x"}, ParamTypes: []string{"bool"}, OptionalNullParams: 1},
		{Name: "f", Params: []string{"x"}, ParamTypes: []string{"int"}},
	} {
		p := ir.Program{Version: ir.Version, PHPVersion: testTarget(t), Classes: []ir.Class{{Name: "X", Methods: []ir.Function{f}}}}
		if out, err := Generate(p); err == nil || len(out) != 0 {
			t.Fatalf("accepted malformed %#v: %v", f, err)
		}
	}
}
