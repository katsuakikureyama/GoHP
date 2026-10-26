package runtime

import "testing"

func TestStrStartsWithBytes(t *testing.T) {
	for _, tc := range []struct {
		haystack, needle string
		want             bool
	}{
		{"", "", true}, {"abc", "", true}, {"", "a", false},
		{"abc", "a", true}, {"abc", "ab", true}, {"abc", "abc", true},
		{"abc", "b", false}, {"abc", "bc", false}, {"abc", "abcd", false},
		{"Abc", "a", false}, {"Abc", "A", true}, {"aaaa", "aaa", true},
		{"猫と犬", "猫", true}, {"猫と犬", "犬", false}, {"猫", "\xe7\x8c", true},
		{"É", "é", false}, {"\x00a\xff", "\x00a", true}, {"a\x00b", "\x00b", false},
		{"\xff\xfe", "\xff", true}, {"\xff\xfe", "\xfe", false},
		{"wp-elements-123", "wp-elements-", true}, {"xwp-elements-123", "wp-elements-", false},
		{" wp-elements-123", "wp-elements-", false}, {"WP-ELEMENTS-123", "wp-elements-", false},
	} {
		if got := StrStartsWith(tc.haystack, tc.needle); got != tc.want {
			t.Fatalf("%q starts with %q: %#v != %v", tc.haystack, tc.needle, got, tc.want)
		}
	}
}

func TestStrStartsWithRejectsNonStrings(t *testing.T) {
	for _, value := range []any{nil, false, int64(0), NewArray(), &struct{}{}, []byte("x")} {
		for _, args := range [][2]any{{value, ""}, {"", value}} {
			func() {
				defer func() {
					if got := recover(); got != "str_starts_with requires two strings" {
						t.Fatalf("guard: %v", got)
					}
				}()
				StrStartsWith(args[0], args[1])
			}()
		}
	}
}
