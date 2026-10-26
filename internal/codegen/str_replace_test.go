package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
)

func TestStrReplaceLowersToRuntimeBoundary(t *testing.T) {
	p, err := frontend.Parse("<?php echo StR_RePlAcE('cat', 'dog', 'cat cat');", testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	generated, err := Generate(p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(generated), "rt.StrReplace(") {
		t.Fatalf("generated Go did not use the runtime str_replace boundary:\n%s", generated)
	}
}

func TestStrReplaceKnownUnsupportedFormsFailBeforeGeneration(t *testing.T) {
	for _, tt := range []struct {
		name string
		src  string
	}{
		{name: "missing_argument", src: "<?php echo str_replace('x', 'y');"},
		{name: "extra_argument", src: "<?php echo str_replace('x', 'y', 'z', 'count');"},
		{name: "integer_search", src: "<?php echo str_replace(1, 'x', 'subject');"},
		{name: "integer_replacement", src: "<?php echo str_replace('s', 1, 'subject');"},
		{name: "integer_subject", src: "<?php echo str_replace('s', 'x', 1);"},
		{name: "boolean_search", src: "<?php echo str_replace(true, 'x', 'subject');"},
		{name: "null_replacement", src: "<?php echo str_replace('s', null, 'subject');"},
		{name: "array_subject", src: "<?php echo str_replace('s', 'x', []);"},
		{name: "redeclare", src: "<?php function StR_RePlAcE($a, $b, $c) { return $a; }"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p, err := frontend.Parse(tt.src, testTarget(t))
			if err == nil {
				_, err = Generate(p)
			}
			if err == nil || !strings.Contains(err.Error(), "PHP00") {
				t.Fatalf("unsupported str_replace source was not rejected with a diagnostic: %v", err)
			}
		})
	}
}
