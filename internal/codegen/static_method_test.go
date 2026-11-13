package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
)

func TestStaticMethodLowering(t *testing.T) {
	src := `<?php
declare(strict_types=1);
final class Box {
public static function add(int $a, int $b) { return $a + $b; }
public static function make() { return new Box(); }
public function id() { return Box::add(1, 2); }
}
echo Box::add(1, 2);
$made = Box::make();
echo $made->id();
`
	p, err := frontend.Parse(src, testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	out, err := Generate(p)
	if err != nil {
		t.Fatal(err)
	}
	text := string(out)
	for _, want := range []string{
		"func S_box_add(args ...any) any {",
		"func S_box_make(args ...any) any {",
		"S_box_add(int64(1), int64(2))",
		"func (self *C_box) M_id(args ...any) any {",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "func (self *C_box) S_") {
		t.Fatalf("static method must not have a receiver:\n%s", text)
	}
	add := text[strings.Index(text, "func S_box_add"):strings.Index(text, "func S_box_make")]
	if strings.Contains(add, "self") || strings.Contains(add, `v["this"]`) {
		t.Fatalf("static method body must not bind a receiver:\n%s", add)
	}
}

func TestStaticMethodExecutionLane(t *testing.T) {
	src := `<?php
final class Tokenizer {
public static function first($value) { return Tokenizer::cut($value); }
public static function cut($value) { return strtok($value, ','); }
}
echo Tokenizer::first('a,b');
`
	p, err := frontend.Parse(src, testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	out, err := Generate(p)
	if err != nil {
		t.Fatal(err)
	}
	text := string(out)
	for _, want := range []string{
		"func S_tokenizer_cut(args ...any) any { return SContext_tokenizer_cut(&rt.Execution{}, args...) }",
		"func SContext_tokenizer_cut(execution *rt.Execution, args ...any) any {",
		"SContext_tokenizer_first(execution,",
		"SContext_tokenizer_cut(execution,",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q:\n%s", want, text)
		}
	}
}

func TestStaticMethodCodegenRejections(t *testing.T) {
	for _, source := range []string{
		`<?php class X { public static function f() { return $this; } }`,
		`<?php class X { public static function __construct() {} }`,
		`<?php class X { public function f() { return 1; } } echo X::f();`,
		`<?php class X { public static function f() { return 1; } } $x = new X(); echo $x->f();`,
	} {
		p, err := frontend.Parse(source, testTarget(t))
		if err != nil {
			continue
		}
		if out, err := Generate(p); err == nil || len(out) != 0 {
			t.Fatalf("static rejection %q: %s %v", source, out, err)
		}
	}
}

func TestStaticMethodMalformedIR(t *testing.T) {
	p, err := frontend.Parse(`<?php class X { public static function f() { return 1; } } echo X::f();`, testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	p.Classes[0].Methods[0].Static = false
	if out, err := Generate(p); err == nil || len(out) != 0 {
		t.Fatalf("non-static method called with :: should reject: %s %v", out, err)
	}
}
