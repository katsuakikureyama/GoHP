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
	"github.com/katsuakikureyama/GoHP/internal/phpversion"
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
		target = phpversion.DefaultTarget
	}
	if err := phpversion.ValidateTarget(target); err != nil {
		t.Fatal(err)
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
	if err := phpversion.ValidateOracle(target, actualPHPVersion); err != nil {
		t.Fatal(err)
	}
	for _, fixture := range []string{"core", "arrays", "legacy_arrays", "max_array_key", "negative_array_keys", "negative_integers", "logical_not", "logical_or", "strict_not_equal", "if_else", "precedence", "strlen", "array_keys", "array_values", "array_key_exists", "array_is_list", "is_array", "array_map", "array_write", "array_append", "rtrim", "sprintf", "str_replace", "ternary_stripslashes", "booleans", "feature_integration", "string_array_integration", "conditional_array_integration", "predicate_conditional_integration", "callback_composition_integration"} {
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
			phpArgs := []string{file}
			if fixture == "booleans" {
				// Set and verify the locale in the same process as the oracle.
				phpArgs = []string{"-r", "if (setlocale(LC_CTYPE, 'C') !== 'C' || setlocale(LC_CTYPE, 0) !== 'C') { throw new RuntimeException('LC_CTYPE must be C'); } require $argv[1];", "--", file}
			}
			want, e := exec.Command(php, phpArgs...).Output()
			if e != nil {
				evidence.Phases["differential"] = phaseFail
				t.Fatal(e)
			}
			native := exec.Command(binary)
			if fixture == "booleans" {
				// Keep the versioned GOHP_CASE payload unchanged; the raw event
				// archive retains this separately recorded oracle precondition.
				t.Log("PHP LC_CTYPE=C selected and verified in the oracle process")
			}
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
		{name: "legacy_array_reference", src: "<?php $value = 1; $a = array(&$value);"},
		{name: "legacy_array_unpacking", src: "<?php $values = [1]; $a = array(...$values);"},
		{name: "legacy_array_dynamic_key", src: "<?php $key = 'key'; $a = array($key => 1);"},
		{name: "legacy_array_empty_hole", src: "<?php $a = array(,);"},
		{name: "legacy_array_mismatched_delimiter", src: "<?php $a = array(1];"},
		{name: "legacy_array_nested_mismatched_delimiter", src: "<?php $a = array(array(1]);"},
		{name: "legacy_array_bareword", src: "<?php $a = array;"},
		{name: "magic_constructor", src: "<?php class X { public function __construct() {} }"},
		{name: "nondecimal_integer", src: "<?php $a = 012;"},
		{name: "negative_integer_variable", src: "<?php $a = 1; echo -$a;"},
		{name: "negative_integer_expression", src: "<?php echo -(1 + 2);"},
		{name: "negative_integer_call", src: "<?php function value() { return 1; } echo -value();"},
		{name: "negative_integer_string", src: "<?php echo -'5';"},
		{name: "negative_integer_repeated_sign", src: "<?php echo - -5;"},
		{name: "negative_integer_decrement", src: "<?php $a = 5; echo --$a;"},
		{name: "negative_integer_subtraction", src: "<?php echo 1 - 2;"},
		{name: "negative_integer_minimum_token", src: "<?php echo -9223372036854775808;"},
		{name: "negative_integer_float", src: "<?php echo -1.5;"},
		{name: "negative_integer_nondecimal", src: "<?php echo -012;"},
		{name: "logical_not_object", src: "<?php class Box {} echo !new Box();"},
		{name: "logical_not_dead_object", src: "<?php class Box {} echo false && !new Box();"},
		{name: "logical_not_result_array", src: "<?php echo count(![]);"},
		{name: "logical_not_literal_key", src: "<?php $a = [!false => 1];"},
		{name: "logical_not_missing_operand", src: "<?php echo !;"},
		{name: "logical_not_equal", src: "<?php echo 1 != 2;"},
		{name: "logical_not_keyword", src: "<?php echo not true;"},
		{name: "logical_not_property_default", src: "<?php class Box { public $value = !false; }"},
		{name: "reference", src: "<?php $a = &$b;"},
		{name: "array_nested_write", src: "<?php $a = [[1]]; $a[0][0] = 2;"},
		{name: "array_append_undefined", src: "<?php $a[] = 2;"},
		{name: "array_append_null", src: "<?php $a = null; $a[] = 2;"},
		{name: "array_append_scalar", src: "<?php $a = 'x'; $a[] = 2;"},
		{name: "array_append_bool", src: "<?php $a = false; $a[] = 2;"},
		{name: "array_append_object", src: "<?php class X {} $a = new X(); $a[] = 2;"},
		{name: "array_append_nested", src: "<?php $a = [[]]; $a[0][] = 2;"},
		{name: "array_append_outer_nested", src: "<?php $a = []; $a[][0] = 2;"},
		{name: "array_append_reference", src: "<?php $a = []; $v = 2; $a[] = &$v;"},
		{name: "array_append_property", src: "<?php class X { public $value = null; } $a = new X(); $a->value[] = 2;"},
		{name: "array_append_compound", src: "<?php $a = []; $a[] += 2;"},
		{name: "array_append_read", src: "<?php $a = []; echo $a[];"},
		{name: "array_append_expression", src: "<?php $a = []; $b = ($a[] = 2);"},
		{name: "array_append_conditional_root", src: "<?php if (true) { $a = []; } $a[] = 2;"},
		{name: "array_write_undefined", src: "<?php $a[0] = 2;"},
		{name: "array_write_null", src: "<?php $a = null; $a[0] = 2;"},
		{name: "array_write_scalar", src: "<?php $a = 'x'; $a[0] = 'y';"},
		{name: "array_write_bool_key", src: "<?php $a = []; $key = true; $a[$key] = 2;"},
		{name: "array_write_null_key", src: "<?php $a = []; $key = null; $a[$key] = 2;"},
		{name: "array_write_key_call", src: "<?php function key_value() { return 0; } $a = []; $a[key_value()] = 2;"},
		{name: "array_write_key_expression", src: "<?php $a = []; $a[1 + 1] = 2;"},
		{name: "array_write_reference", src: "<?php $a = []; $value = 1; $a[0] = &$value;"},
		{name: "array_write_property", src: "<?php class X { public $value = null; } $a = new X(); $a->value[0] = 2;"},
		{name: "array_write_conditional_root", src: "<?php if (true) { $a = []; } $a[0] = 2;"},
		{name: "top_level_return", src: "<?php return 1;"},
		{name: "count_scalar", src: "<?php echo count(1);"},
		{name: "rtrim_missing_argument", src: "<?php echo rtrim('x');"},
		{name: "rtrim_extra_argument", src: "<?php echo rtrim('x', 'x', 'x');"},
		{name: "rtrim_int_input", src: "<?php echo rtrim(1, '1');"},
		{name: "rtrim_bool_input", src: "<?php echo rtrim(true, '1');"},
		{name: "rtrim_null_input", src: "<?php echo rtrim(null, 'x');"},
		{name: "rtrim_array_input", src: "<?php echo rtrim([], 'x');"},
		{name: "rtrim_dynamic_mask", src: "<?php $mask = 'x'; echo rtrim('x', $mask);"},
		{name: "rtrim_range_mask", src: "<?php echo rtrim('xyz', 'a..z');"},
		{name: "rtrim_redeclare", src: "<?php function RtRiM() { return 'x'; }"},
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
		{name: "array_keys_int", src: "<?php $keys = array_keys(1);"},
		{name: "array_keys_string", src: "<?php $keys = array_keys('x');"},
		{name: "array_keys_bool", src: "<?php $keys = array_keys(true);"},
		{name: "array_keys_null", src: "<?php $keys = array_keys(null);"},
		{name: "array_keys_object", src: "<?php class X {} $x = new X(); $keys = array_keys($x);"},
		{name: "array_keys_no_args", src: "<?php $keys = array_keys();"},
		{name: "array_keys_filter", src: "<?php $keys = array_keys(['x' => 1], 1);"},
		{name: "array_keys_strict", src: "<?php $keys = array_keys(['x' => 1], 1, true);"},
		{name: "array_keys_redeclare", src: "<?php function ArRaY_KeYs($value) { return $value; }"},
		{name: "array_keys_echo_result", src: "<?php $keys = array_keys(['x' => 1]); echo $keys;"},
		{name: "array_values_int", src: "<?php $values = array_values(1);"},
		{name: "array_values_string", src: "<?php $values = array_values('x');"},
		{name: "array_values_bool", src: "<?php $values = array_values(true);"},
		{name: "array_values_null", src: "<?php $values = array_values(null);"},
		{name: "array_values_object", src: "<?php class X {} $x = new X(); $values = array_values($x);"},
		{name: "array_values_no_args", src: "<?php $values = array_values();"},
		{name: "array_values_extra_args", src: "<?php $values = array_values(['x'], 'extra');"},
		{name: "array_values_redeclare", src: "<?php function ArRaY_VaLuEs($value) { return $value; }"},
		{name: "array_values_echo_result", src: "<?php $values = array_values(['x' => 1]); echo $values;"},
		{name: "array_key_exists_invalid_bool_key", src: "<?php $exists = array_key_exists(true, ['x' => 1]);"},
		{name: "array_key_exists_invalid_null_key", src: "<?php $exists = array_key_exists(null, ['x' => 1]);"},
		{name: "array_key_exists_invalid_array_key", src: "<?php $exists = array_key_exists([], ['x' => 1]);"},
		{name: "array_key_exists_invalid_object_key", src: "<?php class X {} $key = new X(); $exists = array_key_exists($key, ['x' => 1]);"},
		{name: "array_key_exists_invalid_int_array", src: "<?php $exists = array_key_exists('x', 1);"},
		{name: "array_key_exists_invalid_string_array", src: "<?php $exists = array_key_exists('x', 'not an array');"},
		{name: "array_key_exists_invalid_bool_array", src: "<?php $exists = array_key_exists('x', true);"},
		{name: "array_key_exists_invalid_null_array", src: "<?php $exists = array_key_exists('x', null);"},
		{name: "array_key_exists_invalid_object_array", src: "<?php class X {} $value = new X(); $exists = array_key_exists('x', $value);"},
		{name: "array_key_exists_arity_zero", src: "<?php $exists = array_key_exists();"},
		{name: "array_key_exists_arity_one", src: "<?php $exists = array_key_exists('x');"},
		{name: "array_key_exists_arity_three", src: "<?php $exists = array_key_exists('x', [], 'extra');"},
		{name: "array_key_exists_mixed_case_redeclare", src: "<?php function ArRaY_KeY_ExIsTs($key, $value) { return true; }"},
		{name: "array_key_exists_alias", src: "<?php $exists = key_exists('x', ['x' => 1]);"},
		{name: "array_is_list_int", src: "<?php $list = array_is_list(1);"},
		{name: "array_is_list_string", src: "<?php $list = array_is_list('not an array');"},
		{name: "array_is_list_bool", src: "<?php $list = array_is_list(true);"},
		{name: "array_is_list_null", src: "<?php $list = array_is_list(null);"},
		{name: "array_is_list_object", src: "<?php class X {} $value = new X(); $list = array_is_list($value);"},
		{name: "array_is_list_no_args", src: "<?php $list = array_is_list();"},
		{name: "array_is_list_extra_args", src: "<?php $list = array_is_list([], []);"},
		{name: "array_is_list_redeclare", src: "<?php function ArRaY_Is_LiSt($value) { return true; }"},
		{name: "array_is_list_result_as_array", src: "<?php $count = count(array_is_list([]));"},
		{name: "if_else_unbraced", src: "<?php if (false) {} else echo 'x';"},
		{name: "if_else_if", src: "<?php if (false) {} else if (true) {}"},
		{name: "if_else_duplicate", src: "<?php if (false) {} else {} else {}"},
		{name: "if_else_orphan", src: "<?php else {}"},
		{name: "if_else_alternative", src: "<?php if (false): echo 'a'; else: echo 'b'; endif;"},
		{name: "if_else_local_escape", src: "<?php if (false) {} else { $local = 1; } echo $local;"},
		{name: "if_else_both_local_escape", src: "<?php if (true) { $local = 1; } else { $local = 2; } echo $local;"},
		{name: "if_else_sibling_local", src: "<?php if (true) { $local = 1; } else { echo $local; }"},
		{name: "if_else_sibling_type", src: "<?php $value = 1; if (false) { $value = []; } else { echo count($value); }"},
		{name: "if_else_dead_call", src: "<?php if (true) {} else { echo missing(); }"},
		{name: "if_else_dead_array", src: "<?php if (true) {} else { echo []; }"},
		{name: "if_else_object_condition", src: "<?php class Box {} if (new Box()) {} else {}"},
		{name: "if_elseif", src: "<?php if (false) { echo 'no'; } elseif (true) { echo 'yes'; }"},
		{name: "if_unbraced", src: "<?php if (true) echo 'yes';"},
		{name: "cast_int", src: "<?php echo (int) '1';"},
		{name: "logical_or_object_left", src: "<?php class Box {} echo new Box() || false;"},
		{name: "logical_or_object_right", src: "<?php class Box {} echo false || new Box();"},
		{name: "logical_or_dead_object", src: "<?php class Box {} echo true || new Box();"},
		{name: "logical_or_dead_call", src: "<?php echo true || missing();"},
		{name: "logical_or_dead_variable", src: "<?php echo true || $missing;"},
		{name: "logical_or_missing_left", src: "<?php echo || true;"},
		{name: "logical_or_missing_right", src: "<?php echo false ||;"},
		{name: "logical_or_keyword", src: "<?php echo false or true;"},
		{name: "logical_or_bitwise", src: "<?php echo 1 | 2;"},
		{name: "logical_or_result_array", src: "<?php echo count([] || [1]);"},
		{name: "logical_or_literal_key", src: "<?php $a = [false || true => 1];"},
		{name: "logical_or_property_default", src: "<?php class Box { public $value = false || true; }"},
		{name: "loose_equal", src: "<?php echo 1 == '1';"},
		{name: "strict_equal_array", src: "<?php echo [] === [];"},
		{name: "strict_equal_chain", src: "<?php echo true === false === false;"},
		{name: "strict_not_equal_array_left", src: "<?php echo [] !== null;"},
		{name: "strict_not_equal_array_right", src: "<?php echo null !== [];"},
		{name: "strict_not_equal_object_left", src: "<?php class Box {} echo new Box() !== null;"},
		{name: "strict_not_equal_object_right", src: "<?php class Box {} echo null !== new Box();"},
		{name: "strict_not_equal_dead_array", src: "<?php echo true || ([] !== []);"},
		{name: "strict_not_equal_chain", src: "<?php echo 1 !== 2 !== 3;"},
		{name: "strict_not_equal_chain_equal_first", src: "<?php echo 1 === 2 !== false;"},
		{name: "strict_not_equal_chain_equal_last", src: "<?php echo 1 !== 2 === true;"},
		{name: "strict_not_equal_missing_left", src: "<?php echo !== true;"},
		{name: "strict_not_equal_missing_right", src: "<?php echo false !==;"},
		{name: "strict_not_equal_result_array", src: "<?php echo count(1 !== 2);"},
		{name: "strict_not_equal_literal_key", src: "<?php $a = [1 !== 2 => 0];"},
		{name: "strict_not_equal_property_default", src: "<?php class Box { public $value = 1 !== 2; }"},
		{name: "is_bool_missing_argument", src: "<?php echo is_bool();"},
		{name: "is_bool_extra_argument", src: "<?php echo is_bool(true, false);"},
		{name: "is_string_missing_argument", src: "<?php echo is_string();"},
		{name: "is_string_extra_argument", src: "<?php echo is_string('x', 'y');"},
		{name: "is_bool_redeclare", src: "<?php function Is_Bool() { return true; }"},
		{name: "is_string_redeclare", src: "<?php function is_STRING() { return true; }"},
		{name: "is_array_missing_argument", src: "<?php echo is_array();"},
		{name: "is_array_extra_argument", src: "<?php echo is_array([], []);"},
		{name: "is_array_redeclare", src: "<?php function Is_ArRaY($value) { return true; }"},
		{name: "is_array_bool_as_array", src: "<?php echo count(is_array([]));"},
		{name: "strtolower_missing_argument", src: "<?php echo strtolower();"},
		{name: "strtolower_extra_argument", src: "<?php echo strtolower('X', 'Y');"},
		{name: "strtolower_int_input", src: "<?php echo strtolower(1);"},
		{name: "strtolower_bool_input", src: "<?php echo strtolower(true);"},
		{name: "strtolower_null_input", src: "<?php echo strtolower(null);"},
		{name: "strtolower_array_input", src: "<?php echo strtolower([]);"},
		{name: "strtolower_redeclare", src: "<?php function StRtOlOwEr($value) { return $value; }"},
		{name: "if_branch_local_variable", src: "<?php if (true) { $inside = 'x'; } echo $inside;"},
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

// TestArrayKeysRuntimeTypeGuard covers the dynamic boundary permitted by
// semantic analysis: a function parameter is unknown at generation time, but
// array_keys must still fail explicitly when it receives a scalar. This has no
// native-PHP invocation or PHP error-equivalence claim.
func TestArrayKeysRuntimeTypeGuard(t *testing.T) {
	target := phpTarget(t)
	src, err := os.ReadFile("fixtures/array_keys_runtime_failure.php")
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
	if err == nil || !bytes.Contains(output, []byte("array_keys requires array")) {
		t.Fatalf("expected explicit array_keys runtime failure, got %v\n%s", err, output)
	}
}

// TestArrayValuesRuntimeTypeGuard covers the dynamic boundary permitted by
// semantic analysis: a function parameter is unknown at generation time, but
// array_values must still fail explicitly when it receives a scalar. This has
// no native-PHP invocation or PHP error-equivalence claim.
func TestArrayValuesRuntimeTypeGuard(t *testing.T) {
	target := phpTarget(t)
	src, err := os.ReadFile("fixtures/array_values_runtime_failure.php")
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
	if err == nil || !bytes.Contains(output, []byte("array_values requires array")) {
		t.Fatalf("expected explicit array_values runtime failure, got %v\n%s", err, output)
	}
}

// TestArrayKeyExistsRuntimeTypeGuard covers both dynamic boundaries permitted
// by semantic analysis. Function parameters are unknown at generation time,
// but array_key_exists must reject an invalid key and an invalid array value
// explicitly. This deliberately has no native-PHP invocation or PHP
// error-equivalence claim.
func TestArrayKeyExistsRuntimeTypeGuard(t *testing.T) {
	target := phpTarget(t)
	for _, tt := range []struct {
		name    string
		fixture string
		message string
	}{
		{name: "key", fixture: "array_key_exists_runtime_key_failure.php", message: "array_key_exists requires integer or string key"},
		{name: "array", fixture: "array_key_exists_runtime_array_failure.php", message: "array_key_exists requires array"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			src, err := os.ReadFile(filepath.Join("fixtures", tt.fixture))
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
			if err == nil || !bytes.Contains(output, []byte(tt.message)) {
				t.Fatalf("expected explicit %s runtime failure, got %v\n%s", tt.name, err, output)
			}
		})
	}
}

// TestArrayIsListRuntimeTypeGuard covers the dynamic boundary permitted by
// semantic analysis: a function parameter is unknown at generation time, but
// array_is_list must still fail explicitly when it receives a scalar. This has
// no native-PHP invocation or PHP error-equivalence claim.
func TestArrayIsListRuntimeTypeGuard(t *testing.T) {
	target := phpTarget(t)
	src, err := os.ReadFile("fixtures/array_is_list_runtime_failure.php")
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
	if err == nil || !bytes.Contains(output, []byte("array_is_list requires array")) {
		t.Fatalf("expected explicit array_is_list runtime failure, got %v\n%s", err, output)
	}
}

func sha256Hex(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }
