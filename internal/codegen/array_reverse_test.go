package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
)

func TestArrayReverseLoweringAndSourceMap(t *testing.T) {
	for _, flag := range []string{"", ", true", ", false"} {
		p, err := frontend.Parse("<?php\nfunction reversed(array $a): array { return ArRaY_ReVeRsE($a"+flag+"); }\necho count(reversed([]));", testTarget(t))
		if err != nil {
			t.Fatal(err)
		}
		out, mapping, err := GenerateWithMap(p)
		if err != nil {
			t.Fatal(err)
		}
		want := flag
		if want == "" {
			want = ", false"
		}
		found := false
		for line, text := range strings.Split(string(out), "\n") {
			if strings.Contains(text, `rt.ArrayReverse(v["a"]`+want+")") {
				found = true
				if got, ok := mapping.PHPLine(line + 1); !ok || got != 2 {
					t.Fatalf("mapping = %d, %v", got, ok)
				}
			}
		}
		if !found {
			t.Fatalf("missing lowering:\n%s", out)
		}
	}
}

func TestArrayReverseUnsupportedSource(t *testing.T) {
	for name, source := range map[string]string{
		"missing": "$a = array_reverse();", "extra": "$a = array_reverse([], true, false);",
		"null": "$a = array_reverse(null);", "bool": "$a = array_reverse(false);",
		"int": "$a = array_reverse(1);", "string": "$a = array_reverse('');",
		"object": "class X {} $a = array_reverse(new X());", "callable": "$a = array_reverse(fn() => []);",
		"flag_int": "$a = array_reverse([], 1);", "flag_null": "$a = array_reverse([], null);",
		"flag_string": "$a = array_reverse([], '');", "flag_array": "$a = array_reverse([], []);",
		"undefined": "$a = array_reverse($missing);", "undefined_flag": "$a = array_reverse([], $missing);",
		"redeclare": "function ArRaY_ReVeRsE($a) { return $a; }",
		"named":     "$a = array_reverse(array: []);", "echo_array": "echo array_reverse([]);",
		"dead_invalid": "echo true || array_reverse([], 0);",
	} {
		t.Run(name, func(t *testing.T) {
			p, err := frontend.Parse("<?php "+source, testTarget(t))
			var out []byte
			if err == nil {
				out, err = Generate(p)
			}
			if err == nil || len(out) != 0 || !strings.Contains(err.Error(), "PHP00") {
				t.Fatalf("unsupported source produced %s: %v", out, err)
			}
		})
	}
}
