package runtime

// Callable is an opaque generated capture-free arrow, not a PHP callable registry.
// Clone preserves this immutable identity; mutable argument/results are copied.
type Callable struct {
	arity           int
	call            func(...any) any
	callInExecution func(*Execution, ...any) any
}

// PHPObject reflects that PHP's Closure is an object, without making class
// instances callable or adding closure properties/methods to this subset.
func (callable *Callable) PHPObject() bool {
	return callable != nil && (callable.call != nil || callable.callInExecution != nil) && callable.arity >= 0
}

func NewCallable(arity int, call func(...any) any) *Callable {
	if arity < 0 || call == nil {
		panic("invalid generated callable")
	}
	return &Callable{arity: arity, call: call}
}

func CallableParameter(value any) any {
	callable, ok := value.(*Callable)
	if !ok || !callable.PHPObject() {
		panic("callable requires generated arrow")
	}
	return callable
}

func Invoke(value any, args ...any) any {
	return InvokeWithExecution(&Execution{}, value, args...)
}

// Contextual arrows receive their caller's execution, never retain a creator's
// mutable scan. An arrow can safely cross entry boundaries without leaking it.
func NewCallableWithExecution(arity int, call func(*Execution, ...any) any) *Callable {
	if arity < 0 || call == nil {
		panic("invalid generated callable")
	}
	return &Callable{arity: arity, callInExecution: call}
}

func InvokeWithExecution(execution *Execution, value any, args ...any) any {
	RequireExecution(execution)
	callable := CallableParameter(value).(*Callable)
	if len(args) != callable.arity {
		panic("callable argument count mismatch")
	}
	local := make([]any, len(args))
	for i, arg := range args {
		local[i] = Clone(arg)
	}
	if callable.callInExecution != nil {
		return Clone(callable.callInExecution(execution, local...))
	}
	return Clone(callable.call(local...))
}
