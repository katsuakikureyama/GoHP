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

// This external-module ABI check makes the new lookup protocol observable. A
// receiver need not expose Get or Set: isset must use PHPPropertyLookup only,
// visit each reached component once, and stop when an earlier component is
// absent or null.
func TestPropertyIssetHostEntry(t *testing.T) {
	const source = `<?php
function probe($value) { return isset($value->child->value); }
function multi($value) { return isset($value->first, $value->second->value, $value->third); }
function and_skip($value) { return isset($value->missing) && isset($value->boom->value); }
function or_skip($value) { return isset($value->present) || isset($value->boom->value); }
function choose($value) { return isset($value->child->value) ? 'yes' : 'no'; }
class Box {
    public $child = null;
    public function probe_this() { return isset($this->child->value); }
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
		"go.mod":  []byte(fmt.Sprintf("module example.invalid/property-isset-consumer\n\ngo 1.23\n\nrequire github.com/katsuakikureyama/GoHP v0.0.0\nreplace github.com/katsuakikureyama/GoHP => %q\n", root)),
		"consumer_test.go": []byte(`package main

import (
    "fmt"
    "strings"
    "testing"

    rt "github.com/katsuakikureyama/GoHP/runtime"
)

type lookupSpy struct {
    valid bool
    values map[string]any
    calls  []string
    forbid string
}

func (s *lookupSpy) PHPObject() bool { return s != nil && s.valid }
func (s *lookupSpy) PHPPropertyLookup(name string) (any, bool) {
    if s.forbid == name { panic("unexpected lookup: " + name) }
    s.calls = append(s.calls, name)
    value, ok := s.values[name]
    return value, ok
}

type markerOnly struct{}
func (markerOnly) PHPObject() bool { return true }

type getterOnly struct{}
func (getterOnly) PHPObject() bool { return true }
func (getterOnly) Get(string) any { return "must not be used" }

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
        if got != want { t.Fatalf("got %#v, want %#v", got, want) }
    }
    calls := func(s *lookupSpy, want ...string) {
        t.Helper()
        if got := strings.Join(s.calls, ","); got != strings.Join(want, ",") {
            t.Fatalf("lookups = %q, want %q", got, strings.Join(want, ","))
        }
    }

    leaf := &lookupSpy{valid:true, values:map[string]any{"value":false}}
    root := &lookupSpy{valid:true, values:map[string]any{"child":leaf}}
    check(F_probe(root), true)
    calls(root, "child")
    calls(leaf, "value")
    check(F_choose(root), "yes")

    // A present property whose value is nil stops before the next path part.
    nilChild := &lookupSpy{valid:true, values:map[string]any{"child":nil}}
    check(F_probe(nilChild), false)
    calls(nilChild, "child")

    // A missing property stops without touching the forbidden later receiver.
    absent := &lookupSpy{valid:true, values:map[string]any{"boom":&lookupSpy{valid:true, forbid:"value"}}}
    check(F_and_skip(absent), false)
    calls(absent, "missing")
    present := &lookupSpy{valid:true, values:map[string]any{"present":false, "boom":&lookupSpy{valid:true, forbid:"value"}}}
    check(F_or_skip(present), true)
    calls(present, "present")

    second := &lookupSpy{valid:true, values:map[string]any{"value":""}}
    many := &lookupSpy{valid:true, values:map[string]any{"first":false, "second":second, "third":rt.NewArray()}}
    check(F_multi(many), true)
    calls(many, "first", "second", "third")
    calls(second, "value")

    multiStop := &lookupSpy{valid:true, values:map[string]any{"first":nil, "second":&lookupSpy{valid:true, forbid:"value"}}}
    check(F_multi(multiStop), false)
    calls(multiStop, "first")

    // Scalars, arrays and invalid generated pointers are quiet false; marker
    // or getter lookalikes are rejected rather than silently trusted.
    var typedNil *lookupSpy
    for _, value := range []any{nil, typedNil, int64(1), "text", false, rt.NewArray()} {
        check(F_probe(value), false)
    }
    reject("property isset requires generated property lookup", func() { F_probe(markerOnly{}) })
    reject("property isset requires generated property lookup", func() { F_probe(getterOnly{}) })

    box := New_box().(*C_box)
    box.Set("child", leaf)
    check(box.M_probe_this(), true)
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
		t.Fatalf("external property isset consumer: %v\n%s\n%s", err, output, generated)
	}
}
