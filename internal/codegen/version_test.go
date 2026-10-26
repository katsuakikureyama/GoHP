package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestDisabledTargetsFailBeforeGeneration(t *testing.T) {
	for _, target := range []string{"7.4", "8.2", "8.5", "8.3.35"} {
		t.Run(target, func(t *testing.T) {
			if _, err := frontend.Parse("<?php echo 1;", target); err == nil || !strings.Contains(err.Error(), "PHP003") {
				t.Fatalf("frontend accepted disabled target %s: %v", target, err)
			}
			generated, err := Generate(ir.Program{Version: ir.Version, PHPVersion: target})
			if err == nil || !strings.Contains(err.Error(), "PHP003") || len(generated) != 0 {
				t.Fatalf("IR bypassed target policy: generated=%q err=%v", generated, err)
			}
		})
	}
}

func TestDirectIRPreservesEnabledPHP84Target(t *testing.T) {
	generated, err := Generate(ir.Program{Version: ir.Version, PHPVersion: "8.4"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(generated), "PHP target 8.4") {
		t.Fatalf("generated source lost PHP target: %s", generated)
	}
}
