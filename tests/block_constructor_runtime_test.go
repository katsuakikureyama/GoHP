package tests

import "testing"

func TestBlockConstructorAndArrayCastRuntimeGuards(t *testing.T) {
	checkArrayRuntimeGuards(t, []arrayRuntimeCase{
		{"count_after_constructor_string_property", "<?php class Box { public $value; public function __construct($value) { $this->value = $value; } } $box = new Box('text'); $result = count($box->value);", "count requires array"},
		{"strlen_after_constructor_array_property", "<?php class Box { public $value; public function __construct($value) { $this->value = $value; } } $box = new Box([]); $result = strlen($box->value);", "strlen requires string"},
		{"array_return_after_constructor_null_property", "<?php class Box { public $value; public function __construct($value) { $this->value = $value; } } function result(): array { $box = new Box(null); return $box->value; } $result = result();", "array return requires array"},
		{"strlen_after_generated_object_array_cast", "<?php class Box { public $value; } function forward($value) { return $value; } $box = new Box(); $result = strlen(forward((array) $box));", "strlen requires string"},
	})
}
