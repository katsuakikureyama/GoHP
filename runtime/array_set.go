package runtime

import "math"

// ArraySet returns an updated copy of an array after assigning value at an
// explicit integer or string key. It deliberately excludes append semantics:
// callers must provide a key. Both the source array and assigned nested array
// values are copied using the runtime's finite value-copy model.
func ArraySet(array, key, value any) any {
	input, ok := array.(Array)
	if !ok {
		panic("array_set requires array")
	}

	key = Key(key)
	output := Clone(input).(Array)
	if numericKey, ok := key.(int64); ok && numericKey >= output.next {
		if numericKey == math.MaxInt64 {
			output.nextBlocked = true
		} else {
			output.next = numericKey + 1
		}
	}

	for i := range output.Entries {
		if output.Entries[i].Key == key {
			output.Entries[i].Value = Clone(value)
			return output
		}
	}
	output.Entries = append(output.Entries, Entry{Key: key, Value: Clone(value)})
	return output
}
