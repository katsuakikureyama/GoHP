package storage

import (
	"reflect"
	"testing"
)

func requireArrayRead(t *testing.T, array *ScalarArray, key any, want any, wantPresent bool) {
	t.Helper()
	got, present := array.Read(key)
	if present != wantPresent {
		t.Fatalf("Read(%#v) presence = %v, want %v (value %#v)", key, present, wantPresent, got)
	}
	if got != want {
		t.Fatalf("Read(%#v) value = %#v, want %#v", key, got, want)
	}
}

func requireArrayKeys(t *testing.T, array *ScalarArray, want []any) {
	t.Helper()
	if got := array.Keys(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Keys() = %#v, want %#v", got, want)
	}
}

func TestScalarArrayValuesAndCopy(t *testing.T) {
	var zero ScalarArray
	defer zero.Close()
	requireArrayRead(t, &zero, "missing", nil, false)
	if keys := zero.Keys(); keys != nil {
		t.Fatalf("zero Keys() = %#v, want nil", keys)
	}

	zero.Write("null", nil)
	zero.Write(int64(1), "integer-key")
	zero.Write("1", "string-key")
	zero.Write("bool", true)
	zero.Write("int", int64(-7))
	zero.Write("bool", false) // Replacement retains insertion order.

	requireArrayRead(t, &zero, "null", nil, true)
	requireArrayRead(t, &zero, int64(1), "integer-key", true)
	requireArrayRead(t, &zero, "1", "string-key", true)
	requireArrayRead(t, &zero, "bool", false, true)
	requireArrayRead(t, &zero, "int", int64(-7), true)
	requireArrayKeys(t, &zero, []any{"null", int64(1), "1", "bool", "int"})

	keys := zero.Keys()
	keys[0] = "changed-snapshot"
	requireArrayKeys(t, &zero, []any{"null", int64(1), "1", "bool", "int"})

	copy := zero.Copy()
	defer copy.Close()
	copy.Write("bool", true)
	copy.Write("copy-only", "present")
	requireArrayRead(t, &zero, "bool", false, true)
	requireArrayRead(t, &zero, "copy-only", nil, false)
	requireArrayRead(t, copy, "bool", true, true)
	requireArrayRead(t, copy, "copy-only", "present", true)

	shared := NewSlot("shared")
	defer shared.Unset()
	zero.Bind("bound", shared)
	escaped := zero.Reference("bound")
	defer escaped.Unset()
	if shared.ref == nil {
		t.Fatal("bound scalar did not create a reference cell")
	}
	owners := shared.ref.owners
	boundCopy := zero.Copy()
	defer boundCopy.Close()
	if shared.ref.owners != owners {
		t.Fatalf("Copy retained element reference owners: got %d, want %d", shared.ref.owners, owners)
	}
	requireArrayRead(t, &zero, "bound", "shared", true)
	_ = zero.Keys()
	if shared.ref.owners != owners {
		t.Fatalf("Read or Keys retained element reference owners: got %d, want %d", shared.ref.owners, owners)
	}
	boundCopy.Write("bound", "copy-value")
	requireArrayRead(t, &zero, "bound", "copy-value", true)
	requireRead(t, escaped, "copy-value", true)
	requireArrayRead(t, boundCopy, "bound", "copy-value", true)
}

func TestScalarArrayClose(t *testing.T) {
	array := NewScalarArray()
	array.Write("value", "before-close")
	copy := array.Copy()
	array.Close()
	requireArrayRead(t, copy, "value", "before-close", true)
	copy.Close()

	withEscape := NewScalarArray()
	withEscape.Write("value", "seed")
	escaped := withEscape.Reference("value")
	withEscape.Close()
	escaped.Write("survives-close")
	requireRead(t, escaped, "survives-close", true)
	escaped.Unset()

	source := NewSlot("seed")
	array = NewScalarArray()
	array.Bind("value", source)
	shared := source.ref
	if shared == nil || shared.owners != 2 {
		t.Fatalf("owners after Bind = %#v, want source and one array element", shared)
	}
	copy = array.Copy()
	if shared.owners != 2 {
		t.Fatalf("Copy changed element reference owners to %d, want 2", shared.owners)
	}
	array.Close()
	if shared.owners != 2 {
		t.Fatalf("first container Close changed element reference owners to %d, want 2", shared.owners)
	}
	copy.Close()
	if shared.owners != 1 {
		t.Fatalf("last container Close left element reference owners = %d, want 1", shared.owners)
	}
	copy.Close()
	if shared.owners != 1 {
		t.Fatalf("second Close changed element reference owners to %d, want 1", shared.owners)
	}
	source.Unset()
	if shared.owners != 0 || shared.value != nil {
		t.Fatalf("last external owner left cell = %#v, want zero owners and nil value", shared)
	}
}

func TestScalarArrayRejectsInvalidUse(t *testing.T) {
	array := NewScalarArray()
	array.Write("plain", "before")
	source := NewSlot("shared")
	defer source.Unset()
	array.Bind("bound", source)
	escaped := array.Reference("bound")
	defer escaped.Unset()
	copy := array.Copy()
	defer copy.Close()
	shared := source.ref
	if shared == nil {
		t.Fatal("bound scalar did not create a reference cell")
	}
	owners := shared.owners
	keys := []any{"plain", "bound"}

	checkUnchanged := func() {
		t.Helper()
		requireArrayKeys(t, array, keys)
		requireArrayRead(t, array, "plain", "before", true)
		requireArrayRead(t, array, "bound", "shared", true)
		requireArrayRead(t, copy, "plain", "before", true)
		requireArrayRead(t, copy, "bound", "shared", true)
		requireRead(t, escaped, "shared", true)
		if shared.owners != owners {
			t.Fatalf("invalid operation changed reference owners to %d, want %d", shared.owners, owners)
		}
	}

	nested := NewScalarArray()
	defer nested.Close()
	for _, invalidKey := range []any{nil, false, int(1), float64(1), []int{1}, map[string]int{"key": 1}} {
		requirePanic(t, func() { array.Read(invalidKey) })
		requirePanic(t, func() { array.Write(invalidKey, "new") })
		requirePanic(t, func() { array.Bind(invalidKey, source) })
		requirePanic(t, func() { array.Reference(invalidKey) })
		checkUnchanged()
	}
	for _, invalidValue := range []any{int(1), float64(1), []any{"nested"}, nested, source, (*Slot)(nil)} {
		requirePanic(t, func() { array.Write("bound", invalidValue) })
		requirePanic(t, func() { array.Write("new", invalidValue) })
		checkUnchanged()
	}
	requirePanic(t, func() { array.Bind("bound", nil) })
	requirePanic(t, func() { array.Bind("new", nil) })
	checkUnchanged()

	array.Close()
	for name, action := range map[string]func(){
		"Read":      func() { array.Read("plain") },
		"Keys":      func() { array.Keys() },
		"Write":     func() { array.Write("plain", "after-close") },
		"Bind":      func() { array.Bind("plain", source) },
		"Reference": func() { array.Reference("plain") },
		"Copy":      func() { array.Copy() },
	} {
		t.Run("closed_"+name, func(t *testing.T) { requirePanic(t, action) })
	}
	array.Close()

	var nilArray *ScalarArray
	for name, action := range map[string]func(){
		"Read":      func() { nilArray.Read("plain") },
		"Keys":      func() { nilArray.Keys() },
		"Write":     func() { nilArray.Write("plain", "value") },
		"Bind":      func() { nilArray.Bind("plain", source) },
		"Reference": func() { nilArray.Reference("plain") },
		"Copy":      func() { nilArray.Copy() },
		"Close":     func() { nilArray.Close() },
	} {
		t.Run("nil_"+name, func(t *testing.T) { requirePanic(t, action) })
	}
}
