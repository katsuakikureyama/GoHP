package runtime

// ArrayAll implements the bounded array_all contract: an actual Array and a
// generated two-argument callback are required. The callback receives each
// value as an independent PHP array value copy followed by its stored key.
// Only an actual bool result is accepted; PHP truthiness coercion is outside
// this runtime surface.
func ArrayAll(value any, callback func(any, any) any) bool {
	input, ok := value.(Array)
	if !ok {
		panic("array_all requires an array")
	}
	if callback == nil {
		panic("array_all requires a callback")
	}
	for _, entry := range input.Entries {
		matched, ok := callback(Clone(entry.Value), entry.Key).(bool)
		if !ok {
			panic("array_all callback must return bool")
		}
		if !matched {
			return false
		}
	}
	return true
}
