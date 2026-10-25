package tests

import "testing"

func TestArrayKeyFirstRuntimeGuards(t *testing.T) {
	checkArrayRuntimeGuards(t, []arrayRuntimeCase{
		{"null_argument", "<?php function probe($a) { return array_key_first($a); } echo probe(null);", "array_key_first requires array"},
		{"string_argument", "<?php function probe($a) { return array_key_first($a); } echo probe('');", "array_key_first requires array"},
		{"empty_key_not_index", "<?php $a = []; echo $a[array_key_first($a)];", "unsupported array key type"},
		{"key_not_array", "<?php function probe(): array { return array_key_first(['x' => 1]); } $a = probe();", "array return requires array"},
	})
}
