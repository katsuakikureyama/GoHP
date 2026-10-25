package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestContinueForeachLowering(t *testing.T) {
	const source = `<?php
function collect($groups) {
  $out = '';
  foreach ($groups as $group) {
    foreach ($group as $value) {
      if ($value === 'skip') { CoNtInUe; }
      $out = $out . $value;
    }
    $out = $out . '|';
  }
  return $out;
}
function has_value(array $items): bool {
  foreach ($items as $item) {
    if ($item === 'skip') { continue; }
    return true;
  }
  return false;
}
function retained(array $items): array {
  $out = [];
  foreach ($items as $item) {
    if ($item === 'skip') { continue; }
    $out[] = $item;
  }
  return $out;
}
echo collect([['a', 'skip', 'b'], ['skip', 'c']]);
`
	p, err := frontend.Parse(source, "8.3")
	if err != nil {
		t.Fatal(err)
	}
	stmt := p.Functions[0].Body[1].Body[0].Body[0].Body[0]
	if stmt.Kind != "continue" || stmt.Name != "" || stmt.Expr.Kind != "" || len(stmt.Body) != 0 || stmt.ElseBody != nil {
		t.Fatalf("continue IR = %#v, want a payload-free continue statement", stmt)
	}
	out, mapping, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	text := string(out)
	if got := strings.Count(text, "continue"); got != 3 {
		t.Fatalf("generated continue count = %d, want 3:\n%s", got, out)
	}
	for _, want := range []string{
		`v["items"] = rt.ArrayParameter(args[0])`,
		`return rt.BoolReturn(true)`,
		`return rt.ArrayReturn(v["out"])`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("generated Go missing %q:\n%s", want, out)
		}
	}
	hasValueStart := strings.Index(text, "func F_has_value")
	if hasValueStart < 0 {
		t.Fatalf("missing typed continue function:\n%s", out)
	}
	hasValue := text[hasValueStart:]
	guard := strings.Index(hasValue, `v["items"] = rt.ArrayParameter(args[0])`)
	loop := strings.Index(hasValue, "for _, entry")
	if guard < 0 || loop < 0 || guard > loop {
		t.Fatalf("array parameter guard must precede a continue loop: %s", hasValue)
	}
	for line, generated := range strings.Split(text, "\n") {
		if strings.TrimSpace(generated) == "continue" {
			if got, ok := mapping.PHPLine(line + 1); !ok || got != 6 {
				t.Fatalf("continue maps to PHP line %d, %v; want 6", got, ok)
			}
			return
		}
	}
	t.Fatalf("continue was not source-mapped:\n%s", out)
}

func TestContinueForeachBoundary(t *testing.T) {
	for name, source := range map[string]string{
		"outside_loop":  "$items = []; continue;",
		"numeric_level": "foreach ([] as $item) { continue 2; }",
		"call_form":     "foreach ([] as $item) { continue(); }",
	} {
		t.Run(name, func(t *testing.T) {
			p, err := frontend.Parse("<?php "+source, "8.3")
			if err == nil {
				_, err = Generate(p)
			}
			if err == nil || !strings.Contains(err.Error(), "PHP00") {
				t.Fatalf("unsupported continue source was accepted: %v", err)
			}
		})
	}

	for name, stmt := range map[string]ir.Stmt{
		"name":      {Kind: "continue", Name: "unexpected", Line: 2},
		"expr":      {Kind: "continue", Expr: ir.Expr{Kind: "bool", Text: "true", Line: 2}, Line: 2},
		"target":    {Kind: "continue", Target: ir.Expr{Kind: "var", Text: "value", Line: 2}, Line: 2},
		"body":      {Kind: "continue", Body: []ir.Stmt{{Kind: "echo", Line: 2, Expr: ir.Expr{Kind: "string", Text: "bad", Line: 2}}}, Line: 2},
		"else_body": {Kind: "continue", ElseBody: []ir.Stmt{}, Line: 2},
	} {
		t.Run(name, func(t *testing.T) {
			p := ir.Program{Version: ir.Version, PHPVersion: "8.3", Body: []ir.Stmt{{Kind: "foreach", Line: 1, Expr: ir.Expr{Kind: "array", Line: 1}, Value: "item", Body: []ir.Stmt{stmt}}}}
			out, err := Generate(p)
			if err == nil || len(out) != 0 || !strings.Contains(err.Error(), "IR002") {
				t.Fatalf("malformed continue IR generated output %s: %v", out, err)
			}
		})
	}
}
