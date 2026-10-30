package runtime

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	jsonErrorNone          int64 = 0
	jsonErrorDepth         int64 = 1
	jsonErrorStateMismatch int64 = 2
	jsonErrorControl       int64 = 3
	jsonErrorSyntax        int64 = 4
	jsonErrorUTF8          int64 = 5
	jsonErrorUTF16         int64 = 10
)

// JSONDecodeAssociative implements the bounded json_decode($text, true)
// profile used by generated programs. It accepts an actual string and returns
// only nil, bool, string, int64, and ordered Array values. Decimal/exponent
// numbers and integers outside int64 are deliberately outside this profile.
func (execution *Execution) JSONDecodeAssociative(text any) any {
	RequireExecution(execution)

	source, ok := text.(string)
	if !ok {
		panic("json_decode requires actual string")
	}

	parser := jsonAssociativeParser{text: source, depth: 1}
	value, failure := parser.parseDocument()
	if failure.code != jsonErrorNone {
		execution.jsonError = failure.code
		return nil
	}
	if parser.unsupportedNumber {
		// Preserve the previous JSON error state: this is a profile boundary,
		// rather than PHP's recoverable JSON parser failure.
		panic("json_decode number outside int64 profile")
	}

	execution.jsonError = jsonErrorNone
	return Clone(value)
}

// JSONLastError reports the result of the last JSONDecodeAssociative call for
// this execution. A fresh execution has no JSON error.
func (execution *Execution) JSONLastError() int64 {
	RequireExecution(execution)
	return execution.jsonError
}

// JSONLastErrorMessage is the PHP spelling for the current JSON error code.
func (execution *Execution) JSONLastErrorMessage() string {
	RequireExecution(execution)
	switch execution.jsonError {
	case jsonErrorNone:
		return "No error"
	case jsonErrorDepth:
		return "Maximum stack depth exceeded"
	case jsonErrorStateMismatch:
		return "State mismatch (invalid or malformed JSON)"
	case jsonErrorControl:
		return "Control character error, possibly incorrectly encoded"
	case jsonErrorSyntax:
		return "Syntax error"
	case jsonErrorUTF8:
		return "Malformed UTF-8 characters, possibly incorrectly encoded"
	case jsonErrorUTF16:
		return "Single unpaired UTF-16 surrogate in unicode escape"
	default:
		return "Unknown error"
	}
}

type jsonDecodeFailure struct{ code int64 }

type jsonAssociativeParser struct {
	text              string
	pos, depth        int
	unsupportedNumber bool
}

func (parser *jsonAssociativeParser) parseDocument() (any, jsonDecodeFailure) {
	value, failure := parser.parseValue()
	if failure.code != jsonErrorNone {
		return nil, failure
	}
	if failure = parser.skipWhitespace(); failure.code != jsonErrorNone {
		return nil, failure
	}
	if parser.pos != len(parser.text) {
		return nil, parser.unexpectedByte()
	}
	return value, jsonDecodeFailure{}
}

func (parser *jsonAssociativeParser) parseValue() (any, jsonDecodeFailure) {
	if failure := parser.skipWhitespace(); failure.code != jsonErrorNone {
		return nil, failure
	}
	if parser.pos == len(parser.text) {
		return nil, parser.fail(jsonErrorSyntax)
	}

	switch parser.text[parser.pos] {
	case '{':
		return parser.parseObject()
	case '[':
		return parser.parseArray()
	case '"':
		return parser.parseString()
	case 't':
		return parser.parseLiteral("true", true)
	case 'f':
		return parser.parseLiteral("false", false)
	case 'n':
		return parser.parseLiteral("null", nil)
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return parser.parseNumber()
	default:
		return nil, parser.unexpectedByte()
	}
}

func (parser *jsonAssociativeParser) parseObject() (any, jsonDecodeFailure) {
	if failure := parser.openContainer(); failure.code != jsonErrorNone {
		return nil, failure
	}
	defer func() { parser.depth-- }()
	parser.pos++ // {

	if failure := parser.skipWhitespace(); failure.code != jsonErrorNone {
		return nil, failure
	}
	if parser.pos == len(parser.text) {
		return nil, parser.fail(jsonErrorSyntax)
	}
	if parser.text[parser.pos] == '}' {
		parser.pos++
		return NewArray(), jsonDecodeFailure{}
	}
	if parser.text[parser.pos] == ']' {
		return nil, parser.fail(jsonErrorStateMismatch)
	}

	entries := make([]Entry, 0)
	for {
		if parser.text[parser.pos] != '"' {
			return nil, parser.unexpectedByte()
		}
		key, failure := parser.parseString()
		if failure.code != jsonErrorNone {
			return nil, failure
		}
		if failure = parser.skipWhitespace(); failure.code != jsonErrorNone {
			return nil, failure
		}
		if parser.pos == len(parser.text) {
			return nil, parser.fail(jsonErrorSyntax)
		}
		if parser.text[parser.pos] != ':' {
			return nil, parser.unexpectedByte()
		}
		parser.pos++

		value, failure := parser.parseValue()
		if failure.code != jsonErrorNone {
			return nil, failure
		}
		entries = append(entries, Entry{Key: key, Value: value})

		if failure = parser.skipWhitespace(); failure.code != jsonErrorNone {
			return nil, failure
		}
		if parser.pos == len(parser.text) {
			return nil, parser.fail(jsonErrorSyntax)
		}
		switch parser.text[parser.pos] {
		case '}':
			parser.pos++
			return NewArray(entries...), jsonDecodeFailure{}
		case ']':
			return nil, parser.fail(jsonErrorStateMismatch)
		case ',':
			parser.pos++
			if failure = parser.skipWhitespace(); failure.code != jsonErrorNone {
				return nil, failure
			}
			if parser.pos == len(parser.text) || parser.text[parser.pos] == '}' || parser.text[parser.pos] == ']' {
				return nil, parser.fail(jsonErrorSyntax)
			}
		default:
			return nil, parser.unexpectedByte()
		}
	}
}

func (parser *jsonAssociativeParser) parseArray() (any, jsonDecodeFailure) {
	if failure := parser.openContainer(); failure.code != jsonErrorNone {
		return nil, failure
	}
	defer func() { parser.depth-- }()
	parser.pos++ // [

	if failure := parser.skipWhitespace(); failure.code != jsonErrorNone {
		return nil, failure
	}
	if parser.pos == len(parser.text) {
		return nil, parser.fail(jsonErrorSyntax)
	}
	if parser.text[parser.pos] == ']' {
		parser.pos++
		return NewArray(), jsonDecodeFailure{}
	}
	if parser.text[parser.pos] == '}' {
		return nil, parser.fail(jsonErrorStateMismatch)
	}

	entries := make([]Entry, 0)
	for {
		value, failure := parser.parseValue()
		if failure.code != jsonErrorNone {
			return nil, failure
		}
		entries = append(entries, Entry{Value: value})

		if failure = parser.skipWhitespace(); failure.code != jsonErrorNone {
			return nil, failure
		}
		if parser.pos == len(parser.text) {
			return nil, parser.fail(jsonErrorSyntax)
		}
		switch parser.text[parser.pos] {
		case ']':
			parser.pos++
			return NewArray(entries...), jsonDecodeFailure{}
		case '}':
			return nil, parser.fail(jsonErrorStateMismatch)
		case ',':
			parser.pos++
			if failure = parser.skipWhitespace(); failure.code != jsonErrorNone {
				return nil, failure
			}
			if parser.pos == len(parser.text) || parser.text[parser.pos] == ']' || parser.text[parser.pos] == '}' {
				return nil, parser.fail(jsonErrorSyntax)
			}
		default:
			return nil, parser.unexpectedByte()
		}
	}
}

func (parser *jsonAssociativeParser) parseString() (any, jsonDecodeFailure) {
	parser.pos++ // opening quote
	var decoded strings.Builder
	for parser.pos < len(parser.text) {
		byteValue := parser.text[parser.pos]
		switch byteValue {
		case '"':
			parser.pos++
			return decoded.String(), jsonDecodeFailure{}
		case '\\':
			parser.pos++
			if parser.pos == len(parser.text) {
				return nil, parser.fail(jsonErrorSyntax)
			}
			escape := parser.text[parser.pos]
			parser.pos++
			switch escape {
			case '"', '\\', '/':
				decoded.WriteByte(escape)
			case 'b':
				decoded.WriteByte('\b')
			case 'f':
				decoded.WriteByte('\f')
			case 'n':
				decoded.WriteByte('\n')
			case 'r':
				decoded.WriteByte('\r')
			case 't':
				decoded.WriteByte('\t')
			case 'u':
				unit, failure := parser.parseUnicodeUnit()
				if failure.code != jsonErrorNone {
					return nil, failure
				}
				if unit >= 0xd800 && unit <= 0xdbff {
					if parser.pos+2 > len(parser.text) || parser.text[parser.pos] != '\\' || parser.text[parser.pos+1] != 'u' {
						return nil, parser.fail(jsonErrorUTF16)
					}
					parser.pos += 2
					low, nextFailure := parser.parseUnicodeUnit()
					if nextFailure.code != jsonErrorNone {
						// The complete first surrogate is already an invalid UTF16
						// token unless the full low-surrogate escape matches.
						return nil, parser.fail(jsonErrorUTF16)
					}
					if low < 0xdc00 || low > 0xdfff {
						return nil, parser.fail(jsonErrorUTF16)
					}
					decoded.WriteRune(0x10000 + (rune(unit-0xd800) << 10) + rune(low-0xdc00))
				} else if unit >= 0xdc00 && unit <= 0xdfff {
					return nil, parser.fail(jsonErrorUTF16)
				} else {
					decoded.WriteRune(rune(unit))
				}
			default:
				return nil, parser.fail(jsonErrorSyntax)
			}
		default:
			if byteValue < 0x20 {
				return nil, parser.fail(jsonErrorControl)
			}
			if byteValue < utf8.RuneSelf {
				decoded.WriteByte(byteValue)
				parser.pos++
				continue
			}
			r, size := utf8.DecodeRuneInString(parser.text[parser.pos:])
			if r == utf8.RuneError && size == 1 {
				return nil, parser.fail(jsonErrorUTF8)
			}
			decoded.WriteString(parser.text[parser.pos : parser.pos+size])
			parser.pos += size
		}
	}
	// PHP's scanner classifies end-of-input before a closing quote as its
	// control-character sentinel, rather than a grammar-level syntax error.
	return nil, parser.fail(jsonErrorControl)
}

func (parser *jsonAssociativeParser) parseUnicodeUnit() (uint16, jsonDecodeFailure) {
	if parser.pos+4 > len(parser.text) {
		return 0, parser.fail(jsonErrorSyntax)
	}
	var value uint16
	for i := 0; i < 4; i++ {
		hex, ok := jsonHex(parser.text[parser.pos+i])
		if !ok {
			return 0, parser.fail(jsonErrorSyntax)
		}
		value = value<<4 | uint16(hex)
	}
	parser.pos += 4
	return value, jsonDecodeFailure{}
}

func jsonHex(byteValue byte) (byte, bool) {
	switch {
	case byteValue >= '0' && byteValue <= '9':
		return byteValue - '0', true
	case byteValue >= 'a' && byteValue <= 'f':
		return byteValue - 'a' + 10, true
	case byteValue >= 'A' && byteValue <= 'F':
		return byteValue - 'A' + 10, true
	default:
		return 0, false
	}
}

func (parser *jsonAssociativeParser) parseLiteral(token string, value any) (any, jsonDecodeFailure) {
	if !strings.HasPrefix(parser.text[parser.pos:], token) {
		return nil, parser.fail(jsonErrorSyntax)
	}
	parser.pos += len(token)
	return value, jsonDecodeFailure{}
}

func (parser *jsonAssociativeParser) parseNumber() (any, jsonDecodeFailure) {
	start := parser.pos
	if parser.text[parser.pos] == '-' {
		parser.pos++
		if parser.pos == len(parser.text) {
			return nil, parser.fail(jsonErrorSyntax)
		}
	}
	if parser.text[parser.pos] == '0' {
		parser.pos++
	} else if parser.text[parser.pos] >= '1' && parser.text[parser.pos] <= '9' {
		for parser.pos < len(parser.text) && parser.text[parser.pos] >= '0' && parser.text[parser.pos] <= '9' {
			parser.pos++
		}
	} else {
		return nil, parser.fail(jsonErrorSyntax)
	}

	isDecimal := false
	if parser.pos < len(parser.text) && parser.text[parser.pos] == '.' {
		isDecimal = true
		parser.pos++
		fractionStart := parser.pos
		for parser.pos < len(parser.text) && parser.text[parser.pos] >= '0' && parser.text[parser.pos] <= '9' {
			parser.pos++
		}
		if parser.pos == fractionStart {
			return nil, parser.fail(jsonErrorSyntax)
		}
	}
	if parser.pos < len(parser.text) && (parser.text[parser.pos] == 'e' || parser.text[parser.pos] == 'E') {
		isDecimal = true
		parser.pos++
		if parser.pos < len(parser.text) && (parser.text[parser.pos] == '+' || parser.text[parser.pos] == '-') {
			parser.pos++
		}
		exponentStart := parser.pos
		for parser.pos < len(parser.text) && parser.text[parser.pos] >= '0' && parser.text[parser.pos] <= '9' {
			parser.pos++
		}
		if parser.pos == exponentStart {
			return nil, parser.fail(jsonErrorSyntax)
		}
	}

	token := parser.text[start:parser.pos]
	if isDecimal {
		parser.unsupportedNumber = true
		return nil, jsonDecodeFailure{}
	}
	value, err := strconv.ParseInt(token, 10, 64)
	if err != nil {
		parser.unsupportedNumber = true
		return nil, jsonDecodeFailure{}
	}
	return value, jsonDecodeFailure{}
}

func (parser *jsonAssociativeParser) openContainer() jsonDecodeFailure {
	if parser.depth >= 512 {
		return parser.fail(jsonErrorDepth)
	}
	parser.depth++
	return jsonDecodeFailure{}
}

func (parser *jsonAssociativeParser) skipWhitespace() jsonDecodeFailure {
	for parser.pos < len(parser.text) {
		switch parser.text[parser.pos] {
		case ' ', '\t', '\n', '\r':
			parser.pos++
		default:
			if parser.text[parser.pos] < 0x20 {
				return parser.fail(jsonErrorControl)
			}
			return jsonDecodeFailure{}
		}
	}
	return jsonDecodeFailure{}
}

func (parser *jsonAssociativeParser) unexpectedByte() jsonDecodeFailure {
	if parser.pos == len(parser.text) {
		return parser.fail(jsonErrorSyntax)
	}
	byteValue := parser.text[parser.pos]
	if byteValue == '"' {
		// A lexer scans the whole lookahead string before the grammar can
		// reject its position. Preserve a malformed string's encoding/control
		// error instead of overwriting it with a generic syntax error.
		if _, failure := parser.parseString(); failure.code != jsonErrorNone {
			return failure
		}
		return parser.fail(jsonErrorSyntax)
	}
	if byteValue < 0x20 {
		return parser.fail(jsonErrorControl)
	}
	if byteValue >= utf8.RuneSelf {
		r, size := utf8.DecodeRuneInString(parser.text[parser.pos:])
		if r == utf8.RuneError && size == 1 {
			return parser.fail(jsonErrorUTF8)
		}
	}
	return parser.fail(jsonErrorSyntax)
}

func (parser *jsonAssociativeParser) fail(code int64) jsonDecodeFailure {
	return jsonDecodeFailure{code: code}
}
