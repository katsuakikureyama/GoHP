package runtime

// NullDefaultArgs validates a positional arity range and fills omitted trailing
// arguments with canonical null. Caller argument storage is never modified.
func NullDefaultArgs(args []any, required, total int) []any {
	if required < 0 || total < required || len(args) < required || len(args) > total {
		panic("unsupported argument count")
	}
	if len(args) == total {
		return args
	}
	padded := make([]any, total)
	copy(padded, args)
	return padded
}
