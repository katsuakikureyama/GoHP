package sourceclosure

import (
	"reflect"
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/codegen"
	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestAssembleClosure(t *testing.T) {
	files := map[string]string{
		"main.php":              "<?php\nrequire_once __DIR__ . '/lib/first.php';\nrequire_once __DIR__ . '/lib/second.php';\nrequire_once __DIR__ . '/lib/first.php';\n$x = new Shared('ok');\necho $x->value;\necho helper();\n",
		"lib/first.php":         "<?php\nrequire_once __DIR__ . '/nested/shared.php';\n",
		"lib/second.php":        "<?php\nfunction helper() { return '!'; }\nrequire_once __DIR__ . '/nested/shared.php';\n",
		"lib/nested/shared.php": "<?php\nclass Shared { public $value; public function __construct(string $value) { $this->value = $value; } }\n",
	}
	root, manifest := writeFixture(t, files, []string{"8.3", "8.4"})
	for _, target := range []string{"8.3", "8.4"} {
		program, graph, err := Assemble(manifest, root, "main.php", target, frontend.Options{})
		if err != nil {
			t.Fatal(err)
		}
		if len(program.Sources) != 4 || len(program.Classes) != 1 || len(program.Functions) != 1 || len(program.Body) != 3 || len(graph.Edges) != 5 {
			t.Fatalf("unexpected resolved closure: %+v / %+v", program, graph)
		}
		for _, statement := range program.Body {
			if statement.Kind == frontend.StaticRequireOnce || statement.SourceID != "main.php" {
				t.Fatalf("unresolved or unlocated entry: %+v", statement)
			}
		}
		generated, sourceMap, err := codegen.GenerateWithMap(program)
		if err != nil || len(generated) == 0 || sourceMap.Schema != "gohp.sourcemap/v0.2" {
			t.Fatalf("generate closure: %v / %+v", err, sourceMap)
		}
		seen := map[string]bool{}
		for _, mapping := range sourceMap.Mappings {
			location, ok := sourceMap.PHPLocation(mapping.GeneratedStartLine)
			if !ok || location.SourceID != mapping.SourceID || location.Line != mapping.SourceLine {
				t.Fatalf("lost original closure location: %+v / %+v", mapping, location)
			}
			seen[location.SourceID] = true
		}
		for _, id := range []string{"main.php", "lib/second.php", "lib/nested/shared.php"} {
			if !seen[id] {
				t.Fatalf("missing original file in generated map: %s", id)
			}
		}
	}
}

func TestUnresolvedModuleRejectsCodegen(t *testing.T) {
	module, err := frontend.ParseStaticModule("main.php", "<?php\nrequire_once __DIR__ . '/child.php';", "8.3", frontend.Options{}, false)
	if err != nil || len(module.Program.Body) != 1 || module.Program.Body[0].Kind != frontend.StaticRequireOnce {
		t.Fatalf("load must remain explicit before resolution: %+v / %v", module, err)
	}
	generated, sourceMap, err := codegen.GenerateWithMap(module.Program)
	if err == nil || !strings.Contains(err.Error(), "main.php") || len(generated) != 0 || !reflect.DeepEqual(sourceMap, codegen.SourceMap{}) {
		t.Fatalf("unresolved loads must not be silently dropped: %s / %+v / %v", generated, sourceMap, err)
	}
}

func TestAssembleRejectsEffectsAndLateLoads(t *testing.T) {
	for _, files := range []map[string]string{
		{"main.php": "<?php echo 'early'; require_once __DIR__ . '/child.php';", "child.php": "<?php"},
		{"main.php": "<?php if (true) { require_once __DIR__ . '/child.php'; }", "child.php": "<?php"},
		{"main.php": "<?php require_once __DIR__ . '/child.php'; echo 'ok';", "child.php": "<?php echo 'effect';"},
		{"main.php": "<?php return include('hidden.php');"},
	} {
		root, manifest := writeFixture(t, files, []string{"8.3"})
		program, graph, err := Assemble(manifest, root, "main.php", "8.3", frontend.Options{})
		if err == nil || !reflect.DeepEqual(program, ir.Program{}) || !reflect.DeepEqual(graph, Report{}) {
			t.Fatalf("failed closure leaked partial IR/graph: %+v / %+v / %v", program, graph, err)
		}
	}
}

func TestAssembleDependencyDiagnostic(t *testing.T) {
	files := map[string]string{
		"main.php": "<?php require_once __DIR__ . '/bad.php'; echo broken();",
		"bad.php":  "<?php\nfunction broken() { return unimplemented_function(); }",
	}
	root, manifest := writeFixture(t, files, []string{"8.3"})
	program, _, err := Assemble(manifest, root, "main.php", "8.3", frontend.Options{})
	if err != nil {
		t.Fatal(err)
	}
	generated, _, err := codegen.GenerateWithMap(program)
	if err == nil || !strings.Contains(err.Error(), "bad.php:") || !strings.Contains(err.Error(), "line 2") || len(generated) != 0 {
		t.Fatalf("dependency semantic diagnostic relabeled or ignored: %s / %v", generated, err)
	}
}
