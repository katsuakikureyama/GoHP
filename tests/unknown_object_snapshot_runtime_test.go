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

// TestUnknownObjectSnapshotFixture keeps the committed PHP oracle within the
// intended subset even when a local machine does not have native PHP. Its
// differential stdout comparison is registered separately in the catalog.
func TestUnknownObjectSnapshotFixture(t *testing.T) {
	fixture, err := os.ReadFile("fixtures/unknown_object_snapshot.php")
	if err != nil {
		t.Fatal(err)
	}
	p, err := frontend.Parse(string(fixture), phpTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := codegen.Generate(p); err != nil {
		t.Fatal(err)
	}
}

// TestUnknownObjectSnapshotHostEntry exercises the external-module ABI for an
// untyped cast. Hosts may explicitly implement PublicPropertyObject, but the
// marker contract (including typed nil) and a fresh NewArray snapshot remain
// their responsibility; this is not generated-object provenance detection.
func TestUnknownObjectSnapshotHostEntry(t *testing.T) {
	const source = `<?php
class Trace { public $calls = 0; }
function once($value, Trace $trace) {
    $trace->calls = $trace->calls + 1;
    return $value;
}
function snapshot_unknown($value) { return (array) $value; }
function snapshot_once($value, Trace $trace) { return (array) once($value, $trace); }
function snapshot_property($value) { return (array) $value->block; }
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
		"go.mod":  []byte(fmt.Sprintf("module example.invalid/unknown-object-snapshot-consumer\n\ngo 1.23\n\nrequire github.com/katsuakikureyama/GoHP v0.0.0\nreplace github.com/katsuakikureyama/GoHP => %q\n", root)),
		"consumer_test.go": []byte(`package main

import (
    "fmt"
    "strings"
    "testing"

    rt "github.com/katsuakikureyama/GoHP/runtime"
)

// snapshotHost is a deliberate full opt-in host. Its PHPProperties method
// creates a fresh declaration-ordered array on every call, rather than
// returning a cached array or claiming generated provenance.
type snapshotHost struct {
    valid bool
    name string
    nested rt.Array
    shared any
    calls int
}
func (h *snapshotHost) PHPObject() bool { return h != nil && h.valid }
func (h *snapshotHost) PHPProperties() rt.Array {
    h.calls++
    return rt.NewArray(
        rt.Entry{Key:"name", Value:h.name},
        rt.Entry{Key:"nested", Value:h.nested},
        rt.Entry{Key:"shared", Value:h.shared},
        rt.Entry{Key:"nullable", Value:nil},
    )
}

type markerOnly struct{}
func (markerOnly) PHPObject() bool { return true }

type snapshotOnly struct{}
func (snapshotOnly) PHPProperties() rt.Array { panic("unexpected snapshot-only call") }

type falseMarker struct{}
func (falseMarker) PHPObject() bool { return false }
func (falseMarker) PHPProperties() rt.Array { panic("unexpected false-marker snapshot") }

// propertyHost makes the generated property's single read visible. It is
// intentionally separate from the snapshotted host object.
type propertyHost struct { block any; gets int }
func (h *propertyHost) Get(name string) any {
    if name != "block" { panic("unexpected property: " + name) }
    h.gets++
    return h.block
}

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
    array := func(value any) rt.Array {
        t.Helper()
        got, ok := value.(rt.Array)
        if !ok { t.Fatalf("snapshot = %T, want runtime.Array", value) }
        return got
    }
    entry := func(value rt.Array, key string) any {
        t.Helper()
        for _, item := range value.Entries {
            if item.Key == key { return item.Value }
        }
        t.Fatalf("missing snapshot property %q", key)
        return nil
    }

    shared := &struct{ marker string }{marker:"shared"}
    host := &snapshotHost{
        valid:true,
        name:"original",
        nested:rt.NewArray(rt.Entry{Key:"outer", Value:rt.NewArray(rt.Entry{Key:"value", Value:"original"})}),
        shared:shared,
    }
    first := array(F_snapshot_unknown(host))
    if host.calls != 1 || len(first.Entries) != 4 || first.Entries[0].Key != "name" || first.Entries[1].Key != "nested" || first.Entries[2].Key != "shared" || first.Entries[3].Key != "nullable" {
        t.Fatalf("snapshot was not one fresh declaration-ordered NewArray: %#v", first)
    }
    if entry(first, "shared") != shared { t.Fatal("snapshot lost shared object identity") }
    first.Entries[0].Value = "copy-mutated"
    nested := entry(first, "nested").(rt.Array)
    nested.Entries[0].Value = rt.NewArray(rt.Entry{Key:"value", Value:"copy-mutated"})
    second := array(F_snapshot_unknown(host))
    if host.calls != 2 || entry(second, "name") != "original" || rt.Get(rt.Get(entry(second, "nested"), "outer"), "value") != "original" {
        t.Fatalf("snapshot reuse leaked a prior caller mutation: %#v", second)
    }

    trace := New_trace().(*C_trace)
    array(F_snapshot_once(host, trace))
    if host.calls != 3 || trace.Get("calls") != int64(1) {
        t.Fatalf("effectful receiver = calls:%d trace:%#v, want one each", host.calls, trace)
    }
    property := &propertyHost{block:host}
    array(F_snapshot_property(property))
    if property.gets != 1 || host.calls != 4 {
        t.Fatalf("property getter/snapshot calls = %d/%d, want 1/4", property.gets, host.calls)
    }

    generated := New_trace().(*C_trace)
    generatedSnapshot := array(F_snapshot_unknown(generated))
    if len(generatedSnapshot.Entries) != 1 || generatedSnapshot.Entries[0].Key != "calls" || generatedSnapshot.Entries[0].Value != int64(0) {
        t.Fatalf("generated public-property object did not snapshot: %#v", generatedSnapshot)
    }

    var typedNil *snapshotHost
    var typedNilGenerated *C_trace
    invalid := []any{nil, typedNil, typedNilGenerated, markerOnly{}, snapshotOnly{}, falseMarker{}, false, int64(1), "text", rt.NewArray(), rt.NewCallable(0, func(...any) any { return nil })}
    for _, value := range invalid {
        reject("array cast requires generated public-property object", func() { F_snapshot_unknown(value) })
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
		t.Fatalf("external unknown-object snapshot consumer: %v\n%s\n%s", err, output, generated)
	}
}
