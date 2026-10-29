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

func TestUnknownObjectArrayFixture(t *testing.T) {
	for _, target := range []string{"8.3", "8.4"} {
		t.Run(target, func(t *testing.T) {
			input, err := loadDifferentialInput("tests/fixtures/unknown_object_array.php", target)
			if err != nil {
				t.Fatal(err)
			}
			if _, _, err := codegen.GenerateWithMap(input.program); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// The generated untyped entry must retain the same actual object protocol as
// ObjectToArray, including host values that PHP source itself cannot construct.
func TestUnknownObjectArrayHostEntry(t *testing.T) {
	const source = `<?php
class Box { public $value; }
class Trace { public $calls = 0; public $after = false; }
function operand($value, Trace $trace) {
    $trace->calls = $trace->calls + 1;
    return $value;
}
function snapshot($value, Trace $trace) {
    $result = (array) operand($value, $trace);
    $trace->after = true;
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
		"go.mod":  []byte(fmt.Sprintf("module example.invalid/object-array-consumer\n\ngo 1.23\n\nrequire github.com/katsuakikureyama/GoHP v0.0.0\nreplace github.com/katsuakikureyama/GoHP => %q\n", root)),
		"consumer_test.go": []byte(`package main

import (
    "fmt"
    "testing"
    rt "github.com/katsuakikureyama/GoHP/runtime"
)

type markerOnly struct{}
func (*markerOnly) PHPObject() bool { return true }
type propertiesOnly struct { calls int }
func (h *propertiesOnly) PHPProperties() rt.Array { h.calls++; return rt.NewArray() }
type invalidObject struct { propertiesOnly }
func (*invalidObject) PHPObject() bool { return false }
// An explicit complete structural protocol is opt-in, not proof of provenance.
type propertyHost struct { trace *C_trace; calls int }
func (h *propertyHost) PHPObject() bool { return h != nil }
func (h *propertyHost) PHPProperties() rt.Array {
    if h.trace.Get("calls") != int64(1) || h.trace.Get("after") != false {
        panic("snapshot order changed")
    }
    h.calls++
    return rt.NewArray(rt.Entry{Key:"host", Value:false})
}

func TestConsumer(t *testing.T) {
    reject := func(value any) {
        t.Helper()
        trace := New_trace().(*C_trace)
        defer func() {
            got := recover()
            if fmt.Sprint(got) != "array cast requires generated public-property object" {
                t.Fatalf("panic = %v, want actual-object rejection", got)
            }
            if trace.Get("calls") != int64(1) || trace.Get("after") != false {
                t.Fatalf("operand not once, or following statement ran: %#v", trace)
            }
        }()
        F_snapshot(value, trace)
    }
    var typedNil *C_box
    properties := &propertiesOnly{}
    invalid := &invalidObject{}
    for _, value := range []any{nil, typedNil, false, int64(0), "", "0", rt.NewArray(), rt.NewCallable(0, func(...any) any { return nil }), &markerOnly{}, properties, invalid} {
        reject(value)
    }
    if properties.calls != 0 || invalid.calls != 0 { t.Fatal("invalid host snapshot was called") }

    box := New_box().(*C_box)
    box.Set("value", "before")
    trace := New_trace().(*C_trace)
    result, ok := F_snapshot(box, trace).(rt.Array)
    if !ok || rt.Get(result,"value") != "before" || trace.Get("calls") != int64(1) || trace.Get("after") != true {
        t.Fatalf("generated-object snapshot or sequencing lost: %#v", result)
    }
    box.Set("value", "after")
    if rt.Get(result,"value") != "before" { t.Fatal("snapshot followed later property replacement") }

    hostTrace := New_trace().(*C_trace)
    host := &propertyHost{trace:hostTrace}
    hostResult := F_snapshot(host, hostTrace)
    if host.calls != 1 || rt.Get(hostResult,"host") != false || hostTrace.Get("calls") != int64(1) || hostTrace.Get("after") != true {
        t.Fatal("explicit host protocol, once-only snapshot or statement order lost")
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
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("generated consumer: %v\n%s", err, output)
	}
}
