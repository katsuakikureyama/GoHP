package runtime

// PublicPropertyObject is implemented only by generated class pointers in the
// source subset. PHPProperties builds a fresh declaration-ordered array through
// NewArray (which copies nested arrays but preserves object identifiers).
type PublicPropertyObject interface {
	Object
	PHPProperties() Array
}

// ObjectToArray snapshots generated public properties, not PHP's general
// scalar/array/Closure/private-property cast semantics. Inferred-unknown source
// operands use this same actual-value guard, not a weaker conversion path.
// Host code implementing the entire structural protocol explicitly opts in;
// this is not an unforgeable generated-class provenance check.
func ObjectToArray(value any) any {
	object, ok := value.(PublicPropertyObject)
	if !ok || !object.PHPObject() {
		panic("array cast requires generated public-property object")
	}
	return object.PHPProperties()
}
