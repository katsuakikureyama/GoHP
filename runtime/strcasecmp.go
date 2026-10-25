package runtime

// StrCaseCmp compares actual strings using ASCII-only case folding. The PHP
// 8.3 result is the first differing folded-byte delta, or the length ordering
// when one string is a prefix. No locale, Unicode folding or coercion applies.
func StrCaseCmp(left, right any) any {
	lhs, leftOK := left.(string)
	rhs, rightOK := right.(string)
	if !leftOK || !rightOK {
		panic("strcasecmp requires two strings")
	}
	fold := func(b byte) byte {
		if b >= 'A' && b <= 'Z' {
			return b + ('a' - 'A')
		}
		return b
	}
	for i := 0; i < len(lhs) && i < len(rhs); i++ {
		a, b := fold(lhs[i]), fold(rhs[i])
		if a != b {
			return int64(a) - int64(b)
		}
	}
	if len(lhs) < len(rhs) {
		return int64(-1)
	}
	if len(lhs) > len(rhs) {
		return int64(1)
	}
	return int64(0)
}
