package runtime

import (
	"fmt"
	"strings"
	"testing"
)

func TestJSONAssociativeValuesOrderAndCopies(t *testing.T) {
	execution := &Execution{}
	value := execution.JSONDecodeAssociative(`{"0":{"numbers":[1,2]},"dup":"old","dup":"new","tail":null}`).(Array)
	if got, want := len(value.Entries), 3; got != want {
		t.Fatalf("entries = %d; want %d", got, want)
	}
	if got, want := value.Entries[0].Key, int64(0); got != want {
		t.Fatalf("first key = %#v; want %#v", got, want)
	}
	if got, want := value.Entries[1].Key, "dup"; got != want {
		t.Fatalf("duplicate key position = %#v; want %#v", got, want)
	}
	if got, want := value.Entries[1].Value, "new"; got != want {
		t.Fatalf("duplicate value = %#v; want %#v", got, want)
	}
	if value.Entries[2].Value != nil {
		t.Fatalf("null = %#v; want nil", value.Entries[2].Value)
	}

	nested := value.Entries[0].Value.(Array)
	numbers := nested.Entries[0].Value.(Array)
	if got, want := numbers.Entries[1].Value, int64(2); got != want {
		t.Fatalf("nested integer = %#v; want %#v", got, want)
	}
	copy := Clone(value).(Array)
	copy.Entries[0].Value.(Array).Entries[0].Value.(Array).Entries[0].Value = int64(99)
	if got, want := value.Entries[0].Value.(Array).Entries[0].Value.(Array).Entries[0].Value, int64(1); got != want {
		t.Fatalf("Clone aliased a nested JSON array: got %#v; want %#v", got, want)
	}
	if got := execution.JSONLastError(); got != jsonErrorNone {
		t.Fatalf("last error = %d; want no error", got)
	}
}

func TestJSONAssociativeUnicodeAndParseErrors(t *testing.T) {
	execution := &Execution{}
	if got, want := execution.JSONDecodeAssociative(`"猫\uD83D\uDE3A"`), "猫😺"; got != want {
		t.Fatalf("unicode value = %#v; want %#v", got, want)
	}

	for _, tc := range []struct {
		name, source string
		error        int64
		message      string
	}{
		{"syntax", `{"broken":}`, jsonErrorSyntax, "Syntax error"},
		{"state_mismatch", `[}`, jsonErrorStateMismatch, "State mismatch (invalid or malformed JSON)"},
		{"trailing_comma_wrong_close", `[1,}`, jsonErrorSyntax, "Syntax error"},
		{"missing_colon_wrong_close", `{"key"}`, jsonErrorSyntax, "Syntax error"},
		{"raw_control", "\"line\x01break\"", jsonErrorControl, "Control character error, possibly incorrectly encoded"},
		{"unterminated_string", "\"open", jsonErrorControl, "Control character error, possibly incorrectly encoded"},
		{"invalid_utf8", string([]byte{'"', 0xff, '"'}), jsonErrorUTF8, "Malformed UTF-8 characters, possibly incorrectly encoded"},
		{"unpaired_high_surrogate", `"\uD800"`, jsonErrorUTF16, "Single unpaired UTF-16 surrogate in unicode escape"},
		{"unpaired_low_surrogate", `"\uDC00"`, jsonErrorUTF16, "Single unpaired UTF-16 surrogate in unicode escape"},
		{"high_before_invalid_low", `"\uD800\uZZZZ"`, jsonErrorUTF16, "Single unpaired UTF-16 surrogate in unicode escape"},
		{"trailing_bad_string", "null \"open", jsonErrorControl, "Control character error, possibly incorrectly encoded"},
		{"bad_unicode_escape", `"\u12x4"`, jsonErrorSyntax, "Syntax error"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := execution.JSONDecodeAssociative(tc.source); got != nil {
				t.Fatalf("value = %#v; want nil", got)
			}
			if got := execution.JSONLastError(); got != tc.error {
				t.Fatalf("last error = %d; want %d", got, tc.error)
			}
			if got, want := execution.JSONLastErrorMessage(), tc.message; got != want {
				t.Fatalf("message = %q; want %q", got, want)
			}
		})
	}
}

func TestJSONAssociativeDepthAndNumericProfile(t *testing.T) {
	execution := &Execution{}
	withinLimit := strings.Repeat("[", 511) + "0" + strings.Repeat("]", 511)
	if got := execution.JSONDecodeAssociative(withinLimit); got == nil {
		t.Fatal("511 nested arrays decoded to nil")
	}
	if got := execution.JSONLastError(); got != jsonErrorNone {
		t.Fatalf("511 nested arrays error = %d; want no error", got)
	}
	overLimit := strings.Repeat("[", 512) + "0" + strings.Repeat("]", 512)
	if got := execution.JSONDecodeAssociative(overLimit); got != nil {
		t.Fatalf("512 nested arrays = %#v; want nil", got)
	}
	if got := execution.JSONLastError(); got != jsonErrorDepth {
		t.Fatalf("512 nested arrays error = %d; want depth", got)
	}

	if got, want := execution.JSONDecodeAssociative(`[-9223372036854775808,9223372036854775807]`).(Array), 2; len(got.Entries) != want {
		t.Fatalf("integer boundary entries = %d; want %d", len(got.Entries), want)
	}
	for _, source := range []string{`1.5`, `1e3`, `9223372036854775808`} {
		execution.JSONDecodeAssociative(`{"bad":}`)
		if got := execution.JSONLastError(); got != jsonErrorSyntax {
			t.Fatalf("setup error = %d; want syntax", got)
		}
		jsonMustPanic(t, "number profile "+source, "json_decode number outside int64 profile", func() {
			execution.JSONDecodeAssociative(source)
		})
		if got := execution.JSONLastError(); got != jsonErrorSyntax {
			t.Fatalf("unsupported number changed state to %d; want syntax", got)
		}
	}
	if got := execution.JSONDecodeAssociative(`{"number":1.5,}`); got != nil {
		t.Fatalf("invalid number document = %#v; want nil", got)
	}
	if got := execution.JSONLastError(); got != jsonErrorSyntax {
		t.Fatalf("invalid number document error = %d; want syntax", got)
	}
}

func TestJSONAssociativeExecutionIsolationAndGuards(t *testing.T) {
	first := &Execution{}
	second := &Execution{}
	first.JSONDecodeAssociative(`{`)
	if got := first.JSONLastError(); got != jsonErrorSyntax {
		t.Fatalf("first error = %d; want syntax", got)
	}
	if got := second.JSONLastError(); got != jsonErrorNone {
		t.Fatalf("second fresh error = %d; want no error", got)
	}
	if got, want := first.JSONDecodeAssociative(`true`), true; got != want {
		t.Fatalf("valid reset value = %#v; want %#v", got, want)
	}
	if got := first.JSONLastError(); got != jsonErrorNone {
		t.Fatalf("valid decode did not reset error: %d", got)
	}

	second.JSONDecodeAssociative(`{`)
	jsonMustPanic(t, "non-string source", "json_decode requires actual string", func() {
		second.JSONDecodeAssociative(int64(1))
	})
	if got := second.JSONLastError(); got != jsonErrorSyntax {
		t.Fatalf("non-string source changed state to %d; want syntax", got)
	}
	jsonMustPanic(t, "nil decode", "execution context is required", func() {
		(*Execution)(nil).JSONDecodeAssociative("null")
	})
	jsonMustPanic(t, "nil last error", "execution context is required", func() {
		(*Execution)(nil).JSONLastError()
	})
	jsonMustPanic(t, "nil last error message", "execution context is required", func() {
		(*Execution)(nil).JSONLastErrorMessage()
	})
}

func jsonMustPanic(t *testing.T, name, want string, call func()) {
	t.Helper()
	defer func() {
		if got := recover(); got == nil || !strings.Contains(fmt.Sprint(got), want) {
			t.Fatalf("%s panic = %v; want containing %q", name, got, want)
		}
	}()
	call()
}
