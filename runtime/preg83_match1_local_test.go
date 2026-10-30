package runtime

import (
	"fmt"
	"strings"
	"testing"
)

func TestPreg83LocalMatch1RejectsUnmeasuredInputsWithoutStateChange(t *testing.T) {
	for _, tc := range []struct {
		name                     string
		pattern, subject, offset any
	}{
		{name: "pattern_type", pattern: false, subject: "a", offset: int64(0)},
		{name: "subject_type", pattern: "/a/s", subject: int64(1), offset: int64(0)},
		{name: "offset_type", pattern: "/a/s", subject: "a", offset: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			execution := &Execution{preg83Error: 2}
			defer func() {
				if got := recover(); got == nil || !strings.Contains(fmt.Sprint(got), "actual string/string/int") {
					t.Fatalf("panic = %v, want actual string/string/int", got)
				}
				if got := execution.PregLastError83(); got != 2 {
					t.Fatalf("invalid input changed error state to %d, want 2", got)
				}
			}()
			execution.PregMatch83LocalMatch1(tc.pattern, tc.subject, tc.offset)
		})
	}
}

func TestPreg83LocalMatch1RequiresExecution(t *testing.T) {
	var execution *Execution
	preg83RequirePanic(t, func() {
		execution.PregMatch83LocalMatch1("/a/s", "a", int64(0))
	})
}
