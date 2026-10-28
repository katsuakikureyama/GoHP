package semantic

import "github.com/katsuakikureyama/GoHP/internal/ir"

// Inspect every node before walking a manually constructed property chain.
// General expressions remain excluded even when their result would be typed.
func preg83InputMetadata(input ir.Expr) error {
	for input.Kind == "property" {
		if !validName(input.Text) || input.Params != nil || len(input.Args) != 1 {
			return malformed(input.Line, "preg_match property inputs require named single-receiver nodes")
		}
		input = input.Args[0]
	}
	if input.Args != nil || input.Params != nil {
		return malformed(input.Line, "preg_match inputs cannot carry hidden operands or parameters")
	}
	return nil
}
