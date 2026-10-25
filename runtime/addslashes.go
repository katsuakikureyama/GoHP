package runtime

// AddSlashes implements byte-wise quoting for actual strings only. This is
// not SQL/HTML/shell escaping and does not perform PHP's weak coercions.
func AddSlashes(value any) any {
	input, ok := value.(string)
	if !ok {
		panic("addslashes requires a string argument")
	}
	output := make([]byte, 0, len(input))
	for i := 0; i < len(input); i++ {
		b := input[i]
		switch b {
		case 0:
			output = append(output, '\\', '0')
		case '\'', '"', '\\':
			output = append(output, '\\', b)
		default:
			output = append(output, b)
		}
	}
	return string(output)
}
