package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/codegen"
	"github.com/katsuakikureyama/GoHP/internal/frontend"
)

func TestNamedClassParameterExternalABI(t *testing.T) {
	p, err := frontend.Parse(`<?php
class Box { public $value=1; public function take(Box $box){$box->value=2;return $box;} }
class Other {}
class Holder { public function __construct(Box $box){$box->value=4;} }
function accept(Box $box){return $box;}
function unused(Box $box){return 1;}
function overwrite(Box $box){$box=null;return $box;}
function two(Box $first, Box $second){$first->value=3;return $second;}
function contextual(Box $box){return strtok('a,b',',');}
`, phpTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	generated, err := codegen.Generate(p)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	file, observer, binary := filepath.Join(dir, "main.go"), filepath.Join(dir, "observer.go"), filepath.Join(dir, "app")
	if err := os.WriteFile(file, generated, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(observer, []byte(namedClassParameterObserver), 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("go", "build", "-o", binary, file, observer).CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	cmd := exec.Command(binary)
	cmd.Env = []string{"PATH="}
	if out, err := cmd.CombinedOutput(); err != nil || len(out) != 0 {
		t.Fatalf("external class parameter contract: %v\n%s", err, out)
	}
}

const namedClassParameterObserver = `package main
import (
    "fmt"
    "strings"
    rt "github.com/katsuakikureyama/GoHP/runtime"
)
type lookalike struct{}
func (*lookalike) PHPObject() bool { return true }
func (*lookalike) PHPProperties() rt.Array { return rt.NewArray() }
func (*lookalike) Get(string) any { return int64(1) }
func (*lookalike) Set(string,any) {}
func check(ok bool) { if !ok { panic("external assertion failed") } }
func reject(want string, call func()) {
    defer func(){p:=recover(); if p==nil || !strings.Contains(fmt.Sprint(p),want){panic(fmt.Sprintf("got %v want %s",p,want))}}()
    call()
}
func init() {
    good:=New_box().(*C_box)
    var absent *C_box
    for _, bad:=range []any{nil,absent,New_other(),&lookalike{},int64(1),"box",rt.NewArray()} {
        reject("class parameter requires box",func(){F_accept(bad)})
        reject("class parameter requires box",func(){F_unused(bad)})
        reject("class parameter requires box",func(){F_overwrite(bad)})
        reject("class parameter requires box",func(){good.M_take(bad)})
        reject("class parameter requires box",func(){New_holder(bad)})
        reject("class parameter requires box",func(){F_two(good,bad)})
        check(good.P_value==int64(1)) // failed later argument never enters body
    }
    args:=[]any{good}
    check(F_accept(args...)==good && args[0]==good)
    check(F_overwrite(args...)==nil && args[0]==good && good.P_value==int64(1))
    check(good.M_take(good)==good && good.P_value==int64(2))
    New_holder(good)
    check(good.P_value==int64(4))
    reject("argument count",func(){F_accept()})
    reject("argument count",func(){F_accept(good,good)})
    reject("execution context",func(){FContext_accept(nil,good)})
    context:=&rt.Execution{}
    check(FContext_accept(context,good)==good)
    check(FContext_contextual(context,good)=="a")
    check(context.Strtok(",")=="b")
}
`
