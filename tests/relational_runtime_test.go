package tests

import "testing"

func TestRelationalRuntimeTypeGuards(t *testing.T) {
	checkArrayRuntimeGuards(t, []arrayRuntimeCase{
		{"dynamic_string", "<?php function relation($value) { return $value < 1; } $result = relation('1');", "relational comparison requires int or null operands"},
		{"dynamic_bool", "<?php function relation($value) { return $value > 1; } $result = relation(true);", "relational comparison requires int or null operands"},
		{"dynamic_array", "<?php function relation($value) { return $value < 1; } $result = relation([]);", "relational comparison requires int or null operands"},
		{"null_with_unsupported_rhs", "<?php function relation($value) { return null < $value; } $result = relation('1');", "relational comparison requires int or null operands"},
	})
}
