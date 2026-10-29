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
			!reflect.DeepEqual(module.Requires, []ModuleRequire{{5, "/nested/child.php"}, {6, "/nested/child.php"}}) {
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
