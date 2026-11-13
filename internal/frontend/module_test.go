package frontend

import (
	"reflect"
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestInspectModule(t *testing.T) {
	src := "<?php\nclass Box { public $value; public function get() { return $this->value; } }\n" +
		"function helper($x) { return $x; }\n" +
		"/* require_once __DIR__ . '/not-real.php'; */\n" +
		"require_once __DIR__ . '/nested/child.php';\nrequire_once __DIR__ . '/nested/child.php';\n"
	for _, target := range []string{"8.3", "8.4"} {
		module, err := InspectModule("src/main.php", src, target, Options{})
		if err != nil {
			t.Fatal(err)
		}
		if module.SourceID != "src/main.php" || module.PHPVersion != target || module.RegexProfile != "" {
			t.Fatalf("unexpected module identity: %+v", module)
		}
		if !reflect.DeepEqual(module.Declarations, []ModuleDeclaration{{"class", "Box", 2}, {"function", "helper", 3}}) ||
			!reflect.DeepEqual(module.Requires, []ModuleRequire{{5, "/nested/child.php", ""}, {6, "/nested/child.php", ""}}) {
			t.Fatalf("lost source order or lines: %+v", module)
		}
	}
	module, err := InspectModule("main.php", "<?php", "8.3", Options{ExperimentalPCRE2: true})
	if err != nil || module.RegexProfile != ir.Preg83Profile {
		t.Fatalf("regex admission identity: %+v / %v", module, err)
	}
}

func TestInspectModuleRejectsUnsupported(t *testing.T) {
	cases := []string{
		`echo "side effect";`, `$x = 1;`, `return 1;`,
		`require 'child.php';`, `include_once 'child.php';`,
		`require_once $path;`, `require_once (__DIR__ . '/child.php');`,
		`require_once __DIR__ . '/child.php' . '/other.php';`,
		`require_once __DIR__ . '/child.php'; class Late {}`,
		`require_once __DIR__ . '/child.php'; function late() { return null; }`,
		`if (true) { require_once __DIR__ . '/child.php'; }`,
		`function load() { require_once __DIR__ . '/child.php'; }`,
		`function load() { require_once('child.php'); }`,
		`function load() { return include('child.php'); }`,
		`function load() { return fn() => require_once('child.php'); }`,
		`class C { public function load() { if (true) { REQUIRE_ONCE('child.php'); } } }`,
		`class C { public $x = unknown(); }`,
		`class C { public function invalid() { while (true) { }`,
		`function bad() { return 1.2; }`,
		`"require_once" __DIR__ . '/child.php';`,
		`require_once "__DIR__" . '/child.php';`,
		`require_once __DIR__ "." '/child.php';`,
		`require_once __DIR__ . '/child.php' ";"`,
	}
	for _, suffix := range []string{"", "/", "child.php", "//child.php", "/./child.php", "/../child.php", "/x//child.php", "/x/../child.php", "/x/", "/x\\y.php", "/http:child.php", "/bad\x00.php", "/bad\n.php", "/bad\xff.php"} {
		cases = append(cases, "require_once __DIR__ . '"+suffix+"';")
	}
	for _, body := range cases {
		t.Run(body, func(t *testing.T) {
			module, err := InspectModule("src/bad.php", "<?php\n"+body, "8.3", Options{})
			if err == nil || !strings.Contains(err.Error(), "src/bad.php:") || !reflect.DeepEqual(module, ModuleInspection{}) {
				t.Fatalf("must reject with file identity and no partial module: %+v / %v", module, err)
			}
		})
	}
}

func TestInspectModuleBareRequire(t *testing.T) {
	// Bare require parses into Requires with Kind "require".
	src := "<?php\n" +
		"require __DIR__ . '/dep.php';\n" +
		"require_once __DIR__ . '/other.php';\n"
	module, err := InspectModule("src/main.php", src, "8.3", Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(module.Requires, []ModuleRequire{
		{Line: 2, Suffix: "/dep.php", Kind: "require"},
		{Line: 3, Suffix: "/other.php", Kind: ""},
	}) {
		t.Fatalf("bare require kind not recorded: %+v", module.Requires)
	}
}

func TestInspectModuleRequireOnceKind(t *testing.T) {
	// require_once keeps Kind "" (empty).
	src := "<?php\nrequire_once __DIR__ . '/child.php';\n"
	module, err := InspectModule("src/main.php", src, "8.3", Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(module.Requires, []ModuleRequire{{Line: 2, Suffix: "/child.php", Kind: ""}}) {
		t.Fatalf("require_once kind should be empty: %+v", module.Requires)
	}
}

func TestInspectModuleBareRequireRejects(t *testing.T) {
	cases := []string{
		`require $path;`,
		`require (__DIR__ . '/child.php');`,
		`require __DIR__ . '/child.php' . '/other.php';`,
		`require __DIR__ . '/child.php'; class Late {}`,
		`require __DIR__ . '/child.php'; function late() { return null; }`,
		`if (true) { require __DIR__ . '/child.php'; }`,
		`function load() { require __DIR__ . '/child.php'; }`,
		`require "__DIR__" . '/child.php';`,
		`require __DIR__ "." '/child.php';`,
		`require __DIR__ . '/child.php' ";"`,
	}
	for _, suffix := range []string{"", "/", "child.php", "//child.php", "/./child.php", "/../child.php", "/x//child.php", "/x/../child.php", "/x/", "/x\\y.php", "/http:child.php", "/bad\x00.php", "/bad\n.php", "/bad\xff.php"} {
		cases = append(cases, "require __DIR__ . '"+suffix+"';")
	}
	for _, body := range cases {
		t.Run(body, func(t *testing.T) {
			module, err := InspectModule("src/bad.php", "<?php\n"+body, "8.3", Options{})
			if err == nil || !strings.Contains(err.Error(), "src/bad.php:") || !reflect.DeepEqual(module, ModuleInspection{}) {
				t.Fatalf("must reject with file identity and no partial module: %+v / %v", module, err)
			}
		})
	}
}

func TestParseStaticModuleEntryHoistedDeclarations(t *testing.T) {
	// Entry: load -> body -> function (admitted)
	src1 := "<?php\n" +
		"require __DIR__ . '/lib/dep.php';\n" +
		"$x = late_func();\n" +
		"echo $x;\n" +
		"function late_func() { return 42; }\n"
	pm, err := ParseStaticModule("entry.php", src1, "8.3", Options{}, true)
	if err != nil {
		t.Fatalf("entry load->body->function must be admitted: %v", err)
	}
	if len(pm.Program.Functions) != 1 || pm.Program.Functions[0].Name != "late_func" {
		t.Fatalf("expected late_func in program.Functions: %+v", pm.Program.Functions)
	}
	if len(pm.Inspection.Declarations) != 1 || pm.Inspection.Declarations[0].Name != "late_func" {
		t.Fatalf("expected late_func in declarations: %+v", pm.Inspection.Declarations)
	}

	// Entry: load -> function -> body -> class (admitted)
	src2 := "<?php\n" +
		"require __DIR__ . '/lib/dep.php';\n" +
		"function early_fn() { return 1; }\n" +
		"$o = new LateClass();\n" +
		"echo $o->v;\n" +
		"class LateClass { public $v = 'ok'; }\n"
	pm2, err := ParseStaticModule("entry.php", src2, "8.3", Options{}, true)
	if err != nil {
		t.Fatalf("entry load->function->body->class must be admitted: %v", err)
	}
	if len(pm2.Program.Functions) != 1 || pm2.Program.Functions[0].Name != "early_fn" {
		t.Fatalf("expected early_fn: %+v", pm2.Program.Functions)
	}
	if len(pm2.Program.Classes) != 1 || pm2.Program.Classes[0].Name != "LateClass" {
		t.Fatalf("expected LateClass: %+v", pm2.Program.Classes)
	}

	// Entry: body -> load (still rejected)
	src3 := "<?php\n" +
		"$x = 1;\n" +
		"require __DIR__ . '/lib/dep.php';\n"
	_, err = ParseStaticModule("entry.php", src3, "8.3", Options{}, true)
	if err == nil {
		t.Fatal("entry body->load must still be rejected")
	}

	// Dependency: function after load (still rejected)
	src4 := "<?php\n" +
		"require __DIR__ . '/other.php';\n" +
		"function late() { return null; }\n"
	_, err = ParseStaticModule("dep.php", src4, "8.3", Options{}, false)
	if err == nil {
		t.Fatal("dependency function after load must still be rejected")
	}
	if !strings.Contains(err.Error(), "module declarations must precede") {
		t.Fatalf("expected old error message, got: %v", err)
	}

	// Dependency: class after load (still rejected)
	src5 := "<?php\n" +
		"require __DIR__ . '/other.php';\n" +
		"class Late {}"
	_, err = ParseStaticModule("dep.php", src5, "8.3", Options{}, false)
	if err == nil {
		t.Fatal("dependency class after load must still be rejected")
	}

	// Dependency: final class after load (still rejected)
	src6 := "<?php\n" +
		"require __DIR__ . '/other.php';\n" +
		"final class Late {}"
	_, err = ParseStaticModule("dep.php", src6, "8.3", Options{}, false)
	if err == nil {
		t.Fatal("dependency final class after load must still be rejected")
	}

	// Entry: final class after body (admitted)
	src7 := "<?php\n" +
		"require __DIR__ . '/lib/dep.php';\n" +
		"$o = new FinalLate();\n" +
		"echo $o->v;\n" +
		"final class FinalLate { public $v = 'final'; }\n"
	pm7, err := ParseStaticModule("entry.php", src7, "8.3", Options{}, true)
	if err != nil {
		t.Fatalf("entry body->final class must be admitted: %v", err)
	}
	if len(pm7.Program.Classes) != 1 || pm7.Program.Classes[0].Name != "FinalLate" {
		t.Fatalf("expected FinalLate: %+v", pm7.Program.Classes)
	}
}

func TestInspectModuleLocationsAndIsolation(t *testing.T) {
	for _, id := range []string{"a.php", "sub/b.php"} {
		module, err := InspectModule(id, "<?php\nfunction broken() { throw 1; }", "8.3", Options{})
		if err == nil || !strings.Contains(err.Error(), id+": PHP002 line 2:") || !reflect.DeepEqual(module, ModuleInspection{}) {
			t.Fatalf("must retain distinct original file at the same line: %+v / %v", module, err)
		}
	}
	const load = "<?php require_once __DIR__ . '/child.php';"
	if _, err := ParseWithOptions(load, "8.3", Options{}); err == nil {
		t.Fatal("ordinary single-file Parse must not silently discard a load")
	}
	for _, id := range []string{"", "../bad.php", "/absolute.php", "bad\x00.php"} {
		if module, err := InspectModule(id, "<?php", "8.3", Options{}); err == nil || !reflect.DeepEqual(module, ModuleInspection{}) {
			t.Fatalf("invalid source identity %q: %+v / %v", id, module, err)
		}
	}
	for _, target := range []string{"8.2", "8.5"} {
		if module, err := InspectModule("main.php", "<?php", target, Options{}); err == nil || !reflect.DeepEqual(module, ModuleInspection{}) {
			t.Fatalf("unsupported target %s: %+v / %v", target, module, err)
		}
	}
	if _, err := InspectModule("main.php", "<?php", "8.4", Options{ExperimentalPCRE2: true}); err == nil {
		t.Fatal("PHP83 regex profile cannot leak to8.4 graph metadata")
	}
}
