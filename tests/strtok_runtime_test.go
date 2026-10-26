package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/codegen"
	"github.com/katsuakikureyama/GoHP/internal/frontend"
)

func TestStrtokRuntimeGuards(t *testing.T) {
	checkArrayRuntimeGuards(t, []arrayRuntimeCase{
		{"uninitialized", "<?php echo strtok(',');", "strtok continuation requires initialized state"},
		{"cleared_scan", "<?php $a = strtok(',,,', ','); echo strtok(',');", "strtok continuation requires initialized state"},
		{"dynamic_input", "<?php function f($x) { return strtok($x, ','); } echo f(false);", "strtok requires actual strings"},
		{"dynamic_separator", "<?php function f($x) { return strtok('x', $x); } echo f(null);", "strtok requires actual strings"},
		{"false_result", "<?php echo strlen(strtok('', ','));", "strlen requires string"},
		{"array_return", "<?php function f(): array { return strtok('x', ','); } $a = f();", "array return requires array"},
		{"bool_return", "<?php function f(): bool { return strtok('x', ','); } $a = f();", "bool return requires bool"},
	})
}

// Real generated entry points must isolate requests while explicit same-request
// calls (including methods and escaped arrows) share the caller's scan.
func TestStrtokGeneratedExecutionIsolation(t *testing.T) {
	p, err := frontend.Parse(`<?php
function start($text) { return strtok($text, ','); }
function next_token() { return strtok(','); }
function make_arrow() { return fn() => strtok(','); }
function invoke_arrow(callable $f) { return $f(); }
function pair($text) { $first = start($text); return $first . ':' . next_token(); }
class Scanner { public function next_token() { return strtok(','); } }
class Initializer { public $head; public function __construct($text) { $this->head = strtok($text, ','); } }
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
	const host = `package main
import ("fmt"; "strings"; rt "github.com/katsuakikureyama/GoHP/runtime")
func check(got, want any) { if got != want { panic(fmt.Sprintf("got %v want %v", got, want)) } }
func rejected(fn func(), message string) {
 defer func(){ got:=recover(); if got==nil || !strings.Contains(fmt.Sprint(got),message) {panic(fmt.Sprintf("wrong rejection: %v",got))} }()
 fn()
}
func init() {
 check(F_start("old,leak"),"old")
 rejected(func(){ F_next_token() },"initialized state")
 a,b:=&rt.Execution{},&rt.Execution{}
 check(FContext_start(a,"a,second-a,third-a"),"a")
 check(FContext_start(b,"b,second-b,third-b"),"b")
 arrow:=FContext_make_arrow(a)
 check(FContext_invoke_arrow(b,arrow),"second-b")
 check(FContext_next_token(a),"second-a")
 object:=New_scanner().(*C_scanner)
 check(object.MContext_next_token(a),"third-a")
 check(object.MContext_next_token(b),"third-b")
 rejected(func(){ object.M_next_token() },"initialized state")
 rejected(func(){ FContext_next_token(nil) },"execution context")
 rejected(func(){ F_invoke_arrow(arrow) },"initialized state")
 initialized:=New_initializer("standalone,hidden").(*C_initializer)
 check(initialized.Get("head"),"standalone")
 rejected(func(){ F_next_token() },"initialized state")
 initialized=NewContext_initializer(a,"ctor,shared").(*C_initializer)
 check(initialized.Get("head"),"ctor")
 check(FContext_next_token(a),"shared")
 results:=make(chan any,8)
 for i:=0;i<8;i++ { go func(){ results<-F_pair("first,last") }() }
 for i:=0;i<8;i++ { check(<-results,"first:last") }
 fmt.Print("isolated-and-shared-correctly")
}
`
	if err := os.WriteFile(observer, []byte(host), 0600); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command("go", "build", "-o", binary, file, observer).CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, output)
	}
	run := exec.Command(binary)
	run.Env = []string{"PATH="}
	output, err := run.CombinedOutput()
	if err != nil || string(output) != "isolated-and-shared-correctly" {
		t.Fatalf("execution: %v\n%s", err, output)
	}
}
