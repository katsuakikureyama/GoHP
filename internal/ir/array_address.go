package ir

// ArrayAddressParts recognizes only a variable or a variable-rooted index chain.
// It never treats arbitrary expressions/properties as suppressed-read addresses.
// Name, definition and key-domain checks belong to semantic validation.
func ArrayAddressParts(expr Expr) (Expr, []Expr, bool) {
	var reversed []Expr
	for expr.Kind == "index" {
		if expr.Text != "" || len(expr.Args) != 2 || len(expr.Params) != 0 {
			return Expr{}, nil, false
		}
		reversed = append(reversed, expr.Args[1])
		expr = expr.Args[0]
	}
	if expr.Kind != "var" || len(expr.Args) != 0 || len(expr.Params) != 0 {
		return Expr{}, nil, false
	}
	keys := make([]Expr, len(reversed))
	for i := range reversed {
		keys[len(reversed)-1-i] = reversed[i]
	}
	return expr, keys, true
}
