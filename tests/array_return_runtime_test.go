package tests

import "testing"

func TestArrayReturnRuntimeGuards(t *testing.T) {
	checkArrayRuntimeGuards(t, []arrayRuntimeCase{
		{"unknown_null", "<?php function result($value): array { return $value; } $result = result(null);", "array return requires array"},
		{"unknown_scalar", "<?php function result($value): array { return $value; } $result = result('scalar');", "array return requires array"},
		{"fallthrough", "<?php function result($value): array { if ($value) { return []; } } $result = result(false);", "array return requires array"},
	})
}

func TestArrayReturnElseRuntimeGuards(t *testing.T) {
	checkArrayRuntimeGuards(t, []arrayRuntimeCase{
		{"else_null", "<?php function result($value): array { if (false) { return []; } else { return $value; } } $result = result(null);", "array return requires array"},
		{"elseif_scalar", "<?php function result($value): array { if (false) { return []; } elseif (true) { return $value; } else { return []; } } $result = result('bad');", "array return requires array"},
		{"else_if_scalar", "<?php function result($value): array { if (false) { return []; } else if (true) { return $value; } else { return []; } } $result = result(1);", "array return requires array"},
		{"foreach_else", "<?php function result($items): array { foreach ($items as $item) { if (false) { return []; } else { return $item; } } return []; } $result = result([null]);", "array return requires array"},
		{"else_fallthrough", "<?php function result($value): array { if ($value) { return []; } elseif ($value) { return []; } else {} } $result = result(false);", "array return requires array"},
	})
}
