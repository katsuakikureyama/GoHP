package tests

import "testing"

// Dynamic selector values are checked by the generated runtime boundary. The
// admitted lane accepts actual nonnumeric strings only; PHP's broader scalar
// coercions are intentionally not modeled here.
func TestStringSwitchRuntimeGuards(t *testing.T) {
	const source = `<?php
function select_string_case($value) {
    switch ($value) {
        case 'alpha':
            return 'alpha';
        case 'omega':
            return 'omega';
        default:
            return 'default';
    }
}
`
	const rejection = "switch requires a nonnumeric string"
	checkArrayRuntimeGuards(t, []arrayRuntimeCase{
		{"int", source + "$result = select_string_case(1);", rejection},
		{"null", source + "$result = select_string_case(null);", rejection},
		{"bool", source + "$result = select_string_case(true);", rejection},
		{"array", source + "$result = select_string_case(['alpha']);", rejection},
		{"numeric_whitespace", source + "$result = select_string_case(' 1');", rejection},
		{"numeric_sign", source + "$result = select_string_case('+1');", rejection},
		{"numeric_decimal", source + "$result = select_string_case('1.5');", rejection},
		{"numeric_exponent", source + "$result = select_string_case('1e3');", rejection},
		{"numeric_overflow", source + "$result = select_string_case('9223372036854775808');", rejection},
	})
}
