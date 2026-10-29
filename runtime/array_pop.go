package runtime

// ArrayPopLocal synchronously updates one existing generated local. It does
// not create a reference cell or expose a general mutable array expression.
func ArrayPopLocal(locals map[string]any, name string) any {
	current, present := locals[name]
	if !present {
		panic("array_pop requires an existing local")
	}
	return popArrayValue(current, func(value any) { locals[name] = value })
}

// Pop reads the captured owner's current declared slot, not a retained array
// snapshot. Empty arrays return null without publishing a replacement.
func (target NestedProperty) Pop() any {
	owner := target.object()
	return popArrayValue(owner.Get(target.name), func(value any) { owner.Set(target.name, value) })
}

func popArrayValue(current any, publish func(any)) any {
	array, ok := current.(Array)
	if !ok {
		panic("array_pop requires array")
	}
	if len(array.Entries) == 0 {
		return nil
	}
	// Detach both the remaining entries and the returned nested value. Objects
	// still retain their identity under the existing Clone contract.
	output := Clone(array).(Array)
	last := len(output.Entries) - 1
	entry := output.Entries[last]
	output.Entries[last] = Entry{}
	output.Entries = output.Entries[:last]
	if key, integer := entry.Key.(int64); integer {
		output.index.Pop(key)
	}
	publish(output)
	return entry.Value
}
