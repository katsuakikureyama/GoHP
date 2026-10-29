package runtime

// PublicPropertyObject is the explicit snapshot protocol used by generated
// class pointers. A host may opt in to the same contract; it is not proof of
// generated provenance. PHPObject must reject typed nil. PHPProperties must
// build a fresh declaration-ordered array through NewArray (which copies nested
// arrays but preserves object identifiers).
type PublicPropertyObject interface {
	Object
	PHPProperties() Array
}

// ObjectToArray snapshots generated public properties, not PHP's general
// scalar/array/Closure/private-property cast semantics. Inferred-unknown source
// operands use this same actual-domain guard; the snapshot is requested once.
func ObjectToArray(value any) any {
	object, ok := value.(PublicPropertyObject)
	if !ok || !object.PHPObject() {
		panic("array cast requires generated public-property object")
	}
	return object.PHPProperties()
}
