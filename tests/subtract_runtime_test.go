package tests

import "testing"

func TestIntegerSubtractionRuntimeGuards(t *testing.T) {
	checkArrayRuntimeGuards(t, []arrayRuntimeCase{
		{"dynamic_string", "<?php function f($a) { return $a - 1; } $r=f('2');", "subtraction requires int operands"},
		{"overflow", "<?php function f($a,$b) { return $a-$b; } $r=f(9223372036854775807,-1);", "integer subtraction overflow to float unsupported"},
	})
}
