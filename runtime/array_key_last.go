package runtime

// ArrayKeyLast returns the stored key of the last outer entry, or canonical nil
// for an empty array. It does not sort, inspect values, or change array metadata.
func ArrayKeyLast(value any) any {
	array, ok := value.(Array)
	if !ok {
		panic("array_key_last requires array")
	}
	if len(array.Entries) == 0 {
		return nil
	}
	return array.Entries[len(array.Entries)-1].Key
}
