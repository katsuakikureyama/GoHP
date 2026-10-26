package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
)

func TestLiteralNullIdentityLowering(t *testing.T) {
	const source = `<?php
function dynamic_value($value) { return $value; }
function marker($label, $value) { echo $label; return $value; }
class Box {}
$known = [];
$dynamic = dynamic_value($known);
$object = new Box();
$callable = fn($value) => $value;
$left_equal = null === $known;
$right_equal = $known === null;
$left_not_equal = null !== $dynamic;
$right_not_equal = $dynamic !== null;
$object_equal = null === $object;
$callable_not_equal = $callable !== null;
$once = null === marker('once', $known);
$short_and = false && (null === marker('and', $known));
$short_or = true || (marker('or', $known) !== null);
`
	p, err := frontend.Parse(source, testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	generated, sourceMap, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	text := string(generated)
	for _, want := range []string{
		`rt.IsNull(v["known"])`,
		`(!rt.IsNull(v["dynamic"]))`,
		`rt.IsNull(v["object"])`,
		`(!rt.IsNull(v["callable"]))`,
		`rt.IsNull(F_marker("once", v["known"]))`,
		`rt.Bool(false) && rt.Bool(rt.IsNull(F_marker("and", v["known"])))`,
		`rt.Bool(true) || rt.Bool((!rt.IsNull(F_marker("or", v["known"]))))`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("literal-null identity lowering missing %q:\n%s", want, generated)
		}
	}
	if strings.Contains(text, `rt.Identical(v["known"], nil)`) || strings.Contains(text, `rt.Identical(nil, v["known"])`) {
		t.Fatalf("literal-null array comparisons must not reach Identical:\n%s", generated)
	}
	for line, generatedLine := range strings.Split(text, "\n") {
		if strings.Contains(generatedLine, `v["left_equal"] = rt.Clone(rt.IsNull(v["known"]))`) {
			if got, ok := sourceMap.PHPLine(line + 1); !ok || got != 9 {
				t.Fatalf("literal-null identity maps to PHP line %d, %v; want 9", got, ok)
			}
			return
		}
	}
	t.Fatalf("literal-null identity was not source-mapped:\n%s", generated)
}

func TestLiteralNullIdentityBoundary(t *testing.T) {
	p, err := frontend.Parse("<?php function same($left, $right) { return $left === $right; }", testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	out, err := Generate(p)
	if err != nil || !strings.Contains(string(out), `rt.Identical(v["left"], v["right"])`) {
		t.Fatalf("variable-variable comparison lost the existing runtime guard: %v\n%s", err, out)
	}
	for name, source := range map[string]string{
		"known_arrays":  "$left = []; $right = []; echo $left !== $right;",
		"array_literal": "echo [] === [];",
	} {
		t.Run(name, func(t *testing.T) {
			p, err := frontend.Parse("<?php "+source, testTarget(t))
			if err == nil {
				_, err = Generate(p)
			}
			if err == nil || !strings.Contains(err.Error(), "PHP00") {
				t.Fatalf("general array strict comparison was accepted: %v", err)
			}
		})
	}
}
