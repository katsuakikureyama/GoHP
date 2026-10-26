package codegen

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestOwnedControlFlowSelection(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		owned        bool
	}{
		{
			"if_join_existing_locals",
			`function selected() { $items = []; $flag = true; if ($flag) { $items[] = 1; } else { $items[] = 2; } return $items; }`,
			true,
		},
		{
			"branch_local_does_not_escape",
			`function selected() { $items = []; $flag = true; if ($flag) { $branch = $items; $items[] = 1; } else { $branch = []; $items[] = 2; } return $items; }`,
			true,
		},
		{
			"while_invariant",
			`function selected() { $items = []; $i = 0; while ($i !== 2) { $items[] = $i; $i = $i + 1; } return $items; }`,
			true,
		},
		{
			"nested_while_invariant",
			`function selected() { $items = []; $outer = 0; while ($outer !== 1) { $inner = 0; while ($inner !== 2) { $items[] = $inner; $inner = $inner + 1; } $outer = $outer + 1; } return $items; }`,
			true,
		},
		{
			"if_union_return",
			`function selected() { $items = []; if (true) { $items = 1; } else { $items = []; } return $items; }`,
			true,
		},
		{
			"if_union_is_not_flat",
			`function selected() { $items = []; if (true) { $items = 1; } else { $items = []; } return count($items); }`,
			false,
		},
		{
			"return_dead_tail_does_not_change_join",
			`function selected() { $items = []; if (true) { return $items; $items = 1; } return count($items); }`,
			true,
		},
		{
			"loop_union_entry_bounds_writes",
			`function selected() { $items = true ? [] : 1; $i = 0; while ($i !== 2) { $items = 2; $items = []; $i = $i + 1; } return $items; }`,
			true,
		},
		{
			"loop_type_crossing_then_restore",
			`function selected() { $items = []; while (false) { $items = 1; $items = []; } return $items; }`,
			false,
		},
		{
			"type_crossing_before_continue",
			`function selected() { $items = []; while (false) { $items = 1; continue; $items = []; } return $items; }`,
			false,
		},
		{
			"type_crossing_before_break",
			`function selected() { $items = []; while (false) { $items = 1; break; $items = []; } return $items; }`,
			false,
		},
		{
			"dead_break_type_crossing",
			`function selected() { $items = []; $i = 0; while ($i !== 1) { break; $items = 1; } return $items; }`,
			false,
		},
		{
			"dead_continue_type_crossing",
			`function selected() { $items = []; $i = 0; while ($i !== 1) { continue; $items = 1; } return $items; }`,
			false,
		},
		{
			"nested_loop_type_crossing",
			`function selected() { $items = []; $outer = 0; while ($outer !== 1) { $inner = 0; while ($inner !== 1) { $items = 1; $inner = $inner + 1; } $outer = $outer + 1; } return $items; }`,
			false,
		},
		{
			"literal_false_branch_is_not_optimistic",
			`function selected() { $items = []; if (false) { $items = 1; } return count($items); }`,
			false,
		},
		{
			"literal_false_loop_is_not_optimistic",
			`function selected() { $items = []; while (false) { $items = 1; } return $items; }`,
			false,
		},
		{
			"foreach_stays_legacy",
			`function selected() { $items = []; foreach ([] as $item) { $items[] = $item; } return $items; }`,
			false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, err := frontend.Parse("<?php "+tc.source, testTarget(t))
			if err != nil {
				t.Fatal(err)
			}
			out, err := Generate(p)
			if err != nil {
				t.Fatal(err)
			}
			body := generatedFunction(t, string(out), "F_selected")
			if got := strings.Contains(body, "map[string]*rt.ValueSlot{}"); got != tc.owned {
				t.Fatalf("owned control-flow selection = %v, want %v:\n%s", got, tc.owned, body)
			}
			if !tc.owned && !strings.Contains(body, "v := map[string]any{}") {
				t.Fatalf("legacy fallback missing:\n%s", body)
			}
		})
	}
}

func TestOwnedControlFlowLoweringAndMap(t *testing.T) {
	p, err := frontend.Parse(`<?php
function selected(): array {
    $items = ['old' => 1];
    $flag = true;
    if ($flag) {
        $branch = $items;
        $items['then'] = 2;
    } else {
        $branch = [];
        $items['else'] = 3;
    }
    $i = 0;
    while ($i !== 2) {
        $iteration = $i;
        $items[] = $iteration;
        $i = $i + 1;
        if ($i === 1) {
            continue;
            $items['continue-dead'] = 9;
        }
        break;
        $items['break-dead'] = 10;
    }
    return $items;
}`, testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	out, mapping, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	body := generatedFunction(t, string(out), "F_selected")

	for _, local := range []string{"branch", "flag", "i", "items", "iteration"} {
		declare := `v["` + local + `"] = &rt.ValueSlot{}`
		close := `defer v["` + local + `"].Close()`
		if strings.Count(body, declare) != 1 || strings.Count(body, close) != 1 {
			t.Fatalf("owned local %q is not allocated and closed at function entry:\n%s", local, body)
		}
		if strings.Index(body, declare) > strings.Index(body, "if rt.Bool(") {
			t.Fatalf("owned local %q was allocated after control flow:\n%s", local, body)
		}
	}
	for _, fragment := range []string{
		`temporary.MoveTo(v["branch"])`,
		`v["items"].ArrayWrite(key, value)`,
		`v["items"].ArrayAppend(value)`,
		"for rt.Bool(",
		"continue",
		"break",
		"return rt.ArrayReturn((func() any {",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("owned control-flow lowering missing %q:\n%s", fragment, body)
		}
	}
	for _, dead := range []string{"continue-dead", "break-dead"} {
		if strings.Contains(body, dead) {
			t.Fatalf("dead control-flow tail was emitted (%q):\n%s", dead, body)
		}
	}

	file, err := parser.ParseFile(token.NewFileSet(), "generated.go", out, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "F_selected" {
			continue
		}
		for _, stmt := range fn.Body.List {
			if deferred, ok := stmt.(*ast.DeferStmt); ok {
				if selector, ok := deferred.Call.Fun.(*ast.SelectorExpr); ok {
					if receiver, ok := selector.X.(*ast.Ident); ok && receiver.Name == "temporary" {
						t.Fatal("temporary cleanup escaped an operation into the PHP function lifetime")
					}
				}
			}
		}
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "map[string]*rt.ValueSlot{}") || strings.Contains(line, "defer v[") {
			lineNumber := strings.Count(string(out[:strings.Index(string(out), line)]), "\n") + 1
			if phpLine, ok := mapping.PHPLine(lineNumber); ok {
				t.Fatalf("synthetic lifetime line maps to PHP line %d: %s", phpLine, line)
			}
		}
	}
	for _, check := range []struct {
		fragment string
		phpLines []int
	}{
		{`temporary.MoveTo(v["branch"])`, []int{6, 9}},
		{`v["items"].ArrayWrite(key, value)`, []int{7, 10}},
		{`temporary.MoveTo(v["iteration"])`, []int{14}},
		{`v["items"].ArrayAppend(value)`, []int{15}},
		{`temporary.MoveTo(v["i"])`, []int{12, 16}},
		{"if rt.Bool(", []int{5, 17}},
		{"for rt.Bool(", []int{13}},
		{"continue", []int{18}},
		{"break", []int{21}},
		{"return rt.ArrayReturn((func() any {", []int{24}},
	} {
		found := false
		for index, line := range strings.Split(string(out), "\n") {
			if !strings.Contains(line, check.fragment) {
				continue
			}
			found = true
			got, ok := mapping.PHPLine(index + 1)
			matched := false
			for _, want := range check.phpLines {
				matched = matched || got == want
			}
			if !ok || !matched {
				t.Fatalf("%q maps to %d (%v), want one of %v", check.fragment, got, ok, check.phpLines)
			}
		}
		if !found {
			t.Fatalf("generated source missing %q:\n%s", check.fragment, out)
		}
	}
}

func TestOwnedControlFlowPreservesRejection(t *testing.T) {
	for _, source := range []string{
		`<?php function invalid() { if (true) { $items = []; } return $items; }`,
		`<?php function invalid() { while (false) { $items = []; } return $items; }`,
		`<?php function invalid() { $items = []; $i = 0; while ($i !== 1) { $items[] = $iteration; $iteration = $i; $i = $i + 1; } return $items; }`,
		`<?php function invalid() { $items = []; if (true) { $other =& $items; } return $items; }`,
		`<?php function invalid() { $items = []; while (false) { break; $value = $missing; } return $items; }`,
	} {
		p, err := frontend.Parse(source, testTarget(t))
		if err == nil {
			var out []byte
			out, err = Generate(p)
			if len(out) != 0 {
				t.Fatal("invalid source emitted output")
			}
		}
		if err == nil {
			t.Fatalf("invalid source accepted: %s", source)
		}
	}
	p := ir.Program{Version: ir.Version, PHPVersion: testTarget(t), Functions: []ir.Function{{
		Name: "invalid", Line: 1,
		Body: []ir.Stmt{{
			Kind: "if", Line: 1,
			Expr:     ir.Expr{Kind: "bool", Text: "invalid"},
			Body:     []ir.Stmt{{Kind: "assign", Name: "items", Line: 1, Expr: ir.Expr{Kind: "array"}}},
			ElseBody: []ir.Stmt{{Kind: "assign", Name: "items", Line: 1, Expr: ir.Expr{Kind: "array"}}},
		}},
	}}}
	if out, err := Generate(p); err == nil || !strings.Contains(err.Error(), "IR002") || len(out) != 0 {
		t.Fatalf("malformed control-flow IR accepted: %v", err)
	}
}
