package runtime

// Execution owns mutable PHP state for one logical execution. Use a fresh
// instance for each request; share its pointer through synchronous nested calls.
// It must not be copied or used concurrently. No process-global PHP state exists.
type Execution struct {
	strtokText  string
	strtokPos   int
	strtokReady bool
}

func RequireExecution(execution *Execution) {
	if execution == nil {
		panic("execution context is required")
	}
}

// Strtok accepts only actual strings: (input, delimiters) starts/replaces a
// scan, (delimiters) continues it. Delimiters are bytes, not a substring/runes.
// Warning-producing uninitialized continuation is outside the closed subset.
func (execution *Execution) Strtok(args ...any) any {
	RequireExecution(execution)
	if len(args) != 1 && len(args) != 2 {
		panic("strtok requires one or two strings")
	}
	for _, arg := range args {
		if _, ok := arg.(string); !ok {
			panic("strtok requires actual strings")
		}
	}
	if len(args) == 2 {
		execution.strtokText = args[0].(string)
		execution.strtokPos = 0
		execution.strtokReady = true
	}
	if !execution.strtokReady {
		panic("strtok continuation requires initialized state")
	}
	text, pos := execution.strtokText, execution.strtokPos
	if pos >= len(text) {
		return false
	}
	var delimiters [256]bool
	for _, b := range []byte(args[len(args)-1].(string)) {
		delimiters[b] = true
	}
	for pos < len(text) && delimiters[text[pos]] {
		pos++
	}
	if pos == len(text) {
		// PHP 8.3 clears its scan when the remaining bytes are all delimiters.
		// A further one-argument call would warn; reject that continuation.
		execution.strtokText = ""
		execution.strtokPos = 0
		execution.strtokReady = false
		return false
	}
	start := pos
	for pos < len(text) && !delimiters[text[pos]] {
		pos++
	}
	// Consume exactly one boundary byte. A later call may change delimiters.
	execution.strtokPos = pos
	if pos < len(text) {
		execution.strtokPos++
	}
	return text[start:pos]
}
