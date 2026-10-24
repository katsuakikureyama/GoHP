package tests

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/codegen"
	"github.com/katsuakikureyama/GoHP/internal/frontend"
)

const (
	phaseNotRun          = "not-run"
	phasePass            = "pass"
	phaseFail            = "fail"
	comparisonNotChecked = "not-checked"
)

type differentialEvidence struct {
	CaseID          string            `json:"case_id"`
	PHPTarget       string            `json:"php_target"`
	PHPVersion      string            `json:"php_version"`
	FixtureSHA256   string            `json:"fixture_sha256"`
	GeneratedSHA256 string            `json:"generated_sha256"`
	Phases          map[string]string `json:"phases"`
	Comparisons     map[string]string `json:"comparisons"`
}

func phpTarget(t *testing.T) string {
	t.Helper()
	target := os.Getenv("GOHP_PHP_TARGET")
	if target == "" {
		target = "8.3"
	}
	if target != "7.4" && target != "8.3" {
		t.Fatalf("GOHP_PHP_TARGET = %q, want 7.4 or 8.3", target)
	}
	return target
}

func TestDifferential(t *testing.T) {
	php, e := exec.LookPath("php")
	if e != nil {
		t.Fatal("native PHP is required for differential validation")
	}
	v, e := exec.Command(php, "-r", "echo PHP_VERSION;").Output()
	if e != nil {
		t.Fatal(e)
	}
	actualPHPVersion := string(v)
	target := phpTarget(t)
	if !strings.HasPrefix(actualPHPVersion, target+".") {
		t.Fatalf("reference PHP version %s does not match GOHP_PHP_TARGET %s", actualPHPVersion, target)
	}
	for _, fixture := range []string{"core", "arrays", "max_array_key", "precedence", "strlen"} {
		t.Run(fixture, func(t *testing.T) {
			evidence := differentialEvidence{
				CaseID:     "diff." + fixture,
				PHPTarget:  target,
				PHPVersion: actualPHPVersion,
				Phases: map[string]string{
					"parse": phaseNotRun, "codegen": phaseNotRun, "go-build": phaseNotRun, "go-execute": phaseNotRun, "differential": phaseNotRun,
				},
				Comparisons: map[string]string{
					"stdout": comparisonNotChecked, "exit": comparisonNotChecked, "stderr": comparisonNotChecked, "exceptions": comparisonNotChecked, "side_effects": comparisonNotChecked,
				},
			}
			defer func() {
				encoded, err := json.Marshal(evidence)
				if err != nil {
					t.Errorf("marshal differential evidence: %v", err)
					return
				}
				t.Logf("GOHP_CASE %s", encoded)
			}()

			file := filepath.Join("fixtures", fixture+".php")
			src, e := os.ReadFile(file)
			if e != nil {
				t.Fatal(e)
			}
			evidence.FixtureSHA256 = sha256Hex(src)
			p, e := frontend.Parse(string(src), target)
			if e != nil {
				evidence.Phases["parse"] = phaseFail
				t.Fatal(e)
			}
			evidence.Phases["parse"] = phasePass
			generated, e := codegen.Generate(p)
			if e != nil {
				evidence.Phases["codegen"] = phaseFail
				t.Fatal(e)
			}
			evidence.Phases["codegen"] = phasePass
			evidence.GeneratedSHA256 = sha256Hex(generated)
			dir := t.TempDir()
			out := filepath.Join(dir, "main.go")
			if e = os.WriteFile(out, generated, 0600); e != nil {
				t.Fatal(e)
			}
			binary := filepath.Join(dir, "app")
			build := exec.Command("go", "build", "-o", binary, out)
			if o, e := build.CombinedOutput(); e != nil {
				evidence.Phases["go-build"] = phaseFail
				t.Fatalf("build: %v\n%s\n%s", e, o, generated)
			}
			evidence.Phases["go-build"] = phasePass
			want, e := exec.Command(php, file).Output()
			if e != nil {
				evidence.Phases["differential"] = phaseFail
				t.Fatal(e)
			}
			native := exec.Command(binary)
			native.Env = []string{"PATH="} // The emitted binary must run without discovering PHP.
			got, e := native.Output()
			if e != nil {
				evidence.Phases["go-execute"] = phaseFail
				evidence.Comparisons["exit"] = phaseFail
				t.Fatal(e)
			}
			evidence.Phases["go-execute"] = phasePass
			if !bytes.Equal(got, want) {
				evidence.Phases["differential"] = phaseFail
				evidence.Comparisons["stdout"] = phaseFail
				evidence.Comparisons["exit"] = phasePass
				t.Fatalf("PHP=%q Go=%q", want, got)
			}
			evidence.Phases["differential"] = phasePass
			evidence.Comparisons["stdout"] = phasePass
			evidence.Comparisons["exit"] = phasePass
			t.Logf("PHP %s and generated Go match: %q", target, got)
		})
	}
}
func TestUnsupported(t *testing.T) {
	target := phpTarget(t)
	for _, tt := range []struct {
		name string
		src  string
	}{
		{name: "closure", src: "<?php $f = function ($x) { return $x; };"},
		{name: "include", src: "<?php include 'x.php';"},
		{name: "float", src: "<?php $x = 1.5;"},
		{name: "unknown_function", src: "<?php echo unknown();"},
		{name: "undefined_variable", src: "<?php echo $undefined;"},
		{name: "unknown_class", src: "<?php $x = new Missing();"},
		{name: "function_arity", src: "<?php function x($a) { return $a; } echo x();"},
		{name: "boolean_array_key", src: "<?php $a = [true => 1];"},
		{name: "magic_constructor", src: "<?php class X { public function __construct() {} }"},
		{name: "nondecimal_integer", src: "<?php $a = 012;"},
		{name: "reference", src: "<?php $a = &$b;"},
		{name: "array_write", src: "<?php $a = [1]; $a[0] = 2;"},
		{name: "top_level_return", src: "<?php return 1;"},
		{name: "count_scalar", src: "<?php echo count(1);"},
		{name: "foreach_scalar", src: "<?php foreach (1 as $value) { echo $value; }"},
		{name: "index_scalar", src: "<?php $value = 1; echo $value[0];"},
		{name: "missing_property", src: "<?php class X { public $name = 'x'; } $x = new X(); echo $x->missing;"},
		{name: "method_arity", src: "<?php class X { public function render($value) { return $value; } } $x = new X(); echo $x->render();"},
		{name: "strlen_int", src: "<?php echo strlen(1);"},
		{name: "strlen_bool", src: "<?php echo strlen(true);"},
		{name: "strlen_null", src: "<?php echo strlen(null);"},
		{name: "strlen_array", src: "<?php echo strlen(['x']);"},
		{name: "strlen_object", src: "<?php class X {} $x = new X(); echo strlen($x);"},
		{name: "strlen_no_args", src: "<?php echo strlen();"},
		{name: "strlen_extra_args", src: "<?php echo strlen('x', 'y');"},
		{name: "strlen_redeclare", src: "<?php function StRlEn($value) { return $value; }"},
		{name: "strlen_nested", src: "<?php echo strlen(strlen('x'));"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p, e := frontend.Parse(tt.src, target)
			if e == nil {
				_, e = codegen.Generate(p)
			}
			if e == nil {
				t.Errorf("silently accepted unsupported source: %s", tt.src)
			} else if !strings.Contains(e.Error(), "PHP00") {
				t.Errorf("missing diagnostic code: %v", e)
			}
		})
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
	target := phpTarget(t)
	src, e := os.ReadFile("fixtures/runtime_failure.php")
	if e != nil {
		t.Fatal(e)
	}
	p, e := frontend.Parse(string(src), target)
	if e != nil {
		t.Fatal(e)
	}
	b, sourceMap, e := codegen.GenerateWithMap(p)
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
	// Resolve an actual generated-code stack frame, not a guessed Go line.
	frames := regexp.MustCompile(regexp.QuoteMeta(file)+`:(\d+)`).FindAllSubmatch(o, -1)
	for _, frame := range frames {
		line, err := strconv.Atoi(string(frame[1]))
		if err != nil {
			t.Fatal(err)
		}
		if phpLine, ok := sourceMap.PHPLine(line); ok && phpLine == 3 {
			return
		}
	}
	t.Fatalf("runtime failure has no frame mapping to PHP line 3: %s", o)
}

// TestStrlenRuntimeTypeGuard covers the dynamic boundary permitted by semantic
// analysis: a function parameter is unknown at generation time, but strlen
// must still fail explicitly when it receives an array. This deliberately has
// no native-PHP invocation or PHP error-equivalence claim.
func TestStrlenRuntimeTypeGuard(t *testing.T) {
	target := phpTarget(t)
	src, err := os.ReadFile("fixtures/strlen_runtime_failure.php")
	if err != nil {
		t.Fatal(err)
	}
	p, err := frontend.Parse(string(src), target)
	if err != nil {
		t.Fatal(err)
	}
	generated, err := codegen.Generate(p)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	file := filepath.Join(dir, "main.go")
	if err := os.WriteFile(file, generated, 0600); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(dir, "app")
	if output, err := exec.Command("go", "build", "-o", binary, file).CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s\n%s", err, output, generated)
	}
	run := exec.Command(binary)
	run.Env = []string{"PATH="}
	output, err := run.CombinedOutput()
	if err == nil || !bytes.Contains(output, []byte("strlen requires string")) {
		t.Fatalf("expected explicit strlen runtime failure, got %v\n%s", err, output)
	}
}

func sha256Hex(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }
