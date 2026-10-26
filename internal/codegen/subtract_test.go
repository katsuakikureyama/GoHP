package codegen

import (
	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"strings"
	"testing"
)

func TestIntegerSubtractionLowering(t *testing.T) {
	p, err := frontend.Parse("<?php\n$result = 8 - 3 - 2;\n$other = 8 + 3 - 2;", "8.3")
	if err != nil {
		t.Fatal(err)
	}
	if p.Body[0].Expr.Text != "-" || p.Body[0].Expr.Args[0].Text != "-" || p.Body[1].Expr.Args[0].Text != "+" {
		t.Fatalf("wrong precedence/association: %#v", p.Body)
	}
	out, m, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for line, text := range strings.Split(string(out), "\n") {
		if strings.Contains(text, "rt.Subtract(rt.Subtract(int64(8), int64(3)), int64(2))") {
			found = true
			if got, ok := m.PHPLine(line + 1); !ok || got != 2 {
				t.Fatalf("map=%d,%v", got, ok)
			}
		}
	}
	if !found {
		t.Fatalf("missing left-associated lowering:\n%s", out)
	}
}

func TestIntegerSubtractionBoundary(t *testing.T) {
	for _, source := range []string{"$r='1'-0;", "$r=1-null;", "$r=[]-1;", "$r=true-1;", "$a=3;$r=$a--;", "$r=3--1;", "$a=1;$r=-$a;"} {
		p, err := frontend.Parse("<?php "+source, "8.3")
		var out []byte
		if err == nil {
			out, err = Generate(p)
		}
		if err == nil || len(out) != 0 {
			t.Fatalf("unsupported source %s generated %s: %v", source, out, err)
		}
	}
}
