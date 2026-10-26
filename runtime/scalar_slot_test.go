package runtime

import (
	"strings"
	"testing"
)

type scalarSlotTestObject struct{}

func (*scalarSlotTestObject) PHPObject() bool { return true }

func scalarSlotPanic(t *testing.T, name string, call func()) {
	t.Helper()
	defer func() {
		message, ok := recover().(string)
		if !ok || !strings.Contains(message, "scalar slot") {
			t.Fatalf("%s did not reject with a scalar slot diagnostic: %q", name, message)
		}
	}()
	call()
}

func scalarSlotRead(t *testing.T, slot scalarSlot, want any) {
	t.Helper()
	if got := slot.read(); got != want {
		t.Fatalf("read = %#v, want %#v", got, want)
	}
}

func scalarSlotOwners(t *testing.T, slot scalarSlot, want int) {
	t.Helper()
	if slot.binding == nil || slot.binding.ref == nil {
		t.Fatalf("slot has no ref cell, want %d owners", want)
	}
	if got := slot.binding.ref.owners; got != want {
		t.Fatalf("ref owners = %d, want %d", got, want)
	}
}

func TestScalarSlotValueAndAlias(t *testing.T) {
	value := newScalarSlot()
	for _, want := range []any{int64(-7), "x\x00日本語", false, true, nil} {
		value.assign(want)
		scalarSlotRead(t, value, want)
	}

	value.assign(int64(1))
	copy := newScalarSlot()
	copy.assign(value.read())
	alias := newScalarSlot()
	alias.bind(value)
	value.assign(int64(2))
	scalarSlotRead(t, value, int64(2))
	scalarSlotRead(t, alias, int64(2))
	scalarSlotRead(t, copy, int64(1))

	alias.assign("shared")
	scalarSlotRead(t, value, "shared")
	scalarSlotRead(t, alias, "shared")
	scalarSlotRead(t, copy, int64(1))
	for _, want := range []any{false, true, nil, "", int64(0)} {
		alias.assign(want)
		scalarSlotRead(t, value, want)
		if !value.isDefined() || !alias.isDefined() {
			t.Fatal("aliased scalar/null write lost the defined binding")
		}
	}
}

func TestScalarSlotRebindAndUnset(t *testing.T) {
	origin := newScalarSlot()
	origin.assign(int64(10))
	oldAlias := newScalarSlot()
	oldAlias.bind(origin)

	destination := newScalarSlot()
	destination.assign(int64(20))
	destinationAlias := newScalarSlot()
	destinationAlias.bind(destination)
	destinationCell := destination.binding.ref

	destination.bind(origin)
	scalarSlotRead(t, origin, int64(10))
	scalarSlotRead(t, destination, int64(10))
	scalarSlotRead(t, destinationAlias, int64(20))
	if destinationCell.owners != 1 {
		t.Fatalf("old destination ref owners = %d, want 1", destinationCell.owners)
	}
	scalarSlotOwners(t, origin, 3)

	destination.assign(int64(11))
	scalarSlotRead(t, origin, int64(11))
	scalarSlotRead(t, oldAlias, int64(11))
	destination.unset()
	if destination.isDefined() {
		t.Fatal("unset destination is still defined")
	}
	scalarSlotOwners(t, origin, 2)
	destination.unset()
	scalarSlotRead(t, origin, int64(11))

	destination.assign("new")
	scalarSlotRead(t, destination, "new")
	scalarSlotRead(t, origin, int64(11))
	originCell := origin.binding.ref
	origin.unset()
	oldAlias.unset()
	if originCell.owners != 0 || originCell.value != nil {
		t.Fatalf("released ref cell = %#v owners=%d", originCell.value, originCell.owners)
	}
}

func TestScalarSlotBindingLifetime(t *testing.T) {
	caller := newScalarSlot()
	caller.assign("caller")
	retainedAlias := newScalarSlot()
	retainedAlias.bind(caller)

	localParameter := newScalarSlot()
	localParameter.bind(caller)
	localParameter.assign("mutated")
	scalarSlotRead(t, caller, "mutated")
	scalarSlotRead(t, retainedAlias, "mutated")

	localValue := newScalarSlot()
	localValue.assign("local")
	localParameter.bind(localValue)
	scalarSlotRead(t, caller, "mutated")
	scalarSlotRead(t, retainedAlias, "mutated")
	scalarSlotRead(t, localParameter, "local")

	retainedCell := retainedAlias.binding.ref
	caller.unset()
	scalarSlotRead(t, retainedAlias, "mutated")
	if retainedCell.owners != 1 {
		t.Fatalf("retained ref owners after origin unset = %d, want 1", retainedCell.owners)
	}
	retainedAlias.unset()
	if retainedCell.owners != 0 || retainedCell.value != nil {
		t.Fatalf("final retained release = %#v owners=%d", retainedCell.value, retainedCell.owners)
	}

	localParameter.unset()
	scalarSlotRead(t, localValue, "local")
}

func TestScalarSlotHandleIdentity(t *testing.T) {
	value := newScalarSlot()
	value.assign(int64(1))
	handleCopy := value
	handleCopy.assign(int64(2))
	scalarSlotRead(t, value, int64(2))
	if value.binding != handleCopy.binding {
		t.Fatal("Go handle copy did not retain the same binding")
	}

	value.bind(value)
	if value.binding.ref != nil {
		t.Fatal("self binding converted an ordinary binding to a ref cell")
	}
	alias := newScalarSlot()
	alias.bind(value)
	scalarSlotOwners(t, value, 2)
	aliasCopy := alias
	aliasCopy.bind(value)
	scalarSlotOwners(t, value, 2)
	value.bind(alias)
	scalarSlotOwners(t, value, 2)
	if alias.binding != aliasCopy.binding {
		t.Fatal("alias handle copy did not retain the same binding")
	}
	cell := value.binding.ref
	handleCopy.unset()
	if value.isDefined() || handleCopy.isDefined() {
		t.Fatal("unset through a Go handle copy did not remove the same binding")
	}
	scalarSlotRead(t, alias, int64(2))
	scalarSlotOwners(t, alias, 1)
	aliasCopy.unset()
	if alias.isDefined() || cell.owners != 0 || cell.value != nil {
		t.Fatal("unset through alias handle copy did not release its single owner")
	}
}

func TestScalarSlotUndefinedAndDomain(t *testing.T) {
	var zero scalarSlot
	for name, call := range map[string]func(){
		"isDefined": func() { zero.isDefined() },
		"read":      func() { zero.read() },
		"assign":    func() { zero.assign(int64(1)) },
		"bind":      func() { zero.bind(newScalarSlot()) },
		"unset":     func() { zero.unset() },
	} {
		scalarSlotPanic(t, "zero."+name, call)
	}

	undefined := newScalarSlot()
	if undefined.isDefined() {
		t.Fatal("new slot is defined")
	}
	scalarSlotPanic(t, "undefined read", func() { undefined.read() })
	presentNull := newScalarSlot()
	presentNull.assign(nil)
	if !presentNull.isDefined() || presentNull.read() != nil {
		t.Fatal("present null is not a defined null")
	}

	target := newScalarSlot()
	target.assign("before")
	alias := newScalarSlot()
	alias.bind(target)
	cell := target.binding.ref
	ordinary := newScalarSlot()
	ordinary.assign("ordinary")
	var typedNil *scalarSlotTestObject
	for _, invalid := range []any{
		NewArray(), &scalarSlotTestObject{}, typedNil, NewCallable(0, func(...any) any { return nil }),
		float64(1), int(1), scalarRefCell{}, &scalarRefCell{}, scalarBinding{}, &scalarBinding{}, scalarSlot{}, newScalarSlot(),
	} {
		scalarSlotPanic(t, "invalid scalar assignment", func() { target.assign(invalid) })
		scalarSlotPanic(t, "invalid ordinary assignment", func() { ordinary.assign(invalid) })
		scalarSlotRead(t, ordinary, "ordinary")
		scalarSlotRead(t, target, "before")
		scalarSlotRead(t, alias, "before")
		if target.binding.ref != cell || alias.binding.ref != cell || cell.owners != 2 {
			t.Fatal("invalid assignment changed alias ownership")
		}
	}

	destination := newScalarSlot()
	destination.assign("destination")
	scalarSlotPanic(t, "undefined bind source", func() { destination.bind(undefined) })
	scalarSlotRead(t, destination, "destination")
	scalarSlotPanic(t, "zero bind source", func() { destination.bind(zero) })
	scalarSlotRead(t, destination, "destination")
	scalarSlotPanic(t, "undefined source with aliased destination", func() { target.bind(undefined) })
	scalarSlotPanic(t, "zero source with aliased destination", func() { target.bind(zero) })
	if undefined.isDefined() || target.binding.ref != cell || cell.owners != 2 {
		t.Fatal("failed binding changed source or destination ownership")
	}
	scalarSlotRead(t, target, "before")
	scalarSlotRead(t, alias, "before")
}
