package runtime

import "testing"

func TestDirectPropertyIdentityAndCopies(t *testing.T) {
	object := &nestedTestObject{map[string]any{"value": nil, "items": NewArray()}}
	other := &nestedTestObject{map[string]any{"value": "untouched"}}
	input := NewArray(Entry{Value: NewArray(Entry{Value: "original"})})
	ResolveProperty(object, "value").Assign(input)
	input.Entries[0].Value.(Array).Entries[0].Value = "changed"
	if Get(Get(object.Get("value"), int64(0)), int64(0)) != "original" || other.Get("value") != "untouched" {
		t.Fatal("property assignment lost copy separation or owner identity")
	}
	ResolveProperty(object, "value").Assign(other)
	if object.Get("value") != other {
		t.Fatal("object assignment must preserve identity")
	}
	target := ResolveProperty(object, "items")
	object.Set("items", NewArray(Entry{Value: "rhs"}))
	target.Append(nil)
	items := object.Get("items").(Array)
	if len(items.Entries) != 2 || items.Entries[0].Value != "rhs" || items.Entries[1].Value != nil {
		t.Fatal("append must use the current slot, including null values")
	}
	for _, value := range []any{nil, false, int64(0), ""} {
		ResolveProperty(object, "value").Assign(value)
		if object.Get("value") != value {
			t.Fatalf("declared falsy property lost: %#v", value)
		}
	}
}

func TestDirectPropertyRejectsInvalidTargets(t *testing.T) {
	var typedNil *nestedTestObject
	object := &nestedTestObject{map[string]any{"value": "before", "items": nil}}
	for _, call := range []func(){
		func() { ResolveProperty(nil, "value") },
		func() { ResolveProperty(typedNil, "value") },
		func() { ResolveProperty(int64(1), "value") },
		func() { ResolveProperty(false, "value") },
		func() { ResolveProperty("text", "value") },
		func() { ResolveProperty(NewArray(), "value") },
		func() { ResolveProperty(object, "missing").Assign("must not publish") },
		func() { ResolveProperty(object, "items").Append("must not publish") },
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("unsupported direct property update accepted")
				}
			}()
			call()
		}()
	}
	if len(object.fields) != 2 || object.Get("value") != "before" || object.Get("items") != nil {
		t.Fatal("failed update created or changed a slot")
	}
}
