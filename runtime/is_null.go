package runtime

// IsNull tests only the canonical nil value, not false-like values or Go values
// containing typed nil pointers/slices/maps. It does not probe address existence.
func IsNull(value any) bool { return value == nil }
