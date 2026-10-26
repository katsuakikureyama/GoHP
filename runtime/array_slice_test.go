package runtime

import (
	"math"
	"reflect"
	"testing"
)

func TestArraySliceBounds(t *testing.T) {
	input := NewArray(Entry{Value: "a"}, Entry{Value: "b"}, Entry{Value: "c"}, Entry{Value: "d"})
	for _, tc := range []struct {
		offset int64
		length any
		want   string
	}{
		{0, nil, "abcd"}, {1, nil, "bcd"}, {-2, nil, "cd"}, {-9, nil, "abcd"},
		{4, nil, ""}, {9, nil, ""}, {1, int64(2), "bc"}, {1, int64(0), ""},
		{1, int64(-1), "bc"}, {-3, int64(-2), "b"}, {3, int64(-2), ""},
		{0, int64(math.MinInt64), ""}, {1, int64(math.MaxInt64), "bcd"},
		{math.MinInt64, nil, "abcd"}, {math.MaxInt64, nil, ""},
		{math.MinInt64, int64(math.MinInt64), ""}, {math.MaxInt64, int64(math.MaxInt64), ""},
	} {
		got := ArraySlice(input, tc.offset, tc.length, false).(Array)
		var values string
		for i, entry := range got.Entries {
			values += entry.Value.(string)
			if entry.Key != int64(i) {
				t.Fatalf("unexpected key: %v", entry.Key)
			}
		}
		if values != tc.want {
			t.Fatalf("offset %d length %v: %q != %q", tc.offset, tc.length, values, tc.want)
		}
	}
	for _, offset := range []int64{math.MinInt64, 0, math.MaxInt64} {
		if got := ArraySlice(NewArray(), offset, nil, true).(Array); len(got.Entries) != 0 || ArrayKeyLast(ArrayAppend(got, nil)) != int64(0) {
			t.Fatal("empty selection metadata")
		}
	}
}

func TestArraySliceKeysCopyAndAppend(t *testing.T) {
	input := NewArray(Entry{Key: int64(math.MaxInt64), Value: "excluded"}, Entry{Key: "8", Value: nil}, Entry{Key: "02", Value: false}, Entry{Key: int64(-5), Value: NewArray(Entry{Value: "nested"})}, Entry{Key: "a\x00\xff", Value: "tail"})
	before := Clone(input)
	for _, preserve := range []bool{false, true} {
		got := ArraySlice(input, int64(1), nil, preserve).(Array)
		keys := []any{int64(0), "02", int64(1), "a\x00\xff"}
		next := int64(2)
		if preserve {
			keys = []any{int64(8), "02", int64(-5), "a\x00\xff"}
			next = 9
		}
		if len(got.Entries) != len(keys) {
			t.Fatal("selection length")
		}
		for i, entry := range got.Entries {
			if entry.Key != keys[i] || !reflect.DeepEqual(entry.Value, input.Entries[i+1].Value) {
				t.Fatalf("entry %d = %#v", i, entry)
			}
		}
		if ArrayKeyLast(ArrayAppend(got, nil)) != next {
			t.Fatal("excluded maximum key leaked into result metadata")
		}
		got.Entries[2].Value.(Array).Entries[0].Value = "changed"
		got.Entries[0].Value = "changed"
		if !reflect.DeepEqual(input, before) {
			t.Fatal("aliased input storage")
		}
	}
	negative := NewArray(Entry{Key: int64(-5), Value: nil}, Entry{Key: int64(-10), Value: nil})
	if ArrayKeyLast(ArrayAppend(ArraySlice(negative, int64(0), nil, true), nil)) != int64(-4) {
		t.Fatal("negative next index")
	}
	if ArrayKeyLast(ArrayAppend(ArraySlice(input, int64(0), int64(1), false), nil)) != int64(1) {
		t.Fatal("maximum key should reindex")
	}
	defer func() {
		if got := recover(); got != "array key overflow unsupported" {
			t.Fatalf("preserved maximum: %v", got)
		}
	}()
	ArrayAppend(ArraySlice(input, int64(0), int64(1), true), nil)
}

func TestArraySliceRejectsInvalidDomains(t *testing.T) {
	for position, invalid := range [][]any{
		{nil, false, int64(0), "", &struct{}{}, (*Array)(nil)},
		{nil, false, "0", NewArray(), int(0), float64(0)},
		{false, "0", NewArray(), int(0), float64(0)},
		{nil, int64(0), "", NewArray()},
	} {
		for _, value := range invalid {
			args := []any{NewArray(), int64(math.MaxInt64), nil, false}
			args[position] = value
			want := []string{"array_slice requires array", "array_slice offset requires int", "array_slice length requires int or null", "array_slice preserve_keys requires bool"}[position]
			func() {
				defer func() {
					if got := recover(); got != want {
						t.Fatalf("position %d: %v", position, got)
					}
				}()
				ArraySlice(args[0], args[1], args[2], args[3])
			}()
		}
	}
}
