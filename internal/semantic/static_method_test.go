package semantic

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func parseStaticMethodProgram(t *testing.T, src string) ir.Program {
	t.Helper()
	p, err := frontend.Parse(src, "8.3")
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestStaticMethodValidation(t *testing.T) {
	p := parseStaticMethodProgram(t, `<?php
declare(strict_types=1);
final class Counter {
public static function add(int $a, int $b) { return $a + $b; }
public static function twice(int $value) { return Counter::add($value, $value); }
public static function make($value) { return new Counter(); }
public function id() { return Counter::add(1, 2); }
}
function total(int $a, int $b) { return Counter::add($a, $b); }
echo Counter::add(1, 2);
$made = Counter::make(1);
echo $made->id();
`)
	if err := Validate(p); err != nil {
		t.Fatalf("static methods should validate: %v", err)
	}
}

func TestStaticMethodRejections(t *testing.T) {
	for _, source := range []string{
		`<?php class X { public static function f() { return $this; } }`,
		`<?php class X { public static function __construct() {} }`,
		`<?php class X { public static function f() { return 1; } } echo Y::f();`,
		`<?php class X { public static function f() { return 1; } } echo X::missing();`,
		`<?php class X { public function f() { return 1; } } echo X::f();`,
		`<?php class X { public static function f() { return 1; } } $x = new X(); echo $x->f();`,
		`<?php class X { public static function f() { return 1; } } echo X::f(1);`,
		`<?php class X { public static function f() { return 1; } } echo X::f(); echo X::class;`,
	} {
		p, err := frontend.Parse(source, "8.3")
		if err != nil {
			// Parser-level rejections are also acceptable for these forms.
			continue
		}
		if err := Validate(p); err == nil {
			t.Fatalf("static form should reject: %q", source)
		}
	}
}

func TestStaticCallMalformedIR(t *testing.T) {
	cases := []ir.Expr{
		{Kind: "static_call", Text: "f", Args: nil},
		{Kind: "static_call", Text: "f", Args: []ir.Expr{{Kind: "var", Text: "X"}}},
		{Kind: "static_call", Text: "f", Args: []ir.Expr{{Kind: "class_ref", Text: "bad-name"}}},
		{Kind: "static_call", Text: "f", Args: []ir.Expr{{Kind: "class_ref", Text: "self"}}},
	}
	for _, static := range cases {
		p := ir.Program{
			Version:    ir.Version,
			PHPVersion: "8.3",
			Classes:    []ir.Class{{Name: "X", Methods: []ir.Function{{Name: "f", Static: true}}}},
			Body:       []ir.Stmt{{Kind: "echo", Expr: static}},
		}
		if err := Validate(p); err == nil {
			t.Fatalf("malformed static call should reject: %+v", static)
		}
	}
}

func TestStaticModifierOutsideClassRejects(t *testing.T) {
	p := ir.Program{
		Version:    ir.Version,
		PHPVersion: "8.3",
		Functions:  []ir.Function{{Name: "f", Static: true}},
	}
	if err := Validate(p); err == nil || !strings.Contains(err.Error(), "static modifier") {
		t.Fatalf("top-level static function should reject: %v", err)
	}
}
