package runtime

import "testing"

func TestScalarFrameValuesAndMissing(t *testing.T) {
	frame := NewScalarFrame()
	defer frame.Close()

	for name, want := range map[string]any{
		"null": nil,
		"bool": true,
		"int":  int64(-42),
		"text": "scalar\x00value",
	} {
		frame.Write(name, want)
		if got := frame.Read(name); got != want {
			t.Fatalf("Read(%q) = %#v, want %#v", name, got, want)
		}
	}

	requireScalarFramePanic(t, func() { frame.Read("missing") })

	var other ScalarFrame
	defer other.Close()
	frame.Write("copy", int64(1))
	other.Write("copy", frame.Read("copy"))
	frame.Write("copy", int64(2))
	if got := other.Read("copy"); got != int64(1) {
		t.Fatalf("ordinary scalar copy changed to %#v, want 1", got)
	}
}

func TestScalarFrameRejectsInvalidAndClosedUse(t *testing.T) {
	frame := NewScalarFrame()
	defer frame.Close()
	frame.Write("value", "before")

	for _, invalid := range []any{int(1), float64(1), []any{"array"}, Array{}, &struct{}{}, (*ScalarFrame)(nil), func() {}} {
		requireScalarFramePanic(t, func() { frame.Write("value", invalid) })
		if got := frame.Read("value"); got != "before" {
			t.Fatalf("invalid write changed value to %#v", got)
		}
	}

	frame.Close()
	frame.Close()
	requireScalarFramePanic(t, func() { frame.Read("value") })
	requireScalarFramePanic(t, func() { frame.Write("value", "after") })

	var nilFrame *ScalarFrame
	requireScalarFramePanic(t, func() { nilFrame.Read("value") })
	requireScalarFramePanic(t, func() { nilFrame.Write("value", "after") })
	requireScalarFramePanic(t, func() { nilFrame.Close() })
}

func requireScalarFramePanic(t *testing.T, action func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("operation did not panic")
		}
	}()
	action()
}
