package runtime

// ArrayPushLocal updates one existing generated local binding synchronously.
// Generated code snapshots all value arguments before entering this operation;
// it is not a PHP reference cell, persistent alias or general address API.
func ArrayPushLocal(locals map[string]any, name string, values ...any) int64 {
	current, present := locals[name]
	if !present {
		panic("array_push requires an existing local")
	}
	return pushArrayValues(current, func(value any) { locals[name] = value }, values)
}

// Push reads the current property after argument evaluation, not the value
// seen when its owner/name was resolved as the first call argument.
func (target NestedProperty) Push(values ...any) int64 {
	owner := target.object()
	return pushArrayValues(owner.Get(target.name), func(value any) { owner.Set(target.name, value) }, values)
}

func pushArrayValues(current any, publish func(any), values []any) int64 {
	array, ok := current.(Array)
	if !ok {
		panic("array_push requires array")
	}
	for _, value := range values {
		// Preserve existing next-index and nested-copy/object-identity rules.
		// Publish each successful append; a later overflow does not roll back
		// earlier appends. PHP exception/diagnostic equivalence is not modeled.
		array = ArrayAppend(array, value).(Array)
		publish(array)
	}
	return int64(len(array.Entries))
}
