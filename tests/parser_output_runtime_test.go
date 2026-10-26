package tests

import "testing"

func TestParserOutputRuntimeGuards(t *testing.T) {
	checkArrayRuntimeGuards(t, []arrayRuntimeCase{
		{
			"substr_unknown_value",
			"<?php function forward($value) { return $value; } $result = substr(forward(false), 0);",
			"substr requires string",
		},
		{
			"substr_unknown_offset",
			"<?php function forward($value) { return $value; } $result = substr('text', forward(false));",
			"substr offset requires int",
		},
		{
			"substr_unknown_length",
			"<?php function forward($value) { return $value; } $result = substr('text', 0, forward(false));",
			"substr length requires int or null",
		},
		{
			"property_append_unknown_scalar",
			"<?php class Box { public $items; } function forward($value) { return $value; } $box = new Box(); $box->items = forward('text'); $box->items[] = [];",
			"array append requires array",
		},
		{
			"property_append_unknown_null",
			"<?php class Box { public $items; } function forward($value) { return $value; } $box = new Box(); $box->items = forward(null); $box->items[] = [];",
			"array append requires array",
		},
	})
}
