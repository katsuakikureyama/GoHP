package codegen_test

import (
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/codegen"
	"github.com/katsuakikureyama/GoHP/internal/frontend"
)

func TestPHP84DeclarationProfile(t *testing.T) {
	for _, target := range []string{"8.3", "8.4"} {
		for _, source := range []string{
			"<?php function array_find($x) { return $x; }",
			"<?php function ArRaY_AnY($x) { return $x; }",
			"<?php function array_all($x) { return $x; }",
			"<?php function array_find_key($x) { return $x; }",
			"<?php class _ {}",
		} {
			program, err := frontend.Parse(source, target)
			if err != nil {
				t.Fatal(err)
			}
			_, err = codegen.Generate(program)
			if (err != nil) != (target == "8.4") {
				t.Fatalf("target %s source %s: %v", target, source, err)
			}
		}
	}
}
