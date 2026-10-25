package tests

import "testing"

func TestInArrayStrictRuntimeGuards(t *testing.T) {
	checkArrayRuntimeGuards(t, []arrayRuntimeCase{
		{"array_needle_empty", "<?php function probe($n) { return in_array($n, [], true); } echo probe([]);", "in_array requires a scalar needle"},
		{"object_needle", "<?php class X {} function probe($n) { return in_array($n, [], true); } echo probe(new X());", "in_array requires a scalar needle"},
		{"nonarray_haystack", "<?php function probe($h) { return in_array('x', $h, true); } echo probe('x');", "in_array requires an array haystack"},
		{"invalid_before_match", "<?php function probe($h) { return in_array('x', $h, true); } echo probe([[], 'x']);", "in_array requires scalar haystack values"},
		{"invalid_after_match", "<?php function probe($h) { return in_array('x', $h, true); } echo probe(['x', []]);", "in_array requires scalar haystack values"},
	})
}
