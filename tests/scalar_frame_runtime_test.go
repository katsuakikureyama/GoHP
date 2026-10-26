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

// A real external module catches illegal generated imports of internal/storage
// and checks the callable ABI when owned scalar and legacy value frames coexist.
// PHP equivalence is measured separately by TestDifferential/scalar_frame.
func TestScalarFrameExternalConsumer(t *testing.T) {
	const source = `<?php
function scalar() { $x = 1; $copy = $x; $x = 2; return $copy; }
function fails() { $x = 9223372036854775807; return $x + 1; }
function mixed($value) { return $value; }
function token_start() { return strtok('a,b', ','); }
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
		"go.mod":  []byte(fmt.Sprintf("module example.invalid/frame-consumer\n\ngo 1.23\n\nrequire github.com/katsuakikureyama/GoHP v0.0.0\nreplace github.com/katsuakikureyama/GoHP => %q\n", root)),
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
        if got := F_scalar(); got != int64(1) { t.Fatalf("scalar result = %#v", got) }
    }
    reject("argument count", func() { F_scalar(int64(1)) })
    reject("execution", func() { FContext_scalar(nil) })
    reject("integer overflow", func() { F_fails() })
    if got := F_scalar(); got != int64(1) { t.Fatalf("post-panic result = %#v", got) }
    array := rt.NewArray(rt.Entry{Value:int64(7)})
    copied := F_mixed(array).(rt.Array)
    copied.Entries[0].Value = int64(9)
    if got := rt.Get(array,int64(0)); got != int64(7) { t.Fatalf("legacy copy changed: %#v",got) }
    execution := &rt.Execution{}
    if got := FContext_token_start(execution); got != "a" { t.Fatalf("token start = %#v",got) }
    FContext_scalar(execution)
    if got := execution.Strtok(","); got != "b" { t.Fatalf("scalar frame reset invocation state: %#v",got) }
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
		t.Fatalf("external consumer: %v\n%s\n%s", err, output, generated)
	}
}
