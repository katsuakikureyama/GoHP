package storage

import (
	"math"
	"testing"
)

func TestScalarArrayKeyNormalization(t *testing.T) {
	array := NewScalarArray()
	defer array.Close()

	array.Write(int64(1), "integer")
	array.Write("1", "canonical string replaces integer")
	requireArrayRead(t, array, int64(1), "canonical string replaces integer", true)
	requireArrayRead(t, array, "1", "canonical string replaces integer", true)
	requireArrayKeys(t, array, []any{int64(1)})

	source := NewSlot("bound")
	defer source.Unset()
	array.Bind("2", source)
	alias := array.Reference(int64(2))
	defer alias.Unset()
	alias.Write("through reference")
	requireRead(t, source, "through reference", true)
	requireArrayRead(t, array, "2", "through reference", true)
	requireArrayRead(t, array, int64(2), "through reference", true)

	for _, key := range []string{"01", "+1", "-0"} {
		array.Write(key, key)
		requireArrayRead(t, array, key, key, true)
	}
	requireArrayRead(t, array, int64(0), nil, false)
	requireArrayKeys(t, array, []any{int64(1), int64(2), "01", "+1", "-0"})

	missing := array.Reference("5")
	defer missing.Unset()
	requireRead(t, missing, nil, true)
	requireArrayRead(t, array, int64(5), nil, true)
	if got := array.Append("after missing numeric reference"); got != 6 {
		t.Fatalf("Append() after Reference(\"5\") = %d, want 6", got)
	}
	requireArrayRead(t, array, "6", "after missing numeric reference", true)
	requireArrayKeys(t, array, []any{int64(1), int64(2), "01", "+1", "-0", int64(5), int64(6)})
}

func TestScalarArrayAppendIndex(t *testing.T) {
	t.Run("empty and string-only arrays start at zero", func(t *testing.T) {
		array := NewScalarArray()
		defer array.Close()
		if got := array.Append("zero"); got != 0 {
			t.Fatalf("first Append() = %d, want 0", got)
		}
		if got := array.Append("one"); got != 1 {
			t.Fatalf("second Append() = %d, want 1", got)
		}
		requireArrayRead(t, array, int64(0), "zero", true)
		requireArrayRead(t, array, int64(1), "one", true)

		strings := NewScalarArray()
		defer strings.Close()
		strings.Write("name", "value")
		if got := strings.Append("first integer"); got != 0 {
			t.Fatalf("Append() after a string key = %d, want 0", got)
		}
	})

	t.Run("first negative key establishes its successor", func(t *testing.T) {
		array := NewScalarArray()
		defer array.Close()
		array.Write("name", "prefix")
		array.Write("-4", "negative")
		if got := array.Append("after negative"); got != -3 {
			t.Fatalf("Append() after key -4 = %d, want -3", got)
		}
	})

	t.Run("lower keys and replacements never rewind", func(t *testing.T) {
		array := NewScalarArray()
		defer array.Close()
		array.Write(int64(5), "high")
		array.Write(int64(2), "lower")
		array.Write("5", "replacement")
		if got := array.Append("next"); got != 6 {
			t.Fatalf("Append() after 5, 2 and replacement 5 = %d, want 6", got)
		}
		requireArrayRead(t, array, int64(5), "replacement", true)
	})

	t.Run("separated copies fork their automatic indices", func(t *testing.T) {
		base := NewScalarArray()
		defer base.Close()
		base.Write(int64(8), "seed")
		left := base.Copy()
		defer left.Close()
		right := base.Copy()
		defer right.Close()

		for name, array := range map[string]*ScalarArray{"left": left, "right": right, "base": base} {
			if got := array.Append(name); got != 9 {
				t.Fatalf("%s first Append() = %d, want 9", name, got)
			}
			if got := array.Append(name + " again"); got != 10 {
				t.Fatalf("%s second Append() = %d, want 10", name, got)
			}
		}
		for name, array := range map[string]*ScalarArray{"left": left, "right": right, "base": base} {
			requireArrayRead(t, array, int64(9), name, true)
			requireArrayRead(t, array, int64(10), name+" again", true)
		}
	})

	t.Run("reference-owned elements retain their aliases across append separation", func(t *testing.T) {
		external := NewSlot("seed")
		defer external.Unset()
		first := NewScalarArray()
		defer first.Close()
		first.Bind(int64(0), external)
		second := first.Copy()
		defer second.Close()

		if got := second.Append("new element"); got != 1 {
			t.Fatalf("Append() after key 0 = %d, want 1", got)
		}
		requireArrayRead(t, second, int64(1), "new element", true)
		requireArrayRead(t, first, int64(1), nil, false)
		external.Write("through external alias")
		requireArrayRead(t, first, int64(0), "through external alias", true)
		requireArrayRead(t, second, int64(0), "through external alias", true)
	})

	t.Run("integer boundaries", func(t *testing.T) {
		minimum := NewScalarArray()
		defer minimum.Close()
		minimum.Write(int64(math.MinInt64), "minimum")
		if got := minimum.Append("after minimum"); got != math.MinInt64+1 {
			t.Fatalf("Append() after MinInt64 = %d, want %d", got, int64(math.MinInt64+1))
		}

		maximum := NewScalarArray()
		defer maximum.Close()
		maximum.Write(int64(math.MaxInt64), "maximum")
		requireArrayRead(t, maximum, int64(math.MaxInt64), "maximum", true)
		requirePanic(t, func() { maximum.Append("overflow") })

		lastAvailable := NewScalarArray()
		defer lastAvailable.Close()
		lastAvailable.Write("9223372036854775806", "before maximum")
		if got := lastAvailable.Append("last available"); got != math.MaxInt64 {
			t.Fatalf("last available Append() = %d, want MaxInt64", got)
		}
		requireArrayRead(t, lastAvailable, int64(math.MaxInt64), "last available", true)
		blockedCopy := lastAvailable.Copy()
		defer blockedCopy.Close()
		blockedCopy.Write(int64(-1), "lower key cannot unblock")
		blockedCopy.Write("9223372036854775807", "replacement cannot unblock")
		requirePanic(t, func() { blockedCopy.Append("overflow after copy") })
		requireArrayRead(t, lastAvailable, int64(math.MaxInt64), "last available", true)
	})
}

func TestScalarArrayAppendRejectsWithoutSeparation(t *testing.T) {
	t.Run("invalid payload leaves sharing and the next index intact", func(t *testing.T) {
		external := NewSlot("shared")
		defer external.Unset()
		first := NewScalarArray()
		defer first.Close()
		first.Write(int64(4), "index seed")
		first.Bind("ref", external)
		second := first.Copy()
		defer second.Close()

		requirePanic(t, func() { second.Append(int(1)) })
		requirePanic(t, func() { second.Append([]any{"nested"}) })
		external.Unset()

		if got := second.Append("valid after failures"); got != 5 {
			t.Fatalf("Append() after rejected payloads = %d, want 5", got)
		}
		second.Write("ref", "second only")
		requireArrayRead(t, first, "ref", "shared", true)
		requireArrayRead(t, second, "ref", "second only", true)
	})

	t.Run("overflow leaves a shared reference container attached", func(t *testing.T) {
		external := NewSlot("shared")
		defer external.Unset()
		first := NewScalarArray()
		defer first.Close()
		first.Write(int64(math.MaxInt64), "maximum")
		first.Bind("ref", external)
		second := first.Copy()
		defer second.Close()

		requirePanic(t, func() { second.Append("overflow") })
		requirePanic(t, func() { second.Append("overflow again") })
		external.Unset()

		second.Write("ref", "second only")
		requireArrayRead(t, first, "ref", "shared", true)
		requireArrayRead(t, second, "ref", "second only", true)
		requirePanic(t, func() { second.Append("still overflow") })
		requireArrayRead(t, first, int64(math.MaxInt64), "maximum", true)
		requireArrayRead(t, second, int64(math.MaxInt64), "maximum", true)
	})
}
