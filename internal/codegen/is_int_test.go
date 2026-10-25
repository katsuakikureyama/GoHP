package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
)

func TestIsIntLoweringAndSourceMap(t *testing.T) {
	p, err := frontend.Parse("<?php\nfunction check($value): bool { return Is_InT($value); }\necho check(0);", "8.3")
	if err != nil {
		t.Fatal(err)
	}
	output, mapping, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	for line, content := range strings.Split(string(output), "\n") {
		if strings.Contains(content, `return rt.BoolReturn(rt.IsInt(v["value"]))`) {
			if got, ok := mapping.PHPLine(line + 1); !ok || got != 2 {
				t.Fatalf("is_int return maps to %d (%v), want 2", got, ok)
			}
			return
		}
	}
	t.Fatalf("missing canonical is_int predicate in guarded bool return:\n%s", output)
}
