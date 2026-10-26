package tests

import "testing"

func TestScalarConstructorRuntimeGuards(t *testing.T) {
	checkArrayRuntimeGuards(t, []arrayRuntimeCase{
		{"int_coercion", `<?php function pass($x) {return $x;} class X {public function __construct(int $x, $bad) {count($bad);} } $x = new X(pass('1'), null);`, "int parameter requires exact type"},
		{"string_coercion", `<?php function pass($x) {return $x;} class X {public function __construct(string $x) {} } $x = new X(pass(1));`, "string parameter requires exact type"},
		{"bool_coercion", `<?php function pass($x) {return $x;} class X {public function __construct(bool $x) {} } $x = new X(pass(0));`, "bool parameter requires exact type"},
		{"null", `<?php function pass($x) {return $x;} class X {public function __construct(int $x) {} } $x = new X(pass(null));`, "int parameter requires exact type"},
		{"execution_guard", `<?php function pass($x) {return $x;} class X {public function __construct(string $x) {strtok(',');} } $x = new X(pass(false));`, "string parameter requires exact type"},
		{"local_reassignment", `<?php function pass($x) {return $x;} class X {public function __construct(int $x) {$x = pass(false); echo strlen($x);} } $x = new X(1);`, "strlen requires string"},
	})
}
