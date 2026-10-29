package runtime

import "testing"

type presenceTestObject struct {
	fields map[string]any
	calls  []string
}

func (o *presenceTestObject) PHPObject() bool { return o != nil }
func (o *presenceTestObject) PHPPropertyLookup(name string) (any, bool) {
	o.calls = append(o.calls, name)
	value, ok := o.fields[name]
	return value, ok
}

func TestPropertyIssetQuietPath(t *testing.T) {
	child := &presenceTestObject{fields: map[string]any{"value": false}}
	root := &presenceTestObject{fields: map[string]any{"child": child, "null": nil}}
	if !PropertyIsset(root, "child", "value") || len(root.calls) != 1 || len(child.calls) != 1 {
		t.Fatal("once-only reached-slot lookup")
	}
	if PropertyIsset(root, "missing", "value") || PropertyIsset(root, "null", "value") || len(child.calls) != 1 {
		t.Fatal("missing/null path must stop")
	}
	for _, value := range []any{false, true, int64(0), int64(-1), "", "0", NewArray(), child} {
		child.fields["value"] = value
		if !PropertyIsset(child, "value") {
			t.Fatalf("nonnull value not set: %#v", value)
		}
	}
	for _, value := range []any{nil, false, int64(0), "text", NewArray(), (*presenceTestObject)(nil)} {
		if PropertyIsset(value, "value") {
			t.Fatalf("nonobject root set: %#v", value)
		}
		root.fields["child"] = value
		if PropertyIsset(root, "child", "value") {
			t.Fatalf("nonobject intermediate set: %#v", value)
		}
	}
}

func TestPropertyIssetRejectsForeignHostValues(t *testing.T) {
	for _, value := range []any{struct{}{}, 1, 1.0, (*Array)(nil), &nestedTestObject{fields: map[string]any{}}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("foreign host accepted: %#v", value)
				}
			}()
			PropertyIsset(value, "value")
		}()
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Error("empty path accepted")
			}
		}()
		PropertyIsset(nil)
	}()
}
