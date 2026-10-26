package tests

import "testing"

func TestPublicPropertyWriteRuntimeGuards(t *testing.T) {
	checkArrayRuntimeGuards(t, []arrayRuntimeCase{
		{"count_after_string_write", "<?php class Box { public $value = null; } $box = new Box(); $box->value = 'text'; $result = count($box->value);", "count requires array"},
		{"strlen_after_array_write", "<?php class Box { public $value = null; } $box = new Box(); $box->value = []; $result = strlen($box->value);", "strlen requires string"},
		{"array_return_after_null_write", "<?php class Box { public $value = null; } function result(): array { $box = new Box(); $box->value = null; return $box->value; } $result = result();", "array return requires array"},
	})
}
