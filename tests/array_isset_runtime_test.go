package tests

import "testing"

func TestArrayIssetRuntimeGuards(t *testing.T) {
	checkArrayRuntimeGuards(t, []arrayRuntimeCase{
		{"null_root", "<?php function probe($a) { return isset($a[0]); } echo probe(null);", "isset requires array"},
		{"string_root", "<?php function probe($a) { return isset($a[0]); } echo probe('text');", "isset requires array"},
		{"bool_key", "<?php function probe($key) { $a = []; return isset($a[$key]); } echo probe(false);", "isset requires integer or string key"},
		{"null_key", "<?php function probe($key) { $a = []; return isset($a[$key]); } echo probe(null);", "isset requires integer or string key"},
	})
}
