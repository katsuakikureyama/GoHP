package runtime

// ArrayReturn implements the supported array return-type boundary. A typed
// return owns a detached copy of the returned Array, including nested arrays
// and its automatic-key metadata. Values outside the Array domain fail at the
// boundary rather than silently weakening the declared PHP contract.
func ArrayReturn(value any) any {
	array, ok := value.(Array)
	if !ok {
		panic("array return requires array")
	}
	return Clone(array)
}
