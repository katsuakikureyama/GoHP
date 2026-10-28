package codegen

import (
	"github.com/katsuakikureyama/GoHP/internal/ir"
	"strconv"
	"strings"
)

func (g *generator) nestedProperty(s ir.Stmt, scope map[string]bool) {
	root := s.Target
	var path []string
	for root.Kind == "property" {
		path = append(path, root.Text)
		root = root.Args[0]
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	for i := range path {
		path[i] = strconv.Quote(path[i])
	}
	operation := "Assign"
	if s.Kind == "assign_nested_property_append" {
		operation = "Append"
	}
	if s.Kind == "assign_nested_property_concat" {
		operation = "Concat"
	}
	// Resolve before RHS evaluation. Retain the final object even when RHS
	// replaces a prefix property; read its current final value after the RHS.
	g.emit("{ target := rt.ResolveNestedProperty(%s,%s); value := any(%s); target.%s(value) }\n", g.expr(root, scope), strings.Join(path, ","), g.expr(s.Expr, scope), operation)
}
