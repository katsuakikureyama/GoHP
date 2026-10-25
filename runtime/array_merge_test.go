package runtime

import (
	"math"
	"reflect"
	"testing"
)

func TestArrayMergeKeysValuesAndIsolation(t *testing.T) {
	nested := NewArray(Entry{Key: "child", Value: "before"})
	left := NewArray(Entry{Key: "same", Value: nested}, Entry{Key: int64(-5), Value: false}, Entry{Key: "02", Value: "leading"}, Entry{Key: "2", Value: "integer"})
	right := NewArray(Entry{Key: "same", Value: nil}, Entry{Key: int64(-5), Value: "duplicate integer"}, Entry{Key: "a\x00\xff", Value: nested})
	last := NewArray(Entry{Key: "same", Value: nested}, Entry{Key: "02", Value: false})
	before := []any{Clone(left), Clone(right), Clone(last)}
	result := ArrayMerge(left, right, last).(Array)
	want := []Entry{{"same", nested}, {int64(0), false}, {"02", false}, {int64(1), "integer"}, {int64(2), "duplicate integer"}, {"a\x00\xff", nested}}
	if !reflect.DeepEqual(result.Entries, want) {
		t.Fatalf("merge = %#v, want %#v", result.Entries, want)
	}
	result.Entries[0].Value.(Array).Entries[0].Value = "changed"
	result.Entries[1].Value = "changed"
	if !reflect.DeepEqual([]any{left, right, last}, before) {
		t.Fatal("merge aliases input or nested storage")
	}
	if got := result.Entries[5].Value.(Array).Entries[0].Value; got != "before" {
		t.Fatalf("distinct output entries share nested array: %v", got)
	}
	// A later nested value replaces the entire earlier one, not a recursive merge.
	replaced := ArrayMerge(NewArray(Entry{Key: "x", Value: NewArray(Entry{Key: "old", Value: nil})}), NewArray(Entry{Key: "x", Value: NewArray(Entry{Key: "new", Value: false})})).(Array)
	if got := Get(replaced, "x").(Array); len(got.Entries) != 1 || got.Entries[0].Key != "new" {
		t.Fatal("merge recursively combined nested arrays")
	}
}

func TestArrayMergeEmptySingleAndAppend(t *testing.T) {
	for _, args := range [][]any{nil, {NewArray()}, {NewArray(), NewArray()}} {
		result := ArrayMerge(args...).(Array)
		if len(result.Entries) != 0 || ArrayKeyLast(ArrayAppend(result, nil)) != int64(0) {
			t.Fatalf("empty merge: %#v", result)
		}
	}
	input := NewArray(Entry{Key: int64(math.MaxInt64), Value: nil}, Entry{Key: "s", Value: false}, Entry{Key: int64(math.MinInt64), Value: "min"})
	before := Clone(input)
	result := ArrayMerge(input).(Array)
	if got := ArrayKeys(result); !reflect.DeepEqual(got, NewArray(Entry{Value: int64(0)}, Entry{Value: "s"}, Entry{Value: int64(1)})) {
		t.Fatalf("single array was not renumbered: %#v", got)
	}
	if got := ArrayKeyLast(ArrayAppend(result, nil)); got != int64(2) {
		t.Fatalf("append = %v", got)
	}
	result.Entries[0].Value = "changed"
	if !reflect.DeepEqual(input, before) {
		t.Fatal("single argument aliases input or changes metadata")
	}
	stringsOnly := ArrayMerge(NewArray(Entry{Key: "a", Value: nil}), NewArray(Entry{Key: "a", Value: false}))
	if got := ArrayKeyLast(ArrayAppend(stringsOnly, nil)); got != int64(0) {
		t.Fatalf("string keys affected next index: %v", got)
	}
}

func TestArrayMergeRejectsNonArrays(t *testing.T) {
	for _, value := range []any{nil, false, int64(0), "", &struct{}{}, (*Array)(nil), []any(nil)} {
		for position := 0; position < 3; position++ {
			args := []any{NewArray(), NewArray(), NewArray()}
			args[position] = value
			func() {
				defer func() {
					if got := recover(); got != "array_merge requires arrays" {
						t.Fatalf("argument %d guard = %v", position, got)
					}
				}()
				ArrayMerge(args...)
			}()
		}
	}
}
