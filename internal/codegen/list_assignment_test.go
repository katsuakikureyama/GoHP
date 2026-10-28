package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestListAssignmentLowering(t *testing.T) {
	p, err := frontend.Parse("<?php\nfunction selected() {\n$source = [1, 2];\nlist($source, $source) = $source;\nreturn $source;\n}", testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	s := p.Functions[0].Body[1]
	if s.Kind != "list_assign" || s.Target.Kind != "list" || len(s.Target.Args) != 2 || s.Expr.Text != "source" {
		t.Fatalf("wrong list IR: %+v", s)
	}
	out, mapping, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	text := string(out)
	if strings.Contains(text, "NewScalarFrame") || strings.Contains(text, "rt.ValueSlot") {
		t.Fatal("list scope must retain legacy storage")
	}
	parts := []string{`list1 := rt.ListValues(v["source"], 2)`, `v["source"] = list1[0]`, `v["source"] = list1[1]`}
	last := -1
	for _, part := range parts {
		pos := strings.Index(text, part)
		if pos < 0 || pos <= last || strings.Count(text, part) != 1 {
			t.Fatalf("missing/out-of-order snapshot publication %q:\n%s", part, out)
		}
		last = pos
	}
	for i, line := range strings.Split(text, "\n") {
		if strings.Contains(line, "list1") {
			if got, ok := mapping.PHPLine(i + 1); !ok || got != 4 {
				t.Fatalf("source map %d %v", got, ok)
			}
		}
	}
	// Case-insensitive construct and stateful RHS calls keep Execution routing.
	p, err = frontend.Parse(`<?php function rhs() { return [strtok('a b',' '), strtok(' ')]; } LiSt($a,$b) = rhs(); echo $a . $b;`, testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	out, err = Generate(p)
	if err != nil || !strings.Contains(string(out), "rt.ListValues(FContext_rhs(execution), 2)") {
		t.Fatalf("execution lost: %s %v", out, err)
	}
}

func TestListAssignmentUnsupportedSource(t *testing.T) {
	for _, source := range []string{
		`list()=[];`, `list(,$a)=[1,2];`, `list($a,)=[1];`, `list($a,,$b)=[1,2,3];`,
		`list(0=>$a)=[1];`, `list(list($a))=[[1]];`, `list(&$a)=[1];`, `[$a]=[1];`,
		`list($a[0])=[1];`, `list($a->field)=[1];`, `list($a)=$a;`, `list($this)=[1];`,
		`list($GLOBALS)=[1];`, `list($_hidden)=[1];`, `list($a)=null;`, `list($a)='x';`,
		`list($a)=1;`, `list($a)=true;`, `class C {} list($a)=new C();`, `echo list($a)=[1];`,
		`function LiSt() { return []; }`, `if(true){list($a)=[1];} echo $a;`,
		`while(false){list($a)=[1];} echo $a;`, `if(false){list($a)=[unknown()];}`,
		`class C {public $x;} $c=new C(); list($c)=[1]; $c->x=2;`,
		`class C {public $x;} $c=new C(); while(false){$c->x=2; list($c)=[1];}`,
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

func TestListAssignmentMalformedIR(t *testing.T) {
	for _, mutate := range []func(*ir.Stmt){
		func(s *ir.Stmt) { s.Name = "hidden" }, func(s *ir.Stmt) { s.Key = "hidden" }, func(s *ir.Stmt) { s.Value = "hidden" },
		func(s *ir.Stmt) { s.Body = []ir.Stmt{} }, func(s *ir.Stmt) { s.ElseBody = []ir.Stmt{} },
		func(s *ir.Stmt) { s.Target.Kind = "array" }, func(s *ir.Stmt) { s.Target.Text = "hidden" },
		func(s *ir.Stmt) { s.Target.Params = []string{} }, func(s *ir.Stmt) { s.Target.Args = nil },
		func(s *ir.Stmt) { s.Target.Args[0].Args = []ir.Expr{} }, func(s *ir.Stmt) { s.Target.Args[0].Params = []string{} },
		func(s *ir.Stmt) { s.Target.Args[0].Kind = "string" }, func(s *ir.Stmt) { s.Target.Args[0].Text = "" },
		func(s *ir.Stmt) { s.Expr = ir.Expr{} },
	} {
		p, err := frontend.Parse(`<?php list($a)=[1];`, testTarget(t))
		if err != nil {
			t.Fatal(err)
		}
		mutate(&p.Body[0])
		if out, err := Generate(p); err == nil || len(out) != 0 {
			t.Fatalf("malformed IR generated: %s %v", out, err)
		}
	}
}

func TestListAssignmentInvalidatesPregNull(t *testing.T) {
	for _, write := range []string{`list($matches)=[null];`, `if(true){list($matches)=[null];}`, `while(false){list($matches)=[null];}`} {
		p := preg83Program(t, `<?php $matches=null; `+write+` $result=preg_match('/a/s','a',$matches,PREG_OFFSET_CAPTURE,0);`)
		if out, err := Generate(p); err == nil || len(out) != 0 {
			t.Fatalf("list retained direct null proof: %s %v", out, err)
		}
	}
	p := preg83Program(t, `<?php $matches=null; list($matches)=[null]; $matches=null; $result=preg_match('/a/s','a',$matches,PREG_OFFSET_CAPTURE,0);`)
	if _, err := Generate(p); err != nil {
		t.Fatal(err)
	}
}
