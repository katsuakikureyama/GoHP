package tests

import "testing"

func TestArrayEmptyRuntimeGuards(t *testing.T) {
	checkArrayRuntimeGuards(t, []arrayRuntimeCase{
		{"null", "<?php function probe($a) { return empty($a); } echo probe(null);", "empty requires array or string"},
		{"false", "<?php function probe($a) { return empty($a); } echo probe(false);", "empty requires array or string"},
		{"integer", "<?php function probe($a) { return empty($a); } echo probe(0);", "empty requires array or string"},
		{"arrow", "<?php $probe = fn($a) => empty($a); echo $probe(0);", "empty requires array or string"},
	})
}
