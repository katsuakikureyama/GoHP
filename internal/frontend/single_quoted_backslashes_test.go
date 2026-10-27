package frontend

import (
	"strings"
	"testing"
)

func TestSingleQuotedBackslashes(t *testing.T) {
	for _, tt := range []struct{ name, literal, want string }{
		{"regex", `'\s+\/(?=x)\w*+'`, `\s+\/(?=x)\w*+`},
		{"unknown", `'\F\q\v\e'`, `\F\q\v\e`},
		{"numeric", `'\x41\101\u{41}'`, `\x41\101\u{41}`},
		{"controls", `'\n\r\t\f'`, `\n\r\t\f`},
		{"dollars", `'\$x ${x}'`, `\$x ${x}`},
		{"double_quote", `'\"'`, `\"`},
		{"quote", `'it\'s'`, `it's`},
		{"backslash", `'a\\'`, `a\`},
		{"odd_backslashes", `'\\\''`, `\'`},
		{"even_backslashes", `'\\\\'`, `\\`},
		{"utf8", `'\猫'`, `\猫`},
		{"nul", "'a\\\x00b'", "a\\\x00b"},
		{"invalid_utf8", "'\\\xff'", "\\\xff"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p, err := Parse("<?php\n$value = "+tt.literal+";\necho $value;", testTarget(t))
			if err != nil {
				t.Fatal(err)
			}
			if e := p.Body[0].Expr; e.Kind != "string" || e.Text != tt.want || e.Line != 2 || p.Body[1].Line != 3 {
				t.Fatalf("literal %+v, next line %d; want bytes %v at line 2, next 3", e, p.Body[1].Line, []byte(tt.want))
			}
		})
	}
	for _, newline := range []string{"\n", "\r\n"} {
		want := "a\\" + newline + "b"
		p, err := Parse("<?php\n$value = '"+want+"';\necho $value;", testTarget(t))
		if err != nil {
			t.Fatal(err)
		}
		if p.Body[0].Expr.Text != want || p.Body[1].Line != 4 {
			t.Fatalf("physical newline %q: %+v", newline, p.Body)
		}
	}
}

func TestSingleQuotedUnterminated(t *testing.T) {
	for _, literal := range []string{`'\s`, `'abc\`, `'abc\'`, "'a\\\nb"} {
		_, err := Parse("<?php\necho "+literal, testTarget(t))
		if err == nil || !strings.Contains(err.Error(), "PHP001 line 2: unterminated string") {
			t.Fatalf("unterminated %q: %v", literal, err)
		}
	}
}
