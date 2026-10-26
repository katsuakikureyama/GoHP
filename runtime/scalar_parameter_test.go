package runtime

import "testing"

func TestScalarParameterExactDomain(t *testing.T) {
	for _, tt := range []struct {
		annotation string
		value      any
	}{
		{"int", int64(0)}, {"int", int64(-9223372036854775807)}, {"int", int64(9223372036854775807)},
		{"string", ""}, {"string", "0\x00\f日本語"}, {"bool", false}, {"bool", true},
	} {
		if got := ScalarParameter(tt.value, tt.annotation); got != tt.value {
			t.Fatalf("%s value changed: %#v", tt.annotation, got)
		}
	}
	for _, annotation := range []string{"int", "string", "bool", "float"} {
		for _, value := range []any{nil, int(1), int64(1), "1", true, float64(1), []byte("1"), NewArray(), &struct{}{}, func() {}} {
			valid := annotation == "int" && IsInt(value) || annotation == "string" && IsString(value) || annotation == "bool" && IsBool(value)
			if valid {
				continue
			}
			func() {
				defer func() {
					if recover() == nil {
						t.Errorf("%s accepted %#v", annotation, value)
					}
				}()
				ScalarParameter(value, annotation)
			}()
		}
	}
}
