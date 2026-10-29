package ir

import "strings"

const ClassParameterPrefix = "class:"

// ClassParameterName recognizes a canonical bare named-class annotation, not
// a union, namespace, builtin type or relative self/parent/static type.
// Declaration existence is validated separately after collecting all classes.
func ClassParameterName(annotation string) (string, bool) {
	name, tagged := strings.CutPrefix(annotation, ClassParameterPrefix)
	if !tagged || name == "" || strings.ToLower(name) != name {
		return "", false
	}
	switch name {
	case "array", "callable", "int", "string", "bool", "float", "object", "mixed", "iterable", "resource", "void", "never", "null", "false", "true", "self", "parent", "static":
		return "", false
	}
	for i := range name {
		c := name[i]
		if c != '_' && !(c >= 'a' && c <= 'z') && !(i > 0 && c >= '0' && c <= '9') {
			return "", false
		}
	}
	return name, true
}
