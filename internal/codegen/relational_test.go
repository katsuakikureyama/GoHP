package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestRelationalLowering(t *testing.T) {
	p, err := frontend.Parse("<?php\n$r = 1 + 2 < 4 === true;\n$s = null > -1;", testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	e := p.Body[0].Expr
	if e.Text != "===" || e.Args[0].Text != "<" || e.Args[0].Args[0].Text != "+" {
		t.Fatalf("wrong precedence: %+v", e)
	}
	out, mapping, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "rt.LessThan(rt.Add(int64(1), int64(2)), int64(4))") || !strings.Contains(string(out), "rt.GreaterThan(nil, int64(-1))") {
		t.Fatalf("missing guarded lowering: %s", out)
	}
	for i, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "rt.GreaterThan(") {
			if got, ok := mapping.PHPLine(i + 1); !ok || got != 3 {
				t.Fatalf("source map %d %v", got, ok)
			}
		}
	}
	// Concatenation and unary operators must still bind above the new level.
	for _, source := range []string{"$r = 'a' . 1 < 2;", "$r = !1 < 2;", "$r = (bool) 1 < 2;"} {
		p, err := frontend.Parse("<?php "+source, testTarget(t))
		if err != nil {
			t.Fatal(err)
		}
		if p.Body[0].Expr.Text != "<" {
			t.Fatalf("wrong precedence: %+v", p.Body[0].Expr)
		}
	}
}

func TestRelationalUnsupportedSource(t *testing.T) {
	for _, source := range []string{
		"$r = 1 < 2 < 3;", "$r = 3 > 2 < 1;", "$r = 1 <= 2;", "$r = 1 >= 2;", "$r = 1 <=> 2;", "$r = 1 <> 2;",
		"$r = true < 1;", "$r = 1 > '0';", "$r = null < [];", "$r = [] > null;", "$r = 1 < $missing;",
		"$r = (1 < 2) < 3;", "$r = 3 > (2 > 1);", "if (false) { $r = 1 < []; }",
	} {
		p, err := frontend.Parse("<?php "+source, testTarget(t))
		var out []byte
		if err == nil {
			out, err = Generate(p)
		}
		if err == nil || len(out) != 0 {
			t.Fatalf("accepted %s: %s, %v", source, out, err)
		}
	}
}

func TestRelationalMalformedIR(t *testing.T) {
	for _, mutate := range []func(*ir.Expr){
		func(e *ir.Expr) { e.Args = e.Args[:1] },
		func(e *ir.Expr) { e.Args = append(e.Args, ir.Expr{Kind: "int", Text: "0", Line: 1}) },
		func(e *ir.Expr) { e.Params = []string{"hidden"} },
		func(e *ir.Expr) { e.Text = "<=" },
	} {
		p, err := frontend.Parse("<?php $r=1<2;", testTarget(t))
		if err != nil {
			t.Fatal(err)
		}
		mutate(&p.Body[0].Expr)
		if out, err := Generate(p); err == nil || len(out) != 0 {
			t.Fatalf("malformed IR generated %s: %v", out, err)
		}
	}
}
