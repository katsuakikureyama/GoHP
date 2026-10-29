package runtime

// VariableEmpty implements the defined-local array/string subset of PHP empty.
// It is not undefined-address suppression or general scalar truth conversion.
// The legacy ArrayEmpty entry remains strictly array-only.
func VariableEmpty(value any) bool {
	switch input := value.(type) {
	case Array:
		return ArrayEmpty(input)
	case string:
		return input == "" || input == "0"
	default:
		panic("empty requires array or string")
	}
}
