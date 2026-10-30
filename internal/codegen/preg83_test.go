package codegen

import (
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func preg83Program(t *testing.T, source string) ir.Program {
	t.Helper()
	p, err := frontend.ParseWithOptions(source, "8.3", frontend.Options{ExperimentalPCRE2: true})
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestPreg83CompilerLowering(t *testing.T) {
	program := preg83Program(t, `<?php
function inspect(string $subject, int $offset) {
    $matches = null;
    $result = preg_match('/(?P<word>a)/s', $subject, $matches, PREG_OFFSET_CAPTURE, $offset);
    return count($matches) + $result;
}
$matches = null;
$result = preg_match('/(a)/s', 'a', $matches, PREG_OFFSET_CAPTURE, -1);
echo inspect('a', 0);
echo preg_last_error();`)

	output, mapping, err := GenerateWithMap(program)
	if err != nil {
		t.Fatal(err)
	}
	text := string(output)
	for _, want := range []string{
		"PregMatch83Local",
		"PregLastError83",
		`v["matches"]`,
		"rt.Clone",
		"func FContext_inspect(execution *rt.Execution",
		"FContext_inspect(execution",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("generated regex lowering missing %q:\n%s", want, text)
		}
	}

	var matchedStatement, propagatedCall bool
	for line, generated := range strings.Split(text, "\n") {
		if strings.Contains(generated, "PregMatch83Local") {
			phpLine, ok := mapping.PHPLine(line + 1)
			if !ok || (phpLine != 4 && phpLine != 8) {
				t.Fatalf("regex operation source map = %d, %v; want source line 4 or 8", phpLine, ok)
			}
			matchedStatement = true
		}
		if strings.Contains(generated, "FContext_inspect(execution") && !strings.HasPrefix(strings.TrimSpace(generated), "func ") {
			phpLine, ok := mapping.PHPLine(line + 1)
			if !ok || phpLine != 9 {
				t.Fatalf("context call source map = %d, %v; want source line 9", phpLine, ok)
			}
			propagatedCall = true
		}
	}
	if !matchedStatement || !propagatedCall {
		t.Fatalf("missing mapped regex statement/context call:\n%s", text)
	}
}

func TestPreg83CompilerUnsupportedSource(t *testing.T) {
	for label, source := range map[string]string{
		"ordinary_expression":  ` $matches = null; echo preg_match('/a/s', 'a', $matches, PREG_OFFSET_CAPTURE, 0); `,
		"missing_capture_flag": ` $matches = null; $result = preg_match('/a/s', 'a', $matches, 0); `,
		"wrong_capture_flag":   ` $matches = null; $result = preg_match('/a/s', 'a', $matches, 0, 0); `,
		"missing_offset":       ` $matches = null; $result = preg_match('/a/s', 'a', $matches, PREG_OFFSET_CAPTURE); `,
		"subject_not_string":   ` $matches = null; $result = preg_match('/a/s', 1, $matches, PREG_OFFSET_CAPTURE, 0); `,
		"offset_not_int":       ` $matches = null; $result = preg_match('/a/s', 'a', $matches, PREG_OFFSET_CAPTURE, '0'); `,
		"result_output_alias":  ` $result = null; $result = preg_match('/a/s', 'a', $result, PREG_OFFSET_CAPTURE, 0); `,
		"subject_output_alias": ` $subject = 'a'; $result = preg_match('/a/s', $subject, $subject, PREG_OFFSET_CAPTURE, 0); `,
		"result_input_alias":   ` function probe(string $result) { $matches = null; $result = preg_match('/a/s', $result, $matches, PREG_OFFSET_CAPTURE, 0); } `,
		"uninitialized_output": ` $result = preg_match('/a/s', 'a', $matches, PREG_OFFSET_CAPTURE, 0); `,
		"array_output":         ` $matches = []; $result = preg_match('/a/s', 'a', $matches, PREG_OFFSET_CAPTURE, 0); `,
		"derived_null_output":  ` $seed = null; $matches = $seed; $result = preg_match('/a/s', 'a', $matches, PREG_OFFSET_CAPTURE, 0); `,
		"branch_null_output":   ` if (true) { $matches = null; } $result = preg_match('/a/s', 'a', $matches, PREG_OFFSET_CAPTURE, 0); `,
		"property_output":      ` class Box { public $matches; } $box = new Box(); $result = preg_match('/a/s', 'a', $box->matches, PREG_OFFSET_CAPTURE, 0); `,
		"nested_call_subject":  ` $matches = null; $result = preg_match('/a/s', strtolower('A'), $matches, PREG_OFFSET_CAPTURE, 0); `,
		"wrong_modifier":       ` $matches = null; $result = preg_match('/a/i', 'a', $matches, PREG_OFFSET_CAPTURE, 0); `,
		"last_error_arguments": ` echo preg_last_error(0); `,
		"nested_if":            ` $matches = null; if (true) { $result = preg_match('/a/s', 'a', $matches, PREG_OFFSET_CAPTURE, 0); } `,
		"nested_while":         ` $matches = null; while (false) { $result = preg_match('/a/s', 'a', $matches, PREG_OFFSET_CAPTURE, 0); } `,
		"nested_foreach":       ` $matches = null; foreach ([] as $item) { $result = preg_match('/a/s', 'a', $matches, PREG_OFFSET_CAPTURE, 0); } `,
		"redeclare_match":      ` function PrEg_MaTcH() {} `,
		"redeclare_last_error": ` function PREG_LAST_ERROR() {} `,
	} {
		t.Run(label, func(t *testing.T) {
			program, err := frontend.ParseWithOptions("<?php "+source, "8.3", frontend.Options{ExperimentalPCRE2: true})
			var output []byte
			if err == nil {
				output, err = Generate(program)
			}
			if err == nil || len(output) != 0 || !strings.Contains(err.Error(), "PHP00") {
				t.Fatalf("unsupported regex source generated %q: %v", output, err)
			}
		})
	}
}

func TestPreg83CompilerMalformedIR(t *testing.T) {
	base := func() ir.Program {
		return ir.Program{
			Version:      ir.Version,
			PHPVersion:   "8.3",
			RegexProfile: ir.Preg83Profile,
			Body: []ir.Stmt{
				{Kind: "assign", Name: "matches", Line: 1, Expr: ir.Expr{Kind: "null", Text: "null", Line: 1}},
				{Kind: "preg_match_local", Name: "result", Line: 2, Target: ir.Expr{Kind: "var", Text: "matches", Line: 2}, Expr: ir.Expr{Kind: "call", Text: "preg_match", Line: 2, Args: []ir.Expr{
					{Kind: "string", Text: "/a/s", Line: 2},
					{Kind: "string", Text: "a", Line: 2},
					{Kind: "int", Text: "0", Line: 2},
				}}},
			},
		}
	}
	type mutation struct {
		mutate func(*ir.Program)
		want   string
	}
	for label, test := range map[string]mutation{
		"hidden_statement_payload": {func(p *ir.Program) { p.Body[1].Key = "hidden" }, "IR002"},
		"target_arguments": {func(p *ir.Program) {
			p.Body[1].Target.Args = []ir.Expr{{Kind: "int", Text: "0", Line: 2}}
		}, "IR002"},
		"target_parameters":     {func(p *ir.Program) { p.Body[1].Target.Params = []string{"hidden"} }, "IR002"},
		"expression_parameters": {func(p *ir.Program) { p.Body[1].Expr.Params = []string{"hidden"} }, "IR002"},
		"else_body": {func(p *ir.Program) {
			p.Body[1].ElseBody = []ir.Stmt{{Kind: "echo", Line: 2, Expr: ir.Expr{Kind: "string", Text: "hidden", Line: 2}}}
		}, "IR002"},
		"nested_statement_payload": {func(p *ir.Program) {
			p.Body[1].Body = []ir.Stmt{{Kind: "echo", Line: 2, Expr: ir.Expr{Kind: "string", Text: "hidden", Line: 2}}}
		}, "IR002"},
		"pattern_hidden_payload": {func(p *ir.Program) {
			p.Body[1].Expr.Args[0].Args = []ir.Expr{{Kind: "string", Text: "hidden", Line: 2}}
		}, "IR002"},
		"subject_hidden_payload": {func(p *ir.Program) {
			p.Body[1].Expr.Args[1].Args = []ir.Expr{{Kind: "string", Text: "hidden", Line: 2}}
		}, "IR002"},
		"offset_hidden_payload": {func(p *ir.Program) {
			p.Body[1].Expr.Args[2].Params = []string{"hidden"}
		}, "IR002"},
		"wrong_call_name":      {func(p *ir.Program) { p.Body[1].Expr.Text = "strtok" }, "IR002"},
		"wrong_argument_count": {func(p *ir.Program) { p.Body[1].Expr.Args = p.Body[1].Expr.Args[:2] }, "IR002"},
		"result_output_alias":  {func(p *ir.Program) { p.Body[1].Name = "matches" }, "PHP004"},
		"target_input_alias":   {func(p *ir.Program) { p.Body[1].Expr.Args[1] = p.Body[1].Target }, "PHP004"},
		"nested_control_flow": {func(p *ir.Program) {
			statement := p.Body[1]
			p.Body = append(p.Body[:1], ir.Stmt{Kind: "if", Line: 2, Expr: ir.Expr{Kind: "bool", Text: "true", Line: 2}, Body: []ir.Stmt{statement}})
		}, "PHP004"},
	} {
		t.Run(label, func(t *testing.T) {
			program := base()
			test.mutate(&program)
			output, err := Generate(program)
			if err == nil || len(output) != 0 || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("malformed regex IR generated %q: %v", output, err)
			}
		})
	}
}

func TestPreg83CompilerProfileGate(t *testing.T) {
	const source = `<?php $matches = null; $result = preg_match('/a/s', 'a', $matches, PREG_OFFSET_CAPTURE, 0);`

	program, err := frontend.Parse(source, "8.3")
	if err == nil {
		output, generateErr := Generate(program)
		if generateErr == nil || len(output) != 0 || !strings.Contains(generateErr.Error(), "PHP004") {
			t.Fatalf("default profile generated %q: %v", output, generateErr)
		}
	} else if !strings.Contains(err.Error(), "PHP002") {
		t.Fatalf("default profile parse error = %v, want PHP002 or PHP004 rejection", err)
	}
	program, err = frontend.ParseWithOptions(source, "8.4", frontend.Options{ExperimentalPCRE2: true})
	if err == nil {
		output, generateErr := Generate(program)
		if generateErr == nil || len(output) != 0 || !strings.Contains(generateErr.Error(), "PHP004") {
			t.Fatalf("PHP 8.4 profile generated %q: %v", output, generateErr)
		}
	} else if !strings.Contains(err.Error(), "PHP002") && !strings.Contains(err.Error(), "PHP003: experimental regex profile requires PHP 8.3") {
		t.Fatalf("PHP 8.4 profile parse error = %v, want explicit source/profile rejection", err)
	}
	program, err = frontend.Parse(`<?php echo preg_last_error();`, "8.3")
	if err != nil {
		t.Fatal(err)
	}
	if output, generateErr := Generate(program); generateErr == nil || len(output) != 0 || !strings.Contains(generateErr.Error(), "PHP004") {
		t.Fatalf("bare preg_last_error generated %q: %v", output, generateErr)
	}

	program = ir.Program{Version: ir.Version, PHPVersion: "8.3", RegexProfile: "unknown-profile"}
	if output, err := Generate(program); err == nil || len(output) != 0 || !strings.Contains(err.Error(), "IR002") {
		t.Fatalf("unknown profile generated %q: %v", output, err)
	}
}
