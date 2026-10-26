package tests

import "testing"

func TestCallStatementRuntimeGuards(t *testing.T) {
	checkArrayRuntimeGuards(t, []arrayRuntimeCase{
		{"builtin_domain", "<?php function f($x) { count($x); } f(false);", "count requires array"},
		{"typed_return", "<?php function identity($x) { return $x; } function f(): array { return identity(false); } f();", "array return requires array"},
		{"bool_return", "<?php function identity($x) { return $x; } function f(): bool { return identity('x'); } f();", "bool return requires bool"},
		{"method_body", "<?php class X { public function f($x) { count($x); } } $x = new X(); $x->f(false);", "count requires array"},
		{"arrow_arity", "<?php $f = fn($x) => $x; $f();", "callable argument count mismatch"},
		{"dynamic_callable", "<?php function invoke($f) { $f(); } invoke(null);", "callable requires generated arrow"},
		{"stateful_init", "<?php strtok(',');", "strtok continuation requires initialized state"},
	})
}
