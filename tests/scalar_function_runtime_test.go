package tests

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/codegen"
	"github.com/katsuakikureyama/GoHP/internal/frontend"
)

// These are runtime boundaries for global scalar parameters. They deliberately
// route values through an untyped function so code generation cannot reject the
// call solely from its source literal.
func TestScalarFunctionRuntimeGuards(t *testing.T) {
	checkArrayRuntimeGuards(t, []arrayRuntimeCase{
		{"int_numeric_string", `<?php function pass($value) { return $value; } function accept(int $value) { return $value; } $result = accept(pass('1'));`, "int parameter requires exact type"},
		{"string_integer", `<?php function pass($value) { return $value; } function accept(string $value) { return $value; } $result = accept(pass(1));`, "string parameter requires exact type"},
		{"bool_zero", `<?php function pass($value) { return $value; } function accept(bool $value) { return $value; } $result = accept(pass(0));`, "bool parameter requires exact type"},
		{"int_null", `<?php function pass($value) { return $value; } function accept(int $value) { return $value; } $result = accept(pass(null));`, "int parameter requires exact type"},
		{"int_array", `<?php function pass($value) { return $value; } function accept(int $value) { return $value; } $result = accept(pass([]));`, "int parameter requires exact type"},
		{"int_object", `<?php class Item {} function pass($value) { return $value; } function accept(int $value) { return $value; } $result = accept(pass(new Item()));`, "int parameter requires exact type"},
		{"array_map_wrong_element", `<?php function accept(int $value) { return $value; } $result = array_map('accept', ['wrong']);`, "int parameter requires exact type"},
		// An annotation is an entry guard only. A later scalar reassignment is
		// legal, and its consumer must still perform its own type check.
		{"local_reassignment", `<?php function pass($value) { return $value; } function accept(int $value) { $value = pass(false); return strlen($value); } $result = accept(1);`, "strlen requires string"},
		// This body would fail separately, but the parameter guard must run first.
		{"guard_before_body", `<?php function pass($value) { return $value; } function accept(string $value) { return strtok(','); } $result = accept(pass(false));`, "string parameter requires exact type"},
	})
}

// Compile generated source as an external module, then call its exported Go
// ABI directly. One subprocess covers success, rejection, arity, and explicit
// execution-context behavior without treating Go's types as PHP source types.
func TestScalarFunctionHostEntry(t *testing.T) {
	const source = `<?php
function typed(int $number, string $text, bool $flag) {
    $copy = $number;
    $number = $flag ? $number : 0;
    return $copy;
}
function enable_context() { return strtok('a,b', ','); }
`
	p, err := frontend.Parse(source, phpTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	generated, err := codegen.Generate(p)
	if err != nil {
		t.Fatal(err)
	}
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	files := map[string][]byte{
		"main.go": generated,
		"go.mod":  []byte(fmt.Sprintf("module example.invalid/scalar-function-consumer\n\ngo 1.23\n\nrequire github.com/katsuakikureyama/GoHP v0.0.0\nreplace github.com/katsuakikureyama/GoHP => %q\n", root)),
		"consumer_test.go": []byte(`package main

import (
    "fmt"
    "strings"
    "testing"

    rt "github.com/katsuakikureyama/GoHP/runtime"
)

func TestConsumer(t *testing.T) {
    reject := func(want string, call func()) {
        t.Helper()
        defer func() {
            got := recover()
            if got == nil || !strings.Contains(fmt.Sprint(got), want) {
                t.Fatalf("panic = %v, want %q", got, want)
            }
        }()
        call()
    }
    check := func(got, want any) {
        t.Helper()
        if got != want {
            t.Fatalf("got %#v, want %#v", got, want)
        }
    }

    check(F_typed(int64(7), "seven", true), int64(7))
    check(F_typed(int64(0), "zero", false), int64(0))
    reject("int parameter requires exact type", func() { F_typed(int(7), "x", true) })
    reject("int parameter requires exact type", func() { F_typed(float64(7), "x", true) })
    var nilPointer *int
    reject("int parameter requires exact type", func() { F_typed(nilPointer, "x", true) })
    reject("int parameter requires exact type", func() { F_typed(rt.NewArray(), "x", true) })
    reject("int parameter requires exact type", func() { F_typed(false, "x", true) })
    reject("string parameter requires exact type", func() { F_typed(int64(7), nil, true) })
    reject("bool parameter requires exact type", func() { F_typed(int64(7), "x", int64(1)) })
    reject("argument count", func() { F_typed(int64(1), "x") })
    reject("argument count", func() { F_typed(int64(1), "x", true, "extra") })
    check(F_typed(int64(9), "after-rejection", true), int64(9))

    reject("execution context", func() { FContext_typed(nil, int64(1), "x", true) })
    execution := &rt.Execution{}
    check(FContext_enable_context(execution), "a")
    check(FContext_typed(execution, int64(3), "context", true), int64(3))
    check(execution.Strtok(","), "b")
}
`),
	}
	for name, contents := range files {
		if err := os.WriteFile(filepath.Join(dir, name), contents, 0600); err != nil {
			t.Fatal(err)
		}
	}
	cmd := exec.Command("go", "test", "-count=1", "./...")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off", "GOPROXY=off", "GOSUMDB=off", "GOFLAGS=-mod=mod")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("external consumer: %v\n%s\n%s", err, output, generated)
	}
}
