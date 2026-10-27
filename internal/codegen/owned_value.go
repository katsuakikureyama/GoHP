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

// ownedValueLocals proves a global scope after semantic validation. Parameters
// must have exact scalar entry guards; no host array, unlisted call result,
// alias or nested element can enter it.
// Branches join conservatively; loops require a domain invariant at every write,
// not just at their textual end. Unsupported scopes retain the legacy map.
// This storage proof never broadens PHP admission or definite assignment.
func ownedValueLocals(f ir.Function) ([]string, bool) {
	if !scalarFrameParameters(f) {
		return nil, false
	}
	proof := ownedValueProof{names: map[string]bool{}}
	entry := map[string]valueDomain{}
	for _, parameter := range f.Params {
		// Include even read-only/unused parameters: the entry guard stores
		// every argument before the body. An annotation constrains entry,
		// not later local reassignment outside a loop invariant.
		proof.names[parameter] = true
		entry[parameter] = scalarValue
	}
	if !proof.body(f.Body, entry, nil) || !proof.hasArray {
		return nil, false
	}
	names := make([]string, 0, len(proof.names))
	for name := range proof.names {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, true
}

type ownedValueProof struct {
	names    map[string]bool
	hasArray bool
}

func cloneValueDomains(source map[string]valueDomain) map[string]valueDomain {
	copy := make(map[string]valueDomain, len(source))
	for name, domain := range source {
		copy[name] = domain
	}
	return copy
}

// limits contains the invariant for every active loop-entry local. Branch/body
// locals still need lexical definite assignment before reads on each iteration.
// They get a reusable function-owned slot, never a Go defer inside the loop.
func (proof *ownedValueProof) body(body []ir.Stmt, locals, limits map[string]valueDomain) bool {
	for _, stmt := range body {
		if stmt.Kind == "break" || stmt.Kind == "continue" {
			// Semantic validation checks the enclosing loop and empty payload.
			// Inspect dead tails in a detached environment without rewriting the
			// caller's state at the reachable transfer. Invariants still apply.
			locals = cloneValueDomains(locals)
			continue
		}
		domain := ownedValueDomain(stmt.Expr, locals)
		if domain == 0 {
			return false
		}
		proof.hasArray = proof.hasArray || domain&flatArrayValue != 0
		switch stmt.Kind {
		case "assign":
			if limit, constrained := limits[stmt.Name]; constrained && domain&^limit != 0 {
				return false
			}
			proof.names[stmt.Name] = true
			locals[stmt.Name] = domain
		case "assign_index":
			if domain != scalarValue || locals[stmt.Target.Args[0].Text] != flatArrayValue || ownedValueDomain(stmt.Target.Args[1], locals) != scalarValue {
				return false
			}
		case "assign_append":
			if domain != scalarValue || locals[stmt.Target.Text] != flatArrayValue {
				return false
			}
		case "echo", "call_stmt":
			if domain != scalarValue {
				return false
			}
		case "return":
			locals = cloneValueDomains(locals) // Still inspect unreachable code.
		case "if":
			left, right := cloneValueDomains(locals), cloneValueDomains(locals)
			if !proof.body(stmt.Body, left, limits) || !proof.body(stmt.ElseBody, right, limits) {
				return false
			}
			// Keep the entry state too: no constant/return-sensitive pruning and
			// no promotion of branch-only names, even when both arms define one.
			for name, before := range locals {
				locals[name] = before | left[name] | right[name]
			}
		case "while":
			loopLimits := cloneValueDomains(limits)
			for name, before := range locals {
				if outer, constrained := loopLimits[name]; constrained {
					loopLimits[name] = outer & before
				} else {
					loopLimits[name] = before
				}
			}
			if !proof.body(stmt.Body, cloneValueDomains(locals), loopLimits) {
				return false
			}
			// Every write stays within the header domain, including before any
			// break/continue/return and in nested loops. The unchanged header is
			// therefore a sound upper bound for zero or arbitrarily many trips,
			// subsequent condition evaluations and every post-loop local.
		default:
			return false
		}
	}
	return true
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
