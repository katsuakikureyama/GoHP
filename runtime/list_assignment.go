package runtime

// ListValues implements the admitted quiet positional-list subset. Only real
// arrays with every requested numeric key are accepted. It returns detached
// values before the caller writes any destination, including an RHS alias.
// Missing-key warnings/scalar unpacking and general PHP references are excluded.
func ListValues(value any, count int) []any {
	if count <= 0 {
		panic("list assignment requires targets")
	}
	array, ok := value.(Array)
	if !ok {
		panic("list assignment requires array")
	}
	if count > len(array.Entries) {
		panic("list assignment requires present numeric keys")
	}
	values := make([]any, count)
	for i := range values {
		found := false
		for _, entry := range array.Entries {
			if key, ok := entry.Key.(int64); ok && key == int64(i) {
				values[i], found = Clone(entry.Value), true
				break
			}
		}
		if !found {
			panic("list assignment requires present numeric keys")
		}
	}
	return values
}
