package runtime

// ImportGlobal imports a source global name into the execution context.
// If the name is absent, it is created as present null (matching PHP's
// `global $x` behavior: creates null without warning). This uses the
// sourceGlobals value table, not the host scalar ABI.
func ImportGlobal(execution *Execution, name string) {
	RequireExecution(execution)
	requireScalarGlobalName(name)
	if execution.sourceGlobals == nil {
		execution.sourceGlobals = make(map[string]any)
	}
	if _, present := execution.sourceGlobals[name]; !present {
		execution.sourceGlobals[name] = nil
	}
}

// ReadGlobal returns the current value of a source global.
// The returned value is cloned (arrays are deep-copied; objects retain
// identity). Absent returns nil. This uses the sourceGlobals value table,
// not the host scalar ABI.
func ReadGlobal(execution *Execution, name string) any {
	RequireExecution(execution)
	requireScalarGlobalName(name)
	if execution.sourceGlobals == nil {
		return nil
	}
	value, present := execution.sourceGlobals[name]
	if !present {
		return nil
	}
	return Clone(value)
}

// WriteGlobal writes a value through the source global binding.
// Accepts null, bool, int64, string, rt.Array, generated objects and any
// value the runtime already treats as a PHP value. The value is cloned
// consistently with ordinary `v[...] = rt.Clone(...)` assignment semantics.
// This uses the sourceGlobals value table, not the host scalar ABI.
func WriteGlobal(execution *Execution, name string, value any) {
	RequireExecution(execution)
	requireScalarGlobalName(name)
	if execution.sourceGlobals == nil {
		execution.sourceGlobals = make(map[string]any)
	}
	execution.sourceGlobals[name] = Clone(value)
}

// GlobalArraySet performs an atomic read-modify-write on a source global
// array cell. It sets the key to value within the global's current array.
// If the cell is null or absent, it becomes a new array with the single
// entry. The caller must have already evaluated key and value before calling.
func GlobalArraySet(execution *Execution, name string, key, value any) {
	RequireExecution(execution)
	requireScalarGlobalName(name)
	if execution.sourceGlobals == nil {
		execution.sourceGlobals = make(map[string]any)
	}
	base, present := execution.sourceGlobals[name]
	if !present || base == nil {
		base = Array{}
	}
	execution.sourceGlobals[name] = ArraySet(base, key, value)
}

// GlobalArrayAppend performs an atomic read-modify-write on a source global
// array cell. It appends value at the next integer key. The caller must have
// already evaluated value before calling.
func GlobalArrayAppend(execution *Execution, name string, value any) {
	RequireExecution(execution)
	requireScalarGlobalName(name)
	if execution.sourceGlobals == nil {
		execution.sourceGlobals = make(map[string]any)
	}
	base, present := execution.sourceGlobals[name]
	if !present || base == nil {
		base = Array{}
	}
	execution.sourceGlobals[name] = ArrayAppend(base, value)
}
