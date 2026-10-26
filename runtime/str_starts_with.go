package runtime

import "strings"

// StrStartsWith compares bytes at position zero, not characters or tokens.
// Validate both arguments even when an empty needle would otherwise match.
func StrStartsWith(haystack, needle any) any {
	input, inputOK := haystack.(string)
	search, searchOK := needle.(string)
	if !inputOK || !searchOK {
		panic("str_starts_with requires two strings")
	}
	return strings.HasPrefix(input, search)
}
