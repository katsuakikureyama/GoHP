package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestUnknownObjectSnapshotLoweringAndSourceMap(t *testing.T) {
	const source = `<?php
function snapshot($value): array {
    return (array)$value;
}
function forward($value) { return $value; }
class Box { public $value; }
class Frame { public $block; }
$block = new Box();
$frame = new Frame();
$frame->block = $block;
$stack = [$frame];
$top = $stack[0];
$copy = (array)$top->block;
$callback = fn($value) => (array)$value;
$result = $callback($block);
$returned = (array)forward($block);
`
	p, err := frontend.Parse(source, testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	generated, sourceMap, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	text := string(generated)
	for fragment, wantLine := range map[string]int{
		`return rt.ArrayReturn(rt.ObjectToArray(v["value"]))`:                                            3,
		`v["copy"] = rt.Clone(rt.ObjectToArray((v["top"]).(interface{ Get(string) any }).Get("block")))`: 13,
		`return rt.ObjectToArray(v["value"])`:                                                            14,
		`v["returned"] = rt.Clone(rt.ObjectToArray(F_forward(v["block"])))`:                              16,
	} {
		if strings.Count(text, fragment) != 1 {
			t.Fatalf("expected one %q:\n%s", fragment, generated)
		}
		for line, generatedLine := range strings.Split(text, "\n") {
			if strings.Contains(generatedLine, fragment) {
				if got, ok := sourceMap.PHPLine(line + 1); !ok || got != wantLine {
					t.Fatalf("%q maps to %d/%v, want %d", fragment, got, ok, wantLine)
				}
			}
		}
	}
}

func TestUnknownObjectSnapshotBoundaries(t *testing.T) {
	for name, source := range map[string]string{
		"null":             "echo (array)null;",
		"bool":             "echo (array)false;",
		"int":              "echo (array)1;",
		"string":           "echo (array)'text';",
		"array":            "echo (array)[];",
		"closure":          "$callback = fn() => 1; echo (array)$callback;",
		"typed_array":      "function snapshot(array $value) { return (array)$value; }",
		"typed_scalar":     "function snapshot(string $value) { return (array)$value; }",
		"undefined":        "$copy = (array)$missing;",
		"conditional_only": "class Box {} if (true) { $box = new Box(); } $copy = (array)$box;",
		"dead_invalid":     "echo true || count((array)false);",
	} {
		t.Run(name, func(t *testing.T) {
			p, err := frontend.Parse("<?php "+source, testTarget(t))
			var output []byte
			if err == nil {
				output, err = Generate(p)
			}
			if err == nil || len(output) != 0 || !strings.Contains(err.Error(), "PHP00") {
				t.Fatalf("unsupported cast generated %s: %v", output, err)
			}
		})
	}
}

func TestUnknownObjectSnapshotMalformedIR(t *testing.T) {
	variable := ir.Expr{Kind: "var", Text: "value", Line: 2}
	for name, expression := range map[string]ir.Expr{
		"no_operand":   {Kind: "object_array_cast", Line: 2},
		"two_operands": {Kind: "object_array_cast", Line: 2, Args: []ir.Expr{variable, variable}},
		"text":         {Kind: "object_array_cast", Line: 2, Text: "Box", Args: []ir.Expr{variable}},
		"params":       {Kind: "object_array_cast", Line: 2, Params: []string{}, Args: []ir.Expr{variable}},
	} {
		t.Run(name, func(t *testing.T) {
			p := ir.Program{Version: ir.Version, PHPVersion: testTarget(t), Functions: []ir.Function{{
				Name: "snapshot", Params: []string{"value"}, Line: 1,
				Body: []ir.Stmt{{Kind: "return", Line: 2, Expr: expression}},
			}}}
			output, err := Generate(p)
			if err == nil || len(output) != 0 || !strings.Contains(err.Error(), "IR002") {
				t.Fatalf("malformed unknown cast generated %s: %v", output, err)
			}
		})
	}
}
