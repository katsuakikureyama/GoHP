package sourceinventory

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

const maxTokens = 250_000

type tokenPosition struct {
	kind       string
	start, end int
	line       int
}

// Analyze turns a trusted PHP token_get_all byte tape into lexical facts. It
// deliberately does not parse PHP or resolve names and calls.
func Analyze(path string, source []byte, tokens []Token) (FileInventory, error) {
	if len(tokens) > maxTokens {
		return FileInventory{}, fmt.Errorf("token count %d exceeds limit %d", len(tokens), maxTokens)
	}

	positions, err := validateTape(source, tokens)
	if err != nil {
		return FileInventory{}, err
	}

	hash := sha256.Sum256(source)
	report := FileInventory{
		Path:               path,
		SHA256:             fmt.Sprintf("%x", hash),
		Bytes:              len(source),
		Tokens:             len(tokens),
		TokenKinds:         make([]Frequency, 0),
		Declarations:       make([]Marker, 0),
		Parentheses:        make([]Marker, 0),
		CandidateNames:     make([]NameFrequency, 0),
		NameEvidence:       make([]NameEvidence, 0),
		NameEvidenceStatus: "not-measured",
		GrammarStatus:      "not-validated",
		UnparsedBytes:      len(source),
		SemanticUnits:      nil,
		ResolvedCalls:      nil,
	}

	frequencies := make(map[string]Frequency)
	for _, position := range positions {
		frequency := frequencies[position.kind]
		frequency.Kind = position.kind
		frequency.Count++
		frequency.Bytes += position.end - position.start
		frequencies[position.kind] = frequency
	}
	for _, frequency := range frequencies {
		report.TokenKinds = append(report.TokenKinds, frequency)
	}
	sort.Slice(report.TokenKinds, func(i, j int) bool {
		return report.TokenKinds[i].Kind < report.TokenKinds[j].Kind
	})

	previous, next := significantNeighbors(positions)
	for index, position := range positions {
		switch position.kind {
		case "T_FUNCTION":
			report.Declarations = append(report.Declarations, functionMarker(position, source, positions, next[index]))
		case "T_FN":
			report.Declarations = append(report.Declarations, marker("arrow-marker", position, nil, source))
		}
	}

	nameCounts := make(map[string]int)
	for index, position := range positions {
		if position.kind != "CHAR_28" {
			continue
		}
		kind, name := parenthesisKind(index, positions, previous)
		report.Parentheses = append(report.Parentheses, marker(kind, position, name, source))
		if kind == "unclassified" {
			report.UnclassifiedParentheses++
		}
		if kind == "name-parenthesis" {
			nameCounts[string(source[name.start:name.end])]++
		}
	}
	names := make([]string, 0, len(nameCounts))
	for name := range nameCounts {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		if nameCounts[names[i]] != nameCounts[names[j]] {
			return nameCounts[names[i]] > nameCounts[names[j]]
		}
		return names[i] < names[j]
	})
	for _, name := range names {
		report.CandidateNames = append(report.CandidateNames, nameFrequency([]byte(name), nameCounts[name]))
	}

	return report, nil
}

func validateTape(source []byte, tokens []Token) ([]tokenPosition, error) {
	if len(source) == 0 && len(tokens) == 0 {
		return make([]tokenPosition, 0), nil
	}
	if len(tokens) == 0 {
		return nil, fmt.Errorf("non-empty source has no tokens")
	}

	positions := make([]tokenPosition, 0, len(tokens))
	offset, line := 0, 1
	for index, token := range tokens {
		if token.Bytes <= 0 {
			return nil, fmt.Errorf("token %d has non-positive byte length", index)
		}
		if token.Bytes > len(source)-offset {
			return nil, fmt.Errorf("token %d exceeds source byte coverage", index)
		}
		if !validKind(token.Kind, source, offset, token.Bytes) {
			return nil, fmt.Errorf("token %d has invalid kind %q", index, token.Kind)
		}

		end := offset + token.Bytes
		positions = append(positions, tokenPosition{kind: token.Kind, start: offset, end: end, line: line})
		line += newlineCount(source[offset:end])
		offset = end
	}
	if offset != len(source) {
		return nil, fmt.Errorf("tokens cover %d source bytes, want %d", offset, len(source))
	}
	return positions, nil
}

func validKind(kind string, source []byte, offset, length int) bool {
	if len(kind) > 80 {
		return false
	}
	if strings.HasPrefix(kind, "T_") {
		return validTokenName(kind[2:])
	}
	if len(kind) != len("CHAR_XX") || !strings.HasPrefix(kind, "CHAR_") || length != 1 {
		return false
	}
	return upperHex(kind[5]) && upperHex(kind[6]) && hexByte(kind[5], kind[6]) == source[offset]
}

func newlineCount(bytes []byte) int {
	count := 0
	for _, value := range bytes {
		if value == '\n' {
			count++
		}
	}
	return count
}

func validTokenName(value string) bool {
	if value == "" {
		return false
	}
	for _, character := range value {
		if !(character >= 'A' && character <= 'Z' || character >= '0' && character <= '9' || character == '_') {
			return false
		}
	}
	return true
}

func upperHex(value byte) bool {
	return value >= '0' && value <= '9' || value >= 'A' && value <= 'F'
}

func hexByte(high, low byte) byte {
	return hexNibble(high)<<4 | hexNibble(low)
}

func hexNibble(value byte) byte {
	if value >= '0' && value <= '9' {
		return value - '0'
	}
	return value - 'A' + 10
}

func significantNeighbors(positions []tokenPosition) ([]int, []int) {
	previous := make([]int, len(positions))
	next := make([]int, len(positions))
	last := -1
	for index, position := range positions {
		previous[index] = last
		if !insignificant(position.kind) {
			last = index
		}
	}
	last = -1
	for index := len(positions) - 1; index >= 0; index-- {
		next[index] = last
		if !insignificant(positions[index].kind) {
			last = index
		}
	}
	return previous, next
}

func insignificant(kind string) bool {
	return kind == "T_WHITESPACE" || kind == "T_COMMENT" || kind == "T_DOC_COMMENT"
}

func functionMarker(position tokenPosition, source []byte, positions []tokenPosition, following int) Marker {
	if following < 0 {
		return marker("unclassified-function-marker", position, nil, source)
	}
	nameIndex := following
	if isAmpersand(positions[nameIndex].kind) {
		nameIndex = nextSignificantAfter(nameIndex, positions)
	}
	if nameIndex >= 0 && positions[nameIndex].kind == "T_STRING" {
		return marker("named-function-marker", position, &positions[nameIndex], source)
	}
	if nameIndex >= 0 && positions[nameIndex].kind == "CHAR_28" {
		return marker("anonymous-function-marker", position, nil, source)
	}
	return marker("unclassified-function-marker", position, nil, source)
}

func nextSignificantAfter(index int, positions []tokenPosition) int {
	for index++; index < len(positions); index++ {
		if !insignificant(positions[index].kind) {
			return index
		}
	}
	return -1
}

func parenthesisKind(index int, positions []tokenPosition, previous []int) (string, *tokenPosition) {
	prior := previous[index]
	if isFunctionParameters(prior, positions, previous) {
		return "function-parameters", nil
	}
	if isArrowParameters(prior, positions, previous) {
		return "arrow-parameters", nil
	}
	if prior < 0 {
		return "unclassified", nil
	}

	position := positions[prior]
	if isName(position.kind) {
		beforeName := previous[prior]
		if beforeName >= 0 {
			switch positions[beforeName].kind {
			case "T_NEW":
				return "constructor-name", nil
			case "T_OBJECT_OPERATOR", "T_NULLSAFE_OBJECT_OPERATOR", "T_DOUBLE_COLON":
				return "member-name", nil
			}
		}
		return "name-parenthesis", &position
	}
	if position.kind == "T_VARIABLE" {
		return "variable-parenthesis", nil
	}
	if languageForm(position.kind) {
		return "language-form", nil
	}
	return "unclassified", nil
}

func isFunctionParameters(prior int, positions []tokenPosition, previous []int) bool {
	if prior < 0 {
		return false
	}
	position := positions[prior]
	if position.kind == "T_FUNCTION" {
		return true
	}
	if isAmpersand(position.kind) {
		before := previous[prior]
		return before >= 0 && positions[before].kind == "T_FUNCTION"
	}
	if position.kind != "T_STRING" {
		return false
	}
	beforeName := previous[prior]
	if beforeName >= 0 && isAmpersand(positions[beforeName].kind) {
		beforeName = previous[beforeName]
	}
	return beforeName >= 0 && positions[beforeName].kind == "T_FUNCTION"
}

func isArrowParameters(prior int, positions []tokenPosition, previous []int) bool {
	if prior < 0 {
		return false
	}
	if positions[prior].kind == "T_FN" {
		return true
	}
	if !isAmpersand(positions[prior].kind) {
		return false
	}
	before := previous[prior]
	return before >= 0 && positions[before].kind == "T_FN"
}

func isAmpersand(kind string) bool {
	switch kind {
	case "CHAR_26", "T_AMPERSAND_FOLLOWED_BY_VAR_OR_VARARG", "T_AMPERSAND_NOT_FOLLOWED_BY_VAR_OR_VARARG":
		return true
	default:
		return false
	}
}

func isName(kind string) bool {
	switch kind {
	case "T_STRING", "T_NAME_QUALIFIED", "T_NAME_FULLY_QUALIFIED", "T_NAME_RELATIVE":
		return true
	default:
		return false
	}
}

func languageForm(kind string) bool {
	switch kind {
	case "T_IF", "T_ELSEIF", "T_FOR", "T_FOREACH", "T_WHILE", "T_SWITCH", "T_CATCH", "T_DECLARE", "T_ISSET", "T_EMPTY", "T_UNSET", "T_ARRAY", "T_LIST", "T_EVAL", "T_EXIT", "T_MATCH", "T_USE":
		return true
	default:
		return false
	}
}

func marker(kind string, position tokenPosition, name *tokenPosition, source []byte) Marker {
	marker := Marker{
		Kind:      kind,
		StartByte: position.start,
		EndByte:   position.end,
		Line:      position.line,
	}
	if name != nil {
		bytes := source[name.start:name.end]
		marker.NameBase64 = base64.StdEncoding.EncodeToString(bytes)
		if utf8.Valid(bytes) {
			marker.Name = string(bytes)
		}
	}
	return marker
}

func nameFrequency(name []byte, count int) NameFrequency {
	frequency := NameFrequency{
		NameBase64: base64.StdEncoding.EncodeToString(name),
		Count:      count,
	}
	if utf8.Valid(name) {
		frequency.Name = string(name)
	}
	return frequency
}
