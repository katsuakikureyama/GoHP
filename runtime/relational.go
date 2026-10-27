package runtime

// relationalOperands guards both actual domains before comparing. PHP compares
// null with integers as booleans, not as numeric zero (notably for negatives).
// Int/int comparisons retain all 64 bits without subtraction or float coercion.
func relationalOperands(a, b any) (int64, int64) {
	x, xInt := a.(int64)
	y, yInt := b.(int64)
	if (!xInt && a != nil) || (!yInt && b != nil) {
		panic("relational comparison requires int or null operands")
	}
	if a == nil || b == nil {
		if x != 0 {
			x = 1
		}
		if y != 0 {
			y = 1
		}
	}
	return x, y
}

func LessThan(a, b any) bool {
	x, y := relationalOperands(a, b)
	return x < y
}

func GreaterThan(a, b any) bool {
	x, y := relationalOperands(a, b)
	return x > y
}
