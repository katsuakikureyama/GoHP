package tests

import "testing"

func TestConstructorNullDefaultsRuntimeGuards(t *testing.T) {
	checkArrayRuntimeGuards(t, []arrayRuntimeCase{
		{
			"unknown_string_reaches_count_through_constructor",
			"<?php function forward($value) { return $value; } class Box { public $value; public function __construct($value = null) { $this->value = $value ?? []; } } $box = new Box(forward('text')); $result = count($box->value);",
			"count requires array",
		},
		{
			"unknown_array_reaches_strlen_through_constructor",
			"<?php function forward($value) { return $value; } class Box { public $value; public function __construct($value = null) { $this->value = $value ?? ''; } } $box = new Box(forward([])); $result = strlen($box->value);",
			"strlen requires string",
		},
		{
			"unknown_string_reaches_array_return_through_constructor",
			"<?php function forward($value) { return $value; } class Box { public $value; public function __construct($value = null) { $this->value = $value ?? []; } } function result(): array { $box = new Box(forward('text')); return $box->value; } $result = result();",
			"array return requires array",
		},
	})
}
