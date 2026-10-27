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

// This external-module ABI check covers the generated ValueSlot path for
// exact scalar parameters. Differential PHP output is registered separately.
func TestOwnedParameterExternalConsumer(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("fixtures", "owned_parameters.php"))
	if err != nil {
		t.Fatal(err)
	}
	p, err := frontend.Parse(string(source), phpTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	generated, err := codegen.Generate(p)
	if err != nil {
		t.Fatal(err)
	}
	for _, function := range []string{
		"owned_parameter_trace",
		"owned_parameter_reuse",
		"owned_parameter_return",
		"owned_parameter_read_only",
		"owned_parameter_guard_first",
	} {
		body := generatedFunctionBody(t, generated, "FContext_"+function)
		for _, fragment := range []string{
			"map[string]*rt.ValueSlot",
			`ScalarParameter(args[0], "int")`,
			`ScalarParameter(args[1], "string")`,
			`ScalarParameter(args[2], "bool")`,
			"MoveTo(v[",
		} {
			if !strings.Contains(body, fragment) {
				t.Fatalf("%s did not select the generated owned-parameter lane; missing %q:\n%s", function, fragment, body)
			}
		}
	}
	for _, function := range []string{"legacy_typed_array_identity", "legacy_untyped_identity", "owned_parameter_state_start"} {
		body := generatedFunctionBody(t, generated, "FContext_"+function)
		if strings.Contains(body, "*rt.ValueSlot") {
			t.Fatalf("%s unexpectedly selected the owned-parameter lane:\n%s", function, body)
		}
	}

	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	files := map[string][]byte{
		"main.go": generated,
		"go.mod":  []byte(fmt.Sprintf("module example.invalid/generated-owned-parameter-consumer\n\ngo 1.23\n\nrequire github.com/katsuakikureyama/GoHP v0.0.0\nreplace github.com/katsuakikureyama/GoHP => %q\n", root)),
		"consumer_test.go": []byte(`package main

import (
    "fmt"
    "reflect"
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

    args := []any{int64(2), "seed", true}
    before := append([]any(nil), args...)
    if got := F_owned_parameter_trace(args...); got != "seed:yes:4:2" {
        t.Fatalf("trace = %#v", got)
    }
    if !reflect.DeepEqual(args, before) {
        t.Fatalf("caller args changed: %#v, want %#v", args, before)
    }
    if got := F_owned_parameter_trace(int64(0), "zero", false); got != "zero:no:2:2" {
        t.Fatalf("zero trace = %#v", got)
    }
    reuseArgs := []any{int64(2), "flat", true}
    if got := F_owned_parameter_reuse(reuseArgs...); got != "flat:3:0" {
        t.Fatalf("append reuse = %#v", got)
    }
    if !reflect.DeepEqual(reuseArgs, []any{int64(2), "flat", true}) {
        t.Fatalf("local parameter replacement changed caller args: %#v", reuseArgs)
    }
    if got := F_owned_parameter_reuse(int64(1), "replace", false); got != "replace:2:0" {
        t.Fatalf("replace reuse = %#v", got)
    }
    if got := F_owned_parameter_read_only(int64(99), "word", true); got != "word:yes:2" {
        t.Fatalf("read-only = %#v", got)
    }

    returned := F_owned_parameter_return(int64(2), "owned", true).(rt.Array)
    returned = rt.ArrayAppend(returned, "consumer").(rt.Array)
    if got := rt.ArrayKeyLast(returned); got != int64(2) {
        t.Fatalf("returned append key = %#v", got)
    }
    if got := rt.Get(returned, "label"); got != "owned" {
        t.Fatalf("returned label = %#v", got)
    }
    if got := F_owned_parameter_return(int64(2), "owned", true).(rt.Array); rt.ArrayKeyLast(got) != "flag" || rt.Count(got) != int64(4) {
        t.Fatalf("repeated return shared state: %#v", got)
    }

    reject("int parameter requires exact type", func() { F_owned_parameter_trace(int(2), "seed", true) })
    reject("string parameter requires exact type", func() { F_owned_parameter_trace(int64(2), nil, true) })
    reject("bool parameter requires exact type", func() { F_owned_parameter_trace(int64(2), "seed", int64(1)) })
    reject("argument count", func() { F_owned_parameter_trace(int64(2), "seed") })
    reject("argument count", func() { F_owned_parameter_trace(int64(2), "seed", true, "extra") })
    // A bad third parameter must reject before the body can reach its overflow.
    reject("bool parameter requires exact type", func() { F_owned_parameter_guard_first(int64(1), "x", int64(1)) })
    reject("integer overflow", func() { F_owned_parameter_guard_first(int64(1), "x", true) })
    if got := F_owned_parameter_trace(int64(1), "after", true); got != "after:yes:3:2" {
        t.Fatalf("post-panic trace = %#v", got)
    }

    reject("execution context", func() { FContext_owned_parameter_trace(nil, int64(1), "x", true) })
    execution := &rt.Execution{}
    if got := FContext_owned_parameter_state_start(execution); got != "first" {
        t.Fatalf("stateful start = %#v", got)
    }
    if got := FContext_owned_parameter_trace(execution, int64(1), "shared", true); got != "shared:yes:3:2" {
        t.Fatalf("shared owned result = %#v", got)
    }
    if got := FContext_owned_parameter_state_next(execution); got != "second" {
        t.Fatalf("stateful next = %#v", got)
    }

    nested := rt.NewArray(rt.Entry{Value: rt.NewArray(rt.Entry{Value: "nested"})})
    typed := F_legacy_typed_array_identity(nested).(rt.Array)
    typed.Entries[0].Value = "changed"
    if got := rt.Get(rt.Get(nested, int64(0)), int64(0)); got != "nested" {
        t.Fatalf("typed legacy nested host value changed: %#v", got)
    }
    untyped := F_legacy_untyped_identity(nested).(rt.Array)
    untyped.Entries[0].Value = "changed again"
    if got := rt.Get(rt.Get(nested, int64(0)), int64(0)); got != "nested" {
        t.Fatalf("untyped legacy nested host value changed: %#v", got)
    }
    // Legacy ArrayParameter/Clone remains admissible for a host-edited array
    // whose Entries no longer reconstruct its hidden automatic-key history.
    stale := rt.NewArray(rt.Entry{Key: int64(9), Value: "removed"})
    stale.Entries = nil
    legacy := F_legacy_typed_array_identity(stale).(rt.Array)
    if got := rt.ArrayKeyLast(rt.ArrayAppend(legacy, "legacy")); got != int64(10) {
        t.Fatalf("typed legacy stale history = %#v", got)
    }
    if got := rt.ArrayKeyLast(rt.ArrayAppend(stale, "input")); got != int64(10) {
        t.Fatalf("stale input history changed: %#v", got)
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
		t.Fatalf("external generated owned-parameter consumer: %v\n%s\n%s", err, output, generated)
	}
}
