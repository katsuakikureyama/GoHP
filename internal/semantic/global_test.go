package semantic

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestGlobalNames(t *testing.T) {
	stmts := []ir.Stmt{
		{Kind: "global", Expr: ir.Expr{Args: []ir.Expr{
			{Kind: "var", Text: "a"},
			{Kind: "var", Text: "b"},
		}}},
		{Kind: "global", Expr: ir.Expr{Args: []ir.Expr{
			{Kind: "var", Text: "c"},
		}}},
		{Kind: "assign", Name: "x"},
	}
	names := GlobalNames(stmts)
	if len(names) != 3 || names[0] != "a" || names[1] != "b" || names[2] != "c" {
		t.Fatalf("GlobalNames = %v, want [a b c]", names)
	}
}

func TestGlobalNamesStopsAtNonGlobal(t *testing.T) {
	stmts := []ir.Stmt{
		{Kind: "global", Expr: ir.Expr{Args: []ir.Expr{{Kind: "var", Text: "x"}}}},
		{Kind: "assign", Name: "y"},
		{Kind: "global", Expr: ir.Expr{Args: []ir.Expr{{Kind: "var", Text: "z"}}}},
	}
	names := GlobalNames(stmts)
	if len(names) != 1 || names[0] != "x" {
		t.Fatalf("GlobalNames = %v, want [x]", names)
	}
}

func TestGlobalValidName(t *testing.T) {
	for _, name := range []string{"a", "A", "x1", "foo_bar"} {
		if !validGlobalName(name) {
			t.Errorf("validGlobalName(%q) = false, want true", name)
		}
	}
	for _, name := range []string{"", "this", "GLOBALS", "1x", "x-y", "_private"} {
		if validGlobalName(name) {
			t.Errorf("validGlobalName(%q) = true, want false", name)
		}
	}
}

func makeProgram(body []ir.Stmt, functions []ir.Function) ir.Program {
	return ir.Program{
		Version:    ir.Version,
		PHPVersion: "8.3",
		Body:       body,
		Functions:  functions,
	}
}

func TestGlobalInFunctionPrologue(t *testing.T) {
	// Global at the start of a function body is admitted.
	p := makeProgram(nil, []ir.Function{
		{Name: "f", Body: []ir.Stmt{
			{Kind: "global", Expr: ir.Expr{Args: []ir.Expr{{Kind: "var", Text: "x"}}}},
			{Kind: "assign", Name: "x", Expr: ir.Expr{Kind: "int", Text: "1"}},
			{Kind: "return", Expr: ir.Expr{Kind: "null", Text: "null"}},
		}},
	})
	if err := Validate(p); err != nil {
		t.Fatalf("global in function prologue should be admitted: %v", err)
	}
}

func TestGlobalAfterStatementInFunction(t *testing.T) {
	// Global after another statement in a function is rejected.
	p := makeProgram(nil, []ir.Function{
		{Name: "f", Body: []ir.Stmt{
			{Kind: "assign", Name: "y", Expr: ir.Expr{Kind: "int", Text: "1"}},
			{Kind: "global", Expr: ir.Expr{Args: []ir.Expr{{Kind: "var", Text: "x"}}}},
			{Kind: "return", Expr: ir.Expr{Kind: "null", Text: "null"}},
		}},
	})
	if err := Validate(p); err == nil {
		t.Fatal("global after statement in function should be rejected")
	} else if !strings.Contains(err.Error(), "before other statements") {
		t.Fatalf("wrong error: %v", err)
	}
}

func TestGlobalInMethod(t *testing.T) {
	// Global in a method is rejected.
	p := ir.Program{
		Version:    ir.Version,
		PHPVersion: "8.3",
		Classes: []ir.Class{
			{Name: "C", Methods: []ir.Function{
				{Name: "m", Body: []ir.Stmt{
					{Kind: "global", Expr: ir.Expr{Args: []ir.Expr{{Kind: "var", Text: "x"}}}},
					{Kind: "return", Expr: ir.Expr{Kind: "null", Text: "null"}},
				}},
			}},
		},
	}
	if err := Validate(p); err == nil {
		t.Fatal("global in method should be rejected")
	} else if !strings.Contains(err.Error(), "methods") {
		t.Fatalf("wrong error: %v", err)
	}
}

func TestGlobalInNestedBlock(t *testing.T) {
	// Global inside an if block is rejected.
	p := makeProgram(nil, []ir.Function{
		{Name: "f", Body: []ir.Stmt{
			{Kind: "if", Expr: ir.Expr{Kind: "bool", Text: "true"}, Body: []ir.Stmt{
				{Kind: "global", Expr: ir.Expr{Args: []ir.Expr{{Kind: "var", Text: "x"}}}},
			}},
			{Kind: "return", Expr: ir.Expr{Kind: "null", Text: "null"}},
		}},
	})
	if err := Validate(p); err == nil {
		t.Fatal("global in nested block should be rejected")
	} else if !strings.Contains(err.Error(), "nested block") {
		t.Fatalf("wrong error: %v", err)
	}
}

func TestGlobalInMain(t *testing.T) {
	// Global at main top level is admitted.
	p := makeProgram([]ir.Stmt{
		{Kind: "global", Expr: ir.Expr{Args: []ir.Expr{{Kind: "var", Text: "x"}}}},
		{Kind: "assign", Name: "x", Expr: ir.Expr{Kind: "int", Text: "1"}},
	}, nil)
	if err := Validate(p); err != nil {
		t.Fatalf("global in main should be admitted: %v", err)
	}
}

func TestGlobalReservedName(t *testing.T) {
	// Global with reserved name is rejected.
	p := makeProgram([]ir.Stmt{
		{Kind: "global", Expr: ir.Expr{Args: []ir.Expr{{Kind: "var", Text: "this"}}}},
	}, nil)
	if err := Validate(p); err == nil {
		t.Fatal("global with 'this' should be rejected")
	}
}

func TestGlobalArrayIndexRejected(t *testing.T) {
	// Array index on a global-bound variable is rejected.
	p := makeProgram(nil, []ir.Function{
		{Name: "f", Body: []ir.Stmt{
			{Kind: "global", Expr: ir.Expr{Args: []ir.Expr{{Kind: "var", Text: "x"}}}},
			{Kind: "assign_index", Target: ir.Expr{Kind: "index", Args: []ir.Expr{
				{Kind: "var", Text: "x"},
				{Kind: "int", Text: "0"},
			}}, Expr: ir.Expr{Kind: "int", Text: "1"}},
			{Kind: "return", Expr: ir.Expr{Kind: "null", Text: "null"}},
		}},
	})
	if err := Validate(p); err == nil {
		t.Fatal("array index on global-bound variable should be rejected")
	}
}

func TestGlobalIssetRejected(t *testing.T) {
	// isset on a global-bound variable is rejected.
	p := makeProgram(nil, []ir.Function{
		{Name: "f", Body: []ir.Stmt{
			{Kind: "global", Expr: ir.Expr{Args: []ir.Expr{{Kind: "var", Text: "x"}}}},
			{Kind: "if", Expr: ir.Expr{Kind: "isset_var", Args: []ir.Expr{{Kind: "var", Text: "x"}}}, Body: []ir.Stmt{}},
			{Kind: "return", Expr: ir.Expr{Kind: "null", Text: "null"}},
		}},
	})
	if err := Validate(p); err == nil {
		t.Fatal("isset on global-bound variable should be rejected")
	}
}

func TestGlobalForeachBindingRejected(t *testing.T) {
	// foreach binding a global name is rejected.
	p := makeProgram(nil, []ir.Function{
		{Name: "f", Body: []ir.Stmt{
			{Kind: "global", Expr: ir.Expr{Args: []ir.Expr{{Kind: "var", Text: "x"}}}},
			{Kind: "assign", Name: "a", Expr: ir.Expr{Kind: "array"}},
			{Kind: "foreach", Expr: ir.Expr{Kind: "var", Text: "a"}, Value: "x", Body: []ir.Stmt{}},
			{Kind: "return", Expr: ir.Expr{Kind: "null", Text: "null"}},
		}},
	})
	if err := Validate(p); err == nil {
		t.Fatal("foreach binding on global-bound variable should be rejected")
	}
}

func TestGlobalMultipleNames(t *testing.T) {
	// Two names in one global statement.
	p := makeProgram(nil, []ir.Function{
		{Name: "f", Body: []ir.Stmt{
			{Kind: "global", Expr: ir.Expr{Args: []ir.Expr{
				{Kind: "var", Text: "a"},
				{Kind: "var", Text: "b"},
			}}},
			{Kind: "assign", Name: "a", Expr: ir.Expr{Kind: "int", Text: "1"}},
			{Kind: "assign", Name: "b", Expr: ir.Expr{Kind: "string", Text: "hello"}},
			{Kind: "return", Expr: ir.Expr{Kind: "null", Text: "null"}},
		}},
	})
	if err := Validate(p); err != nil {
		t.Fatalf("two names in global should be admitted: %v", err)
	}
}

func TestGlobalEmptyStmt(t *testing.T) {
	// Global with no names is malformed.
	p := makeProgram([]ir.Stmt{
		{Kind: "global", Expr: ir.Expr{}},
	}, nil)
	if err := Validate(p); err == nil {
		t.Fatal("empty global should be rejected")
	} else if !strings.Contains(err.Error(), "at least one") {
		t.Fatalf("wrong error: %v", err)
	}
}
