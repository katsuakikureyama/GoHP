package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
)

func TestArrayKeyFirstLoweringAndSourceMap(t *testing.T) {
	p, err := frontend.Parse("<?php\nfunction first_key(array $a) { return ArRaY_KeY_FiRsT($a); }\necho first_key([7 => null]);", testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	out, mapping, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	for line, text := range strings.Split(string(out), "\n") {
		if strings.Contains(text, `return rt.Clone(rt.ArrayKeyFirst(v["a"]))`) {
			if got, ok := mapping.PHPLine(line + 1); !ok || got != 2 {
				t.Fatalf("first-key mapping = %d, %v; want 2", got, ok)
			}
			return
		}
	}
	t.Fatalf("missing first-key lowering:\n%s", out)
}

func TestArrayKeyFirstUnsupportedSource(t *testing.T) {
	for name, source := range map[string]string{
		"missing":      "echo array_key_first();",
		"extra":        "echo array_key_first([], []);",
		"null":         "echo array_key_first(null);",
		"bool":         "echo array_key_first(false);",
		"integer":      "echo array_key_first(0);",
		"string":       "echo array_key_first('');",
		"object":       "class X {} echo array_key_first(new X());",
		"callable":     "echo array_key_first(fn() => []);",
		"undefined":    "echo array_key_first($missing);",
		"redeclare":    "function ArRaY_KeY_FiRsT($a) { return null; }",
		"named":        "echo array_key_first(array: []);",
		"dead_invalid": "echo true || array_key_first(null);",
	} {
		t.Run(name, func(t *testing.T) {
			p, err := frontend.Parse("<?php "+source, testTarget(t))
			var out []byte
			if err == nil {
				out, err = Generate(p)
			}
			if err == nil || len(out) != 0 || !strings.Contains(err.Error(), "PHP00") {
				t.Fatalf("unsupported source produced output %s: %v", out, err)
			}
		})
	}
}
