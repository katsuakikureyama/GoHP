package runtime

// VariableEmpty implements PHP empty() for the closed value domain.
// It covers null, bool, int64, string, Array and generated objects/arrows;
// foreign Go values reject explicitly.
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
		// Generated objects and arrows (PHP Closures) are never empty. Any other
		// Go value is outside the PHP value domain and must not be classified
		// silently.
		if IsObject(value) {
			return false
		}
		panic("empty requires a supported PHP value")
	}
}
