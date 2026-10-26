package runtime

// ArrayReverse reverses outer insertion order into detached array storage.
// Stored keys are already normalized and unique; rebuild append metadata from
// the result keys instead of carrying the input's index history forward.
func ArrayReverse(value, preserveKeys any) any {
	input, ok := value.(Array)
	if !ok {
		panic("array_reverse requires array")
	}
	preserve, ok := preserveKeys.(bool)
	if !ok {
		panic("array_reverse preserve_keys requires bool")
	}
	output := Array{}
	var index int64
	for i := len(input.Entries) - 1; i >= 0; i-- {
		entry := input.Entries[i]
		key := entry.Key
		if n, integer := key.(int64); integer {
			if !preserve {
				n = index
				index++
			}
			key = n
			output.index.Observe(n)
		}
		output.Entries = append(output.Entries, Entry{Key: key, Value: Clone(entry.Value)})
	}
	return output
}
