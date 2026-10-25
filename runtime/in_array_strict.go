package runtime

// InArrayStrict implements only scalar membership with PHP's strict comparison.
// Validate the complete input domain before searching, including late values.
func InArrayStrict(needle, haystack any) any {
	if !inArrayScalar(needle) {
		panic("in_array requires a scalar needle")
	}
	array, ok := haystack.(Array)
	if !ok {
		panic("in_array requires an array haystack")
	}
	for _, entry := range array.Entries {
		if !inArrayScalar(entry.Value) {
			panic("in_array requires scalar haystack values")
		}
	}
	for _, entry := range array.Entries {
		if Identical(needle, entry.Value) {
			return true
		}
	}
	return false
}

func inArrayScalar(value any) bool {
	switch value.(type) {
	case nil, bool, int64, string:
		return true
	default:
		return false
	}
}
