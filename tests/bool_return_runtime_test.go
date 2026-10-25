package tests

import "testing"

func TestBoolReturnRuntimeGuards(t *testing.T) {
	checkArrayRuntimeGuards(t, []arrayRuntimeCase{
		{"unknown_null", "<?php function result($value): bool { return $value; } $result = result(null);", "bool return requires bool"},
		{"unknown_int", "<?php function result($value): bool { return $value; } $result = result(0);", "bool return requires bool"},
		{"unknown_string", "<?php function result($value): bool { return $value; } $result = result('true');", "bool return requires bool"},
		{"unknown_array", "<?php function result($value): bool { return $value; } $result = result([]);", "bool return requires bool"},
		{"unknown_object", "<?php class Value {} function result($value): bool { return $value; } $result = result(new Value());", "bool return requires bool"},
		{"fallthrough", "<?php function result($value): bool { if ($value) { return true; } } $result = result(false);", "bool return requires bool"},
	})
}
