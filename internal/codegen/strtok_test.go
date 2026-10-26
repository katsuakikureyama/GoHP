package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestStrtokLoweringAndSourceMap(t *testing.T) {
	p, err := frontend.Parse("<?php\nfunction token($d) { return StRtOk($d); }\n$t = strtok('a,b', ',');\necho token(',');", testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	out, mapping, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	for fragment, want := range map[string]int{`return rt.Clone(execution.Strtok(v["d"]))`: 2, `execution.Strtok("a,b", ",")`: 3, `FContext_token(execution, ",")`: 4} {
		found := false
		for line, text := range strings.Split(string(out), "\n") {
			if strings.Contains(text, fragment) {
				found = true
				if got, ok := mapping.PHPLine(line + 1); !ok || got != want {
					t.Fatalf("mapping %s: %d %v", fragment, got, ok)
				}
			}
		}
		if !found {
			t.Fatalf("missing %s:\n%s", fragment, out)
		}
	}
	if !strings.Contains(string(out), "return FContext_token(&rt.Execution{}, args...)") {
		t.Fatal("missing isolated entry")
	}
}

func TestStrtokUnsupportedSource(t *testing.T) {
	for label, source := range map[string]string{
		"arity_zero": "echo strtok();", "arity_three": "echo strtok('x', ',', 'y');",
		"null_input": "echo strtok(null, ',');", "null_delimiters": "echo strtok('x', null);",
		"bool": "echo strtok(false);", "int": "echo strtok(1);", "array": "echo strtok([], ',');",
		"object": "class X {} echo strtok(new X(), ',');", "arrow": "echo strtok(fn() => ',');",
		"undefined": "echo strtok($missing);", "dead": "echo true || strtok([], ',');",
		"named": "echo strtok(string: 'x', token: ',');", "unpack": "echo strtok(...['x', ',']);",
		"redeclare": "function StRtOk($x) { return $x; }", "builtin_callback": "$a = array_map('strtok', [',']);",
	} {
		t.Run(label, func(t *testing.T) {
			p, err := frontend.Parse("<?php "+source, testTarget(t))
			var out []byte
			if err == nil {
				out, err = Generate(p)
			}
			if err == nil || len(out) != 0 || !strings.Contains(err.Error(), "PHP00") {
				t.Fatalf("unsupported: %s %v", out, err)
			}
		})
	}
}

func TestStrtokMalformedIR(t *testing.T) {
	for _, expr := range []ir.Expr{
		{Kind: "call", Text: "strtok", Args: []ir.Expr{{Kind: "string", Text: ","}}, Params: []string{"hidden"}},
		{Kind: "call", Text: "strtok", Args: []ir.Expr{{Kind: "string", Text: ",", Args: []ir.Expr{{Kind: "null", Text: "null"}}}}},
	} {
		p := ir.Program{Version: ir.Version, PHPVersion: testTarget(t), Body: []ir.Stmt{{Kind: "echo", Expr: expr}}}
		out, err := Generate(p)
		if err == nil || len(out) != 0 || !strings.Contains(err.Error(), "IR002") {
			t.Fatalf("malformed: %s %v", out, err)
		}
	}
}
