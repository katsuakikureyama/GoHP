package runtime

import "testing"

func TestStrlenByteLengthAndIntegerResult(t *testing.T) {
	for _, tt := range []struct {
		name, value string
		want        int64
	}{
		{"empty", "", 0},
		{"ascii", "hello", 5},
		{"whitespace", " \n\t\r", 4},
		{"multibyte", "日本語", 9},
		{"emoji", "🙂", 4},
		{"combining", "e\u0301", 3},
		{"binary", "a\x00\xffb", 4},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := Strlen(tt.value).(int64)
			if !ok || got != tt.want {
				t.Fatalf("strlen(%q) = %v (int64=%v), want %d bytes", tt.value, got, ok, tt.want)
			}
		})
	}
}

func TestStrlenRejectsNonStringWithoutCoercion(t *testing.T) {
	for _, tt := range []struct {
		name  string
		value any
	}{
		{"integer", int64(123)}, {"boolean", true}, {"null", nil},
		{"array", NewArray()}, {"object", &struct{}{}}, {"float", 1.5},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if got := recover(); got != "strlen requires string" {
					t.Fatalf("non-string was not explicitly rejected: %v", got)
				}
			}()
			Strlen(tt.value)
		})
	}
}
