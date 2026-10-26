package phpversion

import "strings"

// ReservedFunction identifies new core functions that cannot be redeclared in
// the selected target. Existing supported builtins are guarded by codegen.
func ReservedFunction(target, name string) bool {
	if target != "8.4" {
		return false
	}
	switch strings.ToLower(name) {
	case "array_find", "array_find_key", "array_any", "array_all":
		return true
	}
	return false
}

// DeprecatedClassName keeps unimplemented diagnostic behavior outside the
// admitted subset instead of silently suppressing PHP 8.4's deprecation.
func DeprecatedClassName(target, name string) bool {
	return target == "8.4" && name == "_"
}
