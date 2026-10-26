package runtime

import "testing"

func TestStrContainsBytes(t *testing.T) {
	for _, tc := range []struct {
		haystack, needle string
		want             bool
	}{
		{"", "", true}, {"abc", "", true}, {"", "a", false},
		{"abc", "a", true}, {"abc", "b", true}, {"abc", "c", true},
		{"abc", "abc", true}, {"abc", "abcd", false}, {"abc", "d", false},
		{"aBc", "B", true}, {"aBc", "b", false}, {"aaaa", "aaa", true},
		{"猫と犬", "と", true}, {"猫", "\x8c", true}, {"É", "é", false},
		{"a\x00b\xffc", "\x00b\xff", true}, {"\xff\xfe", "\xfe\xff", false},
		{"xwp-elements-123", "wp-elements-", true}, {"WP-ELEMENTS-123", "wp-elements-", false},
	} {
		got := StrContains(tc.haystack, tc.needle)
		if got != tc.want {
			t.Fatalf("%q contains %q: %#v != %v", tc.haystack, tc.needle, got, tc.want)
		}
	}
}

func TestStrContainsRejectsNonStrings(t *testing.T) {
	for _, value := range []any{nil, false, int64(0), NewArray(), &struct{}{}, []byte("x")} {
		for _, args := range [][2]any{{value, ""}, {"", value}} {
			func() {
				defer func() {
					if got := recover(); got != "str_contains requires two strings" {
						t.Fatalf("guard: %v", got)
					}
				}()
				StrContains(args[0], args[1])
			}()
		}
	}
}
