package runtime

// BoolReturn enforces the actual-bool-only return subset, not truthiness or
// PHP's weak scalar coercions. Fallthrough nil is rejected by the same guard.
func BoolReturn(value any) any {
	result, ok := value.(bool)
	if !ok {
		panic("bool return requires bool")
	}
	return result
}
