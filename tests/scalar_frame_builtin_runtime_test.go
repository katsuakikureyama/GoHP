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

// An external module verifies that scalar-frame selection remains valid for
// the closed, typed builtin set through the generated callable ABI. The
// guarded function keeps an unknown scalar union inside its owned frame so a
// builtin runtime type failure cannot poison later scalar-frame calls.
func TestScalarFrameBuiltinExternalConsumer(t *testing.T) {
	const source = `<?php
function byte_length(string $value) { return strlen($value); }
function lower(string $value) { return strtolower($value); }
function compare(string $left, string $right) { return strcasecmp($left, $right); }
function trim_suffix(string $value) { return rtrim($value, ' '); }
function slice(string $value, int $offset, int $length) { return substr($value, $offset, $length); }
function contains(string $haystack, string $needle) { return str_contains($haystack, $needle); }
function starts(string $haystack, string $needle) { return str_starts_with($haystack, $needle); }
function bool_value(bool $value) { return is_bool($value); }
function string_value(string $value) { return is_string($value); }
function int_value(int $value) { return is_int($value); }
function null_value(int $value) { return is_null($value); }
function array_value(bool $value) { return is_array($value); }
function object_value(string $value) { return is_object($value); }
function nested(string $value) { return strlen(strtolower($value)); }
function repeated(string $value) { strlen($value); strtolower($value); return strtolower($value); }
function runtime_guard(string $value, bool $use_value) { $candidate = $use_value ? $value : false; return strlen($candidate); }
function token_start() { return strtok('a,b', ','); }
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
		"go.mod":  []byte(fmt.Sprintf("module example.invalid/scalar-frame-builtins-consumer\n\ngo 1.23\n\nrequire github.com/katsuakikureyama/GoHP v0.0.0\nreplace github.com/katsuakikureyama/GoHP => %q\n", root)),
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

    execution := &rt.Execution{}
    check(FContext_token_start(execution), "a")
    for i := 0; i < 2; i++ {
        check(FContext_byte_length(execution, "GoHP"), int64(4))
    }
    check(FContext_lower(execution, "GoHp"), "gohp")
    check(FContext_compare(execution, "GoHp", "goHP"), int64(0))
    check(FContext_trim_suffix(execution, "tail   "), "tail")
    check(FContext_slice(execution, "abcdef", int64(1), int64(3)), "bcd")
    check(FContext_contains(execution, "abcdef", "cde"), true)
    check(FContext_starts(execution, "abcdef", "abc"), true)
    check(FContext_bool_value(execution, true), true)
    check(FContext_string_value(execution, "text"), true)
    check(FContext_int_value(execution, int64(5)), true)
    check(FContext_null_value(execution, int64(5)), false)
    check(FContext_array_value(execution, false), false)
    check(FContext_object_value(execution, "text"), false)
    check(FContext_nested(execution, "GoHP"), int64(4))
    check(FContext_repeated(execution, "GoHP"), "gohp")

    reject("string parameter requires exact type", func() { FContext_lower(execution, int64(1)) })
    reject("int parameter requires exact type", func() { FContext_slice(execution, "abc", "1", int64(1)) })
    reject("bool parameter requires exact type", func() { FContext_bool_value(execution, int64(1)) })
    reject("argument count", func() { FContext_lower(execution) })
    reject("execution context", func() { FContext_lower(nil, "x") })
    reject("strlen requires string", func() { FContext_runtime_guard(execution, "x", false) })
    check(FContext_lower(execution, "After Panic"), "after panic")
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
