package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestPreg83PropertyInputsLowering(t *testing.T) {
	p := preg83Program(t, `<?php
class Source {
    public $document = 'za';
    public $offset = 1;
    public function scan() {
        $matches = null;
        $result = preg_match('/a/s', $this->document, $matches, PREG_OFFSET_CAPTURE, $this->offset);
        return [$result, $matches, preg_last_error()];
    }
}
function nested($holder) {
    $matches = null;
    $result = preg_match('/a/s', $holder->source->document, $matches, PREG_OFFSET_CAPTURE, $holder->source->offset);
    return [$result, $matches];
}`)
	stmt := p.Functions[0].Body[1]
	for i, leaf := range map[int]string{1: "document", 2: "offset"} {
		input := stmt.Expr.Args[i]
		if input.Kind != "property" || input.Text != leaf || input.Args[0].Text != "source" || input.Args[0].Args[0].Text != "holder" {
			t.Fatalf("named input chain not retained: %#v", input)
		}
	}
	out, mapping, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	code := string(out)
	if strings.Contains(code, "NewScalarFrame") || strings.Contains(code, "rt.ValueSlot") {
		t.Fatal("regex captures must retain legacy storage")
	}
	for _, leaf := range []string{"document", "offset"} {
		if strings.Count(code, `.Get("`+leaf+`")`) != 2 {
			t.Fatalf("each named input must be evaluated once per regex statement: %s", code)
		}
	}
	var subject, offset int
	for i, line := range strings.Split(code, "\n") {
		if strings.Contains(line, "Subject :=") {
			subject = i
		}
		if strings.Contains(line, "Offset :=") {
			offset = i
			if subject >= offset {
				t.Fatal("subject snapshot must precede offset")
			}
		}
		if strings.Contains(line, "PregMatch83Local(") {
			if offset >= i || !strings.Contains(line, "Subject,") || !strings.Contains(line, "Offset)") {
				t.Fatal("runtime call must follow both value snapshots")
			}
			if got, ok := mapping.PHPLine(i + 1); !ok || (got != 7 && got != 13) {
				t.Fatalf("regex source map = %d, %v", got, ok)
			}
		}
	}
}

func TestPreg83PropertyInputsUnsupportedSource(t *testing.T) {
	for label, input := range map[string]string{
		"undefined_root":   "$missing->document",
		"unknown_property": "$source->missing",
		"method_call":      "$source->read()",
		"method_prefix":    "$source->read()->document",
		"index_prefix":     "$source[0]->document",
		"indexed_property": "$source->document[0]",
		"dynamic_name":     "$source->$name",
		"arithmetic":       "$source->offset + 1",
		"global_root":      "$GLOBALS->document",
		"superglobal_root": "$_GET->document",
	} {
		t.Run(label, func(t *testing.T) {
			source := `<?php class Source { public $document='a'; public $offset=0; } $source=new Source(); $matches=null; $r=preg_match('/a/s', ` + input + `, $matches, PREG_OFFSET_CAPTURE, $source->offset);`
			p, err := frontend.ParseWithOptions(source, "8.3", frontend.Options{ExperimentalPCRE2: true})
			var out []byte
			if err == nil {
				out, err = Generate(p)
			}
			if err == nil || len(out) != 0 {
				t.Fatalf("unsupported input generated %s: %v", out, err)
			}
		})
	}
	for _, source := range []string{
		`$root=1; $matches=null; $r=preg_match('/a/s',$root->text,$matches,PREG_OFFSET_CAPTURE,0);`,
		`class C{public $text='a';} $r=new C(); $matches=null; $r=preg_match('/a/s',$r->text,$matches,PREG_OFFSET_CAPTURE,0);`,
		`class C{public $offset=0;} $r=new C(); $matches=null; $r=preg_match('/a/s','a',$matches,PREG_OFFSET_CAPTURE,$r->offset);`,
		`$matches=null; $r=preg_match('/a/s',$matches->text,$matches,PREG_OFFSET_CAPTURE,0);`,
		`function scan($text){$matches=null;$r=preg_match('/a/s',$text,$matches,PREG_OFFSET_CAPTURE,0);}`,
		`function scan($offset){$matches=null;$r=preg_match('/a/s','a',$matches,PREG_OFFSET_CAPTURE,$offset);}`,
	} {
		p, err := frontend.ParseWithOptions("<?php "+source, "8.3", frontend.Options{ExperimentalPCRE2: true})
		var out []byte
		if err == nil {
			out, err = Generate(p)
		}
		if err == nil || len(out) != 0 {
			t.Fatalf("alias/domain input generated %s: %v", out, err)
		}
	}
}

func TestPreg83PropertyInputsMalformedIR(t *testing.T) {
	for label, mutate := range map[string]func(*ir.Expr){
		"missing_receiver":  func(e *ir.Expr) { e.Args = nil },
		"extra_receiver":    func(e *ir.Expr) { e.Args = append(e.Args, e.Args[0]) },
		"invalid_name":      func(e *ir.Expr) { e.Text = "bad-name" },
		"leaf_parameters":   func(e *ir.Expr) { e.Params = []string{} },
		"prefix_parameters": func(e *ir.Expr) { e.Args[0].Params = []string{} },
		"root_operands":     func(e *ir.Expr) { e.Args[0].Args[0].Args = []ir.Expr{{Kind: "int", Text: "0"}} },
		"root_parameters":   func(e *ir.Expr) { e.Args[0].Args[0].Params = []string{} },
		"call_root":         func(e *ir.Expr) { e.Args[0].Args[0] = ir.Expr{Kind: "call", Text: "source"} },
		"literal_root":      func(e *ir.Expr) { e.Args[0].Args[0] = ir.Expr{Kind: "string", Text: "source"} },
	} {
		t.Run(label, func(t *testing.T) {
			p := preg83Program(t, `<?php function scan($holder){$matches=null;$r=preg_match('/a/s',$holder->source->text,$matches,PREG_OFFSET_CAPTURE,$holder->source->offset);}`)
			mutate(&p.Functions[0].Body[1].Expr.Args[1])
			if out, err := Generate(p); err == nil || len(out) != 0 {
				t.Fatalf("malformed property input generated %s: %v", out, err)
			}
		})
	}
}
