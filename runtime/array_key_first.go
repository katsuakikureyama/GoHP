package runtime

// ArrayKeyFirst returns the first stored insertion-order key, or canonical nil
// for an empty array. Values and automatic-index state are not inspected/changed.
func ArrayKeyFirst(value any) any {
	array, ok := value.(Array)
	if !ok {
		panic("array_key_first requires array")
	}
	if len(array.Entries) == 0 {
		return nil
	}
	return array.Entries[0].Key
}
