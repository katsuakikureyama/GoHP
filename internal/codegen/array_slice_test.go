package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
)

func TestArraySliceLoweringAndSourceMap(t *testing.T) {
	for _, tc := range []struct{ args, want string }{
		{"$a, 1", `v["a"], int64(1), nil, false`},
		{"$a, 1, null", `v["a"], int64(1), nil, false`},
		{"$a, 1, 2, true", `v["a"], int64(1), int64(2), true`},
	} {
		p, err := frontend.Parse("<?php\nfunction sliced(array $a): array { return ArRaY_SlIcE("+tc.args+"); }\necho count(sliced([]));", testTarget(t))
		if err != nil {
			t.Fatal(err)
		}
		out, mapping, err := GenerateWithMap(p)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for line, text := range strings.Split(string(out), "\n") {
			if strings.Contains(text, "rt.ArraySlice("+tc.want+")") {
				found = true
				if got, ok := mapping.PHPLine(line + 1); !ok || got != 2 {
					t.Fatalf("mapping: %d %v", got, ok)
				}
			}
		}
		if !found {
			t.Fatalf("missing lowering:\n%s", out)
		}
	}
}

func TestArraySliceUnsupportedSource(t *testing.T) {
	for name, source := range map[string]string{
		"zero": "$a = array_slice();", "one": "$a = array_slice([]);", "five": "$a = array_slice([], 0, null, false, 0);",
		"nonarray": "$a = array_slice(null, 0);", "offset_null": "$a = array_slice([], null);",
		"offset_bool": "$a = array_slice([], false);", "offset_string": "$a = array_slice([], '0');",
		"length_bool": "$a = array_slice([], 0, false);", "length_string": "$a = array_slice([], 0, '1');",
		"length_array": "$a = array_slice([], 0, []);", "flag_null": "$a = array_slice([], 0, null, null);",
		"flag_int": "$a = array_slice([], 0, null, 1);", "undefined": "$a = array_slice([], $missing);",
		"redeclare": "function ArRaY_SlIcE($a) { return $a; }", "named": "$a = array_slice(array: [], offset: 0);",
		"unpack": "$a = array_slice(...[[], 0]);", "echo_array": "echo array_slice([], 0);",
		"dead_invalid": "echo true || array_slice([], false);", "builtin_callback": "$a = array_map('array_slice', [[]]);",
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
