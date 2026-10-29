package codegen

import (
	"reflect"
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestNamedClassParameterLowering(t *testing.T) {
	p, err := frontend.Parse(`<?php
function accept(BoX $box, int $number, $tail=null) {
    $box->value = $number;
    return (array) $box;
}
function rebind(Box $box) { $box='changed'; return strlen($box); }
class Holder {
    public $item;
    public function __construct(Box $box) { $this->item=$box; }
    public function take(bOX $box) { $box->value='method'; return $box; }
}
class Box { public $value=0; }
`, testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(p.Functions[0].ParamTypes, []string{"class:box", "int", ""}) || p.Functions[0].OptionalNullParams != 1 {
		t.Fatalf("class parameter metadata: %#v", p.Functions[0])
	}
	if p.Classes[0].Methods[1].ParamTypes[0] != "class:box" {
		t.Fatal("method annotation must be canonical")
	}
	out, mapping, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	code := string(out)
	for _, want := range []string{"args[0].(*C_box)", "!ok || object == nil", `panic("class parameter requires box")`, `v["box"] = object`, "rt.NullDefaultArgs(args, 2, 3)", "rt.ObjectToArray("} {
		if !strings.Contains(code, want) {
			t.Fatalf("missing %q: %s", want, code)
		}
	}
	if strings.Contains(code, "NewScalarFrame") || strings.Contains(code, "rt.ValueSlot") {
		t.Fatal("class parameter scopes must retain legacy storage")
	}
	if strings.Index(code, "args[0].(*C_box)") > strings.Index(code, `.Set("value"`) {
		t.Fatal("class entry guard must precede the body")
	}
	for i, line := range strings.Split(code, "\n") {
		if strings.Contains(line, `.Set("value", v["number"])`) {
			if got, ok := mapping.PHPLine(i + 1); !ok || got != 3 {
				t.Fatalf("body source map: %d %v", got, ok)
			}
		}
	}
}

func TestNamedClassParameterUnsupportedSource(t *testing.T) {
	for _, source := range []string{
		`function f(Missing $x){}`,
		`class Box{} function f(Box $x){} f(null);`,
		`class Box{} class Other{} function f(Box $x){} f(new Other());`,
		`class Box{} function f(Box $x){} if(false){f('bad');}`,
		`class Box{} class Holder{public function f(Box $x){}} $h=new Holder();$h->f([]);`,
		`class Box{} class Other{} class Holder{public function f(Box $x){}} $h=new Holder();$h->f(new Other());`,
		`class Box{} class Holder{public function __construct(Box $x){}} $h=new Holder(false);`,
		`class Box{} class Other{} class Holder{public function __construct(Box $x){}} $h=new Holder(new Other());`,
		`class Box{public $value;} class Other{} function f(Box $x){if(true){$x=new Other();}$x->value=1;}`,
		`class Box{} function f(Box $x=null){}`,
		`class Box{} function f(?Box $x){}`,
		`class Box{} function f(Box|string $x){}`,
		`class Box{} function f(Box &$x){}`,
		`class Box{} function f(Box ...$x){}`,
		`class Box{} function f(self $x){}`,
		`class Box{} function f(parent $x){}`,
		`class Box{} function f(static $x){}`,
		`function f(object $x){}`,
		`class Box{} class Holder{public function f(array $x){}}`,
		`class Box{} class Holder{public function f(int $x){}}`,
		`class Box{} function f(Box $x): Box{return $x;}`,
		`class Box{} $f=fn(Box $x)=>$x;`,
		`class Box{} function f(Box $x){$x->missing=1;}`,
	} {
		p, err := frontend.Parse("<?php "+source, testTarget(t))
		var out []byte
		if err == nil {
			out, err = Generate(p)
		}
		if err == nil || len(out) != 0 {
			t.Fatalf("unsupported source generated %s: %v (%s)", out, err, source)
		}
	}
}

func TestNamedClassParameterMalformedIR(t *testing.T) {
	for _, annotation := range []string{"Box", "class:", "class:Box", "class:bad-name", `class:foo\bar`, "class:?box", "class:box|other", "class:self", "class:int", "class:missing"} {
		p := ir.Program{Version: ir.Version, PHPVersion: testTarget(t), Classes: []ir.Class{{Name: "Box"}}, Functions: []ir.Function{{Name: "accept", Params: []string{"value"}, ParamTypes: []string{annotation}, Line: 1}}}
		if out, err := Generate(p); err == nil || len(out) != 0 {
			t.Fatalf("malformed annotation %q generated %s: %v", annotation, out, err)
		}
	}
	p := ir.Program{Version: ir.Version, PHPVersion: testTarget(t), Classes: []ir.Class{{Name: "Box", Methods: []ir.Function{{Name: "accept", Params: []string{"value"}, ParamTypes: []string{"class:box"}, OptionalNullParams: 1, Line: 1}}}}}
	if out, err := Generate(p); err == nil || len(out) != 0 {
		t.Fatalf("class-typed null default generated %s: %v", out, err)
	}
}
