package tests

import "testing"

func TestStrContainsRuntimeGuards(t *testing.T) {
	checkArrayRuntimeGuards(t, []arrayRuntimeCase{
		{"null_empty_needle", "<?php function probe($h) { return str_contains($h, ''); } echo probe(null);", "str_contains requires two strings"},
		{"array_empty_needle", "<?php function probe($h) { return str_contains($h, ''); } echo probe([]);", "str_contains requires two strings"},
		{"integer_needle", "<?php function probe($n) { return str_contains('', $n); } echo probe(0);", "str_contains requires two strings"},
		{"bool_needle", "<?php function probe($n) { return str_contains('1', $n); } echo probe(true);", "str_contains requires two strings"},
	})
}
