package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
)

func TestStrStartsWithLoweringAndSourceMap(t *testing.T) {
	p, err := frontend.Parse("<?php\nfunction contains($h, $n): bool { return StR_StArTs_WiTh($h, $n); }\necho contains('abc', 'b') === true;", "8.3")
	if err != nil {
		t.Fatal(err)
	}
	out, mapping, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	for line, text := range strings.Split(string(out), "\n") {
		if strings.Contains(text, `rt.StrStartsWith(v["h"], v["n"])`) {
			if got, ok := mapping.PHPLine(line + 1); !ok || got != 2 {
				t.Fatalf("mapping %d %v", got, ok)
			}
			return
		}
	}
	t.Fatalf("missing lowering:\n%s", out)
}

func TestStrStartsWithUnsupportedSource(t *testing.T) {
	for name, source := range map[string]string{
		"zero": "echo str_starts_with();", "one": "echo str_starts_with('a');", "three": "echo str_starts_with('a', 'a', true);",
		"null": "echo str_starts_with(null, '');", "bool": "echo str_starts_with('', false);",
		"integer": "echo str_starts_with(0, '');", "array": "echo str_starts_with('', []);",
		"object": "class X {} echo str_starts_with(new X(), '');", "callable": "echo str_starts_with('', fn() => '');",
		"undefined": "echo str_starts_with($missing, '');", "dead": "echo true || str_starts_with([], '');",
		"named": "echo str_starts_with(haystack: 'abc', needle: 'a');", "unpack": "echo str_starts_with(...['abc', 'a']);",
		"redeclare":        "function StR_StArTs_WiTh($a, $b) { return true; }",
		"builtin_callback": "$a = array_map('str_starts_with', ['x']);",
		"array_return":     "function bad(): array { return str_starts_with('a', ''); }",
	} {
		t.Run(name, func(t *testing.T) {
			p, err := frontend.Parse("<?php "+source, "8.3")
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
