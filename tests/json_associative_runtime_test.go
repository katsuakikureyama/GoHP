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

// The committed fixture is intentionally a PHP-side observation rather than a
// generated serializer. Native PHP remains the differential oracle for it.
func TestJSONAssociativeFixture(t *testing.T) {
	for _, target := range []string{"8.3", "8.4"} {
		t.Run(target, func(t *testing.T) {
			input, err := loadDifferentialInput("tests/fixtures/json_associative.php", target)
			if err != nil {
				t.Fatal(err)
			}
			if _, _, err := codegen.GenerateWithMap(input.program); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// Build a separate module so the assertions use only generated public ABI,
// including FContext_* methods and legacy F_* request wrappers.
func TestJSONAssociativeHostEntry(t *testing.T) {
	const source = `<?php
function decode($text) { return json_decode($text, true); }
function error_code() { return json_last_error(); }
function error_message() { return json_last_error_msg(); }
function through_function($text) { $decoded = decode($text); return json_last_error(); }
function make_arrow() { return fn($text) => json_decode($text, true); }
function invoke_arrow(callable $callback, $text) { return $callback($text); }
function map_decode($text) { return json_decode($text, true); }
function map_values($values) { return array_map('map_decode', $values); }
class JsonDecodeProbe {
    public function decode_error_code($text) { $decoded = json_decode($text, true); return json_last_error(); }
}
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
		"go.mod":  []byte(fmt.Sprintf("module example.invalid/json-associative-consumer\n\ngo 1.23.0\n\nrequire github.com/katsuakikureyama/GoHP v0.0.0\nreplace github.com/katsuakikureyama/GoHP => %q\n", root)),
		"consumer_test.go": []byte(`package main

import (
    "fmt"
    "strings"
    "testing"
    rt "github.com/katsuakikureyama/GoHP/runtime"
)

func TestConsumer(t *testing.T) {
    check := func(got, want any) {
        t.Helper()
        if got != want { t.Fatalf("got %#v (%T), want %#v (%T)", got, got, want, want) }
    }
    expectPanic := func(call func(), want string) {
        t.Helper()
        defer func() {
            got := recover()
            if got == nil || !strings.Contains(fmt.Sprint(got), want) {
                t.Fatalf("panic = %v, want text %q", got, want)
            }
        }()
        call()
    }
    decodeError := func(execution *rt.Execution, input string, want int64, message string) {
        t.Helper()
        if got := FContext_decode(execution, input); got != nil { t.Fatalf("invalid %q = %#v, want nil", input, got) }
        check(FContext_error_code(execution), want)
        check(FContext_error_message(execution), message)
    }

    a, b := &rt.Execution{}, &rt.Execution{}
    check(FContext_error_code(a), int64(0))
    check(FContext_error_message(a), "No error")

    decoded := FContext_decode(a, "{\"b\":1,\"0\":\"zero\",\"b\":3,\"items\":[true,null]}").(rt.Array)
    if len(decoded.Entries) != 3 || decoded.Entries[0].Key != "b" || decoded.Entries[1].Key != int64(0) || decoded.Entries[2].Key != "items" {
        t.Fatalf("object order/key normalization = %#v", decoded)
    }
    check(rt.Get(decoded, "b"), int64(3))
    check(rt.Get(decoded, int64(0)), "zero")
    items := rt.Get(decoded, "items").(rt.Array)
    check(rt.Get(items, int64(0)), true)
    check(rt.Get(items, int64(1)), nil)
    check(FContext_decode(a, "\"text\""), "text")
    check(FContext_decode(a, "true"), true)
    check(FContext_decode(a, "null"), nil)
    check(FContext_error_code(a), int64(0))

    decodeError(a, "{\"broken\":}", int64(4), "Syntax error")
    check(FContext_error_code(b), int64(0))
    check(FContext_error_message(b), "No error")
    check(FContext_through_function(a, "{\"broken\":}"), int64(4))
    probe := New_jsondecodeprobe().(*C_jsondecodeprobe)
    check(probe.MContext_decode_error_code(a, "{\"broken\":}"), int64(4))

    // An escaped arrow executes with its invoking execution, not the creation
    // request. This is the same ABI contract as other stateful builtins.
    check(FContext_decode(a, "null"), nil)
    arrow := FContext_make_arrow(a)
    check(FContext_invoke_arrow(b, arrow, "{\"broken\":}"), nil)
    check(FContext_error_code(b), int64(4))
    check(FContext_error_code(a), int64(0))

    mapped := FContext_map_values(a, rt.NewArray(
        rt.Entry{Value: "{\"id\":1}"}, rt.Entry{Value: "{\"id\":2}"},
    )).(rt.Array)
    check(rt.Get(rt.Get(mapped, int64(0)), "id"), int64(1))
    check(rt.Get(rt.Get(mapped, int64(1)), "id"), int64(2))
    check(FContext_error_code(a), int64(0))
    check(FContext_error_message(a), "No error")

    decodeError(a, "{\"object\":1]", int64(2), "State mismatch (invalid or malformed JSON)")
    decodeError(a, "{\"text\":\"a\nb\"}", int64(3), "Control character error, possibly incorrectly encoded")
    decodeError(a, string([]byte{'"', 0xff, '"'}), int64(5), "Malformed UTF-8 characters, possibly incorrectly encoded")
    decodeError(a, "\"\\uD800\"", int64(10), "Single unpaired UTF-16 surrogate in unicode escape")

    // Unsupported JSON number forms fail before replacing the previous error.
    decodeError(a, "{\"broken\":}", int64(4), "Syntax error")
    for _, bad := range []any{nil, false, int64(1), float64(1.5), rt.NewArray(), rt.NewCallable(0,func(...any)any{return nil})} {
        expectPanic(func() { FContext_decode(a, bad) }, "json_decode requires actual string")
        check(FContext_error_code(a), int64(4))
    }
    expectPanic(func() { FContext_decode(a, "1.5") }, "json_decode number outside int64 profile")
    check(FContext_error_code(a), int64(4))
    check(FContext_error_message(a), "Syntax error")
    expectPanic(func() { FContext_decode(a, "9223372036854775808") }, "json_decode number outside int64 profile")
    check(FContext_error_code(a), int64(4))

    // A valid call resets the state. Standalone wrappers allocate a fresh
    // Execution, and nil contexts are rejected before work is attempted.
    check(FContext_decode(a, "{\"ok\":1}").(rt.Array).Entries[0].Value, int64(1))
    check(FContext_error_code(a), int64(0))
    check(F_decode("{\"broken\":}"), nil)
    check(F_error_code(), int64(0))
    expectPanic(func() { FContext_error_code(nil) }, "execution context")
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
		t.Fatalf("generated consumer: %v\n%s", err, output)
	}
}
