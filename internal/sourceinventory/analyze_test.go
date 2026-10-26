package sourceinventory

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
)

type tapePart struct {
	kind string
	text string
}

func TestAnalyzeLexicalAccounting(t *testing.T) {
	source, tokens := tape(
		tapePart{"T_FUNCTION", "function"}, tapePart{"T_WHITESPACE", " "}, punctuation('&'), tapePart{"T_STRING", "foo"}, punctuation('('), punctuation(')'), tapePart{"T_WHITESPACE", " "}, punctuation('{'), punctuation('}'), tapePart{"T_WHITESPACE", "\n"},
		tapePart{"T_COMMENT", "// comment("}, tapePart{"T_WHITESPACE", "\n"},
		tapePart{"T_STRING", "foo"}, punctuation('('), punctuation(')'), punctuation(';'), tapePart{"T_WHITESPACE", " "},
		tapePart{"T_NEW", "new"}, tapePart{"T_WHITESPACE", " "}, tapePart{"T_STRING", "Bar"}, punctuation('('), punctuation(')'), punctuation(';'), tapePart{"T_WHITESPACE", " "},
		tapePart{"T_VARIABLE", "$obj"}, tapePart{"T_OBJECT_OPERATOR", "->"}, tapePart{"T_STRING", "method"}, punctuation('('), punctuation(')'), punctuation(';'), tapePart{"T_WHITESPACE", " "},
		tapePart{"T_VARIABLE", "$call"}, punctuation('('), punctuation(')'), punctuation(';'), tapePart{"T_WHITESPACE", " "},
		tapePart{"T_IF", "if"}, tapePart{"T_WHITESPACE", " "}, punctuation('('), tapePart{"T_VARIABLE", "$x"}, punctuation(')'), tapePart{"T_WHITESPACE", " "}, punctuation('{'), punctuation('}'),
	)

	report, err := Analyze("fixture.php", source, tokens)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if report.Bytes != len(source) || report.Tokens != len(tokens) {
		t.Fatalf("accounting = %d bytes / %d tokens, want %d / %d", report.Bytes, report.Tokens, len(source), len(tokens))
	}
	if report.GrammarStatus != "not-validated" || report.UnparsedBytes != len(source) {
		t.Fatalf("grammar accounting = %q / %d, want not-validated / %d", report.GrammarStatus, report.UnparsedBytes, len(source))
	}
	if report.SemanticUnits != nil || report.ResolvedCalls != nil {
		t.Fatalf("semantic fields = %v / %v, want nil", report.SemanticUnits, report.ResolvedCalls)
	}
	if len(report.TokenKinds) == 0 || report.TokenKinds[0].Kind >= report.TokenKinds[len(report.TokenKinds)-1].Kind {
		t.Fatalf("token kinds are not sorted: %#v", report.TokenKinds)
	}
	if got := frequency(report.TokenKinds, "T_COMMENT"); got.Count != 1 || got.Bytes != len("// comment(") {
		t.Fatalf("comment frequency = %#v", got)
	}
	if len(report.Declarations) != 1 {
		t.Fatalf("declarations = %#v, want one", report.Declarations)
	}
	declaration := report.Declarations[0]
	if declaration.Kind != "named-function-marker" || declaration.Name != "foo" || declaration.NameBase64 != "Zm9v" || declaration.StartByte != 0 || declaration.EndByte != len("function") || declaration.Line != 1 {
		t.Fatalf("declaration = %#v", declaration)
	}

	gotKinds := markerKinds(report.Parentheses)
	wantKinds := []string{"function-parameters", "name-parenthesis", "constructor-name", "member-name", "variable-parenthesis", "language-form"}
	if strings.Join(gotKinds, ",") != strings.Join(wantKinds, ",") {
		t.Fatalf("parenthesis kinds = %v, want %v", gotKinds, wantKinds)
	}
	if report.Parentheses[1].Line != 3 {
		t.Fatalf("foo() line = %d, want 3", report.Parentheses[1].Line)
	}
	if report.UnclassifiedParentheses != 0 {
		t.Fatalf("unclassified parentheses = %d, want 0", report.UnclassifiedParentheses)
	}
	if len(report.CandidateNames) != 1 || report.CandidateNames[0] != (NameFrequency{Name: "foo", NameBase64: "Zm9v", Count: 1}) {
		t.Fatalf("candidate names = %#v", report.CandidateNames)
	}

	arrowSource, arrowTokens := tape(tapePart{"T_FN", "fn"}, tapePart{"T_WHITESPACE", " "}, punctuation('&'), punctuation('('), punctuation(')'))
	arrow, err := Analyze("arrow.php", arrowSource, arrowTokens)
	if err != nil {
		t.Fatalf("Analyze(arrow) error = %v", err)
	}
	if arrow.Declarations[0].Kind != "arrow-marker" || arrow.Parentheses[0].Kind != "arrow-parameters" {
		t.Fatalf("arrow markers = %#v / %#v", arrow.Declarations, arrow.Parentheses)
	}
	for _, ampersand := range []string{"T_AMPERSAND_FOLLOWED_BY_VAR_OR_VARARG", "T_AMPERSAND_NOT_FOLLOWED_BY_VAR_OR_VARARG"} {
		t.Run(ampersand, func(t *testing.T) {
			functionSource, functionTokens := tape(tapePart{"T_FUNCTION", "function"}, tapePart{"T_WHITESPACE", " "}, tapePart{ampersand, "&"}, tapePart{"T_STRING", "named"}, punctuation('('), punctuation(')'))
			function, err := Analyze("ampersand.php", functionSource, functionTokens)
			if err != nil {
				t.Fatalf("Analyze() error = %v", err)
			}
			if function.Declarations[0].Kind != "named-function-marker" || function.Parentheses[0].Kind != "function-parameters" {
				t.Fatalf("function markers = %#v / %#v", function.Declarations, function.Parentheses)
			}

			arrowSource, arrowTokens := tape(tapePart{"T_FN", "fn"}, tapePart{"T_WHITESPACE", " "}, tapePart{ampersand, "&"}, punctuation('('), punctuation(')'))
			arrow, err := Analyze("arrow-ampersand.php", arrowSource, arrowTokens)
			if err != nil {
				t.Fatalf("Analyze() error = %v", err)
			}
			if arrow.Parentheses[0].Kind != "arrow-parameters" {
				t.Fatalf("arrow parentheses = %#v", arrow.Parentheses)
			}
		})
	}

	nonUTF8Source := []byte{0xff, '('}
	nonUTF8, err := Analyze("bytes.php", nonUTF8Source, []Token{{Kind: "T_STRING", Bytes: 1}, {Kind: "CHAR_28", Bytes: 1}})
	if err != nil {
		t.Fatalf("Analyze(non-UTF-8) error = %v", err)
	}
	if got := nonUTF8.CandidateNames; len(got) != 1 || got[0].Name != "" || got[0].NameBase64 != base64.StdEncoding.EncodeToString([]byte{0xff}) {
		t.Fatalf("non-UTF-8 name = %#v", got)
	}

	unknownSource, unknownTokens := tape(punctuation('='), punctuation('('))
	unknown, err := Analyze("unknown.php", unknownSource, unknownTokens)
	if err != nil {
		t.Fatalf("Analyze(unknown) error = %v", err)
	}
	if unknown.Parentheses[0].Kind != "unclassified" || unknown.UnclassifiedParentheses != 1 {
		t.Fatalf("unknown parenthesis = %#v / %d", unknown.Parentheses, unknown.UnclassifiedParentheses)
	}
}

func TestAnalyzeRejectsInvalidTokens(t *testing.T) {
	tooMany := make([]Token, maxTokens+1)
	cases := []struct {
		name   string
		source []byte
		tokens []Token
	}{
		{"empty source token", nil, []Token{{Kind: "T_STRING", Bytes: 1}}},
		{"zero bytes", []byte("a"), []Token{{Kind: "T_STRING", Bytes: 0}}},
		{"under coverage", []byte("ab"), []Token{{Kind: "T_STRING", Bytes: 1}}},
		{"over coverage", []byte("a"), []Token{{Kind: "T_STRING", Bytes: 2}}},
		{"lowercase native kind", []byte("a"), []Token{{Kind: "T_string", Bytes: 1}}},
		{"long kind", []byte("a"), []Token{{Kind: "T_" + strings.Repeat("A", 79), Bytes: 1}}},
		{"punctuation kind mismatch", []byte("("), []Token{{Kind: "CHAR_29", Bytes: 1}}},
		{"punctuation width", []byte("("), []Token{{Kind: "CHAR_28", Bytes: 2}}},
		{"too many", nil, tooMany},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if _, err := Analyze("invalid.php", test.source, test.tokens); err == nil {
				t.Fatal("Analyze() error = nil")
			}
		})
	}

	report, err := Analyze("empty.php", nil, nil)
	if err != nil {
		t.Fatalf("Analyze(empty) error = %v", err)
	}
	if report.Bytes != 0 || report.Tokens != 0 || len(report.TokenKinds) != 0 || report.TokenKinds == nil || report.Declarations == nil || report.Parentheses == nil || report.CandidateNames == nil {
		t.Fatalf("empty report = %#v", report)
	}
}

func tape(parts ...tapePart) ([]byte, []Token) {
	var source []byte
	tokens := make([]Token, 0, len(parts))
	for _, part := range parts {
		source = append(source, part.text...)
		tokens = append(tokens, Token{Kind: part.kind, Bytes: len(part.text)})
	}
	return source, tokens
}

func punctuation(character byte) tapePart {
	return tapePart{kind: fmt.Sprintf("CHAR_%02X", character), text: string(character)}
}

func frequency(frequencies []Frequency, kind string) Frequency {
	for _, frequency := range frequencies {
		if frequency.Kind == kind {
			return frequency
		}
	}
	return Frequency{}
}

func markerKinds(markers []Marker) []string {
	kinds := make([]string, len(markers))
	for index, marker := range markers {
		kinds[index] = marker.Kind
	}
	return kinds
}
