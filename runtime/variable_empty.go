package runtime

// VariableEmpty implements PHP empty() for the closed value domain.
// It covers null, bool, int64, string, and Array. Objects are not empty.
// The legacy ArrayEmpty entry remains strictly array-only.
func VariableEmpty(value any) bool {
	switch input := value.(type) {
	case nil:
		return true
	case bool:
		return !input
	case int64:
		return input == 0
	case string:
		return input == "" || input == "0"
	case Array:
		return ArrayEmpty(input)
	default:
		// Objects and other runtime values are not empty.
		return false
	}
}
