package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestLiteralParamDefaultsLowering(t *testing.T) {
	const source = `<?php
function greet($name, $times = 1, $prefix = 'Hello') {
  echo $prefix . $name;
}
function flag($value = false) {
  if ($value === false) { echo 'off'; } else { echo 'on'; }
}
function arr($items = []) {
  $items[] = 1;
  echo count($items);
}
class C {
  public function m($x = 10) { echo $x; }
  public function __construct($a = 'a', $b = null) {
    echo $a;
  }
  public static function s($y = true) { echo $y ? 'yes' : 'no'; }
}
greet('world');
greet('world', 2, 'Hi');
flag();
flag(true);
arr();
arr([1, 2]);
$c = new C();
$c->m();
C::s();
`
	p, err := frontend.Parse(source, testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	out, _, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	text := string(out)
	// Check that argsLen is captured and defaults are applied
	if !strings.Contains(text, "argsLen := len(args)") {
		t.Fatal("missing argsLen capture")
	}
	// Check int default
	if !strings.Contains(text, `v["times"] = int64(1)`) {
		t.Fatal("missing int default assignment")
	}
	// Check string default
	if !strings.Contains(text, `v["prefix"] = "Hello"`) {
		t.Fatal("missing string default assignment")
	}
	// Check bool default
	if !strings.Contains(text, `v["value"] = false`) {
		t.Fatal("missing bool default assignment")
	}
	// Check empty array default
	if !strings.Contains(text, `v["items"] = rt.NewArray()`) {
		t.Fatal("missing empty array default assignment")
	}
	// Check constructor default
	if !strings.Contains(text, `v["a"] = "a"`) {
		t.Fatal("missing constructor string default")
	}
	// Check static method default
	if !strings.Contains(text, `v["y"] = true`) {
		t.Fatal("missing static method bool default")
	}
}

func TestLiteralParamDefaultsBoundaries(t *testing.T) {
	for name, source := range map[string]string{
		"negative_int":        `function f($x = -1) {}`,
		"float":               `function f($x = 1.5) {}`,
		"nonempty_array":      `function f($x = [1]) {}`,
		"typed_param_default": `function f(array $x = null) {}`,
		"required_after":      `function f($x = null, $y) {}`,
	} {
		t.Run(name, func(t *testing.T) {
			p, err := frontend.Parse("<?php "+source, testTarget(t))
			var out []byte
			if err == nil {
				out, err = Generate(p)
			}
			if err == nil || len(out) != 0 {
				t.Fatalf("unsupported source generated output %s: %v", out, err)
			}
		})
	}
}

func TestLiteralParamDefaultsMalformedIR(t *testing.T) {
	for name, fn := range map[string]ir.Function{
		"nonempty_array_default": {
			Name: "f", Params: []string{"x"}, OptionalNullParams: 1, Line: 1,
			ParamDefaults: []ir.Expr{{Kind: "array", Line: 1, Args: []ir.Expr{{Kind: "pair", Line: 1}}}},
		},
		"unsupported_kind": {
			Name: "f", Params: []string{"x"}, OptionalNullParams: 1, Line: 1,
			ParamDefaults: []ir.Expr{{Kind: "call", Text: "foo", Line: 1}},
		},
	} {
		t.Run(name, func(t *testing.T) {
			p := ir.Program{Version: ir.Version, PHPVersion: testTarget(t), Functions: []ir.Function{fn}}
			out, err := Generate(p)
			if err == nil || len(out) != 0 {
				t.Fatalf("malformed IR generated output %s: %v", out, err)
			}
		})
	}
}
