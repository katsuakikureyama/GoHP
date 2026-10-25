package runtime

// ArrayAppend adds a value at the next PHP 8.3 integer key of an existing array.
// ArraySet supplies detached root/value copies; no input storage is mutated.
func ArrayAppend(base, value any) any {
	input, ok := base.(Array)
	if !ok {
		panic("array append requires array")
	}
	if input.nextBlocked {
		panic("array key overflow unsupported")
	}
	return ArraySet(input, input.next, value)
}
