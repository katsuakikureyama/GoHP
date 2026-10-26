package codegen

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/katsuakikureyama/GoHP/internal/ir"
)

type valueDomain uint8

const (
	scalarValue valueDomain = 1 << iota
	flatArrayValue
)

// ownedValueLocals proves the entire straight-line, nullary global scope after
// semantic validation. No host array, unlisted call result, alias or nested element can
// enter it. Domains follow all assignments, including dead tails. Unsupported
// scopes retain the legacy map; this proof does not broaden PHP admission.
// Control-flow statements need a separate join/loop proof and are not selected.
func ownedValueLocals(f ir.Function) ([]string, bool) {
	if len(f.Params) != 0 {
		return nil, false
	}
	locals := map[string]valueDomain{}
	hasArray := false
	for _, stmt := range f.Body {
		domain := ownedValueDomain(stmt.Expr, locals)
		if domain == 0 {
			return nil, false
		}
		hasArray = hasArray || domain&flatArrayValue != 0
		switch stmt.Kind {
		case "assign":
			locals[stmt.Name] = domain
		case "assign_index":
			if domain != scalarValue || locals[stmt.Target.Args[0].Text] != flatArrayValue || ownedValueDomain(stmt.Target.Args[1], locals) != scalarValue {
				return nil, false
			}
		case "assign_append":
			if domain != scalarValue || locals[stmt.Target.Text] != flatArrayValue {
				return nil, false
			}
		case "echo", "call_stmt":
			if domain != scalarValue {
				return nil, false
			}
		case "return":
		default:
			return nil, false
		}
	}
	if !hasArray {
		return nil, false
	}
	names := make([]string, 0, len(locals))
	for name := range locals {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, true
}

func ownedValueDomain(e ir.Expr, locals map[string]valueDomain) valueDomain {
	switch e.Kind {
	case "int", "string", "bool", "null", "version_constant":
		return scalarValue
	case "var":
		return locals[e.Text]
	case "array":
		for _, pair := range e.Args {
			if ownedValueDomain(pair.Args[1], locals) != scalarValue {
				return 0
			}
		}
		return flatArrayValue
	case "ternary":
		condition := ownedValueDomain(e.Args[0], locals)
		left, right := ownedValueDomain(e.Args[1], locals), ownedValueDomain(e.Args[2], locals)
		if condition != 0 && left != 0 && right != 0 {
			return left | right
		}
	case "index", "isset_index":
		if ownedValueDomain(e.Args[0], locals) == flatArrayValue && ownedValueDomain(e.Args[1], locals) == scalarValue {
			return scalarValue
		}
	case "bool_cast", "not", "isset_var", "isset_multi":
		for _, arg := range e.Args {
			if ownedValueDomain(arg, locals) == 0 {
				return 0
			}
		}
		return scalarValue
	case "binary":
		for _, arg := range e.Args {
			domain := ownedValueDomain(arg, locals)
			if domain == 0 {
				return 0
			}
			switch e.Text {
			case "+", "-", ".":
				if domain != scalarValue {
					return 0
				}
			case "===", "!==", "&&", "||":
			default:
				return 0
			}
		}
		return scalarValue
	case "call":
		if name(e.Text) == "count" {
			if ownedValueDomain(e.Args[0], locals) == flatArrayValue {
				return scalarValue
			}
			return 0
		}
		// Reuse the explicit pure scalar helper list, but prove operands in
		// this scope rather than assuming that every variable is scalar.
		if !scalarFrameExpr(ir.Expr{Kind: "call", Text: e.Text}) {
			return 0
		}
		for _, arg := range e.Args {
			if ownedValueDomain(arg, locals) != scalarValue {
				return 0
			}
		}
		return scalarValue
	}
	return 0
}

// ownedValueExpr produces exactly one owned result. A variable captures its
// current value without exporting/reimporting; ternaries acquire only in the
// selected arm. Other producers evaluate legacy operands first: Read exports
// and releases each internal owner within the operand, so no operand owner
// survives into a later operand or branch. No call here can mutate a local.
func (g *generator) ownedValueExpr(e ir.Expr, scope map[string]bool) string {
	switch e.Kind {
	case "var":
		g.expr(e, scope) // Preserve codegen's definite-assignment diagnostic.
		return "v[" + strconv.Quote(e.Text) + "].Snapshot()"
	case "ternary":
		return fmt.Sprintf("(func() *rt.ValueTemporary { if rt.Bool(%s) { return %s }; return %s }())",
			g.expr(e.Args[0], scope), g.ownedValueExpr(e.Args[1], scope), g.ownedValueExpr(e.Args[2], scope))
	default:
		return "rt.NewValueTemporary(" + g.expr(e, scope) + ")"
	}
}

// The operation's IIFE is the failure/cleanup boundary, not the PHP function's
// exit. MoveTo consumes on success; Take exports a detached result that remains
// valid after both temporary cleanup and the enclosing locals' deferred Close.
func (g *generator) ownedValueAssignment(name string, e ir.Expr, scope map[string]bool) {
	g.emit("func() { temporary := %s; defer temporary.Close(); temporary.MoveTo(v[%q]) }()\n", g.ownedValueExpr(e, scope), name)
}

func (g *generator) ownedValueResult(e ir.Expr, scope map[string]bool) string {
	return fmt.Sprintf("(func() any { temporary := %s; defer temporary.Close(); return temporary.Take() }())", g.ownedValueExpr(e, scope))
}
