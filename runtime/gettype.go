package runtime

// GetType reports PHP's type label for values in GoHP's closed runtime value
// domain. It deliberately does not reflect over arbitrary Go values: host
// values that have not crossed an explicit runtime boundary are unsupported.
//
// Object is the existing explicit ABI marker for generated class pointers and
// generated arrows. Checking it neither reads properties nor invokes a PHP method
// or callable body.
func GetType(value any) string {
	switch value.(type) {
	case nil:
		return "NULL"
	case bool:
		return "boolean"
	case int64:
		return "integer"
	case string:
		return "string"
	case Array:
		return "array"
	}
	if IsObject(value) {
		return "object"
	}
	panic("gettype requires supported PHP value")
}
