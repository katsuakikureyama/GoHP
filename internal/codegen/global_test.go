package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
)

// Main shares the cell of every name a function imports, so main must reject
// uses that would write a separate local instead of silently diverging.
func TestMainRejectsNonPlainUseOfFunctionImportedGlobal(t *testing.T) {
	src := "<?php\nfunction f() { global $c; return $c; }\n$xs = [1, 2];\nforeach ($xs as $c) { echo $c; }\necho gettype(f());\n"
	p, err := frontend.Parse(src, "8.3")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Generate(p); err == nil || !strings.Contains(err.Error(), "global-bound") {
		t.Fatalf("main foreach binding of a function-imported global must reject, got %v", err)
	}
}

// A read after a function-prologue import must be generated (the imported name
// is defined), and both scopes must go through the shared cell.
func TestGlobalImportReadIsDefinedAndShared(t *testing.T) {
	src := "<?php\nfunction f() { global $n; $n = $n + 1; }\nf();\nglobal $n;\necho $n;\n"
	p, err := frontend.Parse(src, "8.3")
	if err != nil {
		t.Fatal(err)
	}
	out, err := Generate(p)
	if err != nil {
		t.Fatal(err)
	}
	code := string(out)
	if strings.Contains(code, `v["n"]`) || !strings.Contains(code, `rt.ReadGlobal(execution, "n")`) && !strings.Contains(code, `rt.ReadGlobal(execution,"n")`) {
		t.Fatalf("global-bound $n must use the shared cell only:\n%s", code)
	}
}
