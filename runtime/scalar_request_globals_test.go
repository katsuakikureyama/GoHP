package runtime

import "testing"

func TestScalarRequestGlobalBinding(t *testing.T) {
	execution := &Execution{}
	defer execution.Close()
	first := NewScalarFrame()
	nested := NewScalarFrame()
	defer first.Close()
	defer nested.Close()

	if _, present := execution.ReadScalarGlobal("counter"); present {
		t.Fatal("missing lookup created or reported a global")
	}
	first.ImportScalarGlobal(execution, "counter")
	if value, present := execution.ReadScalarGlobal("counter"); !present || value != nil {
		t.Fatalf("missing import global = (%#v, %v), want (nil, true)", value, present)
	}
	first.Write("counter", int64(1))
	nested.ImportScalarGlobal(execution, "counter")
	if got := nested.Read("counter"); got != int64(1) {
		t.Fatalf("nested import = %#v, want 1", got)
	}
	nested.Write("counter", int64(2))
	if got := first.Read("counter"); got != int64(2) {
		t.Fatalf("sibling global write = %#v, want 2", got)
	}
	for name, want := range map[string]any{
		"nullValue": nil, "boolValue": true, "intValue": int64(-42), "byteValue": "\x00\xff",
	} {
		execution.WriteScalarGlobal(name, want)
		if got, present := execution.ReadScalarGlobal(name); !present || got != want {
			t.Fatalf("scalar global %q = (%#v, %v), want %#v", name, got, present, want)
		}
	}

	first.Write("copy", first.Read("counter"))
	first.Write("counter", int64(3))
	if got := first.Read("copy"); got != int64(2) {
		t.Fatalf("ordinary local copy changed to %#v, want 2", got)
	}

	local := NewScalarFrame()
	defer local.Close()
	local.Write("counter", "local")
	if got := local.Read("counter"); got != "local" {
		t.Fatalf("ordinary local = %#v, want local", got)
	}
	if got := first.Read("counter"); got != int64(3) {
		t.Fatalf("ordinary local changed global binding to %#v", got)
	}
	local.ImportScalarGlobal(execution, "counter")
	if got := local.Read("counter"); got != int64(3) {
		t.Fatalf("import did not replace shadowing local: %#v", got)
	}

	first.Close()
	if value, present := execution.ReadScalarGlobal("counter"); !present || value != int64(3) {
		t.Fatalf("frame close changed request global to (%#v, %v)", value, present)
	}
	execution.WriteScalarGlobal("counter", int64(4))
	if got := nested.Read("counter"); got != int64(4) {
		t.Fatalf("host global write did not reach remaining local: %#v", got)
	}
}

func TestScalarRequestGlobalUnsetAndRecreate(t *testing.T) {
	execution := &Execution{}
	defer execution.Close()
	frame := NewScalarFrame()
	defer frame.Close()
	other := NewScalarFrame()
	defer other.Close()

	execution.WriteScalarGlobal("value", "old")
	frame.ImportScalarGlobal(execution, "value")
	other.ImportScalarGlobal(execution, "value")
	execution.UnsetScalarGlobal("value")
	if _, present := execution.ReadScalarGlobal("value"); present {
		t.Fatal("unset global remains present")
	}
	if got := frame.Read("value"); got != "old" {
		t.Fatalf("global unset destroyed retained local alias: %#v", got)
	}

	execution.WriteScalarGlobal("value", "new")
	other.ImportScalarGlobal(execution, "value") // Replace an old live alias directly.
	other.Write("value", "newer")
	frame.Write("value", "retained")
	if value, present := execution.ReadScalarGlobal("value"); !present || value != "newer" {
		t.Fatalf("retained alias reconnected to recreated global: (%#v, %v)", value, present)
	}
	frame.Unset("value")
	frame.Write("value", "local")
	if got := frame.Read("value"); got != "local" {
		t.Fatalf("local after unset = %#v, want local", got)
	}
	if value, present := execution.ReadScalarGlobal("value"); !present || value != "newer" {
		t.Fatalf("local recreation changed global: (%#v, %v)", value, present)
	}
	frame.ImportScalarGlobal(execution, "value")
	if got := frame.Read("value"); got != "newer" {
		t.Fatalf("repeated import = %#v, want new global value", got)
	}
}

func TestScalarRequestIsolationAndClose(t *testing.T) {
	a, b := &Execution{}, &Execution{}
	defer a.Close()
	defer b.Close()
	fa, fb := NewScalarFrame(), NewScalarFrame()
	defer fa.Close()
	defer fb.Close()
	fa.ImportScalarGlobal(a, "same")
	fb.ImportScalarGlobal(b, "same")
	fa.Write("same", "a")
	fb.Write("same", "b")
	if got := fa.Read("same"); got != "a" {
		t.Fatalf("A observed %#v, want a", got)
	}
	if got := fb.Read("same"); got != "b" {
		t.Fatalf("B observed %#v, want b", got)
	}
	ordinary := NewScalarFrame()
	defer ordinary.Close()
	ordinary.Write("local", "still-open")
	b.Strtok("x,y", ",")
	b.JSONDecodeAssociative("{")

	// Establish state from several existing Execution users before Close.
	if got := a.Strtok("a,b", ","); got != "a" {
		t.Fatalf("strtok setup = %#v", got)
	}
	if got := a.JSONLastError(); got != 0 {
		t.Fatalf("initial JSON error = %d", got)
	}
	if got := a.PregLastError83(); got != 0 {
		t.Fatalf("initial PCRE error = %d", got)
	}
	a.Close()
	a.Close()

	for _, action := range []func(){
		func() { RequireExecution(a) },
		func() { a.ReadScalarGlobal("same") },
		func() { a.WriteScalarGlobal("same", "again") },
		func() { a.UnsetScalarGlobal("same") },
		func() { a.Strtok(",") },
		func() { a.JSONLastError() },
		func() { a.JSONLastErrorMessage() },
		func() { a.JSONDecodeAssociative("null") },
		func() { a.PregLastError83() },
		func() { a.PregMatch83Local("~x~", "x", int64(0)) },
		func() { a.PregMatch83LocalMatch1("~x~", "x", int64(0)) },
		func() { InvokeWithExecution(a, NewCallable(0, func(...any) any { return nil })) },
		func() { fa.Read("same") },
		func() { fa.Write("same", "again") },
		func() { fa.ImportScalarGlobal(a, "same") },
		func() { fa.Unset("same") },
	} {
		requireScalarFramePanic(t, action)
	}
	// A's close does not mutate another request or a fresh request.
	if got := fb.Read("same"); got != "b" {
		t.Fatalf("closing A changed B to %#v", got)
	}
	if b.Strtok(",") != "y" || b.JSONLastError() != 4 {
		t.Fatal("closing A changed B's tokenizer/error state")
	}
	c := &Execution{}
	defer c.Close()
	if _, present := c.ReadScalarGlobal("same"); present {
		t.Fatal("fresh request inherited a global")
	}

	if got := ordinary.Read("local"); got != "still-open" {
		t.Fatalf("nonimported frame changed to %#v", got)
	}
}

func TestScalarRequestGlobalGuards(t *testing.T) {
	execution := &Execution{}
	defer execution.Close()
	frame := NewScalarFrame()
	defer frame.Close()
	execution.WriteScalarGlobal("stable", int64(7))
	frame.ImportScalarGlobal(execution, "stable")

	for _, name := range []string{"", "_private", "9lives", "é", "a\x00b", "this", "GLOBALS"} {
		for _, action := range []func(){
			func() { execution.ReadScalarGlobal(name) },
			func() { execution.WriteScalarGlobal(name, int64(1)) },
			func() { execution.UnsetScalarGlobal(name) },
			func() { frame.ImportScalarGlobal(execution, name) },
		} {
			requireScalarFramePanic(t, action)
		}
	}
	for _, value := range []any{int(7), float64(7), []any{"array"}, NewArray(), &struct{}{}, func() {}} {
		requireScalarFramePanic(t, func() { execution.WriteScalarGlobal("stable", value) })
		requireScalarFramePanic(t, func() { execution.WriteScalarGlobal("missing", value) })
		requireScalarFramePanic(t, func() { frame.Write("stable", value) })
	}
	if _, present := execution.ReadScalarGlobal("missing"); present {
		t.Fatal("invalid value created a missing global")
	}
	if value, present := execution.ReadScalarGlobal("stable"); !present || value != int64(7) {
		t.Fatalf("invalid input changed existing global: (%#v, %v)", value, present)
	}
	if got := frame.Read("stable"); got != int64(7) {
		t.Fatalf("invalid input changed existing alias: %#v", got)
	}

	other := &Execution{}
	defer other.Close()
	requireScalarFramePanic(t, func() { frame.ImportScalarGlobal(other, "other") })
	if _, present := other.ReadScalarGlobal("other"); present {
		t.Fatal("cross-request import created a global")
	}
	if got := frame.Read("stable"); got != int64(7) {
		t.Fatalf("cross-request import changed old alias: %#v", got)
	}
	frame.Unset("stable")
	requireScalarFramePanic(t, func() { frame.ImportScalarGlobal(other, "other") })
	if _, present := other.ReadScalarGlobal("other"); present {
		t.Fatal("cross-request import after unset created a global")
	}
	frame.Close()
	requireScalarFramePanic(t, func() { frame.ImportScalarGlobal(other, "other") })

	var nilFrame *ScalarFrame
	requireScalarFramePanic(t, func() { nilFrame.ImportScalarGlobal(execution, "missing") })
	if _, present := execution.ReadScalarGlobal("missing"); present {
		t.Fatal("nil frame import created a global")
	}
	closed := NewScalarFrame()
	closed.Close()
	requireScalarFramePanic(t, func() { closed.ImportScalarGlobal(execution, "missing") })
	if _, present := execution.ReadScalarGlobal("missing"); present {
		t.Fatal("closed frame import created a global")
	}
	var nilExecution *Execution
	requireScalarFramePanic(t, func() { NewScalarFrame().ImportScalarGlobal(nilExecution, "missing") })
	for _, action := range []func(){
		func() { nilExecution.ReadScalarGlobal("missing") },
		func() { nilExecution.WriteScalarGlobal("missing", nil) },
		func() { nilExecution.UnsetScalarGlobal("missing") },
		func() { nilExecution.Close() },
		func() { nilFrame.Unset("missing") },
	} {
		requireScalarFramePanic(t, action)
	}
	if _, present := execution.ReadScalarGlobal("missing"); present {
		t.Fatal("nil execution import created a global")
	}
}
