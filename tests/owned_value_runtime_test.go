package tests

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// This is a real external-module runtime ABI check, not generated PHP evidence.
// A consumer must never need to import internal/storage or release its handles.
func TestOwnedValueExternalConsumer(t *testing.T) {
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	files := map[string]string{
		"go.mod": fmt.Sprintf("module example.invalid/value-consumer\n\ngo 1.23\n\nrequire github.com/katsuakikureyama/GoHP v0.0.0\nreplace github.com/katsuakikureyama/GoHP => %q\n", root),
		"consumer_test.go": `package consumer
import (
    "testing"
    rt "github.com/katsuakikureyama/GoHP/runtime"
)
func TestConsumer(t *testing.T) {
    reject := func(call func()) {
        t.Helper()
        defer func() { if recover() == nil { t.Fatal("expected rejected operation") } }()
        call()
    }
    input := rt.NewArray(rt.Entry{Key:"-5", Value:"before"}, rt.Entry{Key:"label",Value:nil})
    slot := rt.NewValueSlot(input)
    defer slot.Close()
    input.Entries[0].Value = "host changed"
    snapshot := slot.Snapshot()
    defer snapshot.Close()
    if key := slot.ArrayAppend("appended"); key != -4 { t.Fatalf("append key = %d",key) }
    slot.ArrayWrite("-5","after")
    copy := &rt.ValueSlot{}
    defer copy.Close()
    snapshot.MoveTo(copy)
    reject(func() { snapshot.Take() })
    if got := rt.Get(copy.Read(),int64(-5)); got != "before" { t.Fatalf("snapshot = %#v",got) }
    exported := slot.Read().(rt.Array)
    exported.Entries[0].Value = "export changed"
    if got := rt.Get(slot.Read(),int64(-5)); got != "after" { t.Fatalf("export aliased slot: %#v",got) }

    // A taken legacy value survives the producer's deferred cleanup.
    returned := func() any {
        temporary := copy.Snapshot()
        defer temporary.Close()
        copy.Close()
        return temporary.Take()
    }()
    if got := rt.ArrayKeyLast(rt.ArrayAppend(returned,"next")); got != int64(-4) { t.Fatalf("returned index = %#v",got) }
    pending := rt.NewValueTemporary("retry")
    defer pending.Close()
    reject(func() { pending.MoveTo(copy) })
    pending.MoveTo(slot)
    if got := slot.Read(); got != "retry" { t.Fatalf("retry result = %#v",got) }

    // Existing nested-array operations remain usable, outside this new lane.
    nested := rt.NewArray(rt.Entry{Value:rt.NewArray(rt.Entry{Value:"nested"})})
    reject(func() { slot.Write(nested) })
    if got := rt.Get(rt.Get(rt.Clone(nested),int64(0)),int64(0)); got != "nested" { t.Fatalf("legacy nested result = %#v",got) }
    // Removing Entries without updating hidden index is not silently repaired.
    stale := rt.NewArray(rt.Entry{Key:int64(9),Value:"removed"})
    stale.Entries = nil
    reject(func() { slot.Write(stale) })
    if got := rt.ArrayKeyLast(rt.ArrayAppend(stale,"legacy")); got != int64(10) { t.Fatalf("legacy history changed: %#v",got) }
    if got := slot.Read(); got != "retry" { t.Fatalf("failed import changed destination: %#v",got) }
    slot.Close()
    reject(func() { slot.Snapshot() })
}
`,
	}
	for name, contents := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0600); err != nil {
			t.Fatal(err)
		}
	}
	cmd := exec.Command("go", "test", "-count=1", "./...")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off", "GOPROXY=off", "GOSUMDB=off", "GOFLAGS=-mod=mod")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("external value consumer: %v\n%s", err, output)
	}
}
