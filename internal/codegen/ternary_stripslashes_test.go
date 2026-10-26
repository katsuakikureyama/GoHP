package codegen

import (
	"strconv"
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestTernaryAndStripSlashesLowerToRuntimeBoundaries(t *testing.T) {
	p, err := frontend.Parse("<?php echo true ? StRiPsLaShEs('a\\\\qb') : 'unused';", testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	generated, err := Generate(p)
	if err != nil {
		t.Fatal(err)
	}
	text := string(generated)
	for _, want := range []string{"func() any", "rt.Bool(", "rt.StripSlashes("} {
		if !strings.Contains(text, want) {
			t.Fatalf("generated Go did not preserve ternary lowering boundary %q:\n%s", want, generated)
		}
	}
}

func TestTernaryAndStripSlashesKnownUnsupportedFormsFailBeforeGeneration(t *testing.T) {
	for _, tt := range []struct {
		name string
		src  string
	}{
		{name: "short_ternary", src: "<?php echo true ?: 'no';"},
		{name: "null_coalesce", src: "<?php echo true ?? 'no';"},
		{name: "quoted_question_is_not_separator", src: "<?php echo true '?' 1 : 2;"},
		{name: "quoted_colon_is_not_separator", src: "<?php echo true ? 1 ':' 2;"},
		{name: "unparenthesized_ternary_chain", src: "<?php echo true ? 'one' : false ? 'two' : 'three';"},
		{name: "object_condition", src: "<?php class X {} $x = new X(); echo $x ? 'yes' : 'no';"},
		{name: "unselected_arm_is_still_validated", src: "<?php echo true ? 'safe' : stripslashes(1);"},
		{name: "matching_array_arms_keep_static_type", src: "<?php echo stripslashes(true ? [] : []);"},
		{name: "stripslashes_missing_argument", src: "<?php echo stripslashes();"},
		{name: "stripslashes_extra_argument", src: "<?php echo stripslashes('x', 'y');"},
		{name: "stripslashes_static_non_string", src: "<?php echo stripslashes(1);"},
		{name: "stripslashes_redeclare", src: "<?php function StRiPsLaShEs($value) { return $value; }"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p, err := frontend.Parse(tt.src, testTarget(t))
			if err == nil {
				_, err = Generate(p)
			}
			if err == nil || !strings.Contains(err.Error(), "PHP00") {
				t.Fatalf("unsupported ternary/stripslashes source was not rejected with a diagnostic: %v", err)
			}
		})
	}
}

func TestTernaryMalformedIRArityFailsBeforeGeneration(t *testing.T) {
	for _, operands := range []int{0, 2, 4} {
		t.Run("arity_"+strconv.Itoa(operands), func(t *testing.T) {
			p, err := frontend.Parse("<?php echo true ? 'yes' : 'no';", testTarget(t))
			if err != nil {
				t.Fatal(err)
			}
			original := p.Body[0].Expr.Args
			if operands == 4 {
				p.Body[0].Expr.Args = append(append([]ir.Expr{}, original...), original[0])
			} else {
				p.Body[0].Expr.Args = original[:operands]
			}
			_, err = Generate(p)
			if err == nil || !strings.Contains(err.Error(), "IR002") {
				t.Fatalf("ternary malformed IR arity %d was not rejected safely: %v", operands, err)
			}
		})
	}
}

func TestTernaryMixedArmsRemainConservative(t *testing.T) {
	p, err := frontend.Parse("<?php $value = true ? 'string' : []; foreach ($value as $item) { echo $item; }", testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Generate(p); err != nil {
		t.Fatalf("mixed ternary arms should retain unknown rather than an optimistic branch type: %v", err)
	}
}
