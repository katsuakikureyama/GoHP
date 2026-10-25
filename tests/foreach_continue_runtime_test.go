package tests

import "testing"

func TestForeachContinueRuntimeGuards(t *testing.T) {
	checkArrayRuntimeGuards(t, []arrayRuntimeCase{
		{"changed_type_before_continue", "<?php $a = []; foreach ([1] as $v) { $a = false; continue; $a = []; } echo count($a);", "count requires array"},
		{"typed_return_fallthrough", "<?php function f($v): array { foreach ([1] as $x) { if ($v) { continue; } return []; } return $v; } $a = f(true);", "array return requires array"},
	})
}
