package frontend

import (
	"fmt"
	"testing"
)

func TestPHPAttributesRejected(t *testing.T) {
	cases := []struct {
		name string
		src  string
		line int
	}{
		{
			name: "class",
			src:  "<?php\n#[Marker]\nclass Plain {}\n",
			line: 2,
		},
		{
			name: "function",
			src:  "<?php\n\n#[Marker]\nfunction ready() { return 1; }\n",
			line: 3,
		},
		{
			name: "method",
			src:  "<?php\nclass Plain {\n#[Marker]\npublic function ready() { return 1; }\n}\n",
			line: 3,
		},
		{
			name: "property",
			src:  "<?php\nclass Plain {\n#[Marker]\npublic $value;\n}\n",
			line: 3,
		},
		{
			name: "parameter",
			src:  "<?php\nfunction ready(\n#[Marker]\n$value\n) { return $value; }\n",
			line: 3,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.src, testTarget(t))
			want := fmt.Sprintf("PHP002 line %d: PHP attributes unsupported", tt.line)
			if err == nil || err.Error() != want {
				t.Fatalf("attribute must fail closed: got %v, want %q", err, want)
			}
		})
	}
}

func TestPHPAttributesModuleDiagnostics(t *testing.T) {
	_, err := ParseStaticModule(
		"src/attributed.php",
		"<?php\n#[Marker]\nfunction ready() { return 1; }\n",
		testTarget(t),
		Options{},
		false,
	)
	const want = "src/attributed.php: PHP002 line 2: PHP attributes unsupported"
	if err == nil || err.Error() != want {
		t.Fatalf("file-aware attribute diagnostic: got %v, want %q", err, want)
	}
}

func TestHashCommentsAndAttributeLiterals(t *testing.T) {
	const src = "<?php\n" +
		"# ordinary hash comment\n" +
		"# [space after hash remains a comment]\n" +
		"// #[notattribute]\n" +
		"/* #[notattribute] */\n" +
		"$single = '#[literal]';\n" +
		"$double = \"#[literal]\";\n" +
		"echo $single;\n" +
		"echo $double;\n"

	p, err := Parse(src, testTarget(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Body) != 4 {
		t.Fatalf("unexpected statements after comments: %#v", p.Body)
	}
	for _, index := range []int{0, 1} {
		e := p.Body[index].Expr
		if e.Kind != "string" || e.Text != "#[literal]" || e.Line != 6+index {
			t.Fatalf("literal %d: %#v", index, e)
		}
	}
	if p.Body[2].Line != 8 || p.Body[3].Line != 9 {
		t.Fatalf("next statement locations: %#v", p.Body)
	}
}
