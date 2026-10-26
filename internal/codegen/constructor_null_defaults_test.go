package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestConstructorNullDefaults(t *testing.T) {
	const source = `<?php
class Frame {
  public $offset;
  public $leading;
  public function __CONSTRUCT($start, $length, $previous = NULL, $leading = null) {
    $this->offset = $previous ?? $start + $length;
    $this->leading = $leading;
  }
}
$one = new Frame(2, 3);
$two = new Frame(2, 3, null);
$three = new Frame(2, 3, 0, 0);
`
	p, err := frontend.Parse(source, testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	if got := p.Classes[0].Methods[0].OptionalNullParams; got != 2 {
		t.Fatalf("optional count = %d, want 2", got)
	}
	out, mapping, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	text := string(out)
	if !strings.Contains(text, "args = rt.NullDefaultArgs(args, 2, 4)") {
		t.Fatalf("missing runtime min/max guard: %s", out)
	}
	for line, generated := range strings.Split(text, "\n") {
		if strings.Contains(generated, `.Set("offset",`) {
			if got, ok := mapping.PHPLine(line + 1); !ok || got != 6 {
				t.Fatalf("constructor statement mapping = %d, %v; want 6", got, ok)
			}
			return
		}
	}
	t.Fatal("missing constructor statement")
}

func TestConstructorNullDefaultsBoundaries(t *testing.T) {
	for name, source := range map[string]string{
		"below_minimum":           `class C { public function __construct($x, $y = null) {} } $c = new C();`,
		"above_maximum":           `class C { public function __construct($x = null) {} } $c = new C(1, 2);`,
		"required_after_optional": `class C { public function __construct($x = null, $y) {} }`,
		"typed_optional":          `class C { public function __construct(array $x = null) {} }`,
		"nonnull_method":          `class C { public function read($x = false) {} }`,
		"nonpublic":               `class C { private function __construct($x = null) {} }`,
		"direct_reentry":          `class C { public function __construct($x = null) {} } $c = new C(); $c->__construct();`,
	} {
		t.Run(name, func(t *testing.T) {
			p, err := frontend.Parse("<?php "+source, testTarget(t))
			var out []byte
			if err == nil {
				out, err = Generate(p)
			}
			if err == nil || len(out) != 0 {
				t.Fatalf("unsupported constructor source generated output %s: %v", out, err)
			}
		})
	}
}

func TestConstructorNullDefaultsMalformedIR(t *testing.T) {
	for name, count := range map[string]int{"negative": -1, "excessive": 2} {
		t.Run(name, func(t *testing.T) {
			p := ir.Program{Version: ir.Version, PHPVersion: testTarget(t), Classes: []ir.Class{{Name: "C", Methods: []ir.Function{{Name: "__construct", Params: []string{"x"}, OptionalNullParams: count, Line: 1}}}}}
			out, err := Generate(p)
			if err == nil || len(out) != 0 || !strings.Contains(err.Error(), "IR002") {
				t.Fatalf("invalid default count generated output %s: %v", out, err)
			}
		})
	}
}
