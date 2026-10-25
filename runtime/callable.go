package runtime

// Callable is an opaque generated capture-free arrow, not a PHP callable registry.
// Clone preserves this immutable identity; mutable argument/results are copied.
type Callable struct {
	arity int
	call  func(...any) any
}

func NewCallable(arity int, call func(...any) any) *Callable {
	if arity < 0 || call == nil {
		panic("invalid generated callable")
	}
	return &Callable{arity: arity, call: call}
}

func CallableParameter(value any) any {
	callable, ok := value.(*Callable)
	if !ok || callable == nil || callable.call == nil || callable.arity < 0 {
		panic("callable requires generated arrow")
	}
	return callable
}

func Invoke(value any, args ...any) any {
	callable := CallableParameter(value).(*Callable)
	if len(args) != callable.arity {
		panic("callable argument count mismatch")
	}
	local := make([]any, len(args))
	for i, arg := range args {
		local[i] = Clone(arg)
	}
	return Clone(callable.call(local...))
}
