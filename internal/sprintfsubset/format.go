// Package sprintfsubset defines the shared compile-time and runtime format boundary.
package sprintfsubset

// Width accepts only %0Ns with canonical decimal N in [0, 4096]. The bound
// limits padding allocation; it is an explicit subset, not a PHP width limit.
func Width(format string) (int, bool) {
	if len(format) < 4 || len(format) > 7 || format[:2] != "%0" || format[len(format)-1] != 's' {
		return 0, false
	}
	digits := format[2 : len(format)-1]
	if len(digits) > 1 && digits[0] == '0' {
		return 0, false
	}
	width := 0
	for i := range digits {
		if digits[i] < '0' || digits[i] > '9' {
			return 0, false
		}
		width = width*10 + int(digits[i]-'0')
	}
	return width, width <= 4096
}
