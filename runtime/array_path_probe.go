package runtime

// ArrayPathProbe implements suppressed missing/null reads for a defined variable
// and a bounded array-only index path. It is not general PHP offset access:
// every reached intermediate must be Array or canonical nil, and reached keys
// must be int64/string. Keys after a missing/null step are not inspected.
// A present value is detached using the existing PHP-value copy boundary.
func ArrayPathProbe(base any, keys ...any) (any, bool) {
	value := base
	for _, key := range keys {
		if value == nil {
			return nil, false
		}
		array, ok := value.(Array)
		if !ok {
			panic("array path requires array or null")
		}
		switch key.(type) {
		case int64, string:
		default:
			panic("array path requires integer or string key")
		}
		key = Key(key)
		found := false
		for _, entry := range array.Entries {
			if entry.Key == key {
				value, found = entry.Value, true
				break
			}
		}
		if !found {
			return nil, false
		}
	}
	if value == nil {
		return nil, false
	}
	return Clone(value), true
}

func ArrayPathIsset(base any, keys ...any) bool {
	_, present := ArrayPathProbe(base, keys...)
	return present
}
