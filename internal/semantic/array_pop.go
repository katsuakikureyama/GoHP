package semantic

import (
	"strings"

	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func (v validator) arrayPop(stmt ir.Stmt, scope map[string]valueType) error {
	if stmt.Key != "" || stmt.Value != "" || stmt.Body != nil || stmt.ElseBody != nil ||
		stmt.Expr.Kind != "call" || stmt.Expr.Text != "array_pop" || stmt.Expr.Args != nil || stmt.Expr.Params != nil {
		return malformed(stmt.Line, "array_pop requires a target and canonical empty call payload")
	}
	if stmt.Name != "" && (!validName(stmt.Name) || stmt.Name == "this" || stmt.Name == "GLOBALS" || strings.HasPrefix(stmt.Name, "_")) {
		return unsupported(stmt.Line, "array_pop result requires a nonreserved bare local")
	}
	target := stmt.Target
	var base ir.Expr
	switch target.Kind {
	case "var":
		base = target
	case "property":
		if !validName(target.Text) || target.Params != nil || len(target.Args) != 1 {
			return malformed(stmt.Line, "array_pop requires a direct named property target")
		}
		if target.Args[0].Kind != "var" {
			return unsupported(stmt.Line, "array_pop property requires a bare local receiver")
		}
		base = target.Args[0]
	default:
		return unsupported(stmt.Line, "array_pop requires an existing local or direct named property")
	}
	if !validName(base.Text) || base.Args != nil || base.Params != nil {
		return malformed(stmt.Line, "array_pop requires a bare local target root")
	}
	if base.Text == "GLOBALS" || strings.HasPrefix(base.Text, "_") || (target.Kind == "var" && base.Text == "this") {
		return unsupported(stmt.Line, "array_pop requires a nonreserved target")
	}
	if _, exists := scope[base.Text]; !exists {
		return unsupported(stmt.Line, "array_pop requires an existing target root")
	}
	t, err := v.expr(target, scope)
	if err != nil {
		return err
	}
	if t.kind != arrayKind && t.kind != unknownKind {
		return unsupported(stmt.Line, "array_pop requires an array target")
	}
	if target.Kind == "var" {
		scope[base.Text] = valueType{kind: arrayKind}
	}
	if stmt.Name != "" {
		// Mixed element/null result replaces an aliased array binding last.
		scope[stmt.Name] = valueType{kind: unknownKind}
	}
	return nil
}
