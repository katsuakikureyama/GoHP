package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
)

func TestIsArrayLowersToRuntimeBoundary(t *testing.T) {
	p, err := frontend.Parse("<?php function check($value) { return Is_ArRaY($value); } echo check([]);", testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	generated, err := Generate(p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(generated), "rt.IsArray(") {
		t.Fatalf("generated Go did not use the runtime is_array boundary:\n%s", generated)
	}
}
