package runtime

import (
	"strings"
	"testing"
)

func TestStrReplaceStrings(t *testing.T) {
	for _, tt := range []struct {
		name        string
		search      string
		replacement string
		subject     string
		want        string
	}{
		{name: "replace_all_non_overlapping", search: "cat", replacement: "dog", subject: "cat cat", want: "dog dog"},
		{name: "empty_search_unchanged", search: "", replacement: "X", subject: "abc", want: "abc"},
		{name: "overlapping_search", search: "aa", replacement: "X", subject: "aaaa", want: "XX"},
		{name: "inserted_search_not_rescanned", search: "a", replacement: "aa", subject: "a", want: "aa"},
		{name: "empty_replacement", search: "-", replacement: "", subject: "a-b-c", want: "abc"},
		{name: "case_sensitive", search: "A", replacement: "x", subject: "aA", want: "ax"},
		{name: "multibyte", search: "猫", replacement: "犬", subject: "猫と猫", want: "犬と犬"},
		{name: "nul_bytes", search: "\x00", replacement: "X", subject: "a\x00b", want: "aXb"},
		{name: "invalid_utf8_bytes", search: string([]byte{0xff}), replacement: "X", subject: string([]byte{'a', 0xff, 'b'}), want: "aXb"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := StrReplace(tt.search, tt.replacement, tt.subject); got != tt.want {
				t.Fatalf("StrReplace(%q, %q, %q) = %#v, want %#v", tt.search, tt.replacement, tt.subject, got, tt.want)
			}
		})
	}
}

func TestStrReplaceRejectsNonStringInEveryPosition(t *testing.T) {
	for _, tt := range []struct {
		name string
		call func()
	}{
		{name: "search", call: func() { StrReplace(int64(1), "x", "subject") }},
		{name: "replacement", call: func() { StrReplace("s", int64(1), "subject") }},
		{name: "subject", call: func() { StrReplace("s", "x", int64(1)) }},
		{name: "empty_search_still_checks_subject", call: func() { StrReplace("", "x", nil) }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if r == nil || !strings.Contains(r.(string), "str_replace requires three string arguments") {
					t.Fatalf("StrReplace did not reject unsupported input: %v", r)
				}
			}()
			tt.call()
		})
	}
}
