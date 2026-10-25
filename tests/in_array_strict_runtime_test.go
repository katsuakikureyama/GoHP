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

func TestMembershipControlFlowRuntimeGuards(t *testing.T) {
	checkArrayRuntimeGuards(t, []arrayRuntimeCase{
		{"elseif_late_invalid", "<?php function result($items): array { if (false) { return []; } elseif (in_array('x', $items, true)) { return []; } else { return []; } } $result = result(['x', []]);", "in_array requires scalar haystack values"},
		{"else_if_nonarray", "<?php function result($items): array { if (false) { return []; } else if (in_array('x', $items, true)) { return []; } else { return []; } } $result = result(null);", "in_array requires an array haystack"},
		{"else_unknown_return", "<?php function member($items) { return in_array('x', $items, true); } function result($items): array { if (false) { return []; } else { return member($items); } } $result = result(['x']);", "array return requires array"},
	})
}
