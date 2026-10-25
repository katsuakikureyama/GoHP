package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestArrayReturnTypeIsRetainedAndGuardsEveryReturnPath(t *testing.T) {
	const source = `<?php
function guarded($items, $choose): array {
    if ($choose) {
        return [];
    }
    foreach ($items as $item) {
        return [$item];
    }
}
`
	p, err := frontend.Parse(source, "8.3")
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Functions) != 1 || p.Functions[0].ReturnType != "array" {
		t.Fatalf("array annotation was not retained canonically: %#v", p.Functions)
	}
	generated, err := Generate(p)
	if err != nil {
		t.Fatal(err)
	}
	text := string(generated)
	if got := strings.Count(text, "rt.ArrayReturn("); got != 3 {
		t.Fatalf("ArrayReturn lowering count = %d, want returns plus fallthrough:\n%s", got, generated)
	}
	for _, want := range []string{
		"return rt.ArrayReturn(rt.NewArray())",
		"return rt.ArrayReturn(rt.NewArray(rt.Entry{Key: nil, Value: v[\"item\"]}))",
		"return rt.ArrayReturn(nil)",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("typed array return was not guarded as %q:\n%s", want, generated)
		}
	}
}

func TestArrayReturnTypeMixedCaseAndForwardCallLowering(t *testing.T) {
	p, err := frontend.Parse(`<?php
function first($items): ArRaY { return later($items); }
function later($items): array { return $items; }
`, "8.3")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range p.Functions {
		if f.ReturnType != "array" {
			t.Fatalf("%s ReturnType = %q, want array", f.Name, f.ReturnType)
		}
	}
	generated, err := Generate(p)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Count(string(generated), "rt.ArrayReturn("); got != 2 {
		t.Fatalf("typed forward returns were silently unguarded (%d):\n%s", got, generated)
	}
}

func TestArrayReturnTypeMalformedIRAndTypedMethodsFailClosed(t *testing.T) {
	for _, returnType := range []string{"ARRAY", "string", "?array", "array|int"} {
		t.Run(returnType, func(t *testing.T) {
			p := ir.Program{Version: ir.Version, PHPVersion: "8.3", Functions: []ir.Function{{
				Name: "result", ReturnType: returnType, Line: 1,
				Body: []ir.Stmt{{Kind: "return", Line: 1, Expr: ir.Expr{Kind: "array", Line: 1}}},
			}}}
			if _, err := Generate(p); err == nil {
				t.Fatalf("malformed ReturnType %q was accepted", returnType)
			}
		})
	}

	p, err := frontend.Parse(`<?php class Item { public function values(): array { return []; } }`, "8.3")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Generate(p); err == nil {
		t.Fatal("typed method was accepted")
	}
}

func TestArrayReturnUnsupportedSourceFailsClosed(t *testing.T) {
	for _, source := range []string{
		"<?php function result(): string { return ''; }",
		"<?php function result(): ?array { return []; }",
		"<?php function result(): array|int { return []; }",
		"<?php function result(callable $value): array { return $value; }",
		"<?php function result(): array { return; }",
		"<?php function result(): array { return null; }",
		"<?php function result(): array { return false; }",
		"<?php function result(): array { return 0; }",
		"<?php function result(): array { return ''; }",
		"<?php function result(): array { if (false) { return null; } return []; }",
		"<?php class Item {} function result(): array { return new Item(); }",
	} {
		p, err := frontend.Parse(source, "8.3")
		if err == nil {
			var output []byte
			output, err = Generate(p)
			if len(output) != 0 {
				t.Fatalf("unsupported typed source emitted output: %s", source)
			}
		}
		if err == nil || !strings.Contains(err.Error(), "PHP00") {
			t.Fatalf("unsupported array-return source accepted: %s: %v", source, err)
		}
	}
}

func TestArrayReturnRejectsInvalidElseBranches(t *testing.T) {
	for name, body := range map[string]string{
		"dead_else":              "if (true) { return []; } else { return null; }",
		"dead_elseif":            "if (true) { return []; } elseif (false) { return 1; } else { return []; }",
		"dead_else_if":           "if (true) { return []; } else if (false) { return ''; } else { return []; }",
		"nested_foreach_else":    "foreach ([] as $item) { if (true) { return []; } else { return false; } } return [];",
		"independent_else_scope": "$value = ''; if (true) { $value = []; return $value; } else { return $value; }",
	} {
		t.Run(name, func(t *testing.T) {
			p, err := frontend.Parse("<?php function result(): array { "+body+" }", "8.3")
			if err != nil {
				t.Fatal(err)
			}
			output, err := Generate(p)
			if err == nil || !strings.Contains(err.Error(), "array return requires an array value") || len(output) != 0 {
				t.Fatalf("invalid else return must fail before generation: %v\n%s", err, output)
			}
		})
	}
}
