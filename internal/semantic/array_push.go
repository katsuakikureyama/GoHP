package semantic

import (
	"strings"

	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func (v validator) arrayPush(stmt ir.Stmt, scope map[string]valueType) error {
	if stmt.Key != "" || stmt.Value != "" || stmt.Body != nil || stmt.ElseBody != nil ||
		stmt.Expr.Kind != "call" || stmt.Expr.Text != "array_push" || stmt.Expr.Params != nil {
		return malformed(stmt.Line, "array_push requires a target and canonical value arguments")
	}
	if stmt.Name != "" && (!validName(stmt.Name) || stmt.Name == "this" || stmt.Name == "GLOBALS" || strings.HasPrefix(stmt.Name, "_")) {
		return unsupported(stmt.Line, "array_push result requires a nonreserved bare local")
	}
	target := stmt.Target
	var base ir.Expr
	switch target.Kind {
	case "var":
		base = target
	case "property":
		if !validName(target.Text) || target.Params != nil || len(target.Args) != 1 {
			return malformed(stmt.Line, "array_push requires a direct named property target")
		}
		if target.Args[0].Kind != "var" {
			return unsupported(stmt.Line, "array_push property requires a bare local receiver")
		}
		base = target.Args[0]
	default:
		return unsupported(stmt.Line, "array_push requires an existing local or direct named property")
	}
	if !validName(base.Text) || base.Args != nil || base.Params != nil {
		return malformed(stmt.Line, "array_push requires a bare local target root")
	}
	if base.Text == "GLOBALS" || strings.HasPrefix(base.Text, "_") || (target.Kind == "var" && base.Text == "this") {
		return unsupported(stmt.Line, "array_push requires a nonreserved target")
	}
	if _, exists := scope[base.Text]; !exists {
		return unsupported(stmt.Line, "array_push requires an existing target root")
	}
	t, err := v.expr(target, scope)
	if err != nil {
		return err
	}
	if t.kind != arrayKind && t.kind != unknownKind {
		return unsupported(stmt.Line, "array_push requires an array target")
	}
	// Validate values in the original scope; a result local cannot initialize
	// itself or make an undefined target exist during this same statement.
	if _, err := v.exprs(stmt.Expr.Args, scope); err != nil {
		return err
	}
	if target.Kind == "var" {
		scope[base.Text] = valueType{kind: arrayKind}
	}
	if stmt.Name != "" {
		// Last publication wins, including `$items = array_push($items, ...)`.
		scope[stmt.Name] = valueType{kind: intKind}
	}
	return nil
}
