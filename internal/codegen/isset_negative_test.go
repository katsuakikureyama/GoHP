package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
)

func TestIssetNegativeLiteralLowering(t *testing.T) {
	for _, tt := range []struct{ key, canonical string }{
		{"-1", "-1"}, {"-0", "0"}, {"-9223372036854775807", "-9223372036854775807"},
		{"- /* gap\n */ 7", "-7"},
	} {
		t.Run(tt.key, func(t *testing.T) {
			p, err := frontend.Parse("<?php\n$items = [];\necho IsSeT($items["+tt.key+"]);", "8.3")
			if err != nil {
				t.Fatal(err)
			}
			key := p.Body[1].Expr.Args[1]
			if key.Kind != "int" || key.Text != tt.canonical || key.Line != 3 || len(key.Args) != 0 {
				t.Fatalf("unexpected literal key IR: %#v", key)
			}
			output, sourceMap, err := GenerateWithMap(p)
			if err != nil {
				t.Fatal(err)
			}
			needle := `rt.ArrayIsset(v["items"], int64(` + tt.canonical + `))`
			for line, text := range strings.Split(string(output), "\n") {
				if strings.Contains(text, needle) {
					if got, ok := sourceMap.PHPLine(line + 1); !ok || got != 3 {
						t.Fatalf("isset maps to line %d (%v), want 3", got, ok)
					}
					return
				}
			}
			t.Fatalf("missing direct presence probe %q:\n%s", needle, output)
		})
	}
}

func TestIssetNegativeLiteralBoundary(t *testing.T) {
	for _, key := range []string{
		"-$key", "-marker()", "-(1)", "(-1)", "-'1'", "--1", "- -1", "-true",
		"-1 + 0", "-1[0]", "-1.0", "-0x1", "-01", "-1_000", "-9223372036854775808", "-",
	} {
		t.Run(key, func(t *testing.T) {
			source := "<?php function marker() { return 1; } $key = 1; $items = []; echo isset($items[" + key + "]);"
			p, err := frontend.Parse(source, "8.3")
			if err == nil {
				var output []byte
				output, err = Generate(p)
				if len(output) != 0 {
					t.Fatalf("unsupported key produced output: %s", key)
				}
			}
			if err == nil || !strings.Contains(err.Error(), "PHP00") {
				t.Fatalf("unsupported key accepted: %s: %v", key, err)
			}
		})
	}
}
