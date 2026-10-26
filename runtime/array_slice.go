package runtime

// ArraySlice selects by insertion-order position, not by key. Validate even
// empty selections, and avoid negating MinInt64 or adding an unbounded length.
func ArraySlice(value, offsetValue, lengthValue, preserveKeys any) any {
	input, ok := value.(Array)
	if !ok {
		panic("array_slice requires array")
	}
	offset, ok := offsetValue.(int64)
	if !ok {
		panic("array_slice offset requires int")
	}
	var length int64
	if lengthValue != nil {
		length, ok = lengthValue.(int64)
		if !ok {
			panic("array_slice length requires int or null")
		}
	}
	preserve, ok := preserveKeys.(bool)
	if !ok {
		panic("array_slice preserve_keys requires bool")
	}
	n := int64(len(input.Entries))
	start := offset
	if start < 0 {
		if start < -n {
			start = 0
		} else {
			start += n
		}
	} else if start > n {
		start = n
	}
	end := n
	if lengthValue != nil {
		if length >= 0 {
			if length < n-start {
				end = start + length
			}
		} else {
			end = n + length
			if end < start {
				end = start
			}
		}
	}
	output := Array{}
	var index int64
	for _, entry := range input.Entries[start:end] {
		key := entry.Key
		if integer, ok := key.(int64); ok {
			if !preserve {
				integer = index
				index++
			}
			key = integer
			output.noteIntegerKey(integer)
		}
		output.Entries = append(output.Entries, Entry{Key: key, Value: Clone(entry.Value)})
	}
	return output
}
