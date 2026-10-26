package runtime

// ArraySet returns a detached array value for a one-level root binding write.
// It never autovivifies, appends implicitly, or mutates the input's storage.
func ArraySet(base, key, value any) any {
	input, ok := base.(Array)
	if !ok {
		panic("array assignment requires array")
	}
	key = Key(key) // Only actual int64/string keys, using existing normalization.
	output := Clone(input).(Array)
	value = Clone(value)
	for i, entry := range output.Entries {
		if entry.Key == key {
			output.Entries[i].Value = value
			return output
		}
	}
	if n, ok := key.(int64); ok {
		output.index.Observe(n)
	}
	output.Entries = append(output.Entries, Entry{Key: key, Value: value})
	return output
}
