// Package runtime implements the explicitly supported PHP value operations.
// It contains no parser, opcodes, evaluator, or PHP interpreter.
package runtime

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type Entry struct{ Key, Value any }
type Array struct {
	Entries     []Entry
	next        int64
	nextBlocked bool
}

func Key(v any) any {
	switch x := v.(type) {
	case int64:
		return x
	case string:
		if n, e := strconv.ParseInt(x, 10, 64); e == nil && strconv.FormatInt(n, 10) == x {
			return n
		}
		return x
	default:
		panic("unsupported array key type")
	}
}
func NewArray(es ...Entry) Array {
	a := Array{}
	for _, e := range es {
		k := e.Key
		if k == nil {
			if a.nextBlocked {
				panic("array key overflow unsupported")
			}
			k = a.next
		}
		k = Key(k)
		if n, ok := k.(int64); ok && n >= a.next {
			if n == math.MaxInt64 {
				a.nextBlocked = true
			} else {
				a.next = n + 1
			}
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
		b := Array{next: a.next, nextBlocked: a.nextBlocked}
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

func Args(a []any, n int) {
	if len(a) != n {
		panic("unsupported argument count")
	}
}
