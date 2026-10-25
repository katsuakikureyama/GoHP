package tests

import "testing"

func TestArrayMergeRuntimeGuards(t *testing.T) {
	checkArrayRuntimeGuards(t, []arrayRuntimeCase{
		{"first_null", "<?php function probe($a) { return array_merge($a); } $a = probe(null);", "array_merge requires arrays"},
		{"second_string", "<?php function probe($a) { return array_merge([], $a); } $a = probe('');", "array_merge requires arrays"},
		{"third_bool", "<?php function probe($a) { return array_merge([1], [], $a); } $a = probe(false);", "array_merge requires arrays"},
	})
}
