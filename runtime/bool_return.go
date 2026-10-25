package runtime

// BoolReturn enforces the actual-bool subset of a declared PHP bool return.
// This is not Bool's truthiness conversion: weak scalar return coercion is
// deliberately unsupported, and fallthrough/null must not become false.
func BoolReturn(value any) any {
	result, ok := value.(bool)
	if !ok {
		panic("bool return requires bool")
	}
	return result
}
