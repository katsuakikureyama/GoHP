package runtime

import "math"

// Subtract deliberately accepts actual int64 values only, not Number coercion.
// Compare against safe bounds without negating the possibly-MinInt64 RHS.
func Subtract(a, b any) any {
	x, ok := a.(int64)
	if !ok {
		panic("subtraction requires int operands")
	}
	y, ok := b.(int64)
	if !ok {
		panic("subtraction requires int operands")
	}
	if y > 0 && x < math.MinInt64+y || y < 0 && x > math.MaxInt64+y {
		panic("integer subtraction overflow to float unsupported")
	}
	return x - y
}
