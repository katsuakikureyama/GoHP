package runtime

import "testing"

func TestAddSlashesEscapesBytesInOnePass(t *testing.T) {
	input := string([]byte{'A', '\'', '"', '\\', 0, 0xff})
	want := string([]byte{'A', '\\', '\'', '\\', '"', '\\', '\\', '\\', '0', 0xff})
	if got := AddSlashes(input); got != want {
		t.Fatalf("addslashes bytes = %q, want %q", got, want)
	}
	if got := AddSlashes(want); got != string([]byte{'A', '\\', '\\', '\\', '\'', '\\', '\\', '\\', '"', '\\', '\\', '\\', '\\', '\\', '\\', '0', 0xff}) {
		t.Fatalf("addslashes must not rescan inserted backslashes: %q", got)
	}
	for _, unchanged := range []string{"", "猫🦊\n"} {
		if got := AddSlashes(unchanged); got != unchanged {
			t.Fatalf("unquoted bytes changed: %q", got)
		}
	}
}

func TestAddSlashesRejectsNonStrings(t *testing.T) {
	for _, value := range []any{nil, false, int64(1), NewArray(), 1.5} {
		func() {
			defer func() {
				if got := recover(); got != "addslashes requires a string argument" {
					t.Fatalf("non-string panic = %v", got)
				}
			}()
			AddSlashes(value)
		}()
	}
}
