package runtime

import "fmt"

// PregMatch83LocalMatch1 is the generated runtime entry for the
// pcre2-php83-match1-v1 profile. It deliberately has no caller-selectable
// limits: the profile contract fixes Match to one while retaining the B2
// adapter's existing subject, pattern, and engine boundaries.
func (execution *Execution) PregMatch83LocalMatch1(pattern, subject, offset any) Preg83Result {
	RequireExecution(execution)
	p, patternOK := pattern.(string)
	s, subjectOK := subject.(string)
	o, offsetOK := offset.(int64)
	if !patternOK || !subjectOK || !offsetOK {
		panic("experimental preg83 requires actual string/string/int inputs")
	}
	result, err := execution.PregMatch83(p, s, o, Preg83Limits{Match: 1, Depth: 10000, HeapKiB: 131072})
	if err != nil {
		panic(fmt.Errorf("experimental preg83: %w", err))
	}
	return result
}
