package tests

import "testing"

func TestArrowTypedCompositionRuntimeGuards(t *testing.T) {
	checkArrayRuntimeGuards(t, []arrayRuntimeCase{
		{"bool_return_in_elseif", "<?php function run(array $values, callable $callback): bool { if (false) { return true; } elseif (true) { return $callback($values); } return false; } $result = run([1], fn($values) => $values[0]);", "bool return requires bool"},
		{"array_entry_from_arrow", "<?php function take(array $values): bool { return is_array($values); } function run(array $values, callable $callback): bool { return take($callback($values)); } $result = run([false], fn($values) => $values[0]);", "array parameter requires array"},
	})
}
