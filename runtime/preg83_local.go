package runtime

import "fmt"

// PregMatch83Local is the generated-facing B3 entry for the explicit
// pcre2-php83-v1 profile. It does not receive an output location: the compiler
// publishes its detached captures and return value to distinct legacy locals.
// Unsupported inputs, missing backends and unmeasured errors panic explicitly;
// none of them become PHP false. Only the B2 measured outcome set is returned.
func (execution *Execution) PregMatch83Local(pattern, subject, offset any) Preg83Result {
	RequireExecution(execution)
	p, patternOK := pattern.(string)
	s, subjectOK := subject.(string)
	o, offsetOK := offset.(int64)
	if !patternOK || !subjectOK || !offsetOK {
		panic("experimental preg83 requires actual string/string/int inputs")
	}
	result, err := execution.PregMatch83(p, s, o, Preg83Limits{Match: 100000, Depth: 10000, HeapKiB: 131072})
	if err != nil {
		panic(fmt.Errorf("experimental preg83: %w", err))
	}
	return result
}
