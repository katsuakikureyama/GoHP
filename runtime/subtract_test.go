package runtime

import (
	"math"
	"testing"
)

func TestIntegerSubtractionRuntime(t *testing.T) {
	for _, c := range []struct{ x, y, want int64 }{
		{3, 1, 2}, {1, 3, -2}, {0, 0, 0}, {3, -2, 5}, {math.MinInt64, math.MinInt64, 0}, {math.MaxInt64, math.MaxInt64, 0}, {math.MinInt64, 0, math.MinInt64}, {math.MaxInt64, 0, math.MaxInt64},
	} {
		if got := Subtract(c.x, c.y); got != c.want {
			t.Fatalf("%d-%d=%v; want%d", c.x, c.y, got, c.want)
		}
	}
	for _, c := range []struct {
		x, y any
		want string
	}{
		{nil, int64(0), "subtraction requires int operands"},
		{int64(0), "0", "subtraction requires int operands"},
		{int64(math.MinInt64), int64(1), "integer subtraction overflow to float unsupported"},
		{int64(0), int64(math.MinInt64), "integer subtraction overflow to float unsupported"},
		{int64(math.MaxInt64), int64(-1), "integer subtraction overflow to float unsupported"},
	} {
		func() {
			defer func() {
				if got := recover(); got != c.want {
					t.Fatalf("panic=%v; want%q", got, c.want)
				}
			}()
			Subtract(c.x, c.y)
		}()
	}
}
