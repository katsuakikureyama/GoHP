package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
)

func TestSprintfZeroPadLowersToRuntimeBoundary(t *testing.T) {
	p, err := frontend.Parse("<?php function zeroise($number, $threshold) { return sprintf('%0' . $threshold . 's', $number); } echo zeroise(42, 5);", testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	generated, err := Generate(p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(generated), "rt.SprintfZeroPad(") {
		t.Fatalf("generated Go did not use the runtime sprintf boundary:\n%s", generated)
	}
}

func TestSprintfZeroPadKnownUnsupportedFormsFailBeforeGeneration(t *testing.T) {
	for _, tt := range []struct {
		name string
		src  string
	}{
		{name: "missing_argument", src: "<?php echo sprintf('%05s');"},
		{name: "extra_argument", src: "<?php echo sprintf('%05s', 'x', 'y');"},
		{name: "non_string_format", src: "<?php echo sprintf(1, 'x');"},
		{name: "unsupported_directive", src: "<?php echo sprintf('%05d', 'x');"},
		{name: "noncanonical_width", src: "<?php echo sprintf('%005s', 'x');"},
		{name: "width_overflow", src: "<?php echo sprintf('%04097s', 'x');"},
		{name: "bool_value", src: "<?php echo sprintf('%05s', true);"},
		{name: "array_value", src: "<?php echo sprintf('%05s', []);"},
		{name: "redeclare", src: "<?php function SpRiNtF() { return 'x'; }"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p, err := frontend.Parse(tt.src, testTarget(t))
			if err == nil {
				_, err = Generate(p)
			}
			if err == nil || !strings.Contains(err.Error(), "PHP00") {
				t.Fatalf("unsupported sprintf source was not rejected with a diagnostic: %v", err)
			}
		})
	}
}
