package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestNullDefaultParametersLowering(t *testing.T) {
	const source = `<?php
function render(array $items, $prefix = NULL, $suffix = null) {
  if (is_null($prefix)) { $prefix = 'default:'; }
  if (is_null($suffix)) { $suffix = ':end'; }
  return $prefix . count($items) . $suffix;
}
function map_value($value, $suffix = null) {
  if (is_null($suffix)) { return $value . ':mapped'; }
  return $value . $suffix;
}
function forward_render(array $items, $suffix = null) {
  return render($items, null, $suffix);
}
$one = render([1]);
$two = render([1], null);
$three = render([1], 'given:', ':value');
$forwarded = forward_render([1]);
$mapped = array_map('map_value', ['x']);
echo $one . $two . $three . $forwarded . $mapped[0];
`
	p, err := frontend.Parse(source, "8.3")
	if err != nil {
		t.Fatal(err)
	}
	if got := p.Functions[0].OptionalNullParams; got != 2 {
		t.Fatalf("render OptionalNullParams = %d, want 2", got)
	}
	if got := p.Functions[1].OptionalNullParams; got != 1 {
		t.Fatalf("map_value OptionalNullParams = %d, want 1", got)
	}
	out, mapping, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	text := string(out)
	for _, want := range []string{
		"args = rt.NullDefaultArgs(args, 1, 3)",
		"args = rt.NullDefaultArgs(args, 1, 2)",
		`v["items"] = rt.ArrayParameter(args[0])`,
		`F_render(rt.NewArray(rt.Entry{Key: int64(0), Value: int64(1)}))`,
		`F_render(rt.NewArray(rt.Entry{Key: int64(0), Value: int64(1)}), nil)`,
		`rt.ArrayMap(func(value any) any { return F_map_value(value) }, rt.NewArray(rt.Entry{Key: int64(0), Value: "x"}))`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("generated Go missing %q:\n%s", want, out)
		}
	}
	for line, generated := range strings.Split(text, "\n") {
		if strings.Contains(generated, `F_render(rt.NewArray`) {
			if got, ok := mapping.PHPLine(line + 1); !ok || got != 14 {
				t.Fatalf("omitted-default call maps to PHP line %d, %v; want 14", got, ok)
			}
			return
		}
	}
	t.Fatalf("omitted-default call was not source-mapped:\n%s", out)
}

func TestNullDefaultParametersBoundary(t *testing.T) {
	for name, source := range map[string]string{
		"missing_required":        "function f($required, $optional = null) { return $required; } echo f();",
		"too_many":                "function f($optional = null) { return $optional; } echo f(1, 2);",
		"non_null":                "function f($value = false) { return $value; } echo f();",
		"required_after":          "function f($optional = null, $required) { return $required; } echo f(1);",
		"typed_default":           "function f(array $items = null) { return $items; } $result = f();",
		"method_default":          "class C { public function f($value = null) { return $value; } } $c = new C(); echo $c->f();",
		"array_map_below_minimum": "function f($first, $value, $suffix = null) { return $value; } $result = array_map('f', []);",
	} {
		t.Run(name, func(t *testing.T) {
			p, err := frontend.Parse("<?php "+source, "8.3")
			if err == nil {
				_, err = Generate(p)
			}
			if err == nil || !strings.Contains(err.Error(), "PHP00") {
				t.Fatalf("unsupported null-default source was accepted: %v", err)
			}
		})
	}

	for name, function := range map[string]ir.Function{
		"count_exceeds_parameters": {Name: "f", Params: []string{"value"}, OptionalNullParams: 2, Line: 1},
		"typed_optional":           {Name: "f", Params: []string{"value"}, ParamTypes: []string{"array"}, OptionalNullParams: 1, Line: 1},
	} {
		t.Run(name, func(t *testing.T) {
			p := ir.Program{Version: ir.Version, PHPVersion: "8.3", Functions: []ir.Function{function}}
			out, err := Generate(p)
			if err == nil || len(out) != 0 || !strings.Contains(err.Error(), "PHP00") {
				t.Fatalf("malformed null-default IR generated output %s: %v", out, err)
			}
		})
	}
}
