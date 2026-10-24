package runtime

import "strings"

// StrReplace implements only the three-actual-string form, without PHP
// coercions, array substitutions or the optional by-reference count argument.
func StrReplace(search, replacement, subject any) any {
	needle, searchOK := search.(string)
	replace, replacementOK := replacement.(string)
	input, subjectOK := subject.(string)
	if !searchOK || !replacementOK || !subjectOK {
		panic("str_replace requires three string arguments")
	}
	// Unlike Go's empty-needle insertion semantics, PHP leaves the input alone.
	if needle == "" {
		return input
	}
	return strings.ReplaceAll(input, needle, replace)
}
