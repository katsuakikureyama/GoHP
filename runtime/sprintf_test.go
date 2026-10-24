package runtime

import (
	"strings"
	"testing"
)

func TestSprintfZeroPad(t *testing.T) {
	for _, tt := range []struct {
		name   string
		format string
		value  any
		want   string
	}{
		{name: "integer", format: "%05s", value: int64(42), want: "00042"},
		{name: "negative_integer", format: "%05s", value: int64(-42), want: "00-42"},
		{name: "string", format: "%05s", value: "007", want: "00007"},
		{name: "empty_string", format: "%03s", value: "", want: "000"},
		{name: "width_zero", format: "%00s", value: int64(0), want: "0"},
		{name: "no_truncation", format: "%03s", value: "ready", want: "ready"},
		{name: "byte_width", format: "%03s", value: "é", want: "0é"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := SprintfZeroPad(tt.format, tt.value); got != tt.want {
				t.Fatalf("SprintfZeroPad(%q, %#v) = %#v, want %#v", tt.format, tt.value, got, tt.want)
			}
		})
	}

	got := SprintfZeroPad("%04096s", "x").(string)
	if len(got) != 4096 || !strings.HasSuffix(got, "x") || strings.Trim(got[:4095], "0") != "" {
		t.Fatalf("maximum width result is not 4095 leading zero bytes followed by x: len=%d", len(got))
	}
}

func TestSprintfZeroPadRejectsOutsideSubsetAtRuntime(t *testing.T) {
	for _, tt := range []struct {
		name string
		call func()
		want string
	}{
		{name: "format_type", call: func() { SprintfZeroPad(int64(1), "x") }, want: "sprintf format requires string"},
		{name: "directive", call: func() { SprintfZeroPad("%05d", "x") }, want: "sprintf format outside supported"},
		{name: "noncanonical_width", call: func() { SprintfZeroPad("%005s", "x") }, want: "sprintf format outside supported"},
		{name: "width_overflow", call: func() { SprintfZeroPad("%04097s", "x") }, want: "sprintf format outside supported"},
		{name: "value_type", call: func() { SprintfZeroPad("%05s", true) }, want: "sprintf value requires string or integer"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if r == nil || !strings.Contains(r.(string), tt.want) {
					t.Fatalf("SprintfZeroPad did not reject unsupported input: %v", r)
				}
			}()
			tt.call()
		})
	}
}
