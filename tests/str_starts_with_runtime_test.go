package tests

import "testing"

func TestStrStartsWithRuntimeGuards(t *testing.T) {
	checkArrayRuntimeGuards(t, []arrayRuntimeCase{
		{"null_empty_needle", "<?php function probe($h) { return str_starts_with($h, ''); } echo probe(null);", "str_starts_with requires two strings"},
		{"array_empty_needle", "<?php function probe($h) { return str_starts_with($h, ''); } echo probe([]);", "str_starts_with requires two strings"},
		{"integer_needle", "<?php function probe($n) { return str_starts_with('', $n); } echo probe(0);", "str_starts_with requires two strings"},
		{"bool_needle", "<?php function probe($n) { return str_starts_with('1', $n); } echo probe(true);", "str_starts_with requires two strings"},
	})
}
