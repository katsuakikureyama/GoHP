package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestBoolReturnTypeAndSourceMap(t *testing.T) {
	p, err := frontend.Parse(`<?php
function forward($v): BoOl { return predicate($v); }
function predicate($v): bool {
    if ($v) { return TRUE; }
    elseif (false) { return FALSE; }
    else { foreach ([$v] as $item) { return is_bool($item); } }
}
`, testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range p.Functions {
		if f.ReturnType != "bool" {
			t.Fatalf("noncanonical annotation: %#v", f)
		}
	}
	output, sourceMap, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	for needle, wantLine := range map[string]int{
		`return rt.BoolReturn(F_predicate(v["v"]))`:  2,
		`return rt.BoolReturn(true)`:                 4,
		`return rt.BoolReturn(false)`:                5,
		`return rt.BoolReturn(rt.IsBool(v["item"]))`: 6,
	} {
		found := false
		for line, text := range strings.Split(string(output), "\n") {
			if strings.Contains(text, needle) {
				found = true
				if got, ok := sourceMap.PHPLine(line + 1); !ok || got != wantLine {
					t.Fatalf("%s maps to %d (%v), want %d", needle, got, ok, wantLine)
				}
			}
		}
		if !found {
			t.Fatalf("missing guarded return %s:\n%s", needle, output)
		}
	}
	if !strings.Contains(string(output), "return rt.BoolReturn(nil)") {
		t.Fatal("missing fallthrough guard")
	}
}

func TestBoolReturnUnsupportedSource(t *testing.T) {
	for _, source := range []string{
		"<?php function result(): bool { return; }",
		"<?php function result(): bool { return null; }",
		"<?php function result(): bool { return 0; }",
		"<?php function result(): bool { return 1; }",
		"<?php function result(): bool { return 'true'; }",
		"<?php function result(): bool { return []; }",
		"<?php class X {} function result(): bool { return new X(); }",
		"<?php function result(): bool { if (false) { return null; } return true; }",
		"<?php function result(): bool { if (true) { return true; } elseif (false) { return 1; } return false; }",
		"<?php function result(): bool { if (true) { return true; } else { return ''; } }",
		"<?php function result(): bool { foreach ([] as $v) { return []; } return true; }",
		"<?php function result(): boolean { return true; }",
		"<?php function result(): ?bool { return true; }",
		"<?php function result(): bool|int { return true; }",
		"<?php function result(): bool&int { return true; }",
		"<?php function result(): true { return true; }",
		"<?php function result(bool $v): bool { return $v; }",
		"<?php class X { public function result(): bool { return true; } }",
	} {
		t.Run(source, func(t *testing.T) {
			p, err := frontend.Parse(source, testTarget(t))
			if err == nil {
				var output []byte
				output, err = Generate(p)
				if len(output) != 0 {
					t.Fatal("unsupported source emitted output")
				}
			}
			if err == nil || !strings.Contains(err.Error(), "PHP00") {
				t.Fatalf("unsupported source accepted: %v", err)
			}
		})
	}
}

func TestBoolReturnRejectsNoncanonicalIR(t *testing.T) {
	for _, returnType := range []string{"BOOL", "boolean", "?bool", "bool|int", "true"} {
		p := ir.Program{Version: ir.Version, PHPVersion: testTarget(t), Functions: []ir.Function{{
			Name: "result", ReturnType: returnType, Line: 1,
			Body: []ir.Stmt{{Kind: "return", Line: 1, Expr: ir.Expr{Kind: "bool", Text: "true", Line: 1}}},
		}}}
		output, err := Generate(p)
		if err == nil || !strings.Contains(err.Error(), "IR002") || len(output) != 0 {
			t.Fatalf("malformed type %q accepted: %v", returnType, err)
		}
	}
}
