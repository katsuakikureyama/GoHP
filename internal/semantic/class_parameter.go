package semantic

import "strings"

func classParameterType(annotation string) (string, bool) {
	if !strings.HasPrefix(annotation, "class:") {
		return "", false
	}
	name := strings.TrimPrefix(annotation, "class:")
	if !validName(name) || name != strings.ToLower(name) {
		return "", false
	}
	switch name {
	case "int", "string", "bool", "array", "callable", "float", "object", "mixed", "iterable", "void", "never", "null", "true", "false", "self", "parent", "static", "boolean", "integer", "double", "real", "resource":
		return "", false
	}
	return name, true
}

func classArgumentMatches(actual valueType, name string) bool {
	return actual.kind == unknownKind || actual.kind == objectKind && actual.class == name
}
