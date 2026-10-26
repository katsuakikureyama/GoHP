package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
)

func TestStrContainsLoweringAndSourceMap(t *testing.T) {
	p, err := frontend.Parse("<?php\nfunction contains($h, $n): bool { return StR_CoNtAiNs($h, $n); }\necho contains('abc', 'b') === true;", testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	out, mapping, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	for line, text := range strings.Split(string(out), "\n") {
		if strings.Contains(text, `rt.StrContains(v["h"], v["n"])`) {
			if got, ok := mapping.PHPLine(line + 1); !ok || got != 2 {
				t.Fatalf("mapping %d %v", got, ok)
			}
			return
		}
	}
	t.Fatalf("missing lowering:\n%s", out)
}

func TestStrContainsUnsupportedSource(t *testing.T) {
	for name, source := range map[string]string{
		"zero": "echo str_contains();", "one": "echo str_contains('a');", "three": "echo str_contains('a', 'a', true);",
		"null": "echo str_contains(null, '');", "bool": "echo str_contains('', false);",
		"integer": "echo str_contains(0, '');", "array": "echo str_contains('', []);",
		"object": "class X {} echo str_contains(new X(), '');", "callable": "echo str_contains('', fn() => '');",
		"undefined": "echo str_contains($missing, '');", "dead": "echo true || str_contains([], '');",
		"named": "echo str_contains(haystack: 'abc', needle: 'a');", "unpack": "echo str_contains(...['abc', 'a']);",
		"redeclare":        "function StR_CoNtAiNs($a, $b) { return true; }",
		"builtin_callback": "$a = array_map('str_contains', ['x']);",
		"array_return":     "function bad(): array { return str_contains('a', ''); }",
	} {
		t.Run(name, func(t *testing.T) {
			p, err := frontend.Parse("<?php "+source, testTarget(t))
			var out []byte
			if err == nil {
				out, err = Generate(p)
			}
			if err == nil || len(out) != 0 || !strings.Contains(err.Error(), "PHP00") {
				t.Fatalf("unsupported output %s: %v", out, err)
			}
		})
	}
}
