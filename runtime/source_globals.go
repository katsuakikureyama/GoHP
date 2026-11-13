package runtime

// ImportGlobal imports a scalar global name into the execution context.
// If the name is absent, it is created as present null (matching PHP's
// `global $x` behavior: creates null without warning).
func ImportGlobal(execution *Execution, name string) {
	RequireExecution(execution)
	requireScalarGlobalName(name)
	if _, present := execution.ReadScalarGlobal(name); !present {
		execution.WriteScalarGlobal(name, nil)
	}
}

// ReadGlobal returns the current value of a scalar global. Absent returns nil.
func ReadGlobal(execution *Execution, name string) any {
	RequireExecution(execution)
	requireScalarGlobalName(name)
	value, _ := execution.ReadScalarGlobal(name)
	return value
}

// WriteGlobal writes a scalar value through the global binding.
// Non-scalar values (array/object/callable) fail via the existing runtime
// guard/panic path. The value is cloned consistently with ordinary
// `v[...] = rt.Clone(...)` assignment semantics.
func WriteGlobal(execution *Execution, name string, value any) {
	RequireExecution(execution)
	requireScalarGlobalName(name)
	execution.WriteScalarGlobal(name, Clone(value))
}
