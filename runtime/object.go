package runtime

// Object is the explicit identity marker implemented by generated class
// pointers and generated arrows (PHP Closures). A property getter alone does
// not turn an arbitrary Go value into a PHP object. Typed nil pointers are not
// source PHP values and must report false through the marker.
type Object interface {
	PHPObject() bool
}

// IsObject inspects only the supported value representation. It does not read
// properties, invoke PHP methods or convert values to bool.
func IsObject(value any) bool {
	object, ok := value.(Object)
	return ok && object.PHPObject()
}
