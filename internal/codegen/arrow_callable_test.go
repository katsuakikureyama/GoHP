package codegen

import (
	"reflect"
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/frontend"
	"github.com/katsuakikureyama/GoHP/internal/ir"
)

func TestArrowCallableIRAndLowering(t *testing.T) {
	p, err := frontend.Parse(`<?php
function forward(array $items, CaLlAbLe $callback, $extra): bool {
    return later($items, $callback);
}
function later(array $items, callable $callback): bool {
    foreach ($items as $key => $value) {
        if (!$callback($value, $key)) { return false; }
    }
    return true;
}
$callback = Fn($value, $key) => !is_string($key);
echo forward([-5 => 'value'], $callback, null) ? 'yes' : 'no';
`, testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(p.Functions[0].ParamTypes, []string{"array", "callable", ""}) {
		t.Fatalf("annotation alignment: %#v", p.Functions[0])
	}
	arrow := p.Body[0].Expr
	if arrow.Kind != "arrow" || !reflect.DeepEqual(arrow.Params, []string{"value", "key"}) || len(arrow.Args) != 1 {
		t.Fatalf("arrow IR: %#v", arrow)
	}
	output, mapping, err := GenerateWithMap(p)
	if err != nil {
		t.Fatal(err)
	}
	text := string(output)
	for _, want := range []string{`rt.CallableParameter(args[1])`, `rt.NewCallable(2, func(args ...any) any`, `rt.Invoke(v["callback"], v["value"], v["key"])`} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %s:\n%s", want, text)
		}
	}
	for line, content := range strings.Split(text, "\n") {
		if strings.Contains(content, `rt.NewCallable(2,`) {
			if got, ok := mapping.PHPLine(line + 1); !ok || got != 11 {
				t.Fatalf("arrow assignment source map: %d %v", got, ok)
			}
			return
		}
	}
	t.Fatal("missing arrow assignment")
}

func TestArrowCallableUnsupportedSource(t *testing.T) {
	for _, tc := range []struct{ name, source string }{
		{"reserved_function", "function FN() { return true; }"},
		{"capture", "$outer = 1; $cb = fn($x) => $x + $outer;"},
		{"nested_capture", "$cb = fn($x) => fn($y) => $x + $y;"},
		{"undefined_dead_capture", "$cb = fn($x) => false ? $missing : $x;"},
		{"this_capture", "class X { public function make() { return fn() => $this; } }"},
		{"duplicate", "$cb = fn($x, $x) => $x;"},
		{"reserved", "$cb = fn($GLOBALS) => $GLOBALS;"},
		{"typed_arrow", "$cb = fn(array $x) => $x;"},
		{"return_type", "$cb = fn($x): bool => true;"},
		{"default", "$cb = fn($x = 1) => $x;"},
		{"reference", "$cb = fn(&$x) => $x;"},
		{"variadic", "$cb = fn(...$x) => $x;"},
		{"static", "$cb = static fn($x) => $x;"},
		{"anonymous", "$cb = function($x) { return $x; };"},
		{"direct_literal_call", "echo (fn() => true)();"},
		{"index_call", "$cb = [fn() => true]; echo $cb[0]();"},
		{"named_args", "$cb = fn($x) => $x; echo $cb(x: 1);"},
		{"string_call", "$cb = 'is_array'; echo $cb([]);"},
		{"array_call", "$cb = []; $result = $cb();"},
		{"null_call", "$cb = null; $result = $cb();"},
		{"undefined_call", "$result = $cb();"},
		{"string_parameter", "function use_cb(callable $cb) { return true; } $result = use_cb('is_array');"},
		{"null_parameter", "function use_cb(callable $cb) { return true; } $result = use_cb(null);"},
		{"dead_wrong_parameter", "function use_cb(callable $cb) { return true; } if(false) { $result = use_cb([]); }"},
		{"forward_wrong_parameter", "function first() { return later(false); } function later(callable $cb) { return true; }"},
		{"method_parameter", "class X { public function run(callable $cb) { return $cb(); } }"},
		{"nullable_parameter", "function run(?callable $cb) { return true; }"},
		{"callable_return", "function make(): callable { return fn() => true; }"},
		{"array_map_unchanged", "$result = array_map(fn($x) => $x, []);"},
		{"bool_domain", "$cb = fn() => true; echo (bool)$cb;"},
		{"strict_domain", "$cb = fn() => true; echo $cb === $cb;"},
		{"arithmetic_domain", "$cb = fn() => true; echo $cb + 1;"},
		{"echo_domain", "echo fn() => true;"},
		{"array_parameter", "function take(array $x) { return $x; } $result = take(fn() => true);"},
		{"bool_return", "function take(): bool { return fn() => true; }"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, err := frontend.Parse("<?php "+tc.source, testTarget(t))
			if err == nil {
				var output []byte
				output, err = Generate(p)
				if len(output) != 0 {
					t.Fatal("unsupported source generated output")
				}
			}
			if err == nil || !strings.Contains(err.Error(), "PHP00") {
				t.Fatalf("unsupported source accepted: %v", err)
			}
		})
	}
}

func TestArrowCallableMalformedIR(t *testing.T) {
	boolean := ir.Expr{Kind: "bool", Text: "true", Line: 1}
	for _, expr := range []ir.Expr{
		{Kind: "arrow", Line: 1},
		{Kind: "arrow", Text: "bool", Args: []ir.Expr{boolean}, Line: 1},
		{Kind: "arrow", Args: []ir.Expr{boolean, boolean}, Line: 1},
		{Kind: "bool", Text: "true", Params: []string{}, Line: 1},
		{Kind: "invoke", Line: 1},
		{Kind: "invoke", Args: []ir.Expr{boolean}, Line: 1},
		{Kind: "invoke", Text: "registry_name", Args: []ir.Expr{{Kind: "var", Text: "cb", Line: 1}}, Line: 1},
	} {
		p := ir.Program{Version: ir.Version, PHPVersion: testTarget(t), Body: []ir.Stmt{{Kind: "assign", Name: "result", Expr: expr, Line: 1}}}
		if output, err := Generate(p); err == nil || !strings.Contains(err.Error(), "IR002") || len(output) != 0 {
			t.Fatalf("malformed IR accepted: %#v: %v", expr, err)
		}
	}
}
