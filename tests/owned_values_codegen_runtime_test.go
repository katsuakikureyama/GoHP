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

// This is an external-module ABI check for generated ValueSlot functions. The
// fixture's PHP/Go stdout comparison is registered separately in the catalog.
func TestGeneratedOwnedValuesExternalConsumer(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("fixtures", "owned_values.php"))
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
		"owned_copy_isolation",
		"owned_self_assignment",
		"owned_keys_and_null",
		"owned_ternary_snapshot",
		"owned_returned_array",
		"owned_bad_return",
		"owned_overflow",
		"owned_ternary_missing",
		"owned_ternary_overflow",
		"owned_mixed_locals",
	} {
		body := generatedFunctionBody(t, generated, "FContext_"+function)
		for _, fragment := range []string{"map[string]*rt.ValueSlot", "MoveTo(v["} {
			if !strings.Contains(body, fragment) {
				t.Fatalf("%s did not select the generated owned-value lane; missing %q:\n%s", function, fragment, body)
			}
		}
	}
	for _, function := range []string{"owned_copy_isolation", "owned_self_assignment", "owned_ternary_snapshot", "owned_returned_array"} {
		if body := generatedFunctionBody(t, generated, "FContext_"+function); !strings.Contains(body, "Snapshot()") {
			t.Fatalf("%s did not capture an owned local snapshot:\n%s", function, body)
		}
	}
	for _, function := range []string{"legacy_identity", "legacy_state_start"} {
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
		"go.mod":  []byte(fmt.Sprintf("module example.invalid/generated-owned-consumer\n\ngo 1.23\n\nrequire github.com/katsuakikureyama/GoHP v0.0.0\nreplace github.com/katsuakikureyama/GoHP => %q\n", root)),
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
        if got := F_owned_copy_isolation(); got != "before:copy:auto:3:1" { t.Fatalf("copy result = %#v", got) }
    }
    returned := F_owned_returned_array()
    returned = rt.ArrayAppend(returned, "consumer")
    if got := rt.ArrayKeyLast(returned); got != int64(-3) { t.Fatalf("returned append key = %#v", got) }
    if got := rt.Get(returned, int64(-5)); got != "negative" { t.Fatalf("returned value = %#v", got) }
    if got := F_owned_returned_array(); rt.ArrayKeyLast(got) != int64(-4) { t.Fatalf("repeated result shared state: %#v", got) }

    reject("argument count", func() { F_owned_returned_array(int64(1)) })
    reject("execution context", func() { FContext_owned_returned_array(nil) })
    reject("integer overflow", func() { F_owned_overflow() })
    reject("array return requires array", func() { F_owned_bad_return() })
    if got := F_owned_self_assignment(); got != "one:two:2" { t.Fatalf("post-panic owned call = %#v", got) }

    nested := rt.NewArray(rt.Entry{Value:rt.NewArray(rt.Entry{Value:"nested"})})
    copied := F_legacy_identity(nested).(rt.Array)
    copied.Entries[0].Value = "changed"
    if got := rt.Get(rt.Get(nested, int64(0)), int64(0)); got != "nested" { t.Fatalf("legacy nested host value changed: %#v", got) }

    execution := &rt.Execution{}
    if got := FContext_legacy_state_start(execution); got != "first" { t.Fatalf("stateful start = %#v", got) }
    if got := rt.Get(FContext_owned_returned_array(execution),int64(-5)); got != "negative" { t.Fatalf("owned shared-context result = %#v",got) }
    if got := FContext_legacy_state_next(execution); got != "second" { t.Fatalf("stateful next = %#v", got) }
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
		t.Fatalf("external generated owned-value consumer: %v\n%s\n%s", err, output, generated)
	}
}

func generatedFunctionBody(t *testing.T, generated []byte, function string) string {
	t.Helper()
	start := strings.Index(string(generated), "func "+function+"(")
	if start < 0 {
		t.Fatalf("generated function %s is missing:\n%s", function, generated)
	}
	rest := string(generated[start:])
	if end := strings.Index(rest[1:], "\nfunc "); end >= 0 {
		return rest[:end+1]
	}
	return rest
}
