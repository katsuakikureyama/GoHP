package runtime

// IssetValue probes a defined variable, without converting false-like values.
func IssetValue(value any) bool { return value != nil }
