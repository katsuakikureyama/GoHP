package tests

import "testing"

func TestBoolReturnRuntimeGuards(t *testing.T) {
	checkArrayRuntimeGuards(t, []arrayRuntimeCase{
		{"null", "<?php function result($v): bool { return $v; } $result = result(NULL);", "bool return requires bool"},
		{"integer", "<?php function result($v): bool { return $v; } $result = result(1);", "bool return requires bool"},
		{"string_else", "<?php function result($v): bool { if (false) { return true; } else { return $v; } } $result = result('1');", "bool return requires bool"},
		{"array_elseif", "<?php function result($v): bool { if (false) { return true; } elseif (true) { return $v; } return false; } $result = result([]);", "bool return requires bool"},
		{"object_else_if", "<?php class X {} function result($v): bool { if (false) { return true; } else if (true) { return $v; } return false; } $result = result(new X());", "bool return requires bool"},
		{"foreach", "<?php function result($items): bool { foreach ($items as $v) { return $v; } return false; } $result = result([NULL]);", "bool return requires bool"},
		{"fallthrough", "<?php function result($v): bool { if ($v) { return true; } else {} } $result = result(false);", "bool return requires bool"},
	})
}
