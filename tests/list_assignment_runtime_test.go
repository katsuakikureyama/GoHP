package tests

import "testing"

func TestListAssignmentRuntimeGuards(t *testing.T) {
	checkArrayRuntimeGuards(t, []arrayRuntimeCase{
		{"nonarray", "<?php function unpack($value) { list($first) = $value; return $first; } $result = unpack(null);", "list assignment requires array"},
		{"sparse", "<?php function unpack($value) { list($first, $second) = $value; return $first; } $result = unpack([1 => 'missing-zero']);", "list assignment requires present numeric keys"},
		{"string_only", "<?php function unpack($value) { list($first) = $value; return $first; } $result = unpack(['first' => 'value']);", "list assignment requires present numeric keys"},
	})
}
