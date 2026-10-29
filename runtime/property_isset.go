package runtime

// PropertyLookupObject is the nonthrowing, read-only generated-property probe
// protocol. Ordinary Get and Set retain their existing strict semantics.
type PropertyLookupObject interface {
	Object
	PHPPropertyLookup(string) (any, bool)
}

// PropertyIsset probes a named path without copying or eagerly reading slots.
// Missing/null slots and supported nonobject intermediates are quietly false;
// false, zero, empty strings and empty arrays in the final slot are set.
func PropertyIsset(base any, path ...string) bool {
	if len(path) == 0 {
		panic("property isset requires a named path")
	}
	value := base
	for _, name := range path {
		switch value.(type) {
		case nil, int64, string, bool, Array:
			return false
		}
		object, ok := value.(PropertyLookupObject)
		if !ok {
			panic("property isset requires generated property lookup")
		}
		if !object.PHPObject() {
			return false
		}
		var exists bool
		value, exists = object.PHPPropertyLookup(name)
		if !exists || value == nil {
			return false
		}
	}
	return true
}
