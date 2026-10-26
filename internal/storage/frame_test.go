package storage

import "testing"

func TestFrameValueIsolationAndPresence(t *testing.T) {
	var zero Frame
	defer zero.Close()

	value, present := zero.Read("missing")
	if present || value != nil {
		t.Fatalf("Read(missing) = (%#v, %v), want (nil, false)", value, present)
	}
	if zero.slots != nil {
		t.Fatal("Read(missing) allocated a frame slot")
	}
	zero.Unset("missing")
	if zero.slots != nil {
		t.Fatal("Unset(missing) allocated storage")
	}

	zero.Write("null", nil)
	value, present = zero.Read("null")
	if !present || value != nil {
		t.Fatalf("Read(null) = (%#v, %v), want (nil, true)", value, present)
	}

	zero.Write("value", int64(1))
	other := NewFrame()
	defer other.Close()
	other.Write("value", int64(1))
	value, _ = zero.Read("value")
	zero.Write("copy", value)
	zero.Write("value", int64(2))
	value, present = zero.Read("copy")
	if !present || value != int64(1) {
		t.Fatalf("scalar value copy = (%#v, %v), want (1, true)", value, present)
	}

	value, present = other.Read("value")
	if !present || value != int64(1) {
		t.Fatalf("independent frame value = (%#v, %v), want (1, true)", value, present)
	}
}

func TestFrameBindingsAndUnset(t *testing.T) {
	frame := NewFrame()
	defer frame.Close()
	external := NewSlot("outside")
	defer external.Unset()
	frame.Bind("alias", external)
	frame.Write("alias", "mutated")
	requireRead(t, external, "mutated", true)

	replacement := NewSlot("replacement")
	defer replacement.Unset()
	frame.Bind("alias", replacement)
	frame.Write("alias", "rebound")
	requireRead(t, external, "mutated", true)
	requireRead(t, replacement, "rebound", true)

	frame.Unset("alias")
	value, present := frame.Read("alias")
	if present || value != nil {
		t.Fatalf("Read(alias) after Unset = (%#v, %v), want (nil, false)", value, present)
	}
	if _, exists := frame.slots["alias"]; exists {
		t.Fatal("Unset retained the owned slot")
	}

	frame.Write("alias", "fresh")
	requireRead(t, replacement, "rebound", true)

	exported := frame.Reference("export")
	defer exported.Unset()
	if exported == frame.slots["export"] {
		t.Fatal("Reference exposed the frame-owned slot")
	}
	requireRead(t, exported, nil, true)
	exported.Write("through-reference")
	value, present = frame.Read("export")
	if !present || value != "through-reference" {
		t.Fatalf("Read(export) = (%#v, %v), want (through-reference, true)", value, present)
	}
	frame.Unset("export")
	frame.Write("export", "recreated")
	exported.Write("old-binding")
	value, present = frame.Read("export")
	if !present || value != "recreated" {
		t.Fatalf("recreated name = (%#v, %v), want (recreated, true)", value, present)
	}

	frame.Unset("does-not-exist")
}

func TestFrameClosePreservesEscapedReferences(t *testing.T) {
	source := NewSlot("seed")
	defer source.Unset()
	frame := NewFrame()
	defer frame.Close()
	frame.Bind("first", source)
	frame.Bind("second", source)
	escaped := frame.Reference("first")
	defer escaped.Unset()

	shared := source.ref
	if shared == nil || shared.owners != 4 {
		t.Fatalf("shared reference owners before Close = %#v, want 4", shared)
	}

	frame.Close()
	if !frame.closed || frame.slots != nil {
		t.Fatalf("Close state = closed:%v slots:%#v, want closed frame with nil slots", frame.closed, frame.slots)
	}
	if shared.owners != 2 {
		t.Fatalf("shared reference owners after Close = %d, want 2", shared.owners)
	}

	escaped.Write("survives")
	requireRead(t, source, "survives", true)
	frame.Close()
	if shared.owners != 2 {
		t.Fatalf("second Close changed reference owners to %d, want 2", shared.owners)
	}

	escaped.Unset()
	if shared.owners != 1 {
		t.Fatalf("owners after escaped Unset = %d, want 1", shared.owners)
	}
	source.Unset()
	if shared.owners != 0 || shared.value != nil {
		t.Fatalf("last external Unset left cell = %#v, want no owners and nil value", shared)
	}
}

func TestFrameDeferredCloseOnExit(t *testing.T) {
	source := NewSlot("seed")
	defer source.Unset()
	returned := closeFrameOnEarlyReturn(source)
	defer returned.Unset()
	if source.ref == nil || source.ref.owners != 2 {
		t.Fatalf("return-path Close did not release its local binding: %#v", source.ref)
	}
	returned.Write("returned")
	requireRead(t, source, "returned", true)
	returned.Unset()

	panicSource := NewSlot("before-panic")
	defer panicSource.Unset()
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("panic path did not panic")
			}
		}()
		closeFrameOnPanic(panicSource)
	}()
	requireRead(t, panicSource, "panic", true)
	if panicSource.ref == nil || panicSource.ref.owners != 1 {
		t.Fatalf("panic-path Close did not release its local binding: %#v", panicSource.ref)
	}
}

func closeFrameOnEarlyReturn(source *Slot) *Slot {
	frame := NewFrame()
	defer frame.Close()
	frame.Bind("local", source)
	return frame.Reference("local")
}

// This exercises Go cleanup mechanics only; it does not claim PHP exception equivalence.
func closeFrameOnPanic(source *Slot) {
	frame := NewFrame()
	defer frame.Close()
	frame.Bind("local", source)
	frame.Write("local", "panic")
	panic("test panic")
}

func TestFrameClosedAndNilOperationsReject(t *testing.T) {
	frame := NewFrame()
	frame.Write("value", "before-close")
	frame.Close()

	for name, action := range map[string]func(){
		"Read":      func() { frame.Read("value") },
		"Write":     func() { frame.Write("value", "after-close") },
		"Bind":      func() { frame.Bind("value", NewSlot(nil)) },
		"Reference": func() { frame.Reference("value") },
		"Unset":     func() { frame.Unset("value") },
	} {
		t.Run(name, func(t *testing.T) { requirePanic(t, action) })
	}
	if !frame.closed || frame.slots != nil {
		t.Fatalf("closed frame was revived: closed:%v slots:%#v", frame.closed, frame.slots)
	}
	frame.Close()

	var nilFrame *Frame
	for name, action := range map[string]func(){
		"Read":      func() { nilFrame.Read("value") },
		"Write":     func() { nilFrame.Write("value", "value") },
		"Bind":      func() { nilFrame.Bind("value", NewSlot(nil)) },
		"Reference": func() { nilFrame.Reference("value") },
		"Unset":     func() { nilFrame.Unset("value") },
		"Close":     func() { nilFrame.Close() },
	} {
		t.Run("nil_"+name, func(t *testing.T) { requirePanic(t, action) })
	}
}

func TestFrameInvalidWritesPreserveBindings(t *testing.T) {
	frame := NewFrame()
	defer frame.Close()
	source := NewSlot("before")
	defer source.Unset()
	frame.Write("plain", "ordinary")
	frame.Bind("bound", source)
	owned := frame.slots["bound"]
	shared := source.ref

	for _, invalid := range []any{int(1), float64(1), []any{"array"}, struct{ Name string }{"object"}, source, (*Slot)(nil)} {
		requirePanic(t, func() { frame.Write("bound", invalid) })
		requireRead(t, source, "before", true)
		if frame.slots["bound"] != owned || source.ref != shared || shared.owners != 2 {
			t.Fatal("invalid write changed the existing binding")
		}
		requirePanic(t, func() { frame.Write("plain", invalid) })
		value, present := frame.Read("plain")
		if !present || value != "ordinary" {
			t.Fatalf("invalid write changed ordinary value to (%#v, %v)", value, present)
		}

		requirePanic(t, func() { frame.Write("new", invalid) })
		if _, exists := frame.slots["new"]; exists {
			t.Fatal("invalid write allocated a new binding")
		}
	}

	requirePanic(t, func() { frame.Bind("bound", nil) })
	if frame.slots["bound"] != owned || source.ref != shared || shared.owners != 2 {
		t.Fatal("nil Bind changed the existing binding")
	}
	requirePanic(t, func() { frame.Bind("nil-source", nil) })
	if _, exists := frame.slots["nil-source"]; exists {
		t.Fatal("nil Bind allocated a new binding")
	}
}
