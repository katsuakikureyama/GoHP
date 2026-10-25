package runtime

// ArrayEmpty implements only empty($definedArrayVariable). It checks the outer
// entry count without inspecting values or mutating input. Non-arrays are outside
// this subset even when PHP's general empty construct would accept them.
func ArrayEmpty(value any) bool {
	array, ok := value.(Array)
	if !ok {
		panic("empty requires array")
	}
	return len(array.Entries) == 0
}
