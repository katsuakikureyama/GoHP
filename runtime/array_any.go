package runtime

// ArrayAny implements the bounded array_any contract: an actual Array and a
// generated two-argument callback are required.  The callback receives each
// value as an independent PHP array value copy followed by its stored key.
// Only an actual bool result is accepted; PHP truthiness coercion is outside
// this runtime surface.
func ArrayAny(value any, callback func(any, any) any) bool {
	input, ok := value.(Array)
	if !ok {
		panic("array_any requires an array")
	}
	if callback == nil {
		panic("array_any requires a callback")
	}
	for _, entry := range input.Entries {
		matched, ok := callback(Clone(entry.Value), entry.Key).(bool)
		if !ok {
			panic("array_any callback must return bool")
		}
		if matched {
			return true
		}
	}
	return false
}
