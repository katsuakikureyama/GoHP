package runtime

import (
	"fmt"
	"strings"
	"testing"
)

func preg83TestLimits() Preg83Limits {
	return Preg83Limits{Match: 100000, Depth: 10000, HeapKiB: 131072}
}

func preg83RequirePanic(t *testing.T, call func()) {
	t.Helper()
	defer func() {
		if got := recover(); got == nil || !strings.Contains(fmt.Sprint(got), "execution context") {
			t.Fatalf("panic = %v, want execution context", got)
		}
	}()
	call()
}
