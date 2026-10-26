package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
)

func TestElseifNormalizationAndSourceMap(t *testing.T) {
	const php = "<?php\nif ('') {\n echo 'first';\n} elseif ('0') {\n echo 'second';\n} else\nif ('ok') {\n echo 'third';\n} else {\n echo 'fallback';\n}\n"
	p, err := frontend.Parse(php, testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	chain := p.Body[0]
	for _, want := range []int{4, 7} {
		if len(chain.ElseBody) != 1 || chain.ElseBody[0].Kind != "if" || chain.ElseBody[0].Line != want {
			t.Fatalf("expected nested if at PHP line %d, got %#v", want, chain.ElseBody)
		}
		chain = chain.ElseBody[0]
	}
	generated, sourceMap, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	for fragment, want := range map[string]int{
		`if rt.Bool("")`: 2, `rt.Echo("first")`: 3,
		`if rt.Bool("0")`: 4, `rt.Echo("second")`: 5,
		`if rt.Bool("ok")`: 7, `rt.Echo("third")`: 8, `rt.Echo("fallback")`: 10,
	} {
		found := false
		for line, text := range strings.Split(string(generated), "\n") {
			if strings.Contains(text, fragment) {
				found = true
				if got, ok := sourceMap.PHPLine(line + 1); !ok || got != want {
					t.Errorf("%s maps to %d (%v), want %d", fragment, got, ok, want)
				}
			}
		}
		if !found {
			t.Errorf("missing generated statement %s", fragment)
		}
	}
}
