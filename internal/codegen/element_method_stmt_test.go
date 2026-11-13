package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
)

// `$a[$k]->m();` is a statement-level call on a one-level element receiver.
func TestElementMethodCallStatementAdmitted(t *testing.T) {
	src := "<?php\nfinal class B { public $n; public function __construct() { $this->n = 0; } public function bump() { $this->n = $this->n + 1; } }\n" +
		"$a = array();\n$a['x'] = new B();\n$a['x']->bump();\n$a['x']->bump();\necho $a['x']->n;\n"
	p, err := frontend.Parse(src, "8.3")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Generate(p); err != nil {
		t.Fatalf("one-level element method statement must generate: %v", err)
	}
}

func TestElementMethodCallStatementRejections(t *testing.T) {
	for name, src := range map[string]string{
		"nested_element_receiver": "<?php\nfinal class B { public function m() { return 1; } }\n$a = array();\n$a[0] = array();\n$a[0][1] = new B();\n$a[0][1]->m();\n",
		"property_chain_receiver": "<?php\nfinal class B { public $c; public function m() { return 1; } }\n$a = array();\n$a[0] = new B();\n$a[0]->c->m();\n",
	} {
		t.Run(name, func(t *testing.T) {
			p, err := frontend.Parse(src, "8.3")
			if err != nil {
				return // parser-level rejection is also acceptable
			}
			if _, err := Generate(p); err == nil {
				t.Fatalf("%s must reject", name)
			}
		})
	}
}

// The global element method path reads the shared cell, never a separate local.
func TestGlobalElementMethodCallUsesSharedCell(t *testing.T) {
	src := "<?php\nfinal class B { public function m() { return 1; } }\nfunction f() { global $g; $g['k']->m(); }\nglobal $g;\n$g = array();\n$g['k'] = new B();\nf();\n"
	p, err := frontend.Parse(src, "8.3")
	if err != nil {
		t.Fatal(err)
	}
	out, err := Generate(p)
	if err != nil {
		t.Fatal(err)
	}
	if code := string(out); strings.Contains(code, `v["g"]`) {
		t.Fatalf("global-bound $g must not use a local map slot:\n%s", code)
	}
}
