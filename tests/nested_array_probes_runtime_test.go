package tests

import "testing"

func TestNestedArrayProbeRuntimeGuards(t *testing.T) {
	checkArrayRuntimeGuards(t, []arrayRuntimeCase{
		{"coalesce_scalar_intermediate", "<?php function pick($root) { return $root['a']['b'] ?? false; } $result = pick(['a' => false]);", "array path requires array or null"},
		{"isset_string_intermediate", "<?php function present($root) { return isset($root['a']['b']); } $result = present(['a' => 'text']);", "array path requires array or null"},
		{"coalesce_dynamic_key", "<?php function pick($key) { $root = ['a' => []]; return $root['a'][$key] ?? false; } $result = pick(true);", "array path requires integer or string key"},
	})
}
