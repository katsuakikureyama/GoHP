package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
)

func TestIsNullLoweringAndSourceMap(t *testing.T) {
	p, err := frontend.Parse("<?php\nfunction check($value): bool { return Is_NuLl($value); }\necho check(NULL);", testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	output, mapping, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	for line, content := range strings.Split(string(output), "\n") {
		if strings.Contains(content, `return rt.BoolReturn(rt.IsNull(v["value"]))`) {
			if got, ok := mapping.PHPLine(line + 1); !ok || got != 2 {
				t.Fatalf("is_null return maps to %d (%v), want 2", got, ok)
			}
			return
		}
	}
	t.Fatalf("missing canonical is_null predicate in guarded bool return:\n%s", output)
}
