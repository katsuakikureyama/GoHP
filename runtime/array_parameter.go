package runtime

// ArrayParameter checks the entry type and detaches supported array values.
// The annotation constrains the argument, not later local reassignment.
// There is no coercion, reference binding or general object cloning.
func ArrayParameter(value any) any {
	array, ok := value.(Array)
	if !ok {
		panic("array parameter requires array")
	}
	return Clone(array)
}
