package runtime

import "testing"

func TestStrCaseCmpASCIIByteResults(t *testing.T) {
	for _, tt := range []struct {
		name        string
		left, right string
		want        int64
	}{
		{name: "empty", left: "", right: "", want: 0},
		{name: "ascii_case_equal", left: "Alpha", right: "aLPHA", want: 0},
		{name: "first_folded_byte_difference", left: "AZ", right: "ab", want: 24},
		{name: "ascending", left: "a", right: "B", want: -1},
		{name: "descending", left: "B", right: "a", want: 1},
		{name: "left_prefix", left: "abc", right: "abcd", want: -1},
		{name: "right_prefix", left: "abcd", right: "abc", want: 1},
		{name: "nul_case_equal", left: "A\x00B", right: "a\x00b", want: 0},
		{name: "nul_byte_difference", left: "a\x00b", right: "a\x00c", want: -1},
		{name: "raw_high_bytes_are_not_folded", left: string([]byte{0xc4}), right: string([]byte{0xe4}), want: -32},
		{name: "utf8_non_ascii_is_not_folded", left: "Ä", right: "ä", want: -32},
		{name: "ascii_after_utf8_is_folded", left: "ÄA", right: "Äa", want: 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := StrCaseCmp(tt.left, tt.right).(int64)
			if !ok || got != tt.want {
				t.Fatalf("StrCaseCmp(%q, %q) = %v (int64=%v), want %d", tt.left, tt.right, got, ok, tt.want)
			}
		})
	}
}

func TestStrCaseCmpRejectsNonStrings(t *testing.T) {
	invalid := []struct {
		name  string
		value any
	}{
		{name: "nil", value: nil},
		{name: "boolean", value: true},
		{name: "integer", value: int64(1)},
		{name: "array", value: NewArray()},
		{name: "object", value: &struct{}{}},
		{name: "float", value: 1.5},
	}
	for _, tt := range invalid {
		t.Run("left_"+tt.name, func(t *testing.T) {
			defer func() {
				if got := recover(); got != "strcasecmp requires two strings" {
					t.Fatalf("left non-string was not explicitly rejected: %v", got)
				}
			}()
			StrCaseCmp(tt.value, "right")
		})
		t.Run("right_"+tt.name, func(t *testing.T) {
			defer func() {
				if got := recover(); got != "strcasecmp requires two strings" {
					t.Fatalf("right non-string was not explicitly rejected: %v", got)
				}
			}()
			StrCaseCmp("left", tt.value)
		})
	}
}
