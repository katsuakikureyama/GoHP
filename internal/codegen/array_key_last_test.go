package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
)

func TestArrayKeyLastLoweringAndSourceMap(t *testing.T) {
	p, err := frontend.Parse("<?php\nfunction last_key(array $a) { return ArRaY_KeY_LaSt($a); }\necho last_key([7 => null]);", "8.3")
	if err != nil {
		t.Fatal(err)
	}
	out, mapping, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	for line, text := range strings.Split(string(out), "\n") {
		if strings.Contains(text, `return rt.Clone(rt.ArrayKeyLast(v["a"]))`) {
			if got, ok := mapping.PHPLine(line + 1); !ok || got != 2 {
				t.Fatalf("last-key mapping = %d, %v; want 2", got, ok)
			}
			return
		}
	}
	t.Fatalf("missing last-key lowering:\n%s", out)
}

func TestArrayKeyLastUnsupportedSource(t *testing.T) {
	for name, source := range map[string]string{
		"missing":      "echo array_key_last();",
		"extra":        "echo array_key_last([], []);",
		"null":         "echo array_key_last(null);",
		"bool":         "echo array_key_last(false);",
		"integer":      "echo array_key_last(0);",
		"string":       "echo array_key_last('');",
		"object":       "class X {} echo array_key_last(new X());",
		"callable":     "echo array_key_last(fn() => []);",
		"undefined":    "echo array_key_last($missing);",
		"redeclare":    "function ArRaY_KeY_LaSt($a) { return null; }",
		"named":        "echo array_key_last(array: []);",
		"dead_invalid": "echo true || array_key_last(null);",
	} {
		t.Run(name, func(t *testing.T) {
			p, err := frontend.Parse("<?php "+source, "8.3")
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
