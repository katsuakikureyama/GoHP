package codegen

import "github.com/katsuakikureyama/GoHP/internal/ir"

// Scan validated IR, including dead branches and arrow bodies. Every generated
// call in a stateful program must propagate the same explicit execution.
func usesStrtok(p ir.Program) bool {
	var expression func(ir.Expr) bool
	expression = func(e ir.Expr) bool {
		if e.Kind == "call" && name(e.Text) == "strtok" {
			return true
		}
		for _, arg := range e.Args {
			if expression(arg) {
				return true
			}
		}
		return false
	}
	var statements func([]ir.Stmt) bool
	statements = func(stmts []ir.Stmt) bool {
		for _, s := range stmts {
			if expression(s.Expr) || expression(s.Target) || statements(s.Body) || statements(s.ElseBody) {
				return true
			}
		}
		return false
	}
	if statements(p.Body) {
		return true
	}
	for _, f := range p.Functions {
		if statements(f.Body) {
			return true
		}
	}
	for _, c := range p.Classes {
		for _, f := range c.Methods {
			if statements(f.Body) {
				return true
			}
		}
	}
	return false
}
