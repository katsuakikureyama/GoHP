package runtime

import (
	"reflect"
	"testing"
)

func TestNullDefaultArgs(t *testing.T) {
	provided := []any{"first"}
	padded := NullDefaultArgs(provided, 1, 3)
	if want := []any{"first", nil, nil}; !reflect.DeepEqual(padded, want) {
		t.Fatalf("padded arguments = %#v, want %#v", padded, want)
	}
	if want := []any{"first"}; !reflect.DeepEqual(provided, want) {
		t.Fatalf("NullDefaultArgs changed caller arguments: %#v, want %#v", provided, want)
	}
	padded[0] = "changed"
	if provided[0] != "first" {
		t.Fatalf("padded arguments alias caller storage: %#v", provided)
	}

	full := []any{"first", nil, "last"}
	if got := NullDefaultArgs(full, 1, 3); !reflect.DeepEqual(got, full) {
		t.Fatalf("full arguments = %#v, want %#v", got, full)
	}

	for _, tt := range []struct {
		name            string
		args            []any
		required, total int
	}{
		{"negative_required", nil, -1, 0},
		{"total_before_required", nil, 2, 1},
		{"too_few", nil, 1, 2},
		{"too_many", []any{1, 2, 3}, 1, 2},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if got := recover(); got != "unsupported argument count" {
					t.Fatalf("panic = %v, want %q", got, "unsupported argument count")
				}
			}()
			NullDefaultArgs(tt.args, tt.required, tt.total)
		})
	}
}
