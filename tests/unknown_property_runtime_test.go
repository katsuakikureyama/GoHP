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

// Exercise the generated external ABI once, including typed nil and protocol
// lookalikes that PHP source cannot construct. Invalid domains are explicit
// rejections, not a claim of PHP exception or warning equivalence.
func TestUnknownPropertyWriteHostEntry(t *testing.T) {
	// Keep the committed native fixture within the admitted compiler domain
	// even on developer machines without an installed PHP oracle.
	fixture, err := os.ReadFile("fixtures/unknown_property_write.php")
	if err != nil {
		t.Fatal(err)
	}
	fixtureProgram, err := frontend.Parse(string(fixture), phpTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := codegen.Generate(fixtureProgram); err != nil {
		t.Fatal(err)
	}
	const source = `<?php
class Box { public $value = null; public $items = null; }
class Trace { public $mark = 0; public $after = false; }
function rhs(Trace $trace, $value) { $trace->mark = $trace->mark + 1; return $value; }
function write_unknown($target, Trace $trace, $value) {
    $target->value = rhs($trace, $value);
    $trace->after = true;
}
function append_unknown($target, Trace $trace, $value) {
    $target->items[] = rhs($trace, $value);
    $trace->after = true;
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
		"go.mod":  []byte(fmt.Sprintf("module example.invalid/property-write-consumer\n\ngo 1.23\n\nrequire github.com/katsuakikureyama/GoHP v0.0.0\nreplace github.com/katsuakikureyama/GoHP => %q\n", root)),
		"consumer_test.go": []byte(`package main

import (
    "fmt"
    "strings"
    "testing"
    rt "github.com/katsuakikureyama/GoHP/runtime"
)

type setterOnly struct { calls int }
func (s *setterOnly) Set(string, any) { s.calls++ }
type getterSetter struct { calls int }
func (s *getterSetter) Get(string) any { s.calls++; return nil }
func (s *getterSetter) Set(string, any) { s.calls++ }
type invalidObject struct { getterSetter }
func (*invalidObject) PHPObject() bool { return false }
// An explicit full-protocol host is opt-in, as for existing nested writes.
// The marker is a contract, not an unforgeable generated-class certificate.
type propertyHost struct { value any; trace *C_trace; gets, sets int }
func (h *propertyHost) PHPObject() bool { return h != nil }
func (h *propertyHost) Get(name string) any {
    if name != "value" { panic("undefined host property") }
    if h.trace.Get("mark") != int64(1) { panic("guard ran before RHS") }
    h.gets++
    return rt.Clone(h.value)
}
func (h *propertyHost) Set(name string, value any) {
    if name != "value" { panic("undefined host property") }
    h.sets++
    h.value = rt.Clone(value)
}

func TestConsumer(t *testing.T) {
    reject := func(target any, appendValue bool, want string) {
        t.Helper()
        trace := New_trace().(*C_trace)
        defer func() {
            got := recover()
            if got == nil || !strings.Contains(fmt.Sprint(got), want) {
                t.Fatalf("panic = %v, want %q", got, want)
            }
            if trace.Get("mark") != int64(1) || trace.Get("after") != false {
                t.Fatalf("RHS not exactly once before guard, or following write ran: %#v", trace)
            }
        }()
        if appendValue { F_append_unknown(target, trace, "value") } else { F_write_unknown(target, trace, "value") }
    }
    var typedNil *C_box
    setter := &setterOnly{}
    getters := &getterSetter{}
    invalid := &invalidObject{}
    for _, value := range []any{nil, typedNil, false, int64(1), "text", rt.NewArray(), setter, getters, invalid} {
        reject(value, false, "nested property requires object")
    }
    if setter.calls != 0 || getters.calls != 0 || invalid.calls != 0 {
        t.Fatal("non-object protocol lookalike accessed")
    }
    // Trace has no value slot, so it must not gain a dynamic property.
    missing := New_trace().(*C_trace)
    reject(missing, false, "undefined property")
    if _, exists := missing.PHPPropertyLookup("value"); exists { t.Fatal("dynamic slot created") }
    box := New_box().(*C_box)
    reject(box, true, "array append requires array")
    if box.Get("items") != nil { t.Fatal("failed append published") }
    trace := New_trace().(*C_trace)
    input := rt.NewArray(rt.Entry{Value:rt.NewArray(rt.Entry{Value:"original"})})
    F_write_unknown(box, trace, input)
    input.Entries[0].Value.(rt.Array).Entries[0].Value = "mutated"
    if rt.Get(rt.Get(box.Get("value"),int64(0)),int64(0)) != "original" || trace.Get("mark") != int64(1) || trace.Get("after") != true {
        t.Fatal("successful assignment lost copies, RHS count or following write")
    }
    hostTrace := New_trace().(*C_trace)
    host := &propertyHost{trace:hostTrace}
    F_write_unknown(host, hostTrace, false)
    if host.value != false || host.gets != 1 || host.sets != 1 || hostTrace.Get("after") != true {
        t.Fatal("explicit host protocol, once-only check/publication or RHS order lost")
    }
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
		t.Fatalf("external property-write consumer: %v\n%s\n%s", err, output, generated)
	}
}
