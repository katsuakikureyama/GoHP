package tests

import "testing"

func TestArrayEmptyRuntimeGuards(t *testing.T) {
	checkArrayRuntimeGuards(t, []arrayRuntimeCase{
		{"null", "<?php function probe($a) { return empty($a); } echo probe(null);", "empty requires array"},
		{"false", "<?php function probe($a) { return empty($a); } echo probe(false);", "empty requires array"},
		{"string", "<?php function probe($a) { return empty($a); } echo probe('');", "empty requires array"},
		{"arrow", "<?php $probe = fn($a) => empty($a); echo $probe(0);", "empty requires array"},
	})
}
