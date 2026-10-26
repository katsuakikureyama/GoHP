package runtime

// PublicPropertyObject is implemented only by generated class pointers in the
// source subset. PHPProperties builds a fresh declaration-ordered array through
// NewArray (which copies nested arrays but preserves object identifiers).
type PublicPropertyObject interface {
	Object
	PHPProperties() Array
}

// ObjectToArray snapshots generated public properties, not PHP's general
// scalar/array/Closure/private-property cast semantics. Source operands must
// already have a statically known generated class.
func ObjectToArray(value any) any {
	object, ok := value.(PublicPropertyObject)
	if !ok || !object.PHPObject() {
		panic("array cast requires generated public-property object")
	}
	return object.PHPProperties()
}
