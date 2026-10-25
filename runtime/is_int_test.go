package runtime

import (
	"math"
	"testing"
)

func TestIsIntValueDomain(t *testing.T) {
	for _, value := range []int64{math.MinInt64, -1, 0, 1, math.MaxInt64} {
		if !IsInt(value) {
			t.Fatalf("is_int(%d) = false", value)
		}
	}
	for _, value := range []any{
		nil, false, true, "", "0", "42", "-1", NewArray(),
		NewArray(Entry{Key: int64(0), Value: int64(42)}), struct{}{},
		NewCallable(0, func(...any) any { return int64(42) }),
		int(42), int32(42), uint64(42), float64(42), (*int64)(nil),
	} {
		if IsInt(value) {
			t.Fatalf("is_int(%T: %v) = true; must not coerce", value, value)
		}
	}
}
