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

func TestArrayPopFixture(t *testing.T) {
	for _, target := range []string{"8.3", "8.4"} {
		t.Run(target, func(t *testing.T) {
			input, err := loadDifferentialInput("tests/fixtures/array_pop.php", target)
			if err != nil {
				t.Fatal(err)
			}
			if _, _, err := codegen.GenerateWithMap(input.program); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// Exercise the generated external ABI once. These host-only inputs establish
// the guarded runtime domain; they do not claim native warning equivalence.
func TestArrayPopHostEntry(t *testing.T) {
	const source = `<?php
class Box { public $items; }
class Trace { public $after = false; }
function pop_property($target, Trace $trace) {
    $value = array_pop($target->items);
    $trace->after = true;
    return $value;
}
function pop_local($target, Trace $trace) {
    $value = array_pop($target);
    $trace->after = true;
    return $value;
}
function empty_local($target) { $value = array_pop($target); return $value; }
function empty_property($target) { $value = array_pop($target->items); return $value; }
function alias_result($target) { $target = array_pop($target); return $target; }
function loop_alias() {
    $target = ['scalar'];
    $index = 0;
    while ($index < 2) {
        $target = array_pop($target);
        $index = $index + 1;
    }
    return $target;
}
function reject_popped_scalar($target) {
    $popped = array_pop($target);
    $result = array_pop($popped);
    return $result;
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
		"go.mod":  []byte(fmt.Sprintf("module example.invalid/array-pop-consumer\n\ngo 1.23.0\n\nrequire github.com/katsuakikureyama/GoHP v0.0.0\nreplace github.com/katsuakikureyama/GoHP => %q\n", root)),
		"consumer_test.go": []byte(`package main

import (
    "fmt"
    "testing"
    rt "github.com/katsuakikureyama/GoHP/runtime"
)

type markerOnly struct{}
func (*markerOnly) PHPObject() bool { return true }
// This complete structural host is an explicit opt-in, not proof of
// generated-class provenance or an authorization boundary.
type propertyHost struct { value any; gets, sets int }
func (h *propertyHost) PHPObject() bool { return h != nil }
func (h *propertyHost) Get(name string) any {
    if name != "items" { panic("undefined property") }
    h.gets++
    return rt.Clone(h.value)
}
func (h *propertyHost) Set(name string, value any) {
    if name != "items" { panic("undefined property") }
    h.sets++
    h.value = rt.Clone(value)
}

func TestConsumer(t *testing.T) {
    expectPanic := func(call func(), want string) {
        t.Helper()
        defer func() { if got := recover(); fmt.Sprint(got) != want { t.Fatalf("panic = %v, want %q", got,want) } }()
        call()
    }
    checkReject := func(target any, property bool, want string) {
        t.Helper()
        trace := New_trace().(*C_trace)
        expectPanic(func() {
            if property { F_pop_property(target,trace) } else { F_pop_local(target,trace) }
        },want)
        if trace.Get("after") != false { t.Fatal("failure continuation changed") }
    }
    var typedNil *C_box
    for _, bad := range []any{nil,typedNil,false,int64(0),"text",rt.NewArray(),&markerOnly{}} {
        checkReject(bad,true,"nested property requires object")
    }
    for _, bad := range []any{nil,false,int64(0),"text",New_box(),rt.NewCallable(0,func(...any) any{return nil})} {
        checkReject(bad,false,"array_pop requires array")
        expectPanic(func(){F_empty_local(bad)},"array_pop requires array")
    }
    box := New_box().(*C_box)
    checkReject(box,true,"array_pop requires array")
    if box.Get("items") != nil { t.Fatal("invalid property was mutated") }
    expectPanic(func(){F_empty_property(box)},"array_pop requires array")

    input := rt.NewArray(rt.Entry{Value:"first"},rt.Entry{Value:"last"})
    trace := New_trace().(*C_trace)
    if F_pop_local(input,trace) != "last" || trace.Get("after") != true { t.Fatal("local result or continuation lost") }
    if rt.Count(input) != int64(2) || F_empty_local(rt.NewArray()) != nil || F_alias_result(input) != "last" { t.Fatal("caller copy, empty null, or alias destination-last changed") }
    expectPanic(func(){F_reject_popped_scalar(rt.NewArray(rt.Entry{Value:"scalar"}))},"array_pop requires array")
    expectPanic(func(){F_loop_alias()},"array_pop requires array")

    box.Set("items",input)
    trace = New_trace().(*C_trace)
    if F_pop_property(box,trace) != "last" || trace.Get("after") != true { t.Fatal("property result or continuation lost") }
    items, ok := box.Get("items").(rt.Array)
    if !ok || rt.Count(items) != int64(1) || rt.Get(items,int64(0)) != "first" || rt.Count(input) != int64(2) { t.Fatal("property mutation or caller copy lost") }

    host := &propertyHost{value:input}
    hostTrace := New_trace().(*C_trace)
    if F_pop_property(host,hostTrace) != "last" || host.gets != 2 || host.sets != 1 || hostTrace.Get("after") != true { t.Fatal("explicit full host protocol or property order lost") }
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
