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

// This is an external-module ABI check for the bounded generated ValueSlot
// control-flow lane. The fixture's PHP/Go stdout comparison is registered
// separately in the compatibility catalog.
func TestOwnedControlFlowExternalConsumer(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("fixtures", "owned_control_flow.php"))
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
		"owned_nested_if_else",
		"owned_else_only_update",
		"owned_while_zero",
		"owned_while_steps",
		"owned_continue_break",
		"owned_early_return",
		"owned_copy_before_loop",
		"owned_copy_inside_loop",
		"owned_body_only_locals",
		"owned_returned_control_array",
		"owned_nested_loops",
		"owned_loop_overflow",
	} {
		body := generatedFunctionBody(t, generated, "FContext_"+function)
		for _, fragment := range []string{"map[string]*rt.ValueSlot", "MoveTo(v[", "defer v["} {
			if !strings.Contains(body, fragment) {
				t.Fatalf("%s did not select the generated owned control-flow lane; missing %q:\n%s", function, fragment, body)
			}
		}
	}
	for _, function := range []string{"legacy_nested_identity", "owned_control_state_start"} {
		body := generatedFunctionBody(t, generated, "FContext_"+function)
		if strings.Contains(body, "*rt.ValueSlot") {
			t.Fatalf("%s unexpectedly selected the owned-value lane:\n%s", function, body)
		}
	}

	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	files := map[string][]byte{
		"main.go": generated,
		"go.mod":  []byte(fmt.Sprintf("module example.invalid/generated-owned-control-consumer\n\ngo 1.23\n\nrequire github.com/katsuakikureyama/GoHP v0.0.0\nreplace github.com/katsuakikureyama/GoHP => %q\n", root)),
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
            if got == nil || !strings.Contains(fmt.Sprint(got), want) { t.Fatalf("panic = %v, want %q", got, want) }
        }()
        call()
    }
    for i := 0; i < 2; i++ {
        if got := F_owned_nested_if_else(); got != "base:inner-else:2" { t.Fatalf("nested if = %#v", got) }
        if got := F_owned_else_only_update(); got != "base:else:2" { t.Fatalf("else update = %#v", got) }
        if got := F_owned_while_zero(); got != "kept:1" { t.Fatalf("zero loop = %#v", got) }
        if got := F_owned_while_steps(); got != "0:1:2:3" { t.Fatalf("steps = %#v", got) }
        if got := F_owned_continue_break(); got != "1:3:2" { t.Fatalf("continue/break = %#v", got) }
        if got := F_owned_copy_before_loop(); got != "source:copy:0:copy:4:2" { t.Fatalf("copy before = %#v", got) }
        if got := F_owned_copy_inside_loop(); got != "step0:0:2:step1:1:3" { t.Fatalf("copy inside = %#v", got) }
        if got := F_owned_body_only_locals(); got != "zero:one:2" { t.Fatalf("body-only locals = %#v", got) }
        if got := F_owned_nested_loops(); got != "0:0:0:1:1:0:1:1:4" { t.Fatalf("nested loops = %#v", got) }
    }
    early := F_owned_early_return().(rt.Array)
    if got := rt.Get(early, "early"); got != "yes" || rt.Count(early) != int64(1) { t.Fatalf("early return = %#v", early) }

    returned := F_owned_returned_control_array().(rt.Array)
    returned = rt.ArrayAppend(returned, "consumer").(rt.Array)
    if got := rt.ArrayKeyLast(returned); got != int64(-3) { t.Fatalf("returned append key = %#v", got) }
    if got := rt.Get(returned, int64(-5)); got != "negative" { t.Fatalf("returned value = %#v", got) }
    if got := F_owned_returned_control_array().(rt.Array); rt.ArrayKeyLast(got) != "tail" || rt.Count(got) != int64(4) { t.Fatalf("repeated return shared state: %#v", got) }

    reject("argument count", func() { F_owned_returned_control_array(int64(1)) })
    reject("execution context", func() { FContext_owned_returned_control_array(nil) })
    reject("integer overflow", func() { F_owned_loop_overflow() })
    if got := F_owned_while_steps(); got != "0:1:2:3" { t.Fatalf("post-overflow call = %#v", got) }

    execution := &rt.Execution{}
    if got := FContext_owned_control_state_start(execution); got != "first" { t.Fatalf("stateful start = %#v", got) }
    if got := FContext_owned_returned_control_array(execution).(rt.Array); rt.Get(got, int64(-5)) != "negative" { t.Fatalf("owned shared-context result = %#v", got) }
    if got := FContext_owned_control_state_next(execution); got != "second" { t.Fatalf("stateful next = %#v", got) }

    nested := rt.NewArray(rt.Entry{Value:rt.NewArray(rt.Entry{Value:"nested"})})
    copied := F_legacy_nested_identity(nested).(rt.Array)
    copied.Entries[0].Value = "changed"
    if got := rt.Get(rt.Get(nested, int64(0)), int64(0)); got != "nested" { t.Fatalf("legacy nested host value changed: %#v", got) }
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
		t.Fatalf("external generated owned-control consumer: %v\n%s\n%s", err, output, generated)
	}
}
