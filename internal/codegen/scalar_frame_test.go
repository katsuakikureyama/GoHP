package codegen

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestScalarFrameSelection(t *testing.T) {
	for _, tc := range []struct {
		name, source  string
		scalar, value bool
	}{
		{"scalar", `function selected() { $a = 1; $b = $a; $a = 'changed'; return $b; }`, true, false},
		{"control_flow", `function selected() { $i = 0; while ($i !== 3) { $i = $i + 1; if ($i === 1) { continue; } else { break; } } return $i; }`, true, false},
		{"lazy_scalar", `function selected(): bool { $x = null; $y = false; return isset($x, $y) || ($y ? true : !$y); }`, true, false},
		{"bare_return", `function selected() { return; }`, true, false},
		{"fallthrough", `function selected() { $x = 1; }`, true, false},
		{"scalar_parameters", `function selected(int $i, string $s, bool $b) { $copy = $i; $i = 'changed'; return $b ? $s : $i; }`, true, false},
		{"mixed_parameters", `function selected(int $i, $other) { return $i; }`, false, false},
		{"scalar_parameter_array_body", `function selected(int $i) { $i = []; return $i; }`, false, false},
		{"untyped_parameter", `function selected($x) { return $x; }`, false, false},
		{"optional_parameter", `function selected($x = null) { return $x; }`, false, false},
		{"array_parameter", `function selected(array $x) { return $x; }`, false, false},
		{"array_return", `function selected(): array { return []; }`, false, true},
		{"mixed_local", `function selected() { $x = 1; $x = []; return $x; }`, false, true},
		{"array_branch", `function selected() { $x = 1; if (false) { $x = []; } return $x; }`, false, false},
		{"dead_array", `function selected() { return 1; $x = []; }`, false, true},
		{"foreach", `function selected() { foreach ([1] as $x) { return $x; } return null; }`, false, false},
		{"builtin", `function selected() { return strlen('x'); }`, true, false},
		{"user_call", `function other() { return 1; } function selected() { return other(); }`, false, false},
		{"object", `class Item {} function selected() { return new Item(); }`, false, false},
		{"arrow", `function selected() { $f = fn($x) => $x; return $f(1); }`, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, err := frontend.Parse("<?php "+tc.source, testTarget(t))
			if err != nil {
				t.Fatal(err)
			}
			out, err := Generate(p)
			if err != nil {
				t.Fatal(err)
			}
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "generated.go", out, 0)
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Name.Name != "F_selected" {
					continue
				}
				found = true
				body := string(out[fset.Position(fn.Body.Pos()).Offset:fset.Position(fn.Body.End()).Offset])
				if got := strings.Contains(body, "rt.NewScalarFrame()"); got != tc.scalar {
					t.Fatalf("scalar frame = %v, want %v:\n%s", got, tc.scalar, body)
				}
				if got := strings.Contains(body, "map[string]*rt.ValueSlot{}"); got != tc.value {
					t.Fatalf("owned value slots = %v, want %v:\n%s", got, tc.value, body)
				}
				if tc.scalar {
					if strings.Count(body, "defer v.Close()") != 1 || strings.Contains(body, "v[") {
						t.Fatalf("owned frame bypasses read/write or exit cleanup:\n%s", body)
					}
				} else if tc.value {
					if !strings.Contains(body, "defer temporary.Close()") || strings.Contains(body, "v := map[string]any{}") {
						t.Fatalf("owned value slots bypass cleanup or fell back:\n%s", body)
					}
				} else if !strings.Contains(body, "v := map[string]any{}") || strings.Contains(body, "v.Read(") {
					t.Fatalf("legacy value frame changed:\n%s", body)
				}
			}
			if !found {
				t.Fatal("missing selected function")
			}
		})
	}
}

func TestScalarFrameSourceMap(t *testing.T) {
	const source = `<?php
function owned() {
    $x = 1;
    $copy = $x;
    if ($x) { return $copy; }
}
class Box { public function value() { $x = 2; return $x; } }
$x = owned();
`
	p, err := frontend.Parse(source, testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	out, sourceMap, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	for fragment, want := range map[string]int{
		`v.Write("x", rt.Clone(int64(1)))`:       3,
		`v.Write("copy", rt.Clone(v.Read("x")))`: 4,
		`return rt.Clone(v.Read("copy"))`:        5,
		`v["x"] = rt.Clone(int64(2))`:            7,
		`v["x"] = rt.Clone(F_owned())`:           8,
	} {
		found := false
		for index, line := range strings.Split(string(out), "\n") {
			if strings.Contains(line, fragment) {
				found = true
				if got, ok := sourceMap.PHPLine(index + 1); !ok || got != want {
					t.Fatalf("%s maps to %d (%v), want %d", fragment, got, ok, want)
				}
			}
		}
		if !found {
			t.Fatalf("missing %s:\n%s", fragment, out)
		}
	}
	for index, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "rt.NewScalarFrame()") || strings.Contains(line, "defer v.Close()") {
			if got, ok := sourceMap.PHPLine(index + 1); ok {
				t.Fatalf("synthetic frame lifecycle maps to PHP line %d", got)
			}
		}
	}
}

func TestScalarFramePreservesRejection(t *testing.T) {
	for _, source := range []string{
		`<?php function invalid() { return $missing; }`,
		`<?php function invalid() { if (true) { $x = 1; } return $x; }`,
		`<?php function invalid() { $x = 1; $y =& $x; return $y; }`,
		`<?php function invalid() { return true; $x = $missing; }`,
	} {
		p, err := frontend.Parse(source, testTarget(t))
		if err == nil {
			var out []byte
			out, err = Generate(p)
			if len(out) != 0 {
				t.Fatal("invalid source emitted output")
			}
		}
		if err == nil {
			t.Fatalf("invalid source accepted: %s", source)
		}
	}
	p := ir.Program{Version: ir.Version, PHPVersion: testTarget(t), Functions: []ir.Function{{
		Name: "invalid", Line: 1,
		Body: []ir.Stmt{{Kind: "return", Line: 1, Expr: ir.Expr{Kind: "binary", Text: "+", Line: 1}}},
	}}}
	if out, err := Generate(p); err == nil || !strings.Contains(err.Error(), "IR002") || len(out) != 0 {
		t.Fatalf("malformed scalar IR accepted: %v", err)
	}
}
