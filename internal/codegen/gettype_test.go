package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
)

func TestGetTypeLowering(t *testing.T) {
	const source = `<?php
function describe(string $type) { return strtolower(gettype($type)) . ':' . strlen(gettype($type)); }
function scalar(int $number, string $text, bool $flag) { $none = null; return gettype($number) . gettype($text) . gettype($flag) . gettype($none); }
function array_body(array $items) { return gettype($items); }
class Box { public $value = 'text'; }
function object_body() { return gettype(new Box()); }
function callable_body() { $callback = fn($value) => gettype($value); return gettype($callback); }
function trace($value) { echo '<trace>'; return $value; }
$box = new Box();
GetTyPe($box);
echo describe(GetTyPe($box->value));
echo gettype(trace(7));
$token = strtok('first,second', ',');
echo gettype($token);`
	p, err := frontend.Parse(source, testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	generated, mapping, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	text := string(generated)
	for _, want := range []string{
		`rt.StrToLower(rt.GetType(v.Read("type")))`,
		`rt.Strlen(rt.GetType(v.Read("type")))`,
		`_ = rt.GetType(v["box"])`,
		`rt.GetType((v["box"]).(interface{ Get(string) any }).Get("value"))`,
		`rt.GetType(FContext_trace(execution, int64(7)))`,
		`rt.NewCallableWithExecution(1, func(execution *rt.Execution, args ...any) any`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("gettype lowering missing %q:\n%s", want, generated)
		}
	}
	foundSourceMap := false
	for line, generatedLine := range strings.Split(text, "\n") {
		if strings.Contains(generatedLine, `_ = rt.GetType(v["box"])`) {
			if got, ok := mapping.PHPLine(line + 1); !ok || got != 10 {
				t.Fatalf("discarded gettype maps to PHP line %d, %v; want 10", got, ok)
			}
			foundSourceMap = true
			break
		}
	}
	if !foundSourceMap {
		t.Fatalf("discarded gettype was not source-mapped:\n%s", generated)
	}
	for name, wantScalarFrame := range map[string]bool{
		"FContext_describe":      true,
		"FContext_scalar":        true,
		"FContext_array_body":    false,
		"FContext_object_body":   false,
		"FContext_callable_body": false,
	} {
		body := generatedFunction(t, text, name)
		if got := strings.Contains(body, "rt.NewScalarFrame()"); got != wantScalarFrame {
			t.Fatalf("%s scalar frame = %v, want %v:\n%s", name, got, wantScalarFrame, body)
		}
	}
}

func TestGetTypeBoundaries(t *testing.T) {
	for name, source := range map[string]string{
		"missing_argument":        "echo gettype();",
		"extra_argument":          "echo gettype(null, null);",
		"mixedcase_redeclaration": "function GeTtYpE($value) { return $value; }",
		"typed_int_argument":      "function f(int $value) {} f(gettype(null));",
		"typed_bool_return":       "function f(): bool { return gettype(null); }",
		"array_result_use":        "echo count(gettype(null));",
		"undefined_variable":      "echo gettype($missing);",
		"literal_callback":        "$result = array_map('gettype', []);",
	} {
		t.Run(name, func(t *testing.T) {
			p, err := frontend.Parse("<?php "+source, testTarget(t))
			var out []byte
			if err == nil {
				out, err = Generate(p)
			}
			if err == nil || len(out) != 0 || !strings.Contains(err.Error(), "PHP00") {
				t.Fatalf("unsupported gettype source accepted: %s %v", out, err)
			}
		})
	}
}

func generatedFunction(t *testing.T, output, name string) string {
	t.Helper()
	start := strings.Index(output, "func "+name+"(")
	if start < 0 {
		t.Fatalf("generated function %s missing:\n%s", name, output)
	}
	body := output[start:]
	if end := strings.Index(body[1:], "\nfunc "); end >= 0 {
		body = body[:end+1]
	}
	return body
}
