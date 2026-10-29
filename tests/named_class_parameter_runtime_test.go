package tests

import "testing"

func TestNamedClassParameterRuntimeGuards(t *testing.T) {
	checkArrayRuntimeGuards(t, []arrayRuntimeCase{
		{
			"unknown_wrong_class",
			"<?php class Expected {} class Other {} function pass($value) { return $value; } function accept(Expected $value) { return $value; } $result = accept(pass(new Other()));",
			"class parameter requires",
		},
		{
			"unknown_null",
			"<?php class Expected {} function pass($value) { return $value; } function accept(Expected $value) { return $value; } $result = accept(pass(null));",
			"class parameter requires",
		},
		{
			"unknown_scalar",
			"<?php class Expected {} function pass($value) { return $value; } function accept(Expected $value) { return $value; } $result = accept(pass(1));",
			"class parameter requires",
		},
		{
			"unknown_array",
			"<?php class Expected {} function pass($value) { return $value; } function accept(Expected $value) { return $value; } $result = accept(pass([]));",
			"class parameter requires",
		},
		// The entry proof remains required even when the parameter is unused.
		{
			"unused",
			"<?php class Expected {} function pass($value) { return $value; } function accept(Expected $value) { return 'unused'; } $result = accept(pass(null));",
			"class parameter requires",
		},
		// count(pass(null)) is a distinct dynamic body failure. Seeing the class-parameter
		// panic proves validation occurs before a local overwrite and body work.
		{
			"overwrite_before_body",
			"<?php class Expected {} function pass($value) { return $value; } function accept(Expected $value) { $value = new Expected(); count(pass(null)); } $result = accept(pass(null));",
			"class parameter requires",
		},
	})
}
