package codegen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func (g *generator) arrayPush(stmt ir.Stmt, scope map[string]bool) {
	// The dedicated statement is intentionally outside scalar/owned proofs.
	// No embedded mutating expression is added to the generic expression path.
	if g.scalarFrame || g.ownedValues {
		g.fail(stmt.Line, "array_push requires the guarded legacy local scope")
	}
	g.emit("{\n")
	operation := "rt.ArrayPushLocal(v," + strconv.Quote(stmt.Target.Text)
	if stmt.Target.Kind == "property" {
		// Capture owner/name before value arguments, but do not retain a
		// pre-argument array snapshot or check its actual array type yet.
		g.emit("target := rt.ResolveProperty(%s,%q)\n", g.expr(stmt.Target.Args[0], scope), stmt.Target.Text)
		operation = "target.Push("
	}
	var values []string
	for i, value := range stmt.Expr.Args {
		local := fmt.Sprintf("pushValue%d", i)
		g.emit("%s := rt.Clone(%s)\n", local, g.expr(value, scope))
		values = append(values, local)
	}
	if stmt.Target.Kind == "var" && len(values) != 0 {
		operation += ","
	}
	operation += strings.Join(values, ",") + ")"
	if stmt.Name == "" {
		g.emit("_ = %s\n", operation)
	} else {
		g.emit("v[%q] = %s\n", stmt.Name, operation)
		scope[stmt.Name] = true
	}
	g.emit("}\n")
}
