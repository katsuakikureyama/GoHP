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

// This is an external generated-package ABI check. It distinguishes an exact
// generated class pointer from a generic runtime.Object marker and confirms
// that every entry guard runs before the PHP body can alter shared state.
func TestDeclaredClassParameterHostEntry(t *testing.T) {
	const source = `<?php
function ignored(iTeM $item) { return 'unused'; }
function overwritten(Item $item) { $item = new ITEM('replacement'); return $item; }
function mutate(item $item) { $item->value = $item->value . ':mutated'; return $item; }
function guarded(ITEM $item) { return strtok('body-marker,next', ','); }

class Bag {
    public $item;
    public $count;
    public function __construct(iTeM $item, int $count, $tail = null) {
        $this->item = $item;
        $this->count = $count;
        $item->value = $item->value . ':' . $count;
    }
    public function update(ITEM $item) {
        $item->value = $item->value . ':method';
        return $item;
    }
}

class Item {
    public $value;
    public function __construct($value) { $this->value = $value; }
}

class Other {
    public $value;
    public function __construct($value) { $this->value = $value; }
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
		"go.mod":  []byte(fmt.Sprintf("module example.invalid/declared-class-parameter-consumer\n\ngo 1.23\n\nrequire github.com/katsuakikureyama/GoHP v0.0.0\nreplace github.com/katsuakikureyama/GoHP => %q\n", root)),
		"consumer_test.go": []byte(`package main

import (
    "fmt"
    "strings"
    "testing"

    rt "github.com/katsuakikureyama/GoHP/runtime"
)

type fakePHPObject struct{}
func (fakePHPObject) PHPObject() bool { return true }

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

    item := New_item("seed").(*C_item)
    check(F_ignored(item), "unused")
    check(F_mutate(item), item)
    check(item.Get("value"), "seed:mutated")

    replacement := F_overwritten(item).(*C_item)
    if replacement == item { t.Fatal("local reassignment returned caller object") }
    check(replacement.Get("value"), "replacement")
    check(item.Get("value"), "seed:mutated")

    var typedNil *C_item
    invalid := []any{nil, typedNil, New_other("other"), fakePHPObject{}, struct{}{}, int64(1), rt.NewArray()}
    for _, value := range invalid {
        reject("class parameter requires item object", func() { F_ignored(value) })
    }
    reject("argument count", func() { F_ignored() })
    reject("argument count", func() { F_ignored(item, "extra") })
    check(F_ignored(item), "unused")

    // A rejection must occur before guarded() starts its tokenizer body.
    execution := &rt.Execution{}
    reject("class parameter requires item object", func() { FContext_guarded(execution, New_other("wrong")) })
    reject("initialized state", func() { execution.Strtok(",") })
    check(FContext_guarded(execution, item), "body-marker")
    check(execution.Strtok(","), "next")

    reject("execution context", func() { FContext_guarded(nil, item) })
    bag := New_bag(item, int64(7)).(*C_bag)
    check(bag.Get("item"), item)
    check(bag.Get("count"), int64(7))
    check(item.Get("value"), "seed:mutated:7")
    reject("class parameter requires item object", func() { New_bag(New_other("wrong"), int64(2)) })
    reject("int parameter requires exact type", func() { New_bag(item, int(2)) })
    reject("argument count", func() { New_bag(item) })
    reject("argument count", func() { New_bag(item, int64(2), nil, "extra") })
    check(item.Get("value"), "seed:mutated:7")

    reject("class parameter requires item object", func() { bag.M_update(New_other("wrong")) })
    check(item.Get("value"), "seed:mutated:7")
    check(bag.M_update(item), item)
    check(item.Get("value"), "seed:mutated:7:method")
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
		t.Fatalf("external declared-class-parameter consumer: %v\n%s\n%s", err, output, generated)
	}
}
