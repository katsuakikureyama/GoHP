package codegen

import (
	"strconv"

	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func (g *generator) arrayPop(stmt ir.Stmt, scope map[string]bool) {
	if g.scalarFrame || g.ownedValues {
		g.fail(stmt.Line, "array_pop requires the guarded legacy local scope")
	}
	g.emit("{\n")
	operation := "rt.ArrayPopLocal(v," + strconv.Quote(stmt.Target.Text) + ")"
	if stmt.Target.Kind == "property" {
		g.emit("target := rt.ResolveProperty(%s,%q)\n", g.expr(stmt.Target.Args[0], scope), stmt.Target.Text)
		operation = "target.Pop()"
	}
	if stmt.Name == "" {
		g.emit("_ = %s\n", operation)
	} else {
		g.emit("v[%q] = %s\n", stmt.Name, operation)
		scope[stmt.Name] = true
	}
	g.emit("}\n")
}
