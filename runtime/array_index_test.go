package runtime

import (
	"math"
	"reflect"
	"testing"
)

func TestArrayPHP83AutomaticKeys(t *testing.T) {
	for _, tt := range []struct {
		name string
		keys []any // nil means an implicit key, as in an array literal.
		want []any
	}{
		{"first_implicit", []any{nil}, []any{int64(0)}},
		{"negative_integer", []any{int64(-5), nil, nil}, []any{int64(-5), int64(-4), int64(-3)}},
		{"numeric_string", []any{"-5", nil}, []any{int64(-5), int64(-4)}},
		{"string_prefix", []any{"label", "-5", nil}, []any{"label", int64(-5), int64(-4)}},
		{"lower_and_duplicate", []any{"-5", "-9", "-5", nil}, []any{int64(-5), int64(-9), int64(-4)}},
		{"higher_negative", []any{"-5", "-2", nil, nil}, []any{int64(-5), int64(-2), int64(-1), int64(0)}},
		{"no_rewind", []any{int64(2), "-5", nil}, []any{int64(2), int64(-5), int64(3)}},
		{"noncanonical", []any{"-05", nil}, []any{"-05", int64(0)}},
		{"minimum", []any{"-9223372036854775808", nil}, []any{int64(math.MinInt64), int64(math.MinInt64 + 1)}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			entries := make([]Entry, len(tt.keys))
			for i, key := range tt.keys {
				entries[i] = Entry{Key: key, Value: int64(i)}
			}
			got := NewArray(entries...)
			keys := make([]any, len(got.Entries))
			for i, entry := range got.Entries {
				keys[i] = entry.Key
			}
			if !reflect.DeepEqual(keys, tt.want) {
				t.Fatalf("keys = %#v, want %#v", keys, tt.want)
			}
		})
	}
}

func TestArrayIndexStateAfterValueOperations(t *testing.T) {
	identity := func(v any) any { return v }
	for _, tt := range []struct {
		name string
		base Array
		key  any
		next int64
	}{
		{"zero_value", Array{}, "-5", -4},
		{"strings_only", NewArray(Entry{Key: "label"}), "-5", -4},
		{"clone_strings", Clone(NewArray(Entry{Key: "label"})).(Array), "-5", -4},
		{"map_strings", ArrayMap(identity, NewArray(Entry{Key: "label"})).(Array), "-5", -4},
		{"clone_negative", Clone(NewArray(Entry{Key: "-5"})).(Array), "-9", -4},
		{"map_negative", ArrayMap(identity, NewArray(Entry{Key: "-5"})).(Array), "-2", -1},
		{"replace_negative", NewArray(Entry{Key: "-5"}), "-5", -4},
		{"minimum", NewArray(Entry{Key: "label"}), "-9223372036854775808", math.MinInt64 + 1},
		{"keys_reindexed", ArrayKeys(NewArray(Entry{Key: "-5"})).(Array), "-5", 1},
		{"values_reindexed", ArrayValues(NewArray(Entry{Key: "-5"})).(Array), "-5", 1},
		{"empty_keys", ArrayKeys(NewArray()).(Array), "-5", -4},
		{"empty_values", ArrayValues(NewArray()).(Array), "-5", -4},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := ArraySet(tt.base, tt.key, "written").(Array)
			if got.index.Next() != tt.next {
				t.Fatalf("index state = %#v, want next %d", got, tt.next)
			}
			if Get(got, tt.key) != "written" {
				t.Fatal("explicit value lost")
			}
		})
	}
	// Neither a lower insertion nor replacement may clear the overflow guard.
	maximum := NewArray(Entry{Key: int64(math.MaxInt64)})
	for _, copied := range []Array{Clone(maximum).(Array), ArrayMap(identity, maximum).(Array)} {
		for _, key := range []any{"-5", int64(math.MaxInt64)} {
			got := ArraySet(copied, key, "written").(Array)
			assertArrayAppendPanic(t, func() { ArrayAppend(got, nil) }, "array key overflow unsupported")
		}
	}
}
