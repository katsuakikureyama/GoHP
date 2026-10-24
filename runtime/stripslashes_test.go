package runtime

import (
	"strings"
	"testing"
)

func TestStripSlashesRuntime(t *testing.T) {
	for _, tt := range []struct {
		name  string
		input string
		want  string
	}{
		{name: "plain", input: "plain", want: "plain"},
		{name: "escaped_quote", input: "O\\'Brien", want: "O'Brien"},
		{name: "escaped_backslash", input: "a\\\\b", want: "a\\b"},
		{name: "arbitrary_following_byte", input: "a\\qb", want: "aqb"},
		{name: "escaped_ascii_zero", input: "a\\0b", want: "a\x00b"},
		{name: "actual_nul_is_preserved", input: "a\x00b", want: "a\x00b"},
		{name: "trailing_backslash_is_dropped", input: "tail\\", want: "tail"},
		{name: "pairs_are_single_pass", input: "\\\\0", want: "\\0"},
		{name: "invalid_utf8_is_byte_safe", input: string([]byte{'a', '\\', 0xff, 'b'}), want: string([]byte{'a', 0xff, 'b'})},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := StripSlashes(tt.input); got != tt.want {
				t.Fatalf("StripSlashes(%q) = %#v, want %#v", tt.input, got, tt.want)
			}
		})
	}

	defer func() {
		r := recover()
		if r == nil || !strings.Contains(r.(string), "stripslashes requires a string argument") {
			t.Fatalf("StripSlashes did not reject unsupported input: %v", r)
		}
	}()
	StripSlashes(int64(1))
}
