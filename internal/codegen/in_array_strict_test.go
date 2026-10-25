package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
)

func TestInArrayStrictLoweringAndSourceMap(t *testing.T) {
	p, err := frontend.Parse("<?php\n$result = In_ArRaY('x', ['x'], true);\n", "8.3")
	if err != nil {
		t.Fatal(err)
	}
	generated, sourceMap, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	for line, text := range strings.Split(string(generated), "\n") {
		if strings.Contains(text, `rt.InArrayStrict("x", rt.NewArray(`) {
			if got, ok := sourceMap.PHPLine(line + 1); !ok || got != 2 {
				t.Fatalf("in_array maps to PHP line %d, %v; want 2", got, ok)
			}
			return
		}
	}
	t.Fatalf("in_array did not lower through runtime guard:\n%s", generated)
}
