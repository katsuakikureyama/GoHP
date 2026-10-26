package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
)

func TestInArrayStrictLoweringAndSourceMap(t *testing.T) {
	p, err := frontend.Parse("<?php\n$result = In_ArRaY('x', ['x'], true);\n", testTarget(t))
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

func TestMembershipControlFlowRejectsInvalidDeadArms(t *testing.T) {
	for name, body := range map[string]string{
		"elseif_domain":    "if (true) { return []; } elseif (in_array('x', ['x', []], true)) { return []; } return [];",
		"else_if_mode":     "if (true) { return []; } else if (in_array('x', ['x'], false)) { return []; } return [];",
		"else_return_type": "if (true) { return []; } else { return in_array('x', ['x'], true); }",
	} {
		t.Run(name, func(t *testing.T) {
			p, err := frontend.Parse("<?php function result(): array { "+body+" }", testTarget(t))
			if err != nil {
				t.Fatal(err)
			}
			output, err := Generate(p)
			if err == nil || !strings.Contains(err.Error(), "PHP004") || len(output) != 0 {
				t.Fatalf("unsupported dead arm emitted output: %v\n%s", err, output)
			}
		})
	}
}
