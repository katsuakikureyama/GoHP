package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
)

func TestIsObjectLoweringAndSourceMap(t *testing.T) {
	const source = `<?php
function check($value): bool { return Is_ObJeCt($value); }
class Box { public $name = 'box'; }
class EmptyBox {}
$box = new Box();
$empty = new EmptyBox();
$arrow = fn($value) => $value;
echo check($box) ? 'yes' : 'no';
`
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
		`return rt.BoolReturn(rt.IsObject(v["value"]))`,
		`func (self *C_box) PHPObject() bool`,
		`func (self *C_emptybox) PHPObject() bool`,
		`rt.NewCallable(1, func(args ...any) any`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("is_object lowering missing %q:\n%s", want, generated)
		}
	}
	for line, generatedLine := range strings.Split(text, "\n") {
		if strings.Contains(generatedLine, `return rt.BoolReturn(rt.IsObject(v["value"]))`) {
			if got, ok := mapping.PHPLine(line + 1); !ok || got != 2 {
				t.Fatalf("is_object return maps to PHP line %d, %v; want 2", got, ok)
			}
			return
		}
	}
	t.Fatalf("is_object lowering was not source-mapped:\n%s", generated)
}

func TestIsObjectBoundaries(t *testing.T) {
	for name, source := range map[string]string{
		"missing_argument":   "echo is_object();",
		"extra_argument":     "echo is_object(null, null);",
		"undefined_variable": "echo is_object($missing);",
		"redeclaration":      "function Is_ObJeCt($value) { return true; }",
		"builtin_callback":   "$result = array_map('is_object', []);",
		"bool_as_array":      "echo count(is_object(null));",
	} {
		t.Run(name, func(t *testing.T) {
			p, err := frontend.Parse("<?php "+source, testTarget(t))
			if err == nil {
				out, generatedErr := Generate(p)
				err = generatedErr
				if len(out) != 0 && err != nil {
					t.Fatalf("rejected source unexpectedly generated output: %s", out)
				}
			}
			if err == nil || !strings.Contains(err.Error(), "PHP00") {
				t.Fatalf("unsupported is_object source was accepted: %v", err)
			}
		})
	}
}
