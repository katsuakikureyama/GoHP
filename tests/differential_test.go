package tests

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/codegen"
	"github.com/katsuakikureyama/GoHP/internal/frontend"
)

func TestDifferential(t *testing.T) {
	php, e := exec.LookPath("php")
	if e != nil {
		t.Fatal("native PHP is required for differential validation")
	}
	v, e := exec.Command(php, "-r", "echo PHP_MAJOR_VERSION.'.'.PHP_MINOR_VERSION;").Output()
	if e != nil {
		t.Fatal(e)
	}
	target := string(v)
	if target != "7.4" && target != "8.3" {
		t.Fatalf("unsupported reference PHP %s", target)
	}
	for _, fixture := range []string{"core", "arrays", "max_array_key", "precedence"} {
		t.Run(fixture, func(t *testing.T) {
			file := filepath.Join("fixtures", fixture+".php")
			src, e := os.ReadFile(file)
			if e != nil {
				t.Fatal(e)
			}
			p, e := frontend.Parse(string(src), target)
			if e != nil {
				t.Fatal(e)
			}
			generated, e := codegen.Generate(p)
			if e != nil {
				t.Fatal(e)
			}
			dir := t.TempDir()
			out := filepath.Join(dir, "main.go")
			if e = os.WriteFile(out, generated, 0600); e != nil {
				t.Fatal(e)
			}
			binary := filepath.Join(dir, "app")
			build := exec.Command("go", "build", "-o", binary, out)
			if o, e := build.CombinedOutput(); e != nil {
				t.Fatalf("build: %v\n%s\n%s", e, o, generated)
			}
			want, e := exec.Command(php, file).Output()
			if e != nil {
				t.Fatal(e)
			}
			native := exec.Command(binary)
			native.Env = []string{"PATH="} // The emitted binary must run without discovering PHP.
			got, e := native.Output()
			if e != nil {
				t.Fatal(e)
			}
			if !bytes.Equal(got, want) {
				t.Fatalf("PHP=%q Go=%q", want, got)
			}
			t.Logf("PHP %s and generated Go match: %q", target, got)
		})
	}
}
func TestUnsupported(t *testing.T) {
	for _, src := range []string{
		"<?php $f = function ($x) { return $x; };", "<?php include 'x.php';", "<?php $x = 1.5;", "<?php echo unknown();", "<?php echo $undefined;", "<?php $x = new Missing();", "<?php function x($a) { return $a; } echo x();", "<?php $a = [true => 1];", "<?php class X { public function __construct() {} }", "<?php $a = 012;", "<?php $a = &$b;", "<?php $a = [1]; $a[0] = 2;", "<?php return 1;",
		"<?php echo count(1);", "<?php foreach (1 as $value) { echo $value; }", "<?php $value = 1; echo $value[0];", "<?php class X { public $name = 'x'; } $x = new X(); echo $x->missing;", "<?php class X { public function render($value) { return $value; } } $x = new X(); echo $x->render();",
	} {
		p, e := frontend.Parse(src, "8.3")
		if e == nil {
			_, e = codegen.Generate(p)
		}
		if e == nil {
			t.Errorf("silently accepted unsupported source: %s", src)
		} else if !strings.Contains(e.Error(), "PHP00") {
			t.Errorf("missing diagnostic code: %v", e)
		}
	}
}

func TestMalformedIRFailsWithoutPanic(t *testing.T) {
	p, e := frontend.Parse("<?php echo 1 + 2;", "8.3")
	if e != nil {
		t.Fatal(e)
	}
	p.Body[0].Expr.Args = p.Body[0].Expr.Args[:1]
	_, e = codegen.Generate(p)
	if e == nil || !strings.Contains(e.Error(), "IR002") {
		t.Fatalf("malformed IR was not rejected safely: %v", e)
	}
}
func TestRuntimeFailureIsNotPass(t *testing.T) {
	src, e := os.ReadFile("fixtures/runtime_failure.php")
	if e != nil {
		t.Fatal(e)
	}
	p, e := frontend.Parse(string(src), "8.3")
	if e != nil {
		t.Fatal(e)
	}
	b, e := codegen.Generate(p)
	if e != nil {
		t.Fatal(e)
	}
	file := filepath.Join(t.TempDir(), "main.go")
	if e = os.WriteFile(file, b, 0600); e != nil {
		t.Fatal(e)
	}
	o, e := exec.Command("go", "run", file).CombinedOutput()
	if e == nil || !bytes.Contains(o, []byte("undefined array key")) {
		t.Fatalf("expected explicit runtime failure: %v %s", e, o)
	}
}
