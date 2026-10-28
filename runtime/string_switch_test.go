package runtime

import "testing"

func TestSwitchStringDomain(t *testing.T) {
	for _, s := range []string{"alpha", "", " ", "NaN", "INF", "0x10", "1e", "1e2x", "é", "1\x00", "\xff"} {
		if got := SwitchString(s); got != s {
			t.Fatalf("bytes changed: %q", s)
		}
	}
	for _, value := range []any{nil, true, int64(0), NewArray(), "0", "-0", "+1", " 12\t", "\v-1.\r", ".5", "1e3", "1.E+2", "1e9999", "9223372036854775808"} {
		func() {
			defer func() {
				if got := recover(); got != "switch requires a nonnumeric string" {
					t.Fatalf("%#v: panic %v", value, got)
				}
			}()
			SwitchString(value)
		}()
	}
}
