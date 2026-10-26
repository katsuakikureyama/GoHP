package runtime

import "strings"

// StrContains checks case-sensitive bytes, including NUL and invalid UTF-8.
// Both domains are checked before the empty-needle success case.
func StrContains(haystack, needle any) any {
	input, inputOK := haystack.(string)
	search, searchOK := needle.(string)
	if !inputOK || !searchOK {
		panic("str_contains requires two strings")
	}
	return strings.Contains(input, search)
}
