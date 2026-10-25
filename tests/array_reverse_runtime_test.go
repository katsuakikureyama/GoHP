package tests

import "testing"

func TestArrayReverseRuntimeGuards(t *testing.T) {
	checkArrayRuntimeGuards(t, []arrayRuntimeCase{
		{"null_array", "<?php function probe($a) { return array_reverse($a); } $a = probe(null);", "array_reverse requires array"},
		{"string_array", "<?php function probe($a) { return array_reverse($a, true); } $a = probe('');", "array_reverse requires array"},
		{"int_flag_empty", "<?php function probe($p) { return array_reverse([], $p); } $a = probe(1);", "array_reverse preserve_keys requires bool"},
		{"null_flag", "<?php function probe($p) { return array_reverse([1], $p); } $a = probe(null);", "array_reverse preserve_keys requires bool"},
		{"preserved_overflow", "<?php $a = array_reverse([9223372036854775807 => 1], true); $a[] = 2;", "array key overflow unsupported"},
	})
}
