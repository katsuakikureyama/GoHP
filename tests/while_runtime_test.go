package tests

import "testing"

func TestWhileRuntimeGuards(t *testing.T) {
	checkArrayRuntimeGuards(t, []arrayRuntimeCase{
		{"condition_recheck", "<?php $a = [1]; while (count($a)) { $a = false; }", "count requires array"},
		{"continue_recheck", "<?php $a = [1]; while (count($a)) { $a = false; continue; $a = []; }", "count requires array"},
		{"body_recheck", "<?php $a = []; $i = 0; while ($i !== 2) { echo count($a); $a = false; $i = $i + 1; }", "count requires array"},
		{"break_type", "<?php $a = []; while (true) { $a = false; break; $a = []; } echo count($a);", "count requires array"},
		{"return_type", "<?php function f($x): array { while ($x) { break; } return $x; } $a = f(true);", "array return requires array"},
	})
}
