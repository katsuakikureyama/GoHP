package runtime

// Substr implements the byte-oriented bounds of PHP's substr for the actual
// runtime domain. Inputs are validated before applying empty-selection rules.
func Substr(value, offsetValue, lengthValue any) any {
	input, ok := value.(string)
	if !ok {
		panic("substr requires string")
	}
	offset, ok := offsetValue.(int64)
	if !ok {
		panic("substr offset requires int")
	}
	var length int64
	if lengthValue != nil {
		length, ok = lengthValue.(int64)
		if !ok {
			panic("substr length requires int or null")
		}
	}

	n := int64(len(input))
	start := offset
	if start < 0 {
		if start < -n {
			start = 0
		} else {
			start += n
		}
	} else if start > n {
		start = n
	}

	end := n
	if lengthValue != nil {
		if length >= 0 {
			if length < n-start {
				end = start + length
			}
		} else {
			end = n + length
			if end < start {
				end = start
			}
		}
	}
	return input[start:end]
}
