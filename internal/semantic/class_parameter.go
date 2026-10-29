package semantic

import "github.com/katsuakikureyama/GoHP/internal/ir"

func classParameterArgument(annotation string, actual valueType, line int) error {
	if name, namedClass := ir.ClassParameterName(annotation); namedClass && actual.kind != unknownKind {
		if actual.kind != objectKind || actual.class != name {
			return unsupported(line, "class parameter requires "+name)
		}
	}
	return nil
}
