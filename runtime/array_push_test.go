package runtime

import (
	"math"
	"reflect"
	"testing"
)

// arrayPushProbeObject is a declared-slot generated-object stand-in.  Its
// counters let these tests distinguish resolving a location from publishing
// a new value into it.
type arrayPushProbeObject struct {
	fields   map[string]any
	getCalls int
	setCalls int
}

func (o *arrayPushProbeObject) PHPObject() bool { return o != nil }
func (o *arrayPushProbeObject) Get(name string) any {
	o.getCalls++
	value, ok := o.fields[name]
	if !ok {
		panic("undefined property")
	}
	return Clone(value)
}
func (o *arrayPushProbeObject) Set(name string, value any) {
	if _, ok := o.fields[name]; !ok {
		panic("undefined property assignment")
	}
	o.setCalls++
	o.fields[name] = Clone(value)
}

type arrayPushFalsePropertyObject struct{}

func (*arrayPushFalsePropertyObject) PHPObject() bool { return false }
func (*arrayPushFalsePropertyObject) Get(string) any  { return NewArray() }
func (*arrayPushFalsePropertyObject) Set(string, any) {}

func TestArrayPushLocalValues(t *testing.T) {
	t.Run("zero multiple copies and identities", func(t *testing.T) {
		initial := NewArray(
			Entry{Key: "label", Value: "kept"},
			Entry{Key: int64(-5), Value: "negative"},
		)
		locals := map[string]any{"items": initial}
		if got := ArrayPushLocal(locals, "items"); got != 2 {
			t.Fatalf("zero-value result = %d, want 2", got)
		}
		if !reflect.DeepEqual(locals["items"], initial) {
			t.Fatalf("zero-value call changed local: %#v", locals["items"])
		}

		nestedInput := NewArray(Entry{Key: "deep", Value: "before"})
		child := &objectMarker{valid: true}
		if got := ArrayPushLocal(locals, "items", "first", nestedInput, child); got != 5 {
			t.Fatalf("multiple-value result = %d, want 5", got)
		}
		items := locals["items"].(Array)
		wantKeys := []any{"label", int64(-5), int64(-4), int64(-3), int64(-2)}
		for i, want := range wantKeys {
			if got := items.Entries[i].Key; got != want {
				t.Fatalf("key[%d] = %#v, want %#v", i, got, want)
			}
		}
		if items.Entries[2].Value != "first" || items.Entries[4].Value != child {
			t.Fatalf("multiple values lost order or child identity: %#v", items.Entries)
		}
		nestedInput.Entries[0].Value = "changed input"
		if got := Get(Get(items, int64(-3)), "deep"); got != "before" {
			t.Fatalf("appended nested value aliases argument: %#v", got)
		}
	})

	t.Run("self snapshots the pre-write local", func(t *testing.T) {
		initial := NewArray(Entry{Key: "entry", Value: NewArray(Entry{Key: "deep", Value: "before"})})
		locals := map[string]any{"items": initial}
		snapshot := locals["items"]
		if got := ArrayPushLocal(locals, "items", snapshot); got != 2 {
			t.Fatalf("self-push result = %d, want 2", got)
		}
		items := locals["items"].(Array)
		appended := items.Entries[1].Value.(Array)
		appended.Entries[0].Value.(Array).Entries[0].Value = "changed snapshot"
		if got := Get(Get(items, "entry"), "deep"); got != "before" {
			t.Fatalf("self-push aliases current root: %#v", got)
		}
		if got := Get(Get(initial, "entry"), "deep"); got != "before" {
			t.Fatalf("self-push changed original binding: %#v", got)
		}
	})

	t.Run("maximum integer key remains appendable once", func(t *testing.T) {
		locals := map[string]any{"items": NewArray(Entry{Key: int64(math.MaxInt64 - 1), Value: "before"})}
		if got := ArrayPushLocal(locals, "items", "maximum"); got != 2 {
			t.Fatalf("maximum result = %d, want 2", got)
		}
		items := locals["items"].(Array)
		if got := items.Entries[1]; !reflect.DeepEqual(got, Entry{Key: int64(math.MaxInt64), Value: "maximum"}) {
			t.Fatalf("maximum append = %#v", got)
		}
	})
}

func TestArrayPushPropertyValues(t *testing.T) {
	host := &arrayPushProbeObject{fields: map[string]any{
		"items": NewArray(Entry{Key: "old", Value: "resolved value"}),
	}}
	target := ResolveProperty(host, "items")
	if host.getCalls != 1 {
		t.Fatalf("resolution Get calls = %d, want 1", host.getCalls)
	}

	// The handle names an owner and slot; Push must reread that slot rather than
	// use the value observed at resolve time.
	host.Set("items", NewArray(Entry{Key: "replacement", Value: "current value"}))
	setBefore := host.setCalls
	getBefore := host.getCalls
	if got := target.Push(); got != 1 {
		t.Fatalf("zero-value property result = %d, want 1", got)
	}
	if host.getCalls != getBefore+1 || host.setCalls != setBefore {
		t.Fatalf("zero-value property call Get/Set = %d/%d, want %d/%d", host.getCalls, host.setCalls, getBefore+1, setBefore)
	}
	if got := host.fields["items"].(Array).Entries[0].Value; got != "current value" {
		t.Fatalf("property push used stale resolved value: %#v", got)
	}

	nestedInput := NewArray(Entry{Key: "deep", Value: "before"})
	child := &objectMarker{valid: true}
	if got := target.Push("first", nestedInput, child); got != 4 {
		t.Fatalf("multiple-value property result = %d, want 4", got)
	}
	if host.setCalls != setBefore+3 {
		t.Fatalf("multiple property appends published %d Set calls, want 3", host.setCalls-setBefore)
	}
	items := host.fields["items"].(Array)
	if got, want := []any{items.Entries[0].Value, items.Entries[1].Value, items.Entries[2].Value, items.Entries[3].Value}, []any{"current value", "first", NewArray(Entry{Key: "deep", Value: "before"}), child}; !reflect.DeepEqual(got[:3], want[:3]) || got[3] != want[3] {
		t.Fatalf("property values lost order/copy/identity: %#v", items.Entries)
	}
	nestedInput.Entries[0].Value = "changed input"
	if got := Get(Get(items, int64(1)), "deep"); got != "before" {
		t.Fatalf("property append aliases argument: %#v", got)
	}

	assertArrayPushPanic(t, func() { ResolveProperty(host, "missing") }, "undefined property")
	assertArrayPushPanic(t, func() { (NestedProperty{}).Push() }, "nested property requires object")
}

func TestArrayPushRejectsInvalidDomains(t *testing.T) {
	t.Run("locals", func(t *testing.T) {
		assertArrayPushPanic(t, func() { ArrayPushLocal(map[string]any{}, "items") }, "array_push requires an existing local")
		for _, tc := range []struct {
			name  string
			value any
		}{
			{"nil", nil},
			{"boolean", false},
			{"string", "items"},
			{"integer", int64(1)},
			{"closure", func() {}},
			{"object", &objectMarker{valid: true}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				assertArrayPushPanic(t, func() { ArrayPushLocal(map[string]any{"items": tc.value}, "items") }, "array_push requires array")
			})
		}
	})

	t.Run("properties", func(t *testing.T) {
		var typedNil *arrayPushProbeObject
		assertArrayPushPanic(t, func() { ResolveProperty(nil, "items") }, "nested property requires object")
		assertArrayPushPanic(t, func() { ResolveProperty(typedNil, "items") }, "nested property requires object")
		assertArrayPushPanic(t, func() { ResolveProperty(&arrayPushFalsePropertyObject{}, "items") }, "nested property requires object")

		for _, tc := range []struct {
			name  string
			value any
		}{
			{"nil", nil},
			{"boolean", false},
			{"string", "items"},
			{"integer", int64(1)},
			{"closure", func() {}},
			{"object", &objectMarker{valid: true}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				host := &arrayPushProbeObject{fields: map[string]any{"items": tc.value}}
				target := ResolveProperty(host, "items")
				assertArrayPushPanic(t, func() { target.Push() }, "array_push requires array")
				if host.setCalls != 0 {
					t.Fatalf("invalid property published %d Set calls", host.setCalls)
				}
			})
		}
	})

	t.Run("overflow retains successful prefix only", func(t *testing.T) {
		initial := NewArray(Entry{Key: int64(math.MaxInt64 - 1), Value: "before"})
		locals := map[string]any{"items": initial}
		assertArrayPushPanic(t, func() { ArrayPushLocal(locals, "items", "maximum", "overflow", "unreached") }, "array key overflow unsupported")
		localItems := locals["items"].(Array)
		if len(localItems.Entries) != 2 || Get(localItems, int64(math.MaxInt64)) != "maximum" {
			t.Fatalf("local overflow did not retain exactly the successful prefix: %#v", localItems.Entries)
		}

		host := &arrayPushProbeObject{fields: map[string]any{"items": initial}}
		target := ResolveProperty(host, "items")
		setBefore := host.setCalls
		assertArrayPushPanic(t, func() { target.Push("maximum", "overflow", "unreached") }, "array key overflow unsupported")
		propertyItems := host.fields["items"].(Array)
		if len(propertyItems.Entries) != 2 || Get(propertyItems, int64(math.MaxInt64)) != "maximum" {
			t.Fatalf("property overflow did not retain exactly the successful prefix: %#v", propertyItems.Entries)
		}
		if host.setCalls != setBefore+1 {
			t.Fatalf("property overflow published %d values, want 1", host.setCalls-setBefore)
		}
	})
}

func assertArrayPushPanic(t *testing.T, call func(), want string) {
	t.Helper()
	defer func() {
		if got := recover(); got != want {
			t.Fatalf("panic = %v, want %q", got, want)
		}
	}()
	call()
}
