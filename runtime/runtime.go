// Package runtime implements the explicitly supported PHP value operations.
// It contains no parser, opcodes, evaluator, or PHP interpreter.
package runtime

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/katsuakikureyama/GoHP/internal/arraykey"
)

type Entry struct{ Key, Value any }
type Array struct {
	Entries []Entry
	index   arraykey.Index
}

func Key(v any) any { return arraykey.Normalize(v) }

func NewArray(es ...Entry) Array {
	a := Array{}
	for _, e := range es {
		k := e.Key
		if k == nil {
			k = a.index.Next()
		}
		k = Key(k)
		if n, ok := k.(int64); ok {
			a.index.Observe(n)
		}
		found := false
		for i := range a.Entries {
			if a.Entries[i].Key == k {
				a.Entries[i].Value = Clone(e.Value)
				found = true
				break
			}
		}
		if !found {
			a.Entries = append(a.Entries, Entry{k, Clone(e.Value)})
		}
	}
	return a
}
func Clone(v any) any {
	if a, ok := v.(Array); ok {
		b := a
		b.Entries = nil
		for _, e := range a.Entries {
			b.Entries = append(b.Entries, Entry{e.Key, Clone(e.Value)})
		}
		return b
	}
	return v
}
func Get(v, k any) any {
	a, ok := v.(Array)
	if !ok {
		panic("indexing non-array unsupported")
	}
	k = Key(k)
	for _, e := range a.Entries {
		if e.Key == k {
			return Clone(e.Value)
		}
	}
	panic("undefined array key")
}
func Entries(v any) []Entry {
	a, ok := Clone(v).(Array)
	if !ok {
		panic("foreach requires array")
	}
	return a.Entries
}
func Number(v any) int64 {
	switch x := v.(type) {
	case int64:
		return x
	case bool:
		if x {
			return 1
		}
		return 0
	case nil:
		return 0
	case string:
		n, e := strconv.ParseInt(strings.TrimSpace(x), 10, 64)
		if e == nil {
			return n
		}
	}
	panic("numeric conversion outside supported integer subset")
}
func Add(a, b any) any {
	x, y := Number(a), Number(b)
	if y > 0 && x > math.MaxInt64-y || y < 0 && x < math.MinInt64-y {
		panic("integer overflow to float unsupported")
	}
	return x + y
}
func String(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case bool:
		if x {
			return "1"
		}
		return ""
	case int64:
		return strconv.FormatInt(x, 10)
	}
	panic("string conversion outside supported scalar subset")
}
func Concat(a, b any) any { return String(a) + String(b) }
func Echo(v any)          { fmt.Print(String(v)) }
func Count(v any) any {
	a, ok := v.(Array)
	if !ok {
		panic("count requires array")
	}
	return int64(len(a.Entries))
}

// Strlen implements only strlen(string): byte length, not Unicode characters.
// Do not use String here: PHP's scalar/null/object coercions and diagnostics
// are outside this bounded contract. Unknown values fail explicitly at runtime.
func Strlen(v any) any {
	s, ok := v.(string)
	if !ok {
		panic("strlen requires string")
	}
	return int64(len(s))
}

// ArrayKeys implements only the one-argument form. Existing normalized keys
// become values in a fresh, densely indexed array; values are not traversed.
func ArrayKeys(v any) any {
	a, ok := v.(Array)
	if !ok {
		panic("array_keys requires array")
	}
	keys := Array{Entries: make([]Entry, len(a.Entries))}
	for i, e := range a.Entries {
		keys.Entries[i] = Entry{Key: int64(i), Value: e.Key}
		keys.index.Observe(int64(i))
	}
	return keys
}

// ArrayValues reindexes only the outer array. Clone keeps nested array values
// independent using the existing value-copy model; it does not flatten them.
func ArrayValues(v any) any {
	a, ok := v.(Array)
	if !ok {
		panic("array_values requires array")
	}
	values := Array{Entries: make([]Entry, len(a.Entries))}
	for i, e := range a.Entries {
		values.Entries[i] = Entry{Key: int64(i), Value: Clone(e.Value)}
		values.index.Observe(int64(i))
	}
	return values
}

// ArrayKeyExists checks only the first dimension, including entries with nil
// values. Keys use the same integer/string normalization as array construction;
// PHP's other key coercions are deliberately outside this bounded contract.
func ArrayKeyExists(key, value any) bool {
	switch key.(type) {
	case int64, string:
	default:
		panic("array_key_exists requires integer or string key")
	}
	a, ok := value.(Array)
	if !ok {
		panic("array_key_exists requires array")
	}
	key = Key(key)
	for _, entry := range a.Entries {
		if entry.Key == key {
			return true
		}
	}
	return false
}

// ArrayIsList examines only outer keys in insertion order. Values, including
// nested arrays, do not affect list status and are neither read nor copied.
func ArrayIsList(value any) bool {
	a, ok := value.(Array)
	if !ok {
		panic("array_is_list requires array")
	}
	for i, entry := range a.Entries {
		key, ok := entry.Key.(int64)
		if !ok || key != int64(i) {
			return false
		}
	}
	return true
}

func RTrim(v any, mask string) any {
	s, ok := v.(string)
	if !ok {
		panic("rtrim requires string")
	}
	if strings.Contains(mask, "..") {
		panic("rtrim character ranges unsupported")
	}
	var trim [256]bool
	for i := 0; i < len(mask); i++ {
		trim[mask[i]] = true
	}
	end := len(s)
	for end > 0 && trim[s[end-1]] {
		end--
	}
	return s[:end]
}

// Bool implements truthiness only for the closed scalar/array value domain.
// In particular, PHP's string "0" is false but other nonempty strings are true.
func Bool(v any) bool {
	switch x := v.(type) {
	case nil:
		return false
	case bool:
		return x
	case int64:
		return x != 0
	case string:
		return x != "" && x != "0"
	case Array:
		return len(x.Entries) != 0
	default:
		panic("boolean conversion outside supported scalar/array subset")
	}
}

// Identical intentionally excludes PHP array/object identity semantics.
func Identical(a, b any) bool {
	for _, v := range []any{a, b} {
		switch v.(type) {
		case nil, bool, int64, string:
		default:
			panic("strict equality outside supported scalar subset")
		}
	}
	return a == b
}

func IsBool(v any) bool   { _, ok := v.(bool); return ok }
func IsString(v any) bool { _, ok := v.(string); return ok }

// IsArray inspects the value's type, not its contents or truthiness. Native Go
// slices, maps and pointers are not PHP arrays in this runtime's value model.
func IsArray(v any) bool { _, ok := v.(Array); return ok }

// StrToLower uses PHP 8.3 ASCII byte mapping;
// locale-sensitive case mapping and non-string coercion are not supported.
func StrToLower(v any) any {
	s, ok := v.(string)
	if !ok {
		panic("strtolower requires string")
	}
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + ('a' - 'A')
		}
	}
	return string(b)
}

func Args(a []any, n int) {
	if len(a) != n {
		panic("unsupported argument count")
	}
}
