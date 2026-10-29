package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
)

const classParameterDefinitions = `class Item { public $value; }
class Other {}
function accept(Item $item, $extra=null) { $item->value='global'; $item=1; return $item; }
class Host {
public function __construct(iTeM $item) { $item->value='constructor'; }
public function put(ITEM $item, $extra=null) { $item->value='method'; return $item; }
}`

func TestClassParameterLowering(t *testing.T) {
	p, err := frontend.Parse("<?php\n"+classParameterDefinitions, testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	if p.Functions[0].ParamTypes[0] != "class:item" || p.Classes[2].Methods[0].ParamTypes[0] != "class:item" {
		t.Fatalf("class annotation metadata: %+v", p)
	}
	out, mapping, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	text := string(out)
	if strings.Count(text, ".(*C_item)") != 3 || strings.Count(text, "== nil") != 3 || !strings.Contains(text, "class parameter requires item object") {
		t.Fatalf("missing exact nonnull entry guards: %s", out)
	}
	if strings.Contains(text, "NewScalarFrame") || strings.Contains(text, "rt.ValueSlot") {
		t.Fatal("class scopes must remain legacy storage")
	}
	if strings.Index(text, ".(*C_item)") > strings.Index(text, `value := any("global")`) {
		t.Fatal("guard must precede body")
	}
	found := false
	for i, line := range strings.Split(text, "\n") {
		if strings.Contains(line, `value := any("global")`) {
			if got, ok := mapping.PHPLine(i + 1); !ok || got != 4 {
				t.Fatalf("map %d %v", got, ok)
			}
			found = true
		}
	}
	if !found {
		t.Fatal("typed parameter should retain known class for property write")
	}
}

func TestClassParameterUnsupportedSource(t *testing.T) {
	for _, source := range []string{
		`function f(Missing $value){}`,
		`class C { public function f(Missing $value){} }`,
		`class Item {} function f(Item $x=null){}`,
		`class Item {} function f(?Item $x){}`,
		`class Item {} function f(Item|int $x){}`,
		`class Item {} function f(Item &$x){}`,
		`class Item {} function f(\Item $x){}`,
		`class Item {} function f(self $x){}`,
		`class Item {} function f(object $x){}`,
		classParameterDefinitions + ` $x=accept(new Other());`,
		classParameterDefinitions + ` $x=accept(null);`,
		classParameterDefinitions + ` $x=accept([]);`,
		classParameterDefinitions + ` $x=accept(1);`,
		classParameterDefinitions + ` $x=accept(fn()=>true);`,
		classParameterDefinitions + ` $h=new Host(new Other());`,
		classParameterDefinitions + ` $h=new Host(new Item()); $x=$h->put(new Other());`,
		classParameterDefinitions + ` $h=new Host(new Item()); $x=$h->put(null);`,
		`class Item {public $value;} function f(Item $x){$x=1;$x->value=2;}`,
		`class Item {} function f(Item $x){return $x->missing;}`,
	} {
		p, err := frontend.Parse("<?php "+source, testTarget(t))
		var out []byte
		if err == nil {
			out, err = Generate(p)
		}
		if err == nil || len(out) != 0 {
			t.Fatalf("unsupported %s: %s %v", source, out, err)
		}
	}
}

func TestClassParameterMalformedIR(t *testing.T) {
	for _, annotation := range []string{"class:", "class:Item", "class:missing", "class:bad-name", "class:object", "Item"} {
		p, err := frontend.Parse("<?php class Item {} function f(Item $x){}", testTarget(t))
		if err != nil {
			t.Fatal(err)
		}
		p.Functions[0].ParamTypes[0] = annotation
		if out, err := Generate(p); err == nil || len(out) != 0 {
			t.Fatalf("malformed annotation %s: %s %v", annotation, out, err)
		}
	}
}
