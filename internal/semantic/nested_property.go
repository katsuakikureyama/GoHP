package semantic

import (
	"github.com/katsuakikureyama/GoHP/internal/ir"
	"strings"
)

func (v validator) nestedProperty(stmt ir.Stmt, scope map[string]valueType) error {
	if stmt.Name != "" || stmt.Key != "" || stmt.Value != "" || stmt.Body != nil {
		return malformed(stmt.Line, "nested property assignment requires only target and RHS")
	}
	root := stmt.Target
	depth := 0
	for root.Kind == "property" {
		if !validName(root.Text) || root.Params != nil || len(root.Args) != 1 {
			return malformed(stmt.Line, "nested property target requires named property nodes")
		}
		depth++
		root = root.Args[0]
	}
	if depth < 2 || root.Kind != "var" || !validName(root.Text) || root.Args != nil || root.Params != nil {
		return malformed(stmt.Line, "nested property target requires two or more properties rooted in a bare local")
	}
	if root.Text == "GLOBALS" || strings.HasPrefix(root.Text, "_") {
		return unsupported(stmt.Line, "global property roots unsupported")
	}
	// Existing property-read validation checks definitions, known classes and
	// declared prefixes; unknown classes are checked by the runtime protocol.
	if _, err := v.expr(stmt.Target.Args[0], scope); err != nil {
		return err
	}
	rhs, err := v.expr(stmt.Expr, scope)
	if err != nil {
		return err
	}
	if stmt.Kind == "assign_nested_property_concat" && rhs.kind != stringKind && rhs.kind != unknownKind {
		return unsupported(stmt.Line, "nested property concatenation requires strings")
	}
	return nil
}
