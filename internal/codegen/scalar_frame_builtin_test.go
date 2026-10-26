package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestScalarFrameBuiltinSelection(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		owned        bool
	}{
		{"nested", `function f(string $s) { return StRlEn(strtolower(rtrim(substr($s, 0, null)))); }`, true},
		{"compare", `function f(string $s) { return strcasecmp($s, 'x'); }`, true},
		{"contains", `function f(string $s): bool { return str_contains($s, 'x') && str_starts_with($s, 'x'); }`, true},
		{"predicates", `function f(int $i, string $s, bool $b): bool { $n = null; return is_int($i) && is_string($s) && is_bool($b) && is_null($n) && !is_array($s) && !is_object($s); }`, true},
		{"discard", `function f(string $s) { strlen($s); return $s; }`, true},
		{"lazy_unknown", `function f(string $s, bool $b) { $v = $b ? $s : false; return $b && strlen($v); }`, true},
		{"array_operand", `function f() { return is_array([]); }`, false},
		{"object_operand", `class X {} function f() { return is_object(new X()); }`, false},
		{"nested_array", `function f() { return is_int(count([])); }`, false},
		{"user_operand", `function other() { return 'x'; } function f() { return strlen(other()); }`, false},
		{"array_result", `function f(string $s) { return [$s]; }`, false},
		{"unlisted_pure", `function f(string $s) { return addslashes($s); }`, false},
		{"stateful", `function f(string $s) { return strtok($s, ','); }`, false},
		{"discard_user", `function other() { return 'x'; } function f(string $s) { other(); return strlen($s); }`, false},
		{"dead_array", `function f(string $s) { return strlen($s); is_array([]); }`, false},
		{"dead_branch", `function f(string $s) { if(false) { is_array([]); } return strlen($s); }`, false},
		{"untyped", `function f($s) { return strlen($s); }`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, err := frontend.Parse("<?php "+tc.source, testTarget(t))
			if err != nil {
				t.Fatal(err)
			}
			// Other declarations remain available for validation, but do not
			// confuse the storage assertion for the selected function f.
			out, err := Generate(p)
			if err != nil {
				t.Fatal(err)
			}
			s := string(out)
			start := strings.Index(s, "func F_f(")
			if start < 0 {
				t.Fatal("function missing")
			}
			body := s[start:]
			if end := strings.Index(body[1:], "\nfunc "); end >= 0 {
				body = body[:end+1]
			}
			if got := strings.Contains(body, "rt.NewScalarFrame()"); got != tc.owned {
				t.Fatalf("owned = %v, want %v:\n%s", got, tc.owned, body)
			}
		})
	}
}

func TestScalarFrameBuiltinLoweringAndMap(t *testing.T) {
	p, err := frontend.Parse(`<?php
function f(string $text) {
    $n = strlen(strtolower($text));
    substr($text, 0);
    return $n;
}`, testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	out, mapping, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	for fragment, want := range map[string]int{
		`v.Write("n", rt.Clone(rt.Strlen(rt.StrToLower(v.Read("text")))))`: 3,
		`_ = rt.Substr(v.Read("text"), int64(0), nil)`:                     4,
		`return rt.Clone(v.Read("n"))`:                                     5,
	} {
		found := 0
		for i, line := range strings.Split(s, "\n") {
			if strings.Contains(line, fragment) {
				found++
				if got, ok := mapping.PHPLine(i + 1); !ok || got != want {
					t.Fatalf("mapping %s: %d %v", fragment, got, ok)
				}
			}
		}
		if found != 1 {
			t.Fatalf("%s occurs %d times:\n%s", fragment, found, s)
		}
	}
	if strings.Count(s, "defer v.Close()") != 1 {
		t.Fatal("owned function cleanup missing")
	}
}

func TestScalarFrameBuiltinPreservesRejection(t *testing.T) {
	for _, source := range []string{
		`function f() { return strlen(); }`,
		`function f(int $n) { return strtolower($n); }`,
		`function f() { return substr('x', '0'); }`,
		`function f() { return strlen($missing); }`,
		`function f() { return true; strlen(false); }`,
		`function StRlEn(string $s) { return 1; }`,
		`function f() { return missing(); }`,
	} {
		p, err := frontend.Parse("<?php "+source, testTarget(t))
		var out []byte
		if err == nil {
			out, err = Generate(p)
		}
		if err == nil || len(out) != 0 {
			t.Fatalf("accepted %s: %v", source, err)
		}
	}
	p := ir.Program{Version: ir.Version, PHPVersion: testTarget(t), Functions: []ir.Function{{
		Name: "f", Body: []ir.Stmt{{Kind: "return", Expr: ir.Expr{Kind: "call", Text: "strlen", Params: []string{"hidden"}, Args: []ir.Expr{{Kind: "string", Text: "x"}}}}},
	}}}
	if out, err := Generate(p); err == nil || len(out) != 0 {
		t.Fatalf("malformed call accepted: %v", err)
	}
}
