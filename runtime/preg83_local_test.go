package runtime

import (
	"fmt"
	"strings"
	"testing"
)

func TestPreg83LocalRejectsUnmeasuredInputs(t *testing.T) {
	for _, tc := range []struct {
		name                     string
		pattern, subject, offset any
		message                  string
	}{
		{"pattern_type", false, "a", int64(0), "actual string/string/int"},
		{"subject_type", "/a/s", int64(1), int64(0), "actual string/string/int"},
		{"offset_type", "/a/s", "a", "0", "actual string/string/int"},
		{"out_of_range", "/a/s", "a", int64(2), "outside experimental preg83 profile"},
		{"invalid_modifier", "/a/i", "a", int64(0), "outside experimental preg83 profile"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			execution := &Execution{preg83Error: 2}
			defer func() {
				if got := recover(); got == nil || !strings.Contains(fmt.Sprint(got), tc.message) {
					t.Fatalf("panic = %v; want %s", got, tc.message)
				}
				if execution.PregLastError83() != 2 {
					t.Fatal("unsupported failure changed state")
				}
			}()
			execution.PregMatch83Local(tc.pattern, tc.subject, tc.offset)
		})
	}
}
