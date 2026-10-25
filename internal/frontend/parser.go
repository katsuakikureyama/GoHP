// Package frontend parses an intentionally closed PHP subset. Unknown tokens fail closed.
package frontend

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/katsuakikureyama/GoHP/internal/ir"
	"github.com/katsuakikureyama/GoHP/internal/phpversion"
)

type token struct {
	text string
	line int
	kind string
}
type parser struct {
	ts      []token
	pos     int
	version string
}

func lex(s string) ([]token, error) {
	if !strings.HasPrefix(s, "<?php") {
		return nil, fmt.Errorf("PHP001 line 1: expected <?php")
	}
	s = s[5:]
	ts := []token{}
	line := 1
	for i := 0; i < len(s); {
		c := s[i]
		if c == '\n' {
			line++
			i++
			continue
		}
		if c == ' ' || c == '\t' || c == '\r' {
			i++
			continue
		}
		if strings.HasPrefix(s[i:], "//") || c == '#' {
			for i < len(s) && s[i] != '\n' {
				i++
			}
			continue
		}
		if strings.HasPrefix(s[i:], "/*") {
			j := strings.Index(s[i+2:], "*/")
			if j < 0 {
				return nil, fmt.Errorf("PHP001 line %d: unterminated comment", line)
			}
			end := i + 2 + j + 2
			line += strings.Count(s[i:end], "\n")
			i = end
			continue
		}
		start := i
		ln := line
		if c == '\'' || c == '"' {
			quote := c
			i++
			var b strings.Builder
			for i < len(s) && s[i] != quote {
				if s[i] == '\n' {
					line++
				}
				if quote == '"' && s[i] == '$' {
					return nil, fmt.Errorf("PHP002 line %d: string interpolation unsupported", ln)
				}
				if s[i] == '\\' {
					i++
					if i >= len(s) {
						break
					}
					switch s[i] {
					case '\\':
						b.WriteByte('\\')
					case '\'':
						if quote == '\'' {
							b.WriteByte('\'')
						} else {
							b.WriteString("\\'")
						}
					case '"':
						if quote == '"' {
							b.WriteByte('"')
						} else {
							b.WriteString("\\\"")
						}
					case 'n', 't', 'r':
						if quote == '\'' {
							b.WriteByte('\\')
							b.WriteByte(s[i])
						} else {
							b.WriteByte(map[byte]byte{'n': '\n', 't': '\t', 'r': '\r'}[s[i]])
						}
					default:
						return nil, fmt.Errorf("PHP002 line %d: unsupported escape", ln)
					}
					i++
					continue
				}
				b.WriteByte(s[i])
				i++
			}
			if i >= len(s) {
				return nil, fmt.Errorf("PHP001 line %d: unterminated string", ln)
			}
			i++
			ts = append(ts, token{b.String(), ln, "string"})
			continue
		}
		if c == '$' {
			i++
			if i >= len(s) || !identStart(s[i]) {
				return nil, fmt.Errorf("PHP002 line %d: dynamic variable unsupported", ln)
			}
			for i < len(s) && ident(s[i]) {
				i++
			}
			ts = append(ts, token{s[start+1 : i], ln, "var"})
			continue
		}
		if identStart(c) {
			i++
			for i < len(s) && ident(s[i]) {
				i++
			}
			ts = append(ts, token{s[start:i], ln, "id"})
			continue
		}
		if c >= '0' && c <= '9' {
			i++
			for i < len(s) && s[i] >= '0' && s[i] <= '9' {
				i++
			}
			if i < len(s) && s[i] == '.' {
				return nil, fmt.Errorf("PHP002 line %d: floating point literals unsupported", ln)
			}
			v := s[start:i]
			if len(v) > 1 && v[0] == '0' {
				return nil, fmt.Errorf("PHP002 line %d: non-decimal integer unsupported", ln)
			}
			if _, e := strconv.ParseInt(v, 10, 64); e != nil {
				return nil, fmt.Errorf("PHP002 line %d: integer outside int64", ln)
			}
			ts = append(ts, token{v, ln, "int"})
			continue
		}
		matched := false
		// Keep multi-character operators longest-first so === is never split
		// into the unsupported = token followed by ==.
		for _, op := range []string{"===", "!==", "&&", "||", "=>", "->"} {
			if strings.HasPrefix(s[i:], op) {
				ts = append(ts, token{op, ln, "sym"})
				i += len(op)
				matched = true
				break
			}
		}
		if matched {
			continue
		}
		if strings.ContainsRune("(){}[];,=+.-?:!", rune(c)) {
			ts = append(ts, token{string(c), ln, "sym"})
			i++
			continue
		}
		return nil, fmt.Errorf("PHP002 line %d: unsupported token %q", line, c)
	}
	return append(ts, token{"<eof>", line, "eof"}), nil
}
func identStart(c byte) bool  { return c == '_' || unicode.IsLetter(rune(c)) && c < 128 }
func ident(c byte) bool       { return identStart(c) || c >= '0' && c <= '9' }
func (p *parser) peek() token { return p.ts[p.pos] }
func (p *parser) pop() token {
	t := p.peek()
	if t.kind != "eof" {
		p.pos++
	}
	return t
}
func (p *parser) accept(s string) bool {
	if p.peek().text == s && p.peek().kind != "string" && p.peek().kind != "var" {
		p.pop()
		return true
	}
	return false
}
func (p *parser) need(s string) {
	if !p.accept(s) {
		p.fail("expected " + s)
	}
}
func (p *parser) fail(s string) {
	panic(fmt.Errorf("PHP002 line %d: %s (at %q)", p.peek().line, s, p.peek().text))
}
func (p *parser) name(kind string) string {
	t := p.pop()
	if t.kind != kind {
		p.fail("expected " + kind)
	}
	return t.text
}
func Parse(src, version string) (out ir.Program, err error) {
	defer func() {
		if e := recover(); e != nil {
			if x, ok := e.(error); ok {
				err = x
			} else {
				panic(e)
			}
		}
	}()
	if err := phpversion.ValidateTarget(version); err != nil {
		return out, fmt.Errorf("PHP003: %w", err)
	}
	ts, e := lex(src)
	if e != nil {
		return out, e
	}
	p := &parser{ts: ts, version: version}
	out.Version = ir.Version
	out.PHPVersion = version
	for p.peek().kind != "eof" {
		switch {
		case p.accept("function"):
			out.Functions = append(out.Functions, p.function())
		case p.accept("class"):
			out.Classes = append(out.Classes, p.class())
		default:
			out.Body = append(out.Body, p.stmt())
		}
	}
	return out, nil
}
func (p *parser) function() ir.Function {
	f := ir.Function{Name: p.name("id"), Line: p.peek().line}
	p.need("(")
	if !p.accept(")") {
		for {
			f.Params = append(f.Params, p.name("var"))
			if p.accept(")") {
				break
			}
			p.need(",")
		}
	}
	if p.accept(":") {
		t := p.pop()
		returnType := strings.ToLower(t.text)
		if t.kind != "id" || (returnType != "array" && returnType != "bool") {
			p.fail("only array and bool return declarations are supported")
		}
		f.ReturnType = returnType
	}
	f.Body = p.block()
	return f
}
func (p *parser) class() ir.Class {
	c := ir.Class{Name: p.name("id")}
	p.need("{")
	for !p.accept("}") {
		p.need("public")
		if p.accept("function") {
			c.Methods = append(c.Methods, p.function())
		} else {
			n := p.name("var")
			p.need("=")
			e := p.expr(0)
			if e.Kind != "int" && e.Kind != "string" && e.Kind != "bool" && e.Kind != "null" {
				p.fail("property defaults must be scalar literals")
			}
			p.need(";")
			c.Properties = append(c.Properties, ir.Property{Name: n, Value: e})
		}
	}
	return c
}
func (p *parser) block() []ir.Stmt {
	p.need("{")
	var b []ir.Stmt
	for !p.accept("}") {
		if p.peek().kind == "eof" {
			p.fail("unterminated block")
		}
		b = append(b, p.stmt())
	}
	return b
}
func (p *parser) stmt() ir.Stmt {
	s := ir.Stmt{Line: p.peek().line}
	switch {
	case p.accept("echo"):
		s.Kind = "echo"
		s.Expr = p.expr(0)
		p.need(";")
	case p.accept("return"):
		s.Kind = "return"
		s.Expr = p.expr(0)
		p.need(";")
	case p.accept("foreach"):
		s.Kind = "foreach"
		p.need("(")
		s.Expr = p.expr(0)
		p.need("as")
		s.Value = p.name("var")
		if p.accept("=>") {
			s.Key = s.Value
			s.Value = p.name("var")
		}
		p.need(")")
		s.Body = p.block()
	case p.accept("if"):
		s = p.ifStmt(s.Line)
	default:
		if p.peek().kind != "var" {
			p.fail("unsupported statement")
		}
		s.Kind = "assign"
		s.Name = p.name("var")
		if p.peek().kind == "sym" && p.peek().text == "[" {
			p.pop()
			base := ir.Expr{Kind: "var", Text: s.Name, Line: s.Line}
			if p.peek().kind == "sym" && p.peek().text == "]" {
				p.pop()
				s.Kind = "assign_append"
				s.Target = base
			} else {
				key := p.expr(0)
				if p.peek().kind != "sym" || p.peek().text != "]" {
					p.fail("expected array assignment closing bracket")
				}
				p.pop()
				s.Kind = "assign_index"
				s.Target = ir.Expr{Kind: "index", Line: s.Line, Args: []ir.Expr{base, key}}
			}
			s.Name = ""
		}
		if (s.Kind == "assign_index" || s.Kind == "assign_append") && (p.peek().kind != "sym" || p.peek().text != "=") {
			p.fail("array assignment requires one level and =")
		}
		p.need("=")
		s.Expr = p.expr(0)
		p.need(";")
	}
	return s
}

// The leading if/elseif token has already been consumed. Normalize chains to
// nested if statements so validation and codegen share the existing lazy arms.
func (p *parser) ifStmt(line int) ir.Stmt {
	s := ir.Stmt{Kind: "if", Line: line}
	p.need("(")
	s.Expr = p.expr(0)
	p.need(")")
	s.Body = p.block()
	branch := p.peek()
	if p.accept("elseif") {
		s.ElseBody = []ir.Stmt{p.ifStmt(branch.line)}
	} else if p.accept("else") {
		branch = p.peek()
		if p.accept("if") {
			s.ElseBody = []ir.Stmt{p.ifStmt(branch.line)}
		} else {
			s.ElseBody = p.block()
			if s.ElseBody == nil {
				s.ElseBody = []ir.Stmt{}
			}
		}
	}
	return s
}

func isStrictComparison(t token) bool {
	return t.kind == "sym" && (t.text == "===" || t.text == "!==")
}

func (p *parser) prec(t token) int {
	if t.kind == "sym" && t.text == "." {
		return 4
	}
	if t.kind == "sym" && t.text == "+" {
		return 5
	}
	if isStrictComparison(t) {
		return 3
	}
	if t.kind == "sym" && t.text == "&&" {
		return 2
	}
	if t.kind == "sym" && t.text == "||" {
		return 1
	}
	return 0
}
func (p *parser) expr(min int) ir.Expr {
	return p.exprMode(min, true)
}

// The final operand cannot contain an unparenthesized ternary in PHP 8.
// Parentheses call expr again, intentionally resetting this restriction.
func (p *parser) exprMode(min int, allowTernary bool) ir.Expr {
	t := p.pop()
	e := ir.Expr{Line: t.line, Text: t.text}
	switch t.kind {
	case "int", "string", "var":
		e.Kind = t.kind
	case "id":
		if strings.EqualFold(t.text, "isset") {
			// Keep the operand as an address, not an eagerly evaluated index.
			// Only one defined array variable and one literal/variable key are
			// supported; broader PHP isset forms remain outside this subset.
			p.need("(")
			base := p.pop()
			if base.kind != "var" {
				p.fail("isset requires a single indexed variable")
			}
			p.need("[")
			key := p.pop()
			keyExpr := ir.Expr{Kind: key.kind, Text: key.text, Line: key.line}
			if key.kind == "sym" && key.text == "-" {
				keyExpr = p.negativeInteger(key)
			} else if key.kind != "int" && key.kind != "string" && key.kind != "var" {
				p.fail("isset key must be an integer/string literal or variable")
			}
			p.need("]")
			p.need(")")
			e = ir.Expr{Kind: "isset_index", Line: t.line, Args: []ir.Expr{
				{Kind: "var", Text: base.text, Line: base.line},
				keyExpr,
			}}
			break
		}
		// array(...) is a language construct, not a builtin call. Both literal
		// spellings share the same IR and the existing closed semantic boundary.
		if strings.EqualFold(t.text, "array") {
			p.need("(")
			e = p.arrayLiteral(t.line, ")")
			break
		}
		// These three predefined literals are case-insensitive in PHP. Do not
		// fold identifiers globally: variable/property names retain their case.
		literal := strings.ToLower(t.text)
		if literal == "true" || literal == "false" {
			e.Kind, e.Text = "bool", literal
			break
		}
		if literal == "null" {
			e.Kind, e.Text = "null", literal
			break
		}
		switch t.text {
		case "new":
			e.Kind = "new"
			e.Text = p.name("id")
			p.need("(")
			p.need(")")
		default:
			e.Kind = "call"
			p.need("(")
			e.Args = p.args(")")
		}
	case "sym":
		switch t.text {
		case "!":
			e = ir.Expr{Kind: "not", Args: []ir.Expr{p.expr(6)}, Line: t.line}
		case "-":
			e = p.negativeInteger(t)
		case "(":
			// The supported cast spelling is deliberately narrow. Looking ahead
			// keeps ordinary parenthesized expressions intact.
			if p.peek().kind == "id" && strings.EqualFold(p.peek().text, "bool") && p.pos+1 < len(p.ts) && p.ts[p.pos+1].text == ")" {
				p.pop()
				p.need(")")
				e = ir.Expr{Kind: "bool_cast", Args: []ir.Expr{p.expr(6)}, Line: t.line}
			} else {
				e = p.expr(0)
				p.need(")")
			}
		case "[":
			e = p.arrayLiteral(t.line, "]")
		default:
			p.fail("unsupported expression")
		}
	default:
		p.fail("unsupported expression")
	}
	for {
		if p.accept("[") {
			idx := p.expr(0)
			p.need("]")
			e = ir.Expr{Kind: "index", Args: []ir.Expr{e, idx}, Line: t.line}
			continue
		}
		if p.accept("->") {
			n := p.name("id")
			if p.accept("(") {
				e = ir.Expr{Kind: "method", Text: n, Args: append([]ir.Expr{e}, p.args(")")...), Line: t.line}
			} else {
				e = ir.Expr{Kind: "property", Text: n, Args: []ir.Expr{e}, Line: t.line}
			}
			continue
		}
		op := p.peek()
		q := p.prec(op)
		if q == 0 || q < min {
			break
		}
		p.pop()
		right := p.expr(q + 1)
		e = ir.Expr{Kind: "binary", Text: op.text, Args: []ir.Expr{e, right}, Line: op.line}
		if isStrictComparison(op) && isStrictComparison(p.peek()) {
			p.fail("strict comparison is non-associative; use parentheses")
		}
	}
	if min == 0 && p.peek().kind == "sym" && p.peek().text == "?" {
		if !allowTernary {
			p.fail("ternary is non-associative; use parentheses")
		}
		question := p.pop()
		if p.peek().kind == "sym" && p.peek().text == ":" {
			p.fail("short ternary is unsupported")
		}
		yes := p.expr(0)
		if p.peek().kind != "sym" {
			p.fail("expected ternary colon")
		}
		p.need(":")
		no := p.exprMode(0, false)
		e = ir.Expr{Kind: "ternary", Args: []ir.Expr{e, yes, no}, Line: question.line}
	}
	return e
}

// negativeInteger folds only a range-checked decimal token, never an expression.
// The positive token in -9223372036854775808 is outside the integer subset and
// is rejected by lex before folding. Keep the minus token's source position.
func (p *parser) negativeInteger(sign token) ir.Expr {
	if p.peek().kind != "int" {
		p.fail("minus requires a decimal integer literal")
	}
	n := p.pop()
	e := ir.Expr{Kind: "int", Text: "-" + n.text, Line: sign.line}
	if n.text == "0" {
		e.Text = "0"
	}
	return e
}

func (p *parser) arrayLiteral(line int, end string) ir.Expr {
	e := ir.Expr{Kind: "array", Line: line}
	if p.accept(end) {
		return e
	}
	for {
		v := p.expr(0)
		k := ir.Expr{Kind: "auto", Line: v.Line}
		if p.accept("=>") {
			k = v
			v = p.expr(0)
		}
		e.Args = append(e.Args, ir.Expr{Kind: "pair", Args: []ir.Expr{k, v}, Line: v.Line})
		if p.accept(end) {
			return e
		}
		p.need(",")
		if p.accept(end) {
			return e
		}
	}
}

func (p *parser) args(end string) []ir.Expr {
	var a []ir.Expr
	if p.accept(end) {
		return a
	}
	for {
		a = append(a, p.expr(0))
		if p.accept(end) {
			return a
		}
		p.need(",")
	}
}
