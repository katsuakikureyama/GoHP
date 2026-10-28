package runtime

import "testing"

type nestedTestObject struct{ fields map[string]any }

func (o *nestedTestObject) PHPObject() bool { return o != nil }
func (o *nestedTestObject) Get(name string) any {
	value, ok := o.fields[name]
	if !ok {
		panic("undefined property")
	}
	return Clone(value)
}
func (o *nestedTestObject) Set(name string, value any) {
	if _, ok := o.fields[name]; !ok {
		panic("undefined property assignment")
	}
	o.fields[name] = Clone(value)
}

func TestNestedPropertyIdentityAndCopies(t *testing.T) {
	old := &nestedTestObject{map[string]any{"text": "old", "items": NewArray(), "value": nil}}
	root := &nestedTestObject{map[string]any{"child": old}}
	target := ResolveNestedProperty(root, "child", "text")
	replacement := &nestedTestObject{map[string]any{"text": "new"}}
	root.Set("child", replacement)
	old.Set("text", "rhs-updated")
	target.Concat("+suffix")
	if old.Get("text") != "rhs-updated+suffix" || replacement.Get("text") != "new" {
		t.Fatal("resolved owner or current value lost")
	}
	root.Set("child", old)
	input := NewArray(Entry{Value: NewArray(Entry{Value: "before"})})
	ResolveNestedProperty(root, "child", "value").Assign(input)
	input.Entries[0].Value.(Array).Entries[0].Value = "changed"
	if Get(Get(old.Get("value"), int64(0)), int64(0)) != "before" {
		t.Fatal("assigned array aliases input")
	}
	ResolveNestedProperty(root, "child", "items").Append(nil)
	ResolveNestedProperty(root, "child", "items").Append(old)
	items := old.Get("items").(Array)
	if len(items.Entries) != 2 || items.Entries[0].Value != nil || items.Entries[1].Value != old {
		t.Fatal("append value or identity lost")
	}
}

func TestNestedPropertyRejectsInvalidTargets(t *testing.T) {
	var nilObject *nestedTestObject
	root := &nestedTestObject{map[string]any{"child": nil, "text": "root"}}
	for _, call := range []func(){
		func() { ResolveNestedProperty(nil, "child", "text") },
		func() { ResolveNestedProperty(nilObject, "child", "text") },
		func() { ResolveNestedProperty(NewArray(), "child", "text") },
		func() { ResolveNestedProperty(root, "child", "text") },
		func() { ResolveNestedProperty(root, "text") },
		func() { (NestedProperty{}).Assign("x") },
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("unsupported target accepted")
				}
			}()
			call()
		}()
	}
	child := &nestedTestObject{map[string]any{"text": int64(1), "items": nil}}
	root.Set("child", child)
	for _, call := range []func(){
		func() { ResolveNestedProperty(root, "child", "missing") },
		func() { ResolveNestedProperty(root, "child", "text").Concat("x") },
		func() { ResolveNestedProperty(root, "child", "items").Append(1) },
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("unsupported update accepted")
				}
			}()
			call()
		}()
	}
	if child.Get("text") != int64(1) || child.Get("items") != nil {
		t.Fatal("failed update published")
	}
}
