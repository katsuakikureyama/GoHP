package runtime

import (
	"reflect"
	"testing"
)

type objectArrayMarkerOnly struct{}

func (*objectArrayMarkerOnly) PHPObject() bool { return true }

type objectArrayInvalidMarker struct{}

func (*objectArrayInvalidMarker) PHPObject() bool { return false }
func (*objectArrayInvalidMarker) PHPProperties() Array {
	return NewArray()
}

type objectArraySnapshot struct {
	properties Array
	calls      int
}

func (snapshot *objectArraySnapshot) PHPObject() bool { return snapshot != nil }
func (snapshot *objectArraySnapshot) PHPProperties() Array {
	snapshot.calls++
	return snapshot.properties
}

func TestObjectToArrayUsesGeneratedPropertySnapshotOnce(t *testing.T) {
	identity := &objectMarker{valid: true}
	snapshot := &objectArraySnapshot{properties: NewArray(
		Entry{Key: "name", Value: "block"},
		Entry{Key: "nested", Value: NewArray(Entry{Key: int64(-3), Value: "original"})},
		Entry{Key: "object", Value: identity},
	)}

	got, ok := ObjectToArray(snapshot).(Array)
	if !ok {
		t.Fatal("array cast did not return an Array")
	}
	if snapshot.calls != 1 {
		t.Fatalf("PHPProperties calls = %d, want 1", snapshot.calls)
	}
	if !reflect.DeepEqual(got, snapshot.properties) {
		t.Fatalf("array cast = %#v, want %#v", got, snapshot.properties)
	}
	if got.index != snapshot.properties.index {
		t.Fatal("array cast lost property-array metadata")
	}
	if got.Entries[2].Value != identity {
		t.Fatal("array cast did not preserve generated object identity")
	}

}

func TestObjectToArrayRejectsOutsideGeneratedPropertyDomain(t *testing.T) {
	for _, tt := range []struct {
		name  string
		value any
	}{
		{"marker_only", &objectArrayMarkerOnly{}},
		{"getter_only", getterOnly{}},
		{"invalid_marker", &objectArrayInvalidMarker{}},
		{"nil", nil},
		{"typed_nil_snapshot", (*objectArraySnapshot)(nil)},
		{"callable", NewCallable(0, func(...any) any { return nil })},
		{"boolean", true},
		{"integer", int64(1)},
		{"string", "object"},
		{"array", NewArray()},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if got := recover(); got != "array cast requires generated public-property object" {
					t.Fatalf("panic = %v, want %q", got, "array cast requires generated public-property object")
				}
			}()
			ObjectToArray(tt.value)
		})
	}
}
