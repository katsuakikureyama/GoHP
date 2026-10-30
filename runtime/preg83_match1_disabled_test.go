//go:build !gohp_pcre2 || !cgo

package runtime

import (
	"fmt"
	"strings"
	"testing"
)

func TestPreg83LocalMatch1UnavailablePreservesState(t *testing.T) {
	execution := &Execution{preg83Error: 2}
	defer func() {
		problem := recover()
		if problem == nil || !strings.Contains(fmt.Sprint(problem), "unavailable") {
			t.Fatalf("missing backend must fail explicitly: %v", problem)
		}
		if execution.PregLastError83() != 2 {
			t.Fatal("unavailable engine changed prior error state")
		}
	}()
	execution.PregMatch83LocalMatch1(`/a/s`, "a", int64(0))
}
