package tests

import "testing"

func TestArraySliceRuntimeGuards(t *testing.T) {
	checkArrayRuntimeGuards(t, []arrayRuntimeCase{
		{"nonarray", "<?php function probe($a) { return array_slice($a, 0); } $a = probe(null);", "array_slice requires array"},
		{"offset", "<?php function probe($o) { return array_slice([], $o); } $a = probe('0');", "array_slice offset requires int"},
		{"length", "<?php function probe($l) { return array_slice([], 9223372036854775807, $l); } $a = probe(false);", "array_slice length requires int or null"},
		{"flag", "<?php function probe($p) { return array_slice([], 0, 0, $p); } $a = probe(1);", "array_slice preserve_keys requires bool"},
		{"preserved_overflow", "<?php $a = array_slice([9223372036854775807 => 1], 0, null, true); $a[] = 2;", "array key overflow unsupported"},
	})
}
