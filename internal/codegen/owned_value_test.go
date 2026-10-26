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

func TestOwnedValueSelection(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		owned        bool
		scalar       bool
	}{
		{"flat_array", `function selected() { $items = ['old' => 1]; $items['new'] = 2; $items[] = 3; return $items; }`, true, false},
		// The tail remains part of the storage proof even after a return.
		{"dead_flat_tail", `function selected() { return 1; $items = ['tail' => 2]; }`, true, false},
		{"dead_nested_tail", `function selected() { return 1; $items = ['nested' => [2]]; }`, false, false},
		{"dead_unknown_producer", `function other() { return []; } function selected() { return 1; $items = other(); }`, false, false},
		{"scalar_only", `function selected() { $value = 1; return $value; }`, false, true},
		{"parameter", `function selected(array $items) { return $items; }`, false, false},
		{"branch", `function selected() { $items = []; if (false) { $items[] = 1; } return $items; }`, false, false},
		{"while", `function selected() { $items = []; while (false) { $items[] = 1; } return $items; }`, false, false},
		{"foreach", `function selected() { $items = []; foreach ([] as $item) { $items[] = $item; } return $items; }`, false, false},
		{"nested_array", `function selected() { $items = ['nested' => [1]]; return $items; }`, false, false},
		{"array_element_union", `function selected() { $first = []; $choice = true ? 1 : $first; $items = ['choice' => $choice]; return $items; }`, false, false},
		{"stored_union", `function selected() { $first = []; $choice = true ? 1 : $first; return $choice; }`, true, false},
		{"user_call", `function other() { return []; } function selected() { $items = other(); return $items; }`, false, false},
		{"stateful_call", `function selected() { $items = []; $token = strtok('a,b', ','); return $items; }`, false, false},
		{"method", `class Box { public function selected() { $items = []; return $items; } }`, false, false},
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
			var body string
			if tc.name == "method" {
				body = generatedMethod(t, string(out), "selected")
			} else if tc.name == "stateful_call" {
				body = generatedFunction(t, string(out), "FContext_selected")
			} else {
				body = generatedFunction(t, string(out), "F_selected")
			}
			if got := strings.Contains(body, "map[string]*rt.ValueSlot{}"); got != tc.owned {
				t.Fatalf("owned-value selection = %v, want %v:\n%s", got, tc.owned, body)
			}
			if got := strings.Contains(body, "rt.NewScalarFrame()"); got != tc.scalar {
				t.Fatalf("scalar-frame selection = %v, want %v:\n%s", got, tc.scalar, body)
			}
			if tc.owned {
				if !strings.Contains(body, "defer v[") || strings.Contains(body, "v := map[string]any{}") {
					t.Fatalf("owned-value locals are not closed through slots:\n%s", body)
				}
			} else if !tc.scalar && !strings.Contains(body, "v := map[string]any{}") {
				t.Fatalf("fallback did not retain legacy value map:\n%s", body)
			}
		})
	}
}

func TestOwnedValueLoweringAndMap(t *testing.T) {
	p, err := frontend.Parse(`<?php
function selected(): array {
    $items = ['old' => 1];
    $copy = $items;
    $items['old'] = 2;
    $items[] = 3;
    echo $copy['old'];
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
	for _, fragment := range []string{
		"v := map[string]*rt.ValueSlot{}",
		`v["copy"] = &rt.ValueSlot{}`,
		`v["items"] = &rt.ValueSlot{}`,
		"temporary := rt.NewValueTemporary(rt.NewArray(",
		`temporary := v["items"].Snapshot()`,
		"defer temporary.Close()",
		"return temporary.Take()",
		`temporary.MoveTo(v["items"])`,
		`v["items"].ArrayWrite(key, value)`,
		`v["items"].ArrayAppend(value)`,
		"return rt.ArrayReturn((func() any {",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("owned-value lowering missing %q:\n%s", fragment, body)
		}
	}
	// Local owners may live until function exit, but temporary guards must be
	// inside expression IIFEs, never deferred directly by the PHP function.
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
			deferred, ok := stmt.(*ast.DeferStmt)
			if !ok {
				continue
			}
			if selector, ok := deferred.Call.Fun.(*ast.SelectorExpr); ok {
				if receiver, ok := selector.X.(*ast.Ident); ok && receiver.Name == "temporary" {
					t.Fatal("temporary cleanup escaped its operation into the PHP function lifetime")
				}
			}
		}
	}
	for _, line := range strings.Split(body, "\n") {
		if strings.Contains(line, "map[string]*rt.ValueSlot{}") || strings.Contains(line, "defer v[") {
			lineNumber := strings.Count(string(out[:strings.Index(string(out), line)]), "\n") + 1
			if phpLine, ok := mapping.PHPLine(lineNumber); ok {
				t.Fatalf("synthetic lifetime line maps to PHP line %d: %s", phpLine, line)
			}
		}
	}
	for _, check := range []struct {
		fragment string
		phpLine  int
	}{
		{"temporary := rt.NewValueTemporary(rt.NewArray(", 3},
		{`temporary.MoveTo(v["copy"])`, 4},
		{`v["items"].ArrayWrite(key, value)`, 5},
		{`v["items"].ArrayAppend(value)`, 6},
		{`rt.Echo(rt.Get(v["copy"].Read(), "old"))`, 7},
		{"return rt.ArrayReturn((func() any {", 8},
	} {
		found := false
		for index, line := range strings.Split(string(out), "\n") {
			if !strings.Contains(line, check.fragment) {
				continue
			}
			found = true
			if got, ok := mapping.PHPLine(index + 1); !ok || got != check.phpLine {
				t.Fatalf("%q maps to %d (%v), want %d", check.fragment, got, ok, check.phpLine)
			}
		}
		if !found {
			t.Fatalf("generated source missing %q:\n%s", check.fragment, out)
		}
	}
}

func generatedMethod(t *testing.T, output, name string) string {
	t.Helper()
	marker := " M_" + name + "("
	method := strings.Index(output, marker)
	if method < 0 {
		t.Fatalf("generated method %s missing:\n%s", name, output)
	}
	start := strings.LastIndex(output[:method], "\nfunc ")
	if start < 0 {
		t.Fatalf("generated method %s lacks declaration:\n%s", name, output)
	}
	start++
	body := output[start:]
	if end := strings.Index(body[1:], "\nfunc "); end >= 0 {
		body = body[:end+1]
	}
	return body
}

func TestOwnedValuePreservesRejection(t *testing.T) {
	for _, source := range []string{
		`<?php function invalid() { $items = []; return $missing; }`,
		`<?php function invalid() { $items = []; $other =& $items; return $items; }`,
		`<?php function invalid() { $items = []; return $items; $value = $missing; }`,
		`<?php function invalid() { $items = []; return $items; $other =& $items; }`,
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
			Kind: "assign_index", Line: 1,
			Target: ir.Expr{Kind: "index", Line: 1, Args: []ir.Expr{{Kind: "var", Text: "items", Line: 1}}},
			Expr:   ir.Expr{Kind: "int", Text: "1", Line: 1},
		}},
	}}}
	if out, err := Generate(p); err == nil || !strings.Contains(err.Error(), "IR002") || len(out) != 0 {
		t.Fatalf("malformed owned-value IR accepted: %v", err)
	}
}
