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
	if _, present := execution.ReadScalarGlobal("x"); !present {
		t.Fatal("imported global should be present")
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

func TestSourceGlobalNonScalarRejection(t *testing.T) {
	execution := &Execution{}
	defer execution.Close()

	// Arrays must fail via the existing WriteScalarGlobal guard.
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic when writing array to scalar global")
		}
	}()
	WriteGlobal(execution, "arr", NewArray())
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
