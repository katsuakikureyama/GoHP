package runtime

// PropertyObject is the explicit generated-object protocol, including identity
// and declared property access. Getter-only hosts and closures are not writable.
type PropertyObject interface {
	Object
	Get(string) any
	Set(string, any)
}

// NestedProperty keeps the final owner/name, never a copied container or a
// reference to the root variable. It is used synchronously by generated code.
type NestedProperty struct {
	owner PropertyObject
	name  string
}

func propertyObject(value any) PropertyObject {
	object, ok := value.(PropertyObject)
	if !ok || !object.PHPObject() {
		panic("nested property requires object")
	}
	return object
}

func ResolveNestedProperty(value any, path ...string) NestedProperty {
	if len(path) < 2 {
		panic("nested property requires two or more names")
	}
	owner := propertyObject(value)
	for _, name := range path[:len(path)-1] {
		owner = propertyObject(owner.Get(name))
	}
	name := path[len(path)-1]
	owner.Get(name) // Check the declared final property before publication.
	return NestedProperty{owner: owner, name: name}
}

func (target NestedProperty) object() PropertyObject { return propertyObject(target.owner) }
func (target NestedProperty) Assign(value any)       { target.object().Set(target.name, Clone(value)) }
func (target NestedProperty) Append(value any) {
	owner := target.object()
	owner.Set(target.name, ArrayAppend(owner.Get(target.name), value))
}
func (target NestedProperty) Concat(value any) {
	owner := target.object()
	left, ok := owner.Get(target.name).(string)
	right, valid := value.(string)
	if !ok || !valid {
		panic("nested property concatenation requires strings")
	}
	owner.Set(target.name, left+right)
}
