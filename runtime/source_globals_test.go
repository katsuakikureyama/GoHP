package runtime

import "testing"

func TestSourceGlobalImportAndRead(t *testing.T) {
	execution := &Execution{}
	defer execution.Close()

	// Import a missing global -> present null.
	ImportGlobal(execution, "x")
	if v := ReadGlobal(execution, "x"); v != nil {
		t.Fatalf("imported missing global = %#v, want nil", v)
	}
	// Check the sourceGlobals map directly for presence.
	if execution.sourceGlobals == nil {
		t.Fatal("imported global should allocate sourceGlobals map")
	}
	if _, present := execution.sourceGlobals["x"]; !present {
		t.Fatal("imported global should be present in sourceGlobals")
	}

	// Write and read back.
	WriteGlobal(execution, "x", int64(42))
	if v := ReadGlobal(execution, "x"); v != int64(42) {
		t.Fatalf("write/read int64 = %#v, want 42", v)
	}
	WriteGlobal(execution, "x", "hello")
	if v := ReadGlobal(execution, "x"); v != "hello" {
		t.Fatalf("write/read string = %#v, want hello", v)
	}
	WriteGlobal(execution, "x", true)
	if v := ReadGlobal(execution, "x"); v != true {
		t.Fatalf("write/read bool = %#v, want true", v)
	}
	WriteGlobal(execution, "x", nil)
	if v := ReadGlobal(execution, "x"); v != nil {
		t.Fatalf("write/read null = %#v, want nil", v)
	}
}

func TestSourceGlobalArrayAndObject(t *testing.T) {
	execution := &Execution{}
	defer execution.Close()

	// Arrays are now accepted.
	WriteGlobal(execution, "arr", NewArray(Entry{Key: "a", Value: int64(1)}))
	v := ReadGlobal(execution, "arr")
	if v == nil {
		t.Fatal("array global should not be nil")
	}
	arr, ok := v.(Array)
	if !ok {
		t.Fatal("array global should be an Array")
	}
	if len(arr.Entries) != 1 {
		t.Fatalf("array global entries = %d, want 1", len(arr.Entries))
	}

	// ReadGlobal returns a copy — mutating the returned array does not change the cell.
	arr.Entries[0].Value = int64(99)
	v2 := ReadGlobal(execution, "arr")
	arr2 := v2.(Array)
	if arr2.Entries[0].Value != int64(1) {
		t.Fatalf("cell should be unchanged after mutating returned copy: got %v", arr2.Entries[0].Value)
	}
}

func TestSourceGlobalObjectIdentity(t *testing.T) {
	execution := &Execution{}
	defer execution.Close()

	// Objects retain identity through Clone.
	obj := &struct{ X int64 }{X: 42}
	WriteGlobal(execution, "obj", obj)
	v1 := ReadGlobal(execution, "obj")
	v2 := ReadGlobal(execution, "obj")
	if v1 != v2 {
		t.Fatal("object identity should be preserved across reads")
	}
}

func TestSourceGlobalCrossCallSharing(t *testing.T) {
	execution := &Execution{}
	defer execution.Close()

	// Simulate function f() importing and writing a global.
	ImportGlobal(execution, "counter")
	WriteGlobal(execution, "counter", int64(1))

	// Simulate function g() reading the same global.
	if v := ReadGlobal(execution, "counter"); v != int64(1) {
		t.Fatalf("cross-call read = %#v, want 1", v)
	}

	// Simulate main writing, function reading.
	WriteGlobal(execution, "counter", int64(2))
	ImportGlobal(execution, "counter") // re-import in another function
	if v := ReadGlobal(execution, "counter"); v != int64(2) {
		t.Fatalf("main write, function read = %#v, want 2", v)
	}
}

func TestSourceGlobalTwoNames(t *testing.T) {
	execution := &Execution{}
	defer execution.Close()

	ImportGlobal(execution, "a")
	ImportGlobal(execution, "b")
	WriteGlobal(execution, "a", int64(1))
	WriteGlobal(execution, "b", "two")

	if v := ReadGlobal(execution, "a"); v != int64(1) {
		t.Fatalf("a = %#v, want 1", v)
	}
	if v := ReadGlobal(execution, "b"); v != "two" {
		t.Fatalf("b = %#v, want two", v)
	}
}

func TestSourceGlobalClose(t *testing.T) {
	execution := &Execution{}
	ImportGlobal(execution, "x")
	WriteGlobal(execution, "x", int64(1))
	execution.Close()

	// After close, sourceGlobals should be nil.
	if execution.sourceGlobals != nil {
		t.Fatal("Close should clear sourceGlobals")
	}

	// Using a closed execution should panic.
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on closed execution")
		}
	}()
	ReadGlobal(execution, "x")
}

func TestGlobalArraySet(t *testing.T) {
	execution := &Execution{}
	defer execution.Close()

	// Set on absent cell — creates a new array.
	GlobalArraySet(execution, "g", "key", int64(1))
	v := ReadGlobal(execution, "g")
	arr, ok := v.(Array)
	if !ok {
		t.Fatal("should be an array")
	}
	if len(arr.Entries) != 1 || arr.Entries[0].Value != int64(1) {
		t.Fatalf("unexpected array: %#v", arr)
	}

	// Set on existing array.
	GlobalArraySet(execution, "g", "key2", int64(2))
	v = ReadGlobal(execution, "g")
	arr = v.(Array)
	if len(arr.Entries) != 2 {
		t.Fatalf("should have 2 entries, got %d", len(arr.Entries))
	}
}

func TestGlobalArraySetOnNull(t *testing.T) {
	execution := &Execution{}
	defer execution.Close()

	// Import creates null.
	ImportGlobal(execution, "g")
	GlobalArraySet(execution, "g", "k", int64(42))
	v := ReadGlobal(execution, "g")
	arr, ok := v.(Array)
	if !ok {
		t.Fatal("should be an array")
	}
	if arr.Entries[0].Value != int64(42) {
		t.Fatalf("unexpected value: %#v", arr.Entries[0].Value)
	}
}

func TestGlobalArrayAppend(t *testing.T) {
	execution := &Execution{}
	defer execution.Close()

	// Append on absent cell.
	GlobalArrayAppend(execution, "g", int64(1))
	GlobalArrayAppend(execution, "g", int64(2))
	v := ReadGlobal(execution, "g")
	arr, ok := v.(Array)
	if !ok {
		t.Fatal("should be an array")
	}
	if len(arr.Entries) != 2 {
		t.Fatalf("should have 2 entries, got %d", len(arr.Entries))
	}
}
