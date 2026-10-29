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

func TestArrayPushFixture(t *testing.T) {
	for _, target := range []string{"8.3", "8.4"} {
		t.Run(target, func(t *testing.T) {
			input, err := loadDifferentialInput("tests/fixtures/array_push.php", target)
			if err != nil {
				t.Fatal(err)
			}
			if _, _, err := codegen.GenerateWithMap(input.program); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestArrayPushHostEntry(t *testing.T) {
	const source = `<?php
class Box { public $items; }
class Trace { public $calls = 0; public $after = false; }
function tick(Trace $trace, $value) {
    $trace->calls = $trace->calls + 1;
    return $value;
}
function push_property($target, Trace $trace, $value) {
    $count = array_push($target->items, tick($trace, $value), tick($trace, 'last'));
    $trace->after = true;
    return $count;
}
function push_local($target, Trace $trace) {
    $count = array_push($target, tick($trace, 'value'));
    $trace->after = true;
    return $count;
}
function zero_local($target) { $count = array_push($target); return $count; }
function zero_property($target) { $count = array_push($target->items); return $count; }
function alias_result($target) { $target = array_push($target, 'value'); return $target; }
function loop_result() {
    $target = [];
    $index = 0;
    while ($index < 2) {
        $target = array_push($target, 'value');
        $index = $index + 1;
    }
    return $target;
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
		"go.mod":  []byte(fmt.Sprintf("module example.invalid/array-push-consumer\n\ngo 1.23.0\n\nrequire github.com/katsuakikureyama/GoHP v0.0.0\nreplace github.com/katsuakikureyama/GoHP => %q\n", root)),
		"consumer_test.go": []byte(`package main

import (
    "fmt"
    "testing"
    rt "github.com/katsuakikureyama/GoHP/runtime"
)

type markerOnly struct{}
func (*markerOnly) PHPObject() bool { return true }
type propertyHost struct { value any; trace *C_trace; gets, sets int }
func (h *propertyHost) PHPObject() bool { return h != nil }
func (h *propertyHost) Get(name string) any {
    if name != "items" { panic("undefined property") }
    want := int64(0)
    if h.gets > 0 { want = 2 }
    if h.trace.Get("calls") != want { panic("property address/value evaluation order changed") }
    h.gets++
    return rt.Clone(h.value)
}
func (h *propertyHost) Set(name string, value any) {
    if name != "items" || h.trace.Get("calls") != int64(2) { panic("premature publication") }
    h.sets++
    h.value = rt.Clone(value)
}

func TestConsumer(t *testing.T) {
    expectPanic := func(call func(), want string) {
        t.Helper()
        defer func() { if got := recover(); fmt.Sprint(got) != want { t.Fatalf("panic = %v, want %q", got,want) } }()
        call()
    }
    checkReject := func(target any, property bool, calls int64, want string) {
        t.Helper()
        trace := New_trace().(*C_trace)
        expectPanic(func() {
            if property { F_push_property(target,trace,"value") } else { F_push_local(target,trace) }
        },want)
        if trace.Get("calls") != calls || trace.Get("after") != false { t.Fatal("argument order/count or failure continuation changed") }
    }
    var typedNil *C_box
    for _, bad := range []any{nil,typedNil,false,int64(0),"text",rt.NewArray(),&markerOnly{}} {
        checkReject(bad,true,0,"nested property requires object")
    }
    for _, bad := range []any{nil,false,int64(0),"text",New_box(),rt.NewCallable(0,func(...any) any{return nil})} {
        checkReject(bad,false,1,"array_push requires array")
        expectPanic(func(){F_zero_local(bad)},"array_push requires array")
    }
    box := New_box().(*C_box)
    checkReject(box,true,2,"array_push requires array")
    if box.Get("items") != nil { t.Fatal("invalid property was mutated") }
    expectPanic(func(){F_zero_property(box)},"array_push requires array")
    expectPanic(func(){F_loop_result()},"array_push requires array")

    input := rt.NewArray(rt.Entry{Value:"before"})
    trace := New_trace().(*C_trace)
    if F_push_local(input,trace) != int64(2) || trace.Get("calls") != int64(1) || trace.Get("after") != true { t.Fatal("local result or effects lost") }
    if rt.Count(input) != int64(1) || F_zero_local(input) != int64(1) || F_alias_result(input) != int64(2) { t.Fatal("caller copy, zero-value count or result alias changed") }

    box.Set("items",input)
    trace = New_trace().(*C_trace)
    if F_push_property(box,trace,"next") != int64(3) || trace.Get("calls") != int64(2) || trace.Get("after") != true { t.Fatal("generated property result/effects lost") }
    if rt.Get(box.Get("items"),int64(1)) != "next" || rt.Get(box.Get("items"),int64(2)) != "last" || rt.Count(input) != int64(1) { t.Fatal("property append order or caller copy lost") }
    if F_zero_property(box) != int64(3) { t.Fatal("zero-value property count changed") }

    hostTrace := New_trace().(*C_trace)
    host := &propertyHost{value:input,trace:hostTrace}
    if F_push_property(host,hostTrace,"host") != int64(3) || host.gets != 2 || host.sets != 2 || rt.Get(host.value,int64(1)) != "host" { t.Fatal("explicit full host protocol or evaluation/publication order lost") }
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
