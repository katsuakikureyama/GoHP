package runtime

// StripSlashes accepts actual strings only. It consumes backslash pairs once,
// interpreting only an escaped ASCII zero specially, not C-style escapes.
func StripSlashes(value any) any {
	input, ok := value.(string)
	if !ok {
		panic("stripslashes requires a string argument")
	}
	output := make([]byte, 0, len(input))
	for i := 0; i < len(input); i++ {
		b := input[i]
		if b == '\\' {
			i++
			if i == len(input) {
				break
			}
			b = input[i]
			if b == '0' {
				b = 0
			}
		}
		output = append(output, b)
	}
	return string(output)
}
