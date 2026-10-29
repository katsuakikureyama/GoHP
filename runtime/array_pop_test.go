package runtime

import (
	"math"
	"reflect"
	"testing"
)

// arrayPopPartialPropertyObject deliberately has the object marker and a
// getter but not the full writable PropertyObject protocol.
type arrayPopPartialPropertyObject struct{}

func (*arrayPopPartialPropertyObject) PHPObject() bool { return true }
func (*arrayPopPartialPropertyObject) Get(string) any  { return NewArray() }

func TestArrayPopLocalValues(t *testing.T) {
	child := &objectMarker{valid: true}
	initial := NewArray(
		Entry{Key: "first", Value: "kept"},
		Entry{Key: "child", Value: child},
		Entry{Key: int64(7), Value: NewArray(Entry{Key: "deep", Value: "before"})},
	)
	original := Clone(initial).(Array)
	locals := map[string]any{"items": initial}

	got := ArrayPopLocal(locals, "items")
	popped, ok := got.(Array)
	if !ok {
		t.Fatalf("popped value type = %T, want runtime.Array", got)
	}
	if !reflect.DeepEqual(popped, NewArray(Entry{Key: "deep", Value: "before"})) {
		t.Fatalf("popped nested value = %#v", popped)
	}
	remaining := locals["items"].(Array)
	if got, want := remaining.Entries, []Entry{{Key: "first", Value: "kept"}, {Key: "child", Value: child}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("remaining entries = %#v, want %#v", got, want)
	}
	if remaining.Entries[1].Value != child {
		t.Fatal("remaining child object lost identity")
	}
	if !reflect.DeepEqual(initial, original) {
		t.Fatalf("pop changed original local storage: got %#v, want %#v", initial, original)
	}
	remaining.Entries[0].Value = "changed remaining storage"
	if initial.Entries[0].Value != "kept" {
		t.Fatal("remaining array aliases original entry storage")
	}

	popped.Entries[0].Value = "changed returned value"
	if got := Get(initial.Entries[2].Value, "deep"); got != "before" {
		t.Fatalf("popped nested array aliases original storage: %#v", got)
	}
	if got := Get(original.Entries[2].Value, "deep"); got != "before" {
		t.Fatalf("popped nested array aliases snapshot storage: %#v", got)
	}

	// Array-pop's next-key state follows the removed integer key, without
	// rebuilding it from all remaining entries.
	items := NewArray(Entry{Key: int64(-5), Value: "negative"})
	locals = map[string]any{"items": items}
	if got := ArrayPopLocal(locals, "items"); got != "negative" {
		t.Fatalf("pop negative value = %#v, want %q", got, "negative")
	}
	remaining = locals["items"].(Array)
	appended := ArrayAppend(remaining, "again").(Array)
	if got := appended.Entries[0].Key; got != int64(-5) {
		t.Fatalf("append key after popping -5 = %#v, want -5", got)
	}

	empty := NewArray()
	locals = map[string]any{"items": empty}
	if got := ArrayPopLocal(locals, "items"); got != nil {
		t.Fatalf("empty pop = %#v, want nil", got)
	}
	if !reflect.DeepEqual(locals["items"], empty) {
		t.Fatalf("empty pop changed local: got %#v, want %#v", locals["items"], empty)
	}
	if got := ArrayAppend(locals["items"], "first").(Array).Entries[0].Key; got != int64(0) {
		t.Fatalf("append key after string-free empty pop = %#v, want 0", got)
	}

	// The max entry itself is removable and leaves MAX available for exactly
	// one later automatic append.
	locals = map[string]any{"items": NewArray(Entry{Key: int64(math.MaxInt64), Value: "edge"})}
	if got := ArrayPopLocal(locals, "items"); got != "edge" {
		t.Fatalf("max pop = %#v, want edge", got)
	}
	if got := ArrayAppend(locals["items"], "maximum").(Array).Entries[0].Key; got != int64(math.MaxInt64) {
		t.Fatalf("append key after max pop = %#v, want MaxInt64", got)
	}
	locals = map[string]any{"items": NewArray(
		Entry{Key: int64(math.MaxInt64), Value: "maximum kept"},
		Entry{Key: int64(math.MaxInt64 - 1), Value: "lower tail"},
	)}
	if got := ArrayPopLocal(locals, "items"); got != "lower tail" {
		t.Fatalf("out-of-order maximum pop = %#v", got)
	}
	withLower := ArrayAppend(locals["items"], "lower readded").(Array)
	if Get(withLower, int64(math.MaxInt64)) != "maximum kept" || Get(withLower, int64(math.MaxInt64-1)) != "lower readded" {
		t.Fatal("append after predecessor pop changed maximum occupancy")
	}
	assertArrayPopPanic(t, func() { ArrayAppend(withLower, "must not replace maximum") }, "array key overflow unsupported")
}

func TestArrayPopPropertyValues(t *testing.T) {
	host := &arrayPushProbeObject{fields: map[string]any{
		"items": NewArray(Entry{Key: "old", Value: "stale"}),
	}}
	target := ResolveProperty(host, "items")
	if host.getCalls != 1 {
		t.Fatalf("resolution Get calls = %d, want 1", host.getCalls)
	}

	child := &objectMarker{valid: true}
	replacement := NewArray(
		Entry{Key: "child", Value: child},
		Entry{Key: "tail", Value: NewArray(Entry{Key: "deep", Value: "before"})},
	)
	host.Set("items", replacement)
	setBefore := host.setCalls
	getBefore := host.getCalls
	got := target.Pop()
	popped, ok := got.(Array)
	if !ok || Get(popped, "deep") != "before" {
		t.Fatalf("property popped value = %#v, want detached nested array", got)
	}
	if host.getCalls != getBefore+1 || host.setCalls != setBefore+1 {
		t.Fatalf("nonempty Pop Get/Set = %d/%d, want %d/%d", host.getCalls, host.setCalls, getBefore+1, setBefore+1)
	}
	remaining := host.fields["items"].(Array)
	if len(remaining.Entries) != 1 || remaining.Entries[0].Value != child {
		t.Fatalf("property pop remaining value = %#v, want retained child", remaining.Entries)
	}
	popped.Entries[0].Value = "changed returned value"
	if got := Get(replacement.Entries[1].Value, "deep"); got != "before" {
		t.Fatalf("property popped array aliases pre-pop replacement: %#v", got)
	}

	host.Set("items", NewArray())
	setBefore = host.setCalls
	getBefore = host.getCalls
	if got := target.Pop(); got != nil {
		t.Fatalf("empty property pop = %#v, want nil", got)
	}
	if host.getCalls != getBefore+1 || host.setCalls != setBefore {
		t.Fatalf("empty Pop Get/Set = %d/%d, want %d/%d", host.getCalls, host.setCalls, getBefore+1, setBefore)
	}
}

func TestArrayPopRejectsInvalidDomains(t *testing.T) {
	assertArrayPopPanic(t, func() { ArrayPopLocal(map[string]any{}, "items") }, "array_pop requires an existing local")

	for _, tc := range []struct {
		name  string
		value any
	}{
		{"nil", nil},
		{"typed_nil", (*arrayPushProbeObject)(nil)},
		{"boolean", false},
		{"string", "items"},
		{"integer", int64(1)},
		{"closure", func() {}},
		{"object", &objectMarker{valid: true}},
	} {
		t.Run("local_"+tc.name, func(t *testing.T) {
			locals := map[string]any{"items": tc.value}
			before := locals["items"]
			assertArrayPopPanic(t, func() { ArrayPopLocal(locals, "items") }, "array_pop requires array")
			if !sameArrayPopValue(locals["items"], before) {
				t.Fatalf("rejected local pop changed value: got %#v, want %#v", locals["items"], before)
			}
		})
	}

	var typedNil *arrayPushProbeObject
	for _, tc := range []struct {
		name  string
		value any
	}{
		{"nil", nil},
		{"typed_nil", typedNil},
		{"partial", &arrayPopPartialPropertyObject{}},
		{"false_marker", &arrayPushFalsePropertyObject{}},
	} {
		t.Run("property_receiver_"+tc.name, func(t *testing.T) {
			assertArrayPopPanic(t, func() { ResolveProperty(tc.value, "items") }, "nested property requires object")
		})
	}

	for _, tc := range []struct {
		name  string
		value any
	}{
		{"nil", nil},
		{"typed_nil", (*arrayPushProbeObject)(nil)},
		{"boolean", false},
		{"string", "items"},
		{"integer", int64(1)},
		{"closure", func() {}},
		{"object", &objectMarker{valid: true}},
	} {
		t.Run("property_value_"+tc.name, func(t *testing.T) {
			host := &arrayPushProbeObject{fields: map[string]any{"items": tc.value}}
			target := ResolveProperty(host, "items")
			setBefore := host.setCalls
			before := host.fields["items"]
			assertArrayPopPanic(t, func() { target.Pop() }, "array_pop requires array")
			if host.setCalls != setBefore || !sameArrayPopValue(host.fields["items"], before) {
				t.Fatalf("rejected property pop published or changed value: Set=%d, value=%#v", host.setCalls-setBefore, host.fields["items"])
			}
		})
	}
}

func sameArrayPopValue(got, want any) bool {
	gotFunction, gotIsFunction := got.(func())
	wantFunction, wantIsFunction := want.(func())
	if gotIsFunction || wantIsFunction {
		return gotIsFunction && wantIsFunction && reflect.ValueOf(gotFunction).Pointer() == reflect.ValueOf(wantFunction).Pointer()
	}
	return reflect.DeepEqual(got, want)
}

func assertArrayPopPanic(t *testing.T, call func(), want string) {
	t.Helper()
	defer func() {
		if got := recover(); got != want {
			t.Fatalf("panic = %v, want %q", got, want)
		}
	}()
	call()
}
