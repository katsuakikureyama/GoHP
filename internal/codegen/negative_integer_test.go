package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
)

func TestNegativeIntegerLiteralLowering(t *testing.T) {
	p, err := frontend.Parse("<?php\n$items = [-5 => 'first'];\n$items[-4] = 'written';\n$items[] = -3;\n", testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	generated, sourceMap, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	for needle, phpLine := range map[string]int{
		`Key: int64(-5)`: 2,
		`rt.ArraySet(v["items"], int64(-4), "written")`: 3,
		`rt.ArrayAppend(v["items"], int64(-3))`:         4,
	} {
		found := false
		for line, text := range strings.Split(string(generated), "\n") {
			if !strings.Contains(text, needle) {
				continue
			}
			found = true
			if got, ok := sourceMap.PHPLine(line + 1); !ok || got != phpLine {
				t.Fatalf("%s maps to PHP line %d, %v; want %d", needle, got, ok, phpLine)
			}
		}
		if !found {
			t.Fatalf("missing %s in generated Go:\n%s", needle, generated)
		}
	}
}
