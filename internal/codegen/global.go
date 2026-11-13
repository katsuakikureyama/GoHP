package codegen

import "github.com/katsuakikureyama/GoHP/internal/ir"

// hasGlobalStmts scans a program for any global statements.
func hasGlobalStmts(p ir.Program) bool {
	var scan func([]ir.Stmt) bool
	scan = func(stmts []ir.Stmt) bool {
		for _, s := range stmts {
			if s.Kind == "global" {
				return true
			}
			if scan(s.Body) || scan(s.ElseBody) {
				return true
			}
		}
		return false
	}
	if scan(p.Body) {
		return true
	}
	for _, f := range p.Functions {
		if scan(f.Body) {
			return true
		}
	}
	return false
}

// globalNamesInScope returns the set of global-bound names for a given
// function body or main body. For functions, only the leading prologue
// global statements count. For main, all global statements contribute.
func globalNamesInScope(stmts []ir.Stmt, inFunc bool) map[string]bool {
	globals := make(map[string]bool)
	if inFunc {
		// Only leading prologue globals.
		for _, s := range stmts {
			if s.Kind != "global" {
				break
			}
			for _, arg := range s.Expr.Args {
				globals[arg.Text] = true
			}
		}
	} else {
		// Main: all global statements.
		for _, s := range stmts {
			if s.Kind == "global" {
				for _, arg := range s.Expr.Args {
					globals[arg.Text] = true
				}
			}
		}
	}
	return globals
}
