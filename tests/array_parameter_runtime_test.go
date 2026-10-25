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

func TestArrayParameterRuntimeGuards(t *testing.T) {
	checkArrayRuntimeGuards(t, []arrayRuntimeCase{
		{"unknown_null", "<?php function pass($value) { return $value; } function accept(array $value) { return $value; } $result = accept(pass(null));", "array parameter requires array"},
		{"unknown_boolean", "<?php function pass($value) { return $value; } function accept(array $value) { return $value; } $result = accept(pass(true));", "array parameter requires array"},
		{"unknown_integer", "<?php function pass($value) { return $value; } function accept(array $value) { return $value; } $result = accept(pass(1));", "array parameter requires array"},
		{"unknown_string", "<?php function pass($value) { return $value; } function accept(array $value) { return $value; } $result = accept(pass('items'));", "array parameter requires array"},
		{"unknown_object", "<?php class Item {} function pass($value) { return $value; } function accept(array $value) { return $value; } $result = accept(pass(new Item()));", "array parameter requires array"},
		{"array_map_scalar_element", "<?php function callback(array $value): array { return $value; } $result = array_map('callback', [1]);", "array parameter requires array"},
	})
}

func TestArrayParameterGuardPrecedesFunctionBody(t *testing.T) {
	source := "<?php function pass($value) { return $value; } function accept(array $value) { echo 'body-ran'; return $value; } $result = accept(pass(null));"
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
	if err == nil || !bytes.Contains(output, []byte("array parameter requires array")) {
		t.Fatalf("expected explicit runtime rejection, got %v\n%s", err, output)
	}
	if bytes.Contains(output, []byte("body-ran")) {
		t.Fatalf("typed parameter body ran before its guard: %s", output)
	}
}
