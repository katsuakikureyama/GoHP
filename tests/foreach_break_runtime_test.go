package tests

import "testing"

func TestForeachBreakRuntimeGuards(t *testing.T) {
	checkArrayRuntimeGuards(t, []arrayRuntimeCase{
		{"changed_type_before_break", "<?php $a = []; foreach ([1] as $v) { $a = false; break; $a = []; } echo count($a);", "count requires array"},
		{"conditional_break_type", "<?php $a = []; foreach ([1] as $v) { if ($v) { $a = false; break; $a = []; } } echo count($a);", "count requires array"},
		{"typed_return_after_break", "<?php function f($v): array { foreach ([1] as $x) { if ($v) { break; } return []; } return $v; } $a = f(true);", "array return requires array"},
	})
}
