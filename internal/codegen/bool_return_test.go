package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestBoolReturnTypeRetainedAndGuarded(t *testing.T) {
	const source = `<?php
function predicate($items, $choose): BoOl {
    if ($choose) {
        return false;
    }
    foreach ($items as $item) {
        return $item;
    }
}
function forward($value): bool { return later($value); }
function later($value): bool { return $value; }
`
	p, err := frontend.Parse(source, "8.3")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range p.Functions {
		if f.ReturnType != "bool" {
			t.Fatalf("lost canonical bool type: %#v", f)
		}
	}
	generated, mapping, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	text := string(generated)
	if strings.Count(text, "rt.BoolReturn(") != 5 {
		t.Fatalf("not all explicit/implicit returns guarded:\n%s", text)
	}
	for _, want := range []string{"return rt.BoolReturn(false)", "return rt.BoolReturn(v[\"item\"])", "return rt.BoolReturn(nil)", "return rt.BoolReturn(F_later(v[\"value\"]))"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %s:\n%s", want, text)
		}
	}
	for line, text := range strings.Split(text, "\n") {
		if strings.Contains(text, "return rt.BoolReturn(false)") {
			if actual, ok := mapping.PHPLine(line + 1); !ok || actual != 4 {
				t.Fatalf("return source line = %d, %v", actual, ok)
			}
			return
		}
	}
	t.Fatal("no mapped bool return")
}

func TestBoolReturnUnsupportedSourceFailsClosed(t *testing.T) {
	for _, tc := range []struct{ name, source string }{
		{"bare", "function predicate(): bool { return; }"},
		{"null", "function predicate(): bool { return null; }"},
		{"int", "function predicate(): bool { return 1; }"},
		{"string", "function predicate(): bool { return '0'; }"},
		{"array", "function predicate(): bool { return []; }"},
		{"object", "class Item {} function predicate(): bool { return new Item(); }"},
		{"dead_branch", "function predicate(): bool { if (false) { return 1; } return true; }"},
		{"nullable", "function predicate(): ?bool { return true; }"},
		{"union", "function predicate(): bool|int { return true; }"},
		{"intersection", "function predicate(): bool&int { return true; }"},
		{"alias", "function predicate(): boolean { return true; }"},
		{"literal_type", "function predicate(): true { return true; }"},
		{"typed_parameter", "function predicate(bool $value): bool { return $value; }"},
		{"typed_method", "class Item { public function predicate(): bool { return true; } }"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, err := frontend.Parse("<?php "+tc.source, "8.3")
			if err == nil {
				var output []byte
				output, err = Generate(p)
				if len(output) != 0 {
					t.Fatal("unsupported source produced Go output")
				}
			}
			if err == nil || !strings.Contains(err.Error(), "PHP00") {
				t.Fatalf("unsupported source accepted: %v", err)
			}
		})
	}
}

func TestBoolReturnMalformedIRFailsClosed(t *testing.T) {
	for _, returnType := range []string{"BOOL", "boolean", "?bool", "bool|int", "true"} {
		t.Run(returnType, func(t *testing.T) {
			p := ir.Program{Version: ir.Version, PHPVersion: "8.3", Functions: []ir.Function{{
				Name: "predicate", ReturnType: returnType, Line: 1,
				Body: []ir.Stmt{{Kind: "return", Line: 1, Expr: ir.Expr{Kind: "bool", Text: "true", Line: 1}}},
			}}}
			if out, err := Generate(p); err == nil || len(out) != 0 {
				t.Fatalf("invalid ReturnType accepted: %s %v", out, err)
			}
		})
	}
}
