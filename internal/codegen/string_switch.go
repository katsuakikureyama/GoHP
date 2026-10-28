package codegen

import (
	"fmt"
	"strconv"

	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func (g *generator) stringSwitch(s ir.Stmt, scope map[string]bool, inFunc bool, returnType string) {
	g.serial++
	index := fmt.Sprintf("switchIndex%d", g.serial)
	defaultIndex := -1
	for i, clause := range s.Body {
		if clause.Kind == "default" {
			defaultIndex = i
		}
	}
	g.emit("%s := %d\nswitch rt.SwitchString(%s) {\n", index, defaultIndex, g.expr(s.Expr, scope))
	seen := map[string]bool{}
	for i, clause := range s.Body {
		if clause.Kind == "case" && !seen[clause.Expr.Text] {
			g.emit("case %s: %s=%d\n", strconv.Quote(clause.Expr.Text), index, i)
			seen[clause.Expr.Text] = true
		}
	}
	// Integer dispatch admits duplicate PHP labels (first match), and preserves
	// every source clause for subsequent fallthrough, including default.
	g.emit("}\nswitch %s {\n", index)
	for i, clause := range s.Body {
		g.emit("case %d:\n", i)
		inner := map[string]bool{}
		for name, defined := range scope {
			inner[name] = defined
		}
		previous := g.suppress
		terminated := false
		for _, statement := range clause.Body {
			g.stmts([]ir.Stmt{statement}, inner, inFunc, returnType)
			if switchTransfer(statement) {
				terminated = true
				g.suppress = true
			}
		}
		g.suppress = previous
		if !terminated && i+1 < len(s.Body) {
			g.emit("fallthrough\n")
		}
	}
	g.emit("}\n")
}

// Only unconditional transfers suppress fallthrough. Loops and nested switches
// may finish normally; a break inside them does not exit this switch.
func switchTransfer(s ir.Stmt) bool {
	switch s.Kind {
	case "return", "break", "continue":
		return true
	case "if":
		return switchBodyTransfers(s.Body) && s.ElseBody != nil && switchBodyTransfers(s.ElseBody)
	}
	return false
}

func switchBodyTransfers(body []ir.Stmt) bool {
	for _, s := range body {
		if switchTransfer(s) {
			return true
		}
	}
	return false
}
