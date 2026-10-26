package runtime

import (
	"fmt"
	"strings"
	"testing"
)

func TestStrtokByteScanning(t *testing.T) {
	for _, tt := range []struct {
		name, text, delimiters string
		want                   []any
	}{
		{"empty", "", ",", []any{false, false}},
		{"empty_delimiters", " a,b\x00c ", "", []any{" a,b\x00c ", false, false}},
		{"repeated", ",;a,,b;", ",;", []any{"a", "b", false}},
		{"all_delimiters", ",,,", ",", []any{false}},
		{"zero", "0,x", ",", []any{"0", "x", false}},
		{"nul_delimiter", "a\x00b", "\x00", []any{"a", "b", false}},
		{"nul_data", "a\x00b,c", ",", []any{"a\x00b", "c", false}},
		{"byte_not_rune", "éx", "\xc3", []any{"\xa9x", false}},
		{"invalid_utf8", "\xffa\xfeb", "\xff\xfe", []any{"a", "b", false}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			s := &Execution{}
			for i, want := range tt.want {
				var got any
				if i == 0 {
					got = s.Strtok(tt.text, tt.delimiters)
				} else {
					got = s.Strtok(tt.delimiters)
				}
				if got != want {
					t.Fatalf("token %d: %#v, want %#v", i, got, want)
				}
			}
		})
	}
}

func TestStrtokExecutionIsolationAndReset(t *testing.T) {
	a, b := &Execution{}, &Execution{}
	check := func(got, want any) {
		t.Helper()
		if got != want {
			t.Fatalf("got %#v want %#v", got, want)
		}
	}
	check(a.Strtok("a,b;c", ","), "a")
	check(b.Strtok("x/y", "/"), "x")
	check(a.Strtok(";"), "b") // Changing separator preserves current cursor.
	check(b.Strtok("/"), "y")
	check(a.Strtok(";"), "c")
	check(a.Strtok("new:scan", ":"), "new")
	check(a.Strtok(":"), "scan")
	check(b.Strtok("/"), false)
	// One boundary byte is already consumed; newly non-delimiter bytes remain.
	check(a.Strtok("a,,b", ","), "a")
	check(a.Strtok(""), ",b")
}

func TestStrtokRejectsInvalidStateAndDomains(t *testing.T) {
	panics := func(want string, call func()) {
		t.Helper()
		func() {
			defer func() {
				if got := recover(); got == nil || !strings.Contains(fmt.Sprint(got), want) {
					t.Fatalf("panic %v, want %s", got, want)
				}
			}()
			call()
		}()
	}
	panics("execution context", func() { var s *Execution; s.Strtok("a", ",") })
	panics("initialized state", func() { (&Execution{}).Strtok(",") })
	s := &Execution{}
	if s.Strtok("a,b", ",") != "a" {
		t.Fatal("initial token")
	}
	for _, args := range [][]any{nil, {"a", ",", ";"}, {nil}, {true}, {int64(1)}, {NewArray()}, {"new", nil}, {false, ","}} {
		panics("strtok requires", func() { s.Strtok(args...) })
	}
	if s.Strtok(",") != "b" {
		t.Fatal("invalid arguments changed scan")
	}
	if s.Strtok(" ,,, ", " ,") != false {
		t.Fatal("all separators")
	}
	panics("initialized state", func() { s.Strtok(",") })
	if s.Strtok("reset", ",") != "reset" {
		t.Fatal("reset after exhausted separators")
	}
}

func TestContextualCallableUsesCallerExecution(t *testing.T) {
	arrow := NewCallableWithExecution(1, func(s *Execution, args ...any) any { return s.Strtok(args[0]) })
	a, b := &Execution{}, &Execution{}
	a.Strtok("a,next-a", ",")
	b.Strtok("b,next-b", ",")
	if !IsObject(arrow) || Clone(arrow) != arrow {
		t.Fatal("arrow identity")
	}
	if InvokeWithExecution(a, arrow, ",") != "next-a" || InvokeWithExecution(b, arrow, ",") != "next-b" {
		t.Fatal("caller context not used")
	}
	plain := NewCallable(1, func(args ...any) any { return args[0] })
	if InvokeWithExecution(a, plain, "legacy") != "legacy" {
		t.Fatal("legacy callable")
	}
	defer func() {
		if got := recover(); got == nil || !strings.Contains(fmt.Sprint(got), "initialized state") {
			t.Fatalf("standalone invocation must not inherit state: %v", got)
		}
	}()
	Invoke(arrow, ",")
}
