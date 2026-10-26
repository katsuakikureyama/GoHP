package codegen

import "github.com/katsuakikureyama/GoHP/internal/ir"

// scalarFrameBody is a conservative storage-domain proof, not PHP admission or
// type validation. It runs only after semantic.Validate, for global functions
// with no parameters or all-exact-scalar parameters. Every local value must
// originate in a scalar literal, guarded parameter or scalar operation. Unknown
// forms (including calls outside the explicit pure-builtin list) retain the
// existing value-map implementation.
// Inspect dead tails and both branches too; never infer safety from reachability.
func scalarFrameBody(body []ir.Stmt) bool {
	for _, stmt := range body {
		switch stmt.Kind {
		case "assign", "echo", "return", "call_stmt":
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

// Untyped inputs can carry arrays, objects or callables even if a particular
// caller passes scalars. Selection must hold for host entries and every call.
func scalarFrameParameters(function ir.Function) bool {
	if len(function.Params) == 0 {
		return true
	}
	if len(function.ParamTypes) != len(function.Params) {
		return false
	}
	for _, annotation := range function.ParamTypes {
		switch annotation {
		case "int", "string", "bool":
		default:
			return false
		}
	}
	return true
}

func scalarFrameExpr(expr ir.Expr) bool {
	switch expr.Kind {
	case "int", "string", "bool", "null", "var":
		// Definite assignment is still enforced by validation/codegen. Entry
		// guards and the absence of non-scalar producers keep locals scalar.
	case "bool_cast", "not", "ternary", "isset_var", "isset_multi":
	case "call":
		// These existing runtime helpers return scalar values or panic. This
		// is a storage-domain allowlist, not a general purity/type inference.
		// Semantic validation and normal lowering retain arity/domain guards
		// and reject builtin redeclarations. Inspect every operand below:
		// even a scalar-returning predicate must not admit an array producer.
		switch name(expr.Text) {
		case "strlen", "strtolower", "strcasecmp", "rtrim", "substr",
			"str_contains", "str_starts_with", "is_array", "is_bool",
			"is_int", "is_null", "is_string", "is_object":
		default:
			return false
		}
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
