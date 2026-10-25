package runtime

// ArrayMap implements only one input array with a statically selected,
// one-value callback. Keys, order and next-index metadata are preserved.
func ArrayMap(callback func(any) any, value any) any {
	input, ok := value.(Array)
	if !ok {
		panic("array_map requires an array")
	}
	if callback == nil {
		panic("array_map requires a callback")
	}
	output := input
	output.Entries = make([]Entry, len(input.Entries))
	for i, entry := range input.Entries {
		output.Entries[i] = Entry{Key: entry.Key, Value: Clone(callback(Clone(entry.Value)))}
	}
	return output
}
