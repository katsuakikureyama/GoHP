package codegen

import "github.com/katsuakikureyama/GoHP/internal/ir"

// scalarFrameBody is a conservative storage-domain proof, not PHP admission or
// type validation. It runs only after semantic.Validate, for parameter-free
// global functions. Every value entering locals must originate in a scalar
// literal or an operation whose operands and result remain scalar. Unknown
// forms (including all calls) retain the existing value-map implementation.
// Inspect dead tails and both branches too; never infer safety from reachability.
func scalarFrameBody(body []ir.Stmt) bool {
	for _, stmt := range body {
		switch stmt.Kind {
		case "assign", "echo", "return":
			if !scalarFrameExpr(stmt.Expr) {
				return false
			}
		case "if", "while":
			if !scalarFrameExpr(stmt.Expr) || !scalarFrameBody(stmt.Body) || !scalarFrameBody(stmt.ElseBody) {
				return false
			}
		case "break", "continue":
			// Payload and enclosing-loop checks belong to semantic.Validate.
		default:
			return false
		}
	}
	return true
}

func scalarFrameExpr(expr ir.Expr) bool {
	switch expr.Kind {
	case "int", "string", "bool", "null", "var":
		// Definite assignment is still enforced by validation/codegen. With no
		// parameters or non-scalar producers, every defined local is scalar.
	case "bool_cast", "not", "ternary", "isset_var", "isset_multi":
	case "binary":
		switch expr.Text {
		case "+", "-", ".", "===", "!==", "&&", "||":
		default:
			return false
		}
	default:
		return false
	}
	for _, arg := range expr.Args {
		if !scalarFrameExpr(arg) {
			return false
		}
	}
	return true
}
