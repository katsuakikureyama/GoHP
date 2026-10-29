package tests

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/codegen"
	"github.com/katsuakikureyama/GoHP/internal/frontend"
)

// TestEmptyStringHostEntry exercises the generated external-module ABI.  It
// intentionally supplies host values through an untyped parameter so source
// semantic checks cannot discard the actual-domain guard.
func TestEmptyStringHostEntry(t *testing.T) {
	const source = `<?php
function empty_host($value): bool { return empty($value); }
function empty_typed_host(string $value): bool { return empty($value); }
function empty_lazy_host($value, bool $skip): bool { return $skip || empty($value); }
`
	p, err := frontend.Parse(source, phpTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	generated, err := codegen.Generate(p)
	if err != nil {
		t.Fatal(err)
	}
	for _, function := range []string{"empty_host", "empty_typed_host", "empty_lazy_host"} {
		body := generatedFunctionBody(t, generated, "FContext_"+function)
		if !strings.Contains(body, "rt.EmptyValue(") {
			t.Fatalf("%s did not lower through the bounded runtime guard:\n%s", function, body)
		}
	}

	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	files := map[string][]byte{
		"main.go": generated,
		"go.mod":  []byte(fmt.Sprintf("module example.invalid/empty-string-consumer\n\ngo 1.23\n\nrequire github.com/katsuakikureyama/GoHP v0.0.0\nreplace github.com/katsuakikureyama/GoHP => %q\n", root)),
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

    for _, value := range []string{"", "0"} {
        check(F_empty_host(value), true)
        check(F_empty_typed_host(value), true)
    }
    for _, value := range []string{"00", "0.0", " 0", "0 ", " ", "-0", "雪", "\x00", "a\x00b", "\xff"} {
        check(F_empty_host(value), false)
        check(F_empty_typed_host(value), false)
    }

    emptyArray := rt.NewArray()
    oneArray := rt.NewArray(rt.Entry{Key: "item", Value: nil})
    check(F_empty_host(emptyArray), true)
    check(F_empty_host(oneArray), false)
    if len(oneArray.Entries) != 1 || oneArray.Entries[0].Key != "item" || oneArray.Entries[0].Value != nil {
        t.Fatalf("empty mutated caller array: %#v", oneArray)
    }

    for _, value := range []any{
        nil, false, true, int64(0), int(0), float64(0),
        &struct{}{}, []string{"not", "php"}, map[string]int{"not": 1},
        rt.NewCallable(0, func(...any) any { return nil }),
    } {
        reject("empty requires array or string", func() { F_empty_host(value) })
    }
    reject("string parameter requires exact type", func() { F_empty_typed_host(emptyArray) })
    reject("string parameter requires exact type", func() { F_empty_typed_host(int64(0)) })

    // The false right-hand branch is never evaluated, so an otherwise invalid
    // unknown value cannot reach EmptyValue.  The true branch still does.
    check(F_empty_lazy_host(int64(0), true), true)
    check(F_empty_lazy_host("", false), true)
    reject("empty requires array or string", func() { F_empty_lazy_host(int64(0), false) })
    check(F_empty_host("after rejection"), false)
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
		t.Fatalf("external generated empty-string consumer: %v\n%s\n%s", err, output, generated)
	}
}
