package runtime

// EmptyValue implements only empty($definedBareLocal) for the measured array
// and byte-string domains. It is not PHP's general value/address suppression.
func EmptyValue(value any) bool {
	switch value := value.(type) {
	case Array:
		return len(value.Entries) == 0
	case string:
		return value == "" || value == "0"
	default:
		panic("empty requires array or string")
	}
}
