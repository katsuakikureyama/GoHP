package runtime

// ArrayMerge joins outer entries without modifying its inputs. Integer keys
// become consecutive indexes; repeated string keys replace values in place.
func ArrayMerge(values ...any) any {
	output := Array{}
	positions := make(map[string]int)
	var index int64
	for _, value := range values {
		input, ok := value.(Array)
		if !ok {
			panic("array_merge requires arrays")
		}
		for _, entry := range input.Entries {
			key := entry.Key
			if text, ok := key.(string); ok {
				if position, exists := positions[text]; exists {
					output.Entries[position].Value = Clone(entry.Value)
					continue
				}
				positions[text] = len(output.Entries)
			} else {
				// Array storage already contains normalized int64/string keys.
				key = index
				output.noteIntegerKey(index)
				index++
			}
			output.Entries = append(output.Entries, Entry{Key: key, Value: Clone(entry.Value)})
		}
	}
	return output
}
