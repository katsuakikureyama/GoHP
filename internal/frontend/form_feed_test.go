package frontend

import (
	"strings"
	"testing"
)

func TestFormFeedLiteralIR(t *testing.T) {
	for _, tt := range []struct{ source, want string }{
		{`"\f"`, string([]byte{12})},
		{`'\f'`, `\f`},
		{`"\\f"`, `\f`},
		{`"\\\f"`, string([]byte{'\\', 12})},
		{`"a\f\fb"`, string([]byte{'a', 12, 12, 'b'})},
		{`" \t\f\r\n"`, string([]byte{32, 9, 12, 13, 10})},
		{"\"a" + string([]byte{12}) + "b\"", string([]byte{'a', 12, 'b'})},
	} {
		t.Run(tt.source, func(t *testing.T) {
			p, err := Parse("<?php\n$value = "+tt.source+";\necho $value;", testTarget(t))
			if err != nil {
				t.Fatal(err)
			}
			e := p.Body[0].Expr
			if e.Kind != "string" || e.Text != tt.want || e.Line != 2 || p.Body[1].Line != 3 {
				t.Fatalf("literal %#v, next line %d; want bytes %v at line 2, next 3", e, p.Body[1].Line, []byte(tt.want))
			}
		})
	}
	// Physical LF inside a literal still advances source locations; decoded FF does not.
	p, err := Parse("<?php\n$value = \"a\n\\fb\";\necho $value;", testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	if p.Body[0].Expr.Text != "a\n\fb" || p.Body[1].Line != 4 {
		t.Fatalf("multiline literal: %#v", p.Body)
	}
}

func TestFormFeedEscapeBoundary(t *testing.T) {
	// Other PHP escape forms remain outside this closed subset, not silently decoded.
	for _, literal := range []string{`"\F"`, `"\v"`, `"\e"`, `"\x0c"`, `"\014"`, `"\u{c}"`, `"\$x"`, `"\q"`, `"\f$x"`, `'\F'`} {
		t.Run(literal, func(t *testing.T) {
			_, err := Parse("<?php echo "+literal+";", testTarget(t))
			if err == nil || !strings.Contains(err.Error(), "PHP002") {
				t.Fatalf("unsupported literal %s: %v", literal, err)
			}
		})
	}
	for _, source := range []string{"<?php echo \"\\f", "<?php echo '\\f", "<?php echo \"\\f\\"} {
		if _, err := Parse(source, testTarget(t)); err == nil || !strings.Contains(err.Error(), "PHP001") {
			t.Fatalf("unterminated %q: %v", source, err)
		}
	}
}
