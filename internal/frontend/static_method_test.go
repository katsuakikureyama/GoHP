package frontend

import (
	"strings"
	"testing"
)

func TestParseStrictTypesDeclareAndFinalClass(t *testing.T) {
	for _, target := range []string{"8.3", "8.4"} {
		p, err := Parse("<?php\ndeclare( strict_types = 1 );\nfinal class Box {\npublic static function make(int $value) { return Box::make($value); }\n}\n", target)
		if err != nil {
			t.Fatalf("%s: %v", target, err)
		}
		if len(p.Classes) != 1 || p.Classes[0].Name != "Box" || len(p.Classes[0].Methods) != 1 {
			t.Fatalf("%s: unexpected declarations: %+v", target, p.Classes)
		}
		method := p.Classes[0].Methods[0]
		if !method.Static || method.Name != "make" || method.ParamTypes[0] != "int" {
			t.Fatalf("%s: static method metadata: %+v", target, method)
		}
		if len(method.Body) != 1 || method.Body[0].Expr.Kind != "static_call" || method.Body[0].Expr.Text != "make" {
			t.Fatalf("%s: static call body: %+v", target, method.Body)
		}
		if len(method.Body[0].Expr.Args) != 2 || method.Body[0].Expr.Args[0].Kind != "class_ref" || method.Body[0].Expr.Args[0].Text != "Box" {
			t.Fatalf("%s: static call receiver: %+v", target, method.Body[0].Expr)
		}
	}
}

func TestParseStaticCallStatement(t *testing.T) {
	p, err := Parse("<?php\nfinal class Box { public static function ping() { return 1; } }\nBox::ping();\n", testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Body) != 1 || p.Body[0].Kind != "call_stmt" || p.Body[0].Expr.Kind != "static_call" {
		t.Fatalf("standalone static call: %+v", p.Body)
	}
}

func TestParseStrictTypesDeclareRejections(t *testing.T) {
	for _, source := range []string{
		"<?php\ndeclare(ticks=1);\n",
		"<?php\ndeclare(encoding='UTF-8');\n",
		"<?php\ndeclare(strict_types=0);\n",
		"<?php\ndeclare(strict_types=1) {}\n",
		"<?php\n$value = 1;\ndeclare(strict_types=1);\n",
		"<?php\ndeclare(strict_types=1);\ndeclare(strict_types=1);\n",
		"<?php\nfunction f() { declare(strict_types=1); }\n",
	} {
		if _, err := Parse(source, testTarget(t)); err == nil || !strings.Contains(err.Error(), "PHP002") {
			t.Fatalf("declare form must fail closed: %q -> %v", source, err)
		}
	}
}

func TestParseStaticModifierRejections(t *testing.T) {
	for _, source := range []string{
		"<?php\nclass X { static public function f() {} }\n",
		"<?php\nclass X { static function f() {} }\n",
		"<?php\nclass X { public static $value; }\n",
		"<?php\nclass X { public final function f() {} }\n",
		"<?php\nabstract class X {}\n",
		"<?php\nreadonly class X {}\n",
		"<?php\nclass X extends Y {}\n",
		"<?php\nclass X implements Y {}\n",
		"<?php\nfinal function f() {}\n",
	} {
		if _, err := Parse(source, testTarget(t)); err == nil || !strings.Contains(err.Error(), "PHP002") {
			t.Fatalf("final/static modifier must fail closed: %q -> %v", source, err)
		}
	}
}

func TestParseStaticMemberRejections(t *testing.T) {
	for _, source := range []string{
		"<?php\nclass X { public static function f() { return 1; } }\nself::f();\n",
		"<?php\nclass X { public static function f() { return 1; } }\nstatic::f();\n",
		"<?php\nclass X { public static function f() { return 1; } }\nparent::f();\n",
		"<?php\nclass X { public static function f() { return 1; } }\n$value = new X();\n$value::f();\n",
		"<?php\nclass X { public $value = 1; }\necho X::$value;\n",
		"<?php\nclass X {}\necho X::CONSTANT;\n",
		"<?php\nclass X {}\necho X::class;\n",
	} {
		if _, err := Parse(source, testTarget(t)); err == nil || !strings.Contains(err.Error(), "PHP002") {
			t.Fatalf("static member form must fail closed: %q -> %v", source, err)
		}
	}
}

func TestParseStaticModuleDeclareAndFinal(t *testing.T) {
	source := "<?php\ndeclare( strict_types = 1 );\nfinal class X {}\n"
	for _, target := range []string{"8.3", "8.4"} {
		parsed, err := ParseStaticModule("src/dependency.php", source, target, Options{}, false)
		if err != nil {
			t.Fatalf("%s: %v", target, err)
		}
		if len(parsed.Inspection.Declarations) != 1 || parsed.Inspection.Declarations[0].Kind != "class" || parsed.Inspection.Declarations[0].Name != "X" {
			t.Fatalf("%s: declarations: %+v", target, parsed.Inspection.Declarations)
		}
		if len(parsed.Program.Classes) != 1 || parsed.Program.Classes[0].Name != "X" {
			t.Fatalf("%s: classes: %+v", target, parsed.Program.Classes)
		}
	}
	if _, err := ParseStaticModule("src/dependency.php", "<?php\nfinal class X {}\ndeclare(strict_types=1);\n", testTarget(t), Options{}, false); err == nil || !strings.Contains(err.Error(), "PHP002") {
		t.Fatalf("late declare must fail closed: %v", err)
	}
}
