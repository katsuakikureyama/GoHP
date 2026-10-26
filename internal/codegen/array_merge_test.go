package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
)

func TestArrayMergeLoweringAndSourceMap(t *testing.T) {
	for _, tc := range []struct{ source, want string }{
		{"", ""}, {"$a", `v["a"]`},
		{"$a, $b", `v["a"], v["b"]`},
		{"$a, $b, $a", `v["a"], v["b"], v["a"]`},
	} {
		p, err := frontend.Parse("<?php\nfunction merged(array $a, array $b): array { return ArRaY_MeRgE("+tc.source+"); }\necho count(merged([], []));", testTarget(t))
		if err != nil {
			t.Fatal(err)
		}
		out, mapping, err := GenerateWithMap(p)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for line, text := range strings.Split(string(out), "\n") {
			if strings.Contains(text, "rt.ArrayMerge("+tc.want+")") {
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

func TestArrayMergeUnsupportedSource(t *testing.T) {
	for name, source := range map[string]string{
		"null": "$a = array_merge(null);", "bool": "$a = array_merge(false);",
		"int": "$a = array_merge(1);", "string": "$a = array_merge('');",
		"object": "class X {} $a = array_merge(new X());", "callable": "$a = array_merge(fn() => []);",
		"second": "$a = array_merge([], false);", "third": "$a = array_merge([], [], 0);",
		"undefined": "$a = array_merge([], $missing);", "redeclare": "function ArRaY_MeRgE($a) { return $a; }",
		"named": "$a = array_merge(arrays: []);", "unpack": "$a = array_merge(...[[]]);",
		"echo_array": "echo array_merge();", "dead_invalid": "echo true || array_merge([], null);",
		"builtin_callback": "$a = array_map('array_merge', [[]]);",
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
