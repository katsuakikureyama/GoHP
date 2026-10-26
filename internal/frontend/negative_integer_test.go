package frontend

import "testing"

func TestNegativeIntegerLiteralIR(t *testing.T) {
	for _, tt := range []struct {
		source string
		text   string
		line   int
	}{
		{"<?php $value = -5;", "-5", 1},
		{"<?php $value = -0;", "0", 1},
		{"<?php $value = -9223372036854775807;", "-9223372036854775807", 1},
		{"<?php\n$value = - /* gap\n */ 5;", "-5", 2},
	} {
		t.Run(tt.source, func(t *testing.T) {
			p, err := Parse(tt.source, testTarget(t))
			if err != nil {
				t.Fatal(err)
			}
			e := p.Body[0].Expr
			if e.Kind != "int" || e.Text != tt.text || e.Line != tt.line || len(e.Args) != 0 {
				t.Fatalf("negative literal IR = %#v, want int %s at line %d", e, tt.text, tt.line)
			}
		})
	}
}
