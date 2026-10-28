package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestStringSwitchLowering(t *testing.T) {
	p, err := frontend.Parse("<?php\nfunction f($x){\nSwItCh($x){\ncase 'a': return 'one'; echo 'dead';\ncase 'a': break;\ndefault: echo 'tail';\n}\n}", testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	if s := p.Functions[0].Body[0]; s.Kind != "switch" || len(s.Body) != 3 || s.Body[0].Kind != "case" || s.Body[2].Kind != "default" {
		t.Fatalf("IR: %+v", s)
	}
	out, mapping, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	text := string(out)
	if strings.Count(text, `case "a":`) != 1 || strings.Contains(text, `rt.Echo("dead")`) || strings.Contains(text, "fallthrough") || strings.Contains(text, "NewScalarFrame") || strings.Contains(text, "rt.ValueSlot") {
		t.Fatalf("wrong switch lowering: %s", out)
	}
	for i, line := range strings.Split(text, "\n") {
		if strings.Contains(line, "rt.SwitchString(") {
			if got, ok := mapping.PHPLine(i + 1); !ok || got != 3 {
				t.Fatalf("map %d %v", got, ok)
			}
		}
	}
}

func TestStringSwitchUnsupportedSource(t *testing.T) {
	for _, source := range []string{
		`switch(1){default:break;}`, `switch(null){}`, `switch('1'){}`, `switch('x'){case '1':break;}`,
		`switch('x'){case 1:break;}`, `switch('x'){case strtolower('X'):break;}`, `switch('x'):endswitch;`,
		`switch('x'){default:default:}`, `switch('x'){case 'x';break;}`, `switch('x'){echo 1;}`,
		`switch('x'){case 'x':continue;}`, `while(false){switch('x'){default:if(true){continue;}}}`,
		`switch('x'){default:break 1;}`, `switch('x'){default:continue 2;}`, `case 'x':echo 1;`,
		`switch('x'){default:$new=1;} echo $new;`,
		`class C{public $x;} $c=new C(); switch('x'){case 'x':$c=1; case 'y':$c->x=2;}`,
		`class C{public $x;} $c=new C(); switch('x'){default:$c=1;} $c->x=2;`,
		`function SwItCh(){}`, `function CaSe(){}`, `function DeFaUlT(){}`,
		`switch('x'){default:break; echo unknown();}`,
	} {
		p, err := frontend.Parse("<?php "+source, testTarget(t))
		var out []byte
		if err == nil {
			out, err = Generate(p)
		}
		if err == nil || len(out) != 0 {
			t.Fatalf("accepted %s: %s %v", source, out, err)
		}
	}
}

func TestStringSwitchMalformedIR(t *testing.T) {
	for _, mutate := range []func(*ir.Stmt){
		func(s *ir.Stmt) { s.Name = "hidden" }, func(s *ir.Stmt) { s.Key = "hidden" }, func(s *ir.Stmt) { s.Target = ir.Expr{Kind: "var", Text: "x"} }, func(s *ir.Stmt) { s.ElseBody = []ir.Stmt{} },
		func(s *ir.Stmt) { s.Expr = ir.Expr{} }, func(s *ir.Stmt) { s.Body[0].Kind = "echo" }, func(s *ir.Stmt) { s.Body[0].Expr.Args = []ir.Expr{} },
		func(s *ir.Stmt) { s.Body[0].Expr.Params = []string{} }, func(s *ir.Stmt) { s.Body[0].Name = "hidden" }, func(s *ir.Stmt) { s.Body[0].ElseBody = []ir.Stmt{} },
		func(s *ir.Stmt) { s.Body = append(s.Body, ir.Stmt{Kind: "default"}, ir.Stmt{Kind: "default"}) },
	} {
		p, err := frontend.Parse(`<?php switch('x'){case 'x':break;}`, testTarget(t))
		if err != nil {
			t.Fatal(err)
		}
		mutate(&p.Body[0])
		if out, err := Generate(p); err == nil || len(out) != 0 {
			t.Fatalf("malformed IR generated %s %v", out, err)
		}
	}
}

func TestStringSwitchInvalidatesPregNull(t *testing.T) {
	p := preg83Program(t, `<?php $matches=null; switch('x'){case 'x':$matches=null;} $r=preg_match('/a/s','a',$matches,PREG_OFFSET_CAPTURE,0);`)
	if out, err := Generate(p); err == nil || len(out) != 0 {
		t.Fatalf("stale direct-null fact: %s %v", out, err)
	}
}
