package runtime

// IsInt checks the canonical PHP integer representation, without numeric
// conversion. Other Go numeric types are not PHP values in this runtime.
func IsInt(value any) bool { _, ok := value.(int64); return ok }
