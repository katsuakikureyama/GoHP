package runtime

// ArrayIsset implements the bounded isset($array[$key]) form. It checks only
// a defined, one-level Array binding with int64/string keys; nil is treated as
// not set, while every other supported stored value (including false and an
// empty Array) is set. The array is read without copying or mutating it.
func ArrayIsset(base, key any) bool {
	a, ok := base.(Array)
	if !ok {
		panic("isset requires array")
	}
	switch key.(type) {
	case int64, string:
	default:
		panic("isset requires integer or string key")
	}
	key = Key(key)
	for _, entry := range a.Entries {
		if entry.Key == key {
			return entry.Value != nil
		}
	}
	return false
}
