package runtime

import (
	"math"
	"testing"
)

func TestRelationalIntNull(t *testing.T) {
	for _, c := range []struct {
		a, b          any
		less, greater bool
	}{
		{int64(1), int64(2), true, false},
		{int64(2), int64(1), false, true},
		{int64(-1), int64(-1), false, false},
		{int64(math.MinInt64), int64(math.MaxInt64), true, false},
		{int64(math.MaxInt64), int64(math.MinInt64), false, true},
		{nil, nil, false, false},
		{nil, int64(0), false, false},
		{int64(0), nil, false, false},
		{nil, int64(-1), true, false},
		{nil, int64(1), true, false},
		{int64(-1), nil, false, true},
		{int64(1), nil, false, true},
	} {
		if LessThan(c.a, c.b) != c.less || GreaterThan(c.a, c.b) != c.greater {
			t.Fatalf("wrong relation for %#v and %#v", c.a, c.b)
		}
	}
}

func TestRelationalRejectsOtherDomains(t *testing.T) {
	for _, bad := range []any{true, false, "1", int(1), float64(1), NewArray(), []any{}, (*int)(nil)} {
		for _, valid := range []any{nil, int64(0)} {
			for _, operands := range [][2]any{{bad, valid}, {valid, bad}} {
				for _, compare := range []func(any, any) bool{LessThan, GreaterThan} {
					func() {
						defer func() {
							if got := recover(); got != "relational comparison requires int or null operands" {
								t.Fatalf("operands %#v: panic %v", operands, got)
							}
						}()
						compare(operands[0], operands[1])
					}()
				}
			}
		}
	}
}
