package runtime

// ScalarParameter admits exact scalar values only, without PHP weak coercion.
// The annotation applies at entry, not to later local or property assignments.
func ScalarParameter(value any, annotation string) any {
	valid := false
	switch annotation {
	case "int":
		_, valid = value.(int64)
	case "string":
		_, valid = value.(string)
	case "bool":
		_, valid = value.(bool)
	default:
		panic("unsupported scalar parameter annotation")
	}
	if !valid {
		panic(annotation + " parameter requires exact type")
	}
	return value
}
