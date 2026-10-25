package runtime

import "testing"

func TestInArrayStrictMatchesOnlyIdenticalSupportedScalars(t *testing.T) {
	haystack := NewArray(
		Entry{Key: int64(9), Value: "1"},
		Entry{Key: "integer", Value: int64(1)},
		Entry{Key: "false", Value: false},
		Entry{Key: "true", Value: true},
		Entry{Key: "null", Value: nil},
		Entry{Key: "utf8", Value: "東京"},
		Entry{Key: "binary", Value: "A\x00B"},
		Entry{Key: "key-only", Value: "stored-under-a-key"},
		Entry{Key: "ascii", Value: "CaseSensitive"},
	)

	for _, tt := range []struct {
		name   string
		needle any
		want   bool
	}{
		{name: "string", needle: "1", want: true},
		{name: "integer", needle: int64(1), want: true},
		{name: "false", needle: false, want: true},
		{name: "true", needle: true, want: true},
		{name: "null", needle: nil, want: true},
		{name: "utf8", needle: "東京", want: true},
		{name: "binary", needle: "A\x00B", want: true},
		{name: "value_with_ignored_key", needle: "stored-under-a-key", want: true},
		{name: "string_not_integer", needle: "01", want: false},
		{name: "integer_not_string", needle: int64(2), want: false},
		{name: "key_is_not_searched", needle: "key-only", want: false},
		{name: "ascii_case_sensitive", needle: "casesensitive", want: false},
		{name: "trailing_space_is_distinct", needle: "東京 ", want: false},
		{name: "binary_difference", needle: "A\x00C", want: false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := InArrayStrict(tt.needle, haystack).(bool)
			if !ok || got != tt.want {
				t.Fatalf("InArrayStrict(%#v, haystack) = %v (bool=%v), want %v", tt.needle, got, ok, tt.want)
			}
		})
	}

	if InArrayStrict(int64(1), NewArray(Entry{Key: "one", Value: "1"})) != false {
		t.Fatal("integer needle matched equal-looking string")
	}
	if InArrayStrict("1", NewArray(Entry{Key: "one", Value: int64(1)})) != false {
		t.Fatal("string needle matched equal-looking integer")
	}
	if InArrayStrict(false, NewArray(Entry{Key: "zero", Value: int64(0)})) != false {
		t.Fatal("boolean needle matched equal-looking integer")
	}
	if InArrayStrict(nil, NewArray(Entry{Key: "empty", Value: ""})) != false {
		t.Fatal("nil needle matched empty string")
	}
	if got := InArrayStrict("missing", NewArray()); got != false {
		t.Fatalf("empty haystack result = %v, want false", got)
	}
}

func TestInArrayStrictRejectsValuesOutsideItsClosedDomain(t *testing.T) {
	for _, tt := range []struct {
		name string
		call func()
		want string
	}{
		{
			name: "array_needle_even_when_haystack_empty",
			call: func() { InArrayStrict(NewArray(), NewArray()) },
			want: "in_array requires a scalar needle",
		},
		{
			name: "object_needle",
			call: func() { InArrayStrict(&struct{}{}, NewArray()) },
			want: "in_array requires a scalar needle",
		},
		{
			name: "float_needle",
			call: func() { InArrayStrict(1.5, NewArray()) },
			want: "in_array requires a scalar needle",
		},
		{
			name: "non_array_haystack",
			call: func() { InArrayStrict("value", "not an array") },
			want: "in_array requires an array haystack",
		},
		{
			name: "nested_array_value",
			call: func() { InArrayStrict("value", NewArray(Entry{Key: int64(0), Value: NewArray()})) },
			want: "in_array requires scalar haystack values",
		},
		{
			name: "object_value",
			call: func() { InArrayStrict("value", NewArray(Entry{Key: int64(0), Value: &struct{}{}})) },
			want: "in_array requires scalar haystack values",
		},
		{
			name: "late_unsupported_value_after_match",
			call: func() {
				InArrayStrict("match", NewArray(
					Entry{Key: "first", Value: "match"},
					Entry{Key: "later", Value: NewArray()},
				))
			},
			want: "in_array requires scalar haystack values",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if got := recover(); got != tt.want {
					t.Fatalf("panic = %v, want %q", got, tt.want)
				}
			}()
			tt.call()
		})
	}
}
