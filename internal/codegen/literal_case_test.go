package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
)

func TestLiteralCaseCanonicalIRAndSourceMap(t *testing.T) {
	for _, spelling := range []string{"true", "TRUE", "TrUe", "false", "FALSE", "FaLsE", "null", "NULL", "NuLl"} {
		t.Run(spelling, func(t *testing.T) {
			p, err := frontend.Parse("<?php\n$value = "+spelling+";", "8.3")
			if err != nil {
				t.Fatal(err)
			}
			e := p.Body[0].Expr
			canonical, kind, goValue := strings.ToLower(spelling), "bool", strings.ToLower(spelling)
			if canonical == "null" {
				kind, goValue = "null", "nil"
			}
			if e.Kind != kind || e.Text != canonical || e.Line != 2 || len(e.Args) != 0 {
				t.Fatalf("unexpected canonical literal: %#v", e)
			}
			output, sourceMap, err := GenerateWithMap(p)
			if err != nil {
				t.Fatal(err)
			}
			for line, text := range strings.Split(string(output), "\n") {
				if strings.Contains(text, `v["value"] = rt.Clone(`+goValue+`)`) {
					if got, ok := sourceMap.PHPLine(line + 1); !ok || got != 2 {
						t.Fatalf("literal source line = %d (%v), want 2", got, ok)
					}
					return
				}
			}
			t.Fatalf("missing canonical literal lowering:\n%s", output)
		})
	}
}

func TestLiteralCaseDoesNotBypassDomains(t *testing.T) {
	for _, source := range []string{
		"<?php echo in_array('x', ['x'], FALSE);",
		"<?php echo in_array('x', ['x'], NULL);",
		"<?php $strict = TRUE; echo in_array('x', ['x'], $strict);",
		"<?php $items = [TRUE => 1];",
		"<?php $items = []; echo isset($items[FALSE]);",
		"<?php function result(): array { return NULL; }",
		"<?php echo strlen(TRUE);",
		"<?php echo TRUE();",
		"<?php echo NULL();",
		"<?php echo TRUE_VALUE;",
		"<?php IF (TRUE) { echo 1; }",
		"<?php $TRUE = 1; echo $true;",
	} {
		t.Run(source, func(t *testing.T) {
			p, err := frontend.Parse(source, "8.3")
			if err == nil {
				var output []byte
				output, err = Generate(p)
				if len(output) != 0 {
					t.Fatal("invalid source emitted output")
				}
			}
			if err == nil || !strings.Contains(err.Error(), "PHP00") {
				t.Fatalf("outside-subset source accepted: %v", err)
			}
		})
	}
}
