package codegen

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestVersionConstantLowering(t *testing.T) {
	source := versionConstantFixture(t)
	program, err := frontend.Parse(source, "8.3")
	if err != nil {
		t.Fatal(err)
	}

	constants, calls := versionConstantExpressions(program)
	if len(constants) == 0 {
		t.Fatal("fixture did not retain predefined version constants in IR")
	}
	for _, constant := range constants {
		if constant.Kind != "version_constant" || (constant.Text != "PHP_MAJOR_VERSION" && constant.Text != "PHP_MINOR_VERSION") || len(constant.Args) != 0 || len(constant.Params) != 0 {
			t.Fatalf("noncanonical version constant IR: %#v", constant)
		}
	}
	if !hasVersionConstantAt(constants, "PHP_MAJOR_VERSION", 8) || !hasVersionConstantAt(constants, "PHP_MINOR_VERSION", 8) {
		t.Fatalf("version identity lost its symbolic source expressions: %#v", constants)
	}
	if !hasCallAt(calls, "PHP_MINOR_VERSION", 32) {
		t.Fatalf("same-spelled function call was lowered as a constant: %#v", calls)
	}

	var first83 []byte
	for _, tc := range []struct {
		target, minor string
	}{
		{"8.3", "3"},
		{"8.4", "4"},
		{"8.3", "3"},
	} {
		// Retarget this exact parsed IR. Generation must not cache the earlier
		// target or replace the symbolic nodes in the program.
		program.PHPVersion = tc.target
		generated, sourceMap, err := GenerateWithMap(program)
		if err != nil {
			t.Fatalf("target %s: %v", tc.target, err)
		}
		text := string(generated)
		for _, want := range []string{
			"int64(8)",
			"int64(" + tc.minor + ")",
			"F_typed_increment(int64(" + tc.minor + "))",
			"rt.NewScalarFrame()",
			"F_php_minor_version()",
		} {
			if !strings.Contains(text, want) {
				t.Fatalf("target %s missing %q:\n%s", tc.target, want, generated)
			}
		}
		expectVersionConstantMapped(t, generated, sourceMap, "F_typed_increment(int64("+tc.minor+"))", 28)
		expectVersionConstantMapped(t, generated, sourceMap, "rt.Identical(int64("+tc.minor+"), int64(3))", 17)
		identityStart := strings.Index(text, "func F_version_identity(")
		if identityStart < 0 {
			t.Fatal("missing version_identity function")
		}
		identityBody := strings.SplitN(text[identityStart:], "\nfunc ", 2)[0]
		if !strings.Contains(identityBody, "rt.NewScalarFrame()") {
			t.Fatalf("version constant stopped closed scalar frame selection: %s", identityBody)
		}

		if tc.target == "8.3" {
			if first83 == nil {
				first83 = generated
			} else if !bytes.Equal(first83, generated) {
				t.Fatalf("8.3 generation leaked target state across retargeting:\nfirst:\n%s\nsecond:\n%s", first83, generated)
			}
		} else if bytes.Equal(first83, generated) {
			t.Fatal("8.3 and 8.4 generation were identical despite different minor constants")
		}
	}

	constants, calls = versionConstantExpressions(program)
	if !hasVersionConstantAt(constants, "PHP_MINOR_VERSION", 8) || !hasCallAt(calls, "PHP_MINOR_VERSION", 32) {
		t.Fatalf("generation mutated symbolic IR: constants=%#v calls=%#v", constants, calls)
	}
}

func TestVersionConstantBoundaries(t *testing.T) {
	for name, source := range map[string]string{
		"case_variant_major": "<?php echo php_major_version;",
		"case_variant_minor": "<?php echo Php_MINOR_VERSION;",
		"other_version":      "<?php echo PHP_VERSION;",
		"version_id":         "<?php echo PHP_VERSION_ID;",
		"release":            "<?php echo PHP_RELEASE_VERSION;",
		"extra":              "<?php echo PHP_EXTRA_VERSION;",
		"undefined":          "<?php echo GOHP_UNKNOWN_CONSTANT;",
		"define":             "<?php define('GOHP_DEFINED', 1);",
		"constant_function":  "<?php echo constant('PHP_MINOR_VERSION');",
		"qualified":          "<?php echo Namespace\\PHP_MINOR_VERSION;",
		"global_qualified":   "<?php echo \\PHP_MINOR_VERSION;",
		"declaration":        "<?php const LOCAL_VERSION = 3;",
		"property_default":   "<?php class Box { public $value = PHP_MINOR_VERSION; }",
		"array_key":          "<?php $items = [PHP_MINOR_VERSION => 'value'];",
		"array_write_key":    "<?php $items = []; $items[PHP_MINOR_VERSION] = 'value';",
		"unary_negative":     "<?php echo -PHP_MINOR_VERSION;",
		"default_argument":   "<?php function f($value = PHP_MINOR_VERSION) { return $value; }",
		"isset_key":          "<?php $items = []; echo isset($items[PHP_MINOR_VERSION]);",
		"coalesce_key":       "<?php $items = []; echo $items[PHP_MINOR_VERSION] ?? 0;",
		"string_parameter":   "<?php function f(string $value) {} f(PHP_MINOR_VERSION);",
		"undeclared_call":    "<?php echo PHP_MINOR_VERSION();",
	} {
		t.Run(name, func(t *testing.T) {
			program, err := frontend.Parse(source, "8.3")
			var output []byte
			if err == nil {
				output, err = Generate(program)
			}
			if err == nil || len(output) != 0 {
				t.Fatalf("outside-subset form unexpectedly generated %q: %v", output, err)
			}
		})
	}

	program, err := frontend.Parse(`<?php
function PHP_MINOR_VERSION() { return 90; }
echo PHP_MINOR_VERSION;
echo PHP_MINOR_VERSION();
`, "8.3")
	if err != nil {
		t.Fatal(err)
	}
	constants, calls := versionConstantExpressions(program)
	if !hasVersionConstantAt(constants, "PHP_MINOR_VERSION", 3) || !hasCallAt(calls, "PHP_MINOR_VERSION", 4) {
		t.Fatalf("constant/function namespaces were not separated: constants=%#v calls=%#v", constants, calls)
	}
	if output, err := Generate(program); err != nil || !strings.Contains(string(output), "F_php_minor_version()") {
		t.Fatalf("same-named function and constant were not both accepted: %v\n%s", err, output)
	}
}

func TestVersionConstantMalformedIR(t *testing.T) {
	valid := ir.Expr{Kind: "version_constant", Text: "PHP_MINOR_VERSION", Line: 1}
	for name, program := range map[string]ir.Program{
		"unknown_name": {
			Version: ir.Version, PHPVersion: "8.3",
			Body: []ir.Stmt{{Kind: "echo", Line: 1, Expr: ir.Expr{Kind: "version_constant", Text: "PHP_VERSION", Line: 1}}},
		},
		"extra_args": {
			Version: ir.Version, PHPVersion: "8.3",
			Body: []ir.Stmt{{Kind: "echo", Line: 1, Expr: ir.Expr{Kind: "version_constant", Text: "PHP_MINOR_VERSION", Line: 1, Args: []ir.Expr{{Kind: "int", Text: "1", Line: 1}}}}},
		},
		"extra_params": {
			Version: ir.Version, PHPVersion: "8.3",
			Body: []ir.Stmt{{Kind: "echo", Line: 1, Expr: ir.Expr{Kind: "version_constant", Text: "PHP_MINOR_VERSION", Line: 1, Params: []string{"unexpected"}}}},
		},
		"invalid_target": {
			Version: ir.Version, PHPVersion: "8.5",
			Body: []ir.Stmt{{Kind: "echo", Line: 1, Expr: valid}},
		},
	} {
		t.Run(name, func(t *testing.T) {
			output, err := Generate(program)
			if err == nil || len(output) != 0 {
				t.Fatalf("malformed version constant IR generated %q: %v", output, err)
			}
		})
	}
}

func versionConstantFixture(t *testing.T) string {
	t.Helper()
	source, err := os.ReadFile(filepath.Join("..", "..", "tests", "fixtures", "version_constants.php"))
	if err != nil {
		t.Fatal(err)
	}
	return string(source)
}

func versionConstantExpressions(program ir.Program) (constants, calls []ir.Expr) {
	var expression func(ir.Expr)
	expression = func(expr ir.Expr) {
		if expr.Kind == "version_constant" {
			constants = append(constants, expr)
		}
		if expr.Kind == "call" {
			calls = append(calls, expr)
		}
		for _, arg := range expr.Args {
			expression(arg)
		}
	}
	var statements func([]ir.Stmt)
	statements = func(body []ir.Stmt) {
		for _, statement := range body {
			expression(statement.Expr)
			expression(statement.Target)
			statements(statement.Body)
			statements(statement.ElseBody)
		}
	}
	statements(program.Body)
	for _, function := range program.Functions {
		statements(function.Body)
	}
	for _, class := range program.Classes {
		for _, property := range class.Properties {
			expression(property.Value)
		}
		for _, method := range class.Methods {
			statements(method.Body)
		}
	}
	return constants, calls
}

func hasVersionConstantAt(expressions []ir.Expr, name string, line int) bool {
	for _, expression := range expressions {
		if expression.Text == name && expression.Line == line {
			return true
		}
	}
	return false
}

func hasCallAt(expressions []ir.Expr, name string, line int) bool {
	for _, expression := range expressions {
		if expression.Text == name && expression.Line == line && len(expression.Args) == 0 {
			return true
		}
	}
	return false
}

func expectVersionConstantMapped(t *testing.T, generated []byte, sourceMap SourceMap, fragment string, wantLine int) {
	t.Helper()
	for line, text := range strings.Split(string(generated), "\n") {
		if !strings.Contains(text, fragment) {
			continue
		}
		if got, ok := sourceMap.PHPLine(line + 1); !ok || got != wantLine {
			t.Fatalf("%q maps to PHP line %d (%v), want %d", fragment, got, ok, wantLine)
		}
		return
	}
	t.Fatalf("missing generated fragment %q:\n%s", fragment, generated)
}
