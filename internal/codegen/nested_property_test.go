package codegen

import (
	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
	"strings"
	"testing"
)

func TestNestedPropertyLowering(t *testing.T) {
	p, err := frontend.Parse("<?php\nfunction update($frame) {\n$frame->block->text .= strtok('a,b', ',');\n$frame->block->items[] = null;\n$frame->block->value = [1];\nreturn 1;\n}", testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	out, mapping, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	code := string(out)
	if strings.Index(code, "value := any(execution.Strtok(") > strings.Index(code, "target := rt.ResolveNestedProperty(") {
		t.Fatal("RHS must precede writable-path resolution")
	}
	for _, want := range []string{`rt.ResolveNestedProperty(v["frame"], "block", "text")`, `execution.Strtok(`, `target.Concat(value)`, `target.Append(value)`, `target.Assign(value)`} {
		if !strings.Contains(code, want) {
			t.Fatalf("missing %s in %s", want, out)
		}
	}
	if strings.Contains(code, "NewScalarFrame") || strings.Contains(code, "rt.ValueSlot") {
		t.Fatal("nested updates require legacy storage")
	}
	for i, line := range strings.Split(code, "\n") {
		if strings.Contains(line, `"block", "text"`) || strings.Contains(line, "target.Concat(") {
			if got, ok := mapping.PHPLine(i + 1); !ok || got != 3 {
				t.Fatalf("source map %d %v", got, ok)
			}
		}
	}
}

func TestNestedPropertyUnsupportedSource(t *testing.T) {
	for _, source := range []string{
		`function f($a){$a->text.='x';}`,
		`function f($a){$a->child->text . = 'x';}`,
		`function f($a){$a->child->text += 1;}`, `function f($a){$a->child->text .= 1;}`,
		`function f($a){$a->child->items[0]='x';}`, `function f($a){$a->child[0]->text='x';}`,
		`function f($a){$a->child()->text='x';}`, `function f($a){$a->child->$field='x';}`,
		`$missing->child->text='x';`, `$GLOBALS->child->text='x';`,
		`$_GET->child->text='x';`, `$a=1; $a->child->text='x';`,
		`class C{} $c=new C();$c->missing->text='x';`,
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

func TestNestedPropertyMalformedIR(t *testing.T) {
	for _, mutate := range []func(*ir.Stmt){
		func(s *ir.Stmt) { s.Name = "hidden" }, func(s *ir.Stmt) { s.Key = "hidden" }, func(s *ir.Stmt) { s.Value = "hidden" },
		func(s *ir.Stmt) { s.Body = []ir.Stmt{} }, func(s *ir.Stmt) { s.ElseBody = []ir.Stmt{} },
		func(s *ir.Stmt) { s.Target.Params = []string{} }, func(s *ir.Stmt) { s.Target.Args = append(s.Target.Args, s.Target.Args[0]) },
		func(s *ir.Stmt) { s.Target.Text = "bad-name" }, func(s *ir.Stmt) { s.Target = s.Target.Args[0] },
		func(s *ir.Stmt) { s.Target.Args[0].Args[0] = ir.Expr{Kind: "call", Text: "root"} },
		func(s *ir.Stmt) { s.Kind = "assign_property" },
	} {
		p, err := frontend.Parse(`<?php function f($a){$a->child->text='x';}`, testTarget(t))
		if err != nil {
			t.Fatal(err)
		}
		mutate(&p.Functions[0].Body[0])
		if out, err := Generate(p); err == nil || len(out) != 0 {
			t.Fatalf("malformed target generated %s %v", out, err)
		}
	}
}
