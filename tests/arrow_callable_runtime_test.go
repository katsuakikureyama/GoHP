package tests

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/codegen"
	"github.com/katsuakikureyama/GoHP/internal/frontend"
)

func TestArrowCallableRuntimeGuards(t *testing.T) {
	checkArrayRuntimeGuards(t, []arrayRuntimeCase{
		{"unknown_null_callable_parameter", "<?php function pass($value) { return $value; } function accept(callable $callback): bool { return $callback(); } $result = accept(pass(null));", "callable requires generated arrow"},
		{"unknown_string_callable_parameter", "<?php function pass($value) { return $value; } function accept(callable $callback): bool { return $callback(); } $result = accept(pass('callback'));", "callable requires generated arrow"},
		{"unknown_array_callable_parameter", "<?php function pass($value) { return $value; } function accept(callable $callback): bool { return $callback(); } $result = accept(pass([]));", "callable requires generated arrow"},
		{"unknown_invalid_variable_callee", "<?php function pass($value) { return $value; } $callback = pass(null); $result = $callback();", "callable requires generated arrow"},
		{"unknown_variable_callee_wrong_arity", "<?php function pass($value) { return $value; } $callback = pass(fn($value) => $value); $result = $callback();", "callable argument count mismatch"},
		{"extra_arguments", "<?php $callback = fn() => true; $result = $callback(1);", "callable argument count mismatch"},
		{"unknown_callable_bool", "<?php function pass($value) { return $value; } echo (bool)pass(fn() => true);", "boolean conversion outside supported scalar/array subset"},
		{"unknown_callable_strict", "<?php function pass($value) { return $value; } echo pass(fn() => true) === false;", "strict equality outside supported scalar subset"},
	})
}

func TestArrowCallableParameterGuardPrecedesFunctionBody(t *testing.T) {
	source := "<?php function pass($value) { return $value; } function accept(callable $callback): bool { echo 'body-ran'; return true; } $result = accept(pass(null));"
	p, err := frontend.Parse(source, phpTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	generated, err := codegen.Generate(p)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	file, binary := filepath.Join(dir, "main.go"), filepath.Join(dir, "app")
	if err := os.WriteFile(file, generated, 0600); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command("go", "build", "-o", binary, file).CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, output)
	}
	run := exec.Command(binary)
	run.Env = []string{"PATH="}
	output, err := run.CombinedOutput()
	if err == nil || !bytes.Contains(output, []byte("callable requires generated arrow")) {
		t.Fatalf("expected explicit runtime rejection, got %v\n%s", err, output)
	}
	if bytes.Contains(output, []byte("body-ran")) {
		t.Fatalf("typed parameter body ran before its guard: %s", output)
	}
}
