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
	ts                []token
	pos               int
	version           string
	experimentalPCRE2 bool
}

// Options selects syntax that is intentionally outside the default closed
// frontend. Enabling it only records source admission; it never installs or
// enables a PCRE2 backend for a generated Go build.
type Options struct {
	ExperimentalPCRE2 bool
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
		if strings.HasPrefix(s[i:], "#[") {
			return nil, fmt.Errorf("PHP002 line %d: PHP attributes unsupported", line)
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
					// Single quotes decode only backslash and quote. Every other
					// escaped byte remains literal, including a physical newline.
					if quote == '\'' {
						if s[i] != '\\' && s[i] != '\'' {
							b.WriteByte('\\')
						}
						if s[i] == '\n' {
							line++
						}
						b.WriteByte(s[i])
						i++
						continue
					}
					switch s[i] {
					case '\\':
						b.WriteByte('\\')
					case '\'':
						b.WriteString("\\'")
					case '"':
						b.WriteByte('"')
					case 'n', 't', 'r', 'f':
						b.WriteByte(map[byte]byte{'n': '\n', 't': '\t', 'r': '\r', 'f': '\f'}[s[i]])
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
		for _, op := range []string{"===", "!==", "&&", "||", "??", "=>", "->", "--", ".="} {
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
		if strings.ContainsRune("(){}[];,=+.-?:!<>", rune(c)) {
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
func Parse(src, version string) (ir.Program, error) {
	return ParseWithOptions(src, version, Options{})
}

// ParseWithOptions parses the closed subset plus explicitly selected
// experiments. Semantic validation remains the authority for target/profile
// admission and for malformed manually constructed IR.
func ParseWithOptions(src, version string, options Options) (out ir.Program, err error) {
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
	p := &parser{ts: ts, version: version, experimentalPCRE2: options.ExperimentalPCRE2}
	out.Version = ir.Version
	out.PHPVersion = version
	if options.ExperimentalPCRE2 {
		out.RegexProfile = ir.Preg83Profile
	}
	for p.peek().kind != "eof" {
		switch {
		case p.accept("function"):
			out.Functions = append(out.Functions, p.function(false))
		case p.accept("class"):
			out.Classes = append(out.Classes, p.class())
		default:
			out.Body = append(out.Body, p.stmt())
		}
	}
	return out, nil
}
func (p *parser) function(method bool) ir.Function {
	f := ir.Function{Name: p.name("id"), Line: p.peek().line}
	p.need("(")
	var paramTypes []string
	typed := false
	if !p.accept(")") {
		for {
			paramType := ""
			if p.peek().kind == "id" {
				annotation := p.pop()
				paramType = strings.ToLower(annotation.text)
				switch paramType {
				case "array", "callable":
				case "int", "string", "bool":
					if method && !strings.EqualFold(f.Name, "__construct") {
						p.fail("scalar parameter declarations require a global function or public constructor")
					}
				case "float", "object", "mixed", "iterable", "void", "never", "null", "false", "true", "self", "parent", "static", "boolean", "integer", "double", "real", "resource":
					p.fail("unsupported parameter declaration")
				default:
					paramType = "class:" + paramType
				}
				typed = true
			}
			f.Params = append(f.Params, p.name("var"))
			paramTypes = append(paramTypes, paramType)
			if p.accept("=") {
				value := p.pop()
				if value.kind != "id" || !strings.EqualFold(value.text, "null") {
					p.fail("only null parameter defaults are supported")
				}
				f.OptionalNullParams++
			} else if f.OptionalNullParams != 0 {
				p.fail("required parameters cannot follow optional parameters")
			}
			if p.accept(")") {
				break
			}
			p.need(",")
		}
	}
	if typed {
		f.ParamTypes = paramTypes
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
	line := p.peek().line
	c := ir.Class{Name: p.name("id"), Line: line}
	p.need("{")
	for !p.accept("}") {
		p.need("public")
		if p.accept("function") {
			c.Methods = append(c.Methods, p.function(true))
		} else {
			propertyLine := p.peek().line
			n := p.name("var")
			if p.accept(";") {
				// An untyped public property without an initializer is null.
				c.Properties = append(c.Properties, ir.Property{Name: n, Value: ir.Expr{Kind: "null", Text: "null", Line: propertyLine}})
				continue
			}
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
	case p.peek().kind == "id" && strings.EqualFold(p.peek().text, "switch"):
		p.pop()
		s.Kind = "switch"
		p.need("(")
		s.Expr = p.expr(0)
		p.need(")")
		p.need("{")
		for !p.accept("}") {
			label := p.pop()
			clause := ir.Stmt{Kind: strings.ToLower(label.text), Line: label.line}
			if label.kind != "id" || (clause.Kind != "case" && clause.Kind != "default") {
				p.fail("switch requires case or default")
			}
			if clause.Kind == "case" {
				value := p.pop()
				if value.kind != "string" {
					p.fail("switch case requires a literal string")
				}
				clause.Expr = ir.Expr{Kind: "string", Text: value.text, Line: value.line}
			}
			p.need(":")
			for p.peek().text != "}" && !(p.peek().kind == "id" && (strings.EqualFold(p.peek().text, "case") || strings.EqualFold(p.peek().text, "default"))) {
				if p.peek().kind == "eof" {
					p.fail("unterminated switch")
				}
				clause.Body = append(clause.Body, p.stmt())
			}
			s.Body = append(s.Body, clause)
		}
	case p.peek().kind == "id" && strings.EqualFold(p.peek().text, "list"):
		p.pop()
		s.Kind = "list_assign"
		s.Target = ir.Expr{Kind: "list", Line: s.Line}
		p.need("(")
		for {
			line := p.peek().line
			name := p.name("var") // Nonempty, no holes, keyed or nested forms.
			s.Target.Args = append(s.Target.Args, ir.Expr{Kind: "var", Text: name, Line: line})
			if p.accept(")") {
				break
			}
			p.need(",")
		}
		p.need("=")
		s.Expr = p.expr(0)
		p.need(";")
	case p.peek().kind == "id" && (strings.EqualFold(p.peek().text, "continue") || strings.EqualFold(p.peek().text, "break")):
		s.Kind = strings.ToLower(p.pop().text)
		p.need(";") // Only bare loop transfers; no explicit level.
	case p.accept("echo"):
		s.Kind = "echo"
		s.Expr = p.expr(0)
		p.need(";")
	case p.accept("return"):
		s.Kind = "return"
		if p.peek().kind == "sym" && p.peek().text == ";" {
			s.Expr = ir.Expr{Kind: "null", Text: "null", Line: s.Line}
		} else {
			s.Expr = p.expr(0)
		}
		p.need(";")
	case p.peek().kind == "id" && strings.EqualFold(p.peek().text, "while"):
		p.pop()
		s.Kind = "while"
		p.need("(")
		s.Expr = p.expr(0)
		p.need(")")
		s.Body = p.block()
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
	case p.startsCallStatement():
		s.Kind = "call_stmt"
		s.Expr = p.expr(0)
		p.need(";")
	default:
		if p.peek().kind != "var" {
			p.fail("unsupported statement")
		}
		s.Kind = "assign"
		s.Name = p.name("var")
		if p.accept("->") {
			property := p.name("id")
			s.Kind = "assign_property"
			s.Target = ir.Expr{Kind: "property", Text: property, Line: s.Line,
				Args: []ir.Expr{{Kind: "var", Text: s.Name, Line: s.Line}}}
			s.Name = ""
			nested := false
			for p.accept("->") {
				nested = true
				s.Target = ir.Expr{Kind: "property", Text: p.name("id"), Line: s.Line, Args: []ir.Expr{s.Target}}
			}
			if nested {
				s.Kind = "assign_nested_property"
			}
			if p.accept("[") {
				p.need("]")
				s.Kind = "assign_property_append"
				if nested {
					s.Kind = "assign_nested_property_append"
				}
			}
			if nested && s.Kind != "assign_nested_property_append" && p.accept(".=") {
				s.Kind = "assign_nested_property_concat"
			}
			if s.Kind != "assign_nested_property_concat" && (p.peek().kind != "sym" || p.peek().text != "=") {
				p.fail("property assignment requires one declared property and =")
			}
		} else if p.peek().kind == "sym" && p.peek().text == "[" {
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
		if s.Kind != "assign_nested_property_concat" {
			p.need("=")
		}
		if s.Kind == "assign" && p.experimentalPCRE2 && p.startsPregMatchLocal() {
			s = p.pregMatchLocal(s)
		} else {
			s.Expr = p.expr(0)
		}
		p.need(";")
	}
	// Keep the writable first argument out of the ordinary value-call path.
	// Only standalone calls and a whole RHS assigned to a bare local enter
	// this bounded form; embedded calls remain fail-closed in semantic checks.
	if (s.Kind == "assign" || s.Kind == "call_stmt") && s.Expr.Kind == "call" &&
		strings.EqualFold(s.Expr.Text, "array_push") && len(s.Expr.Args) != 0 {
		s.Kind = "array_push"
		s.Target = s.Expr.Args[0]
		s.Expr.Text = "array_push"
		s.Expr.Args = s.Expr.Args[1:]
	}
	if (s.Kind == "assign" || s.Kind == "call_stmt") && s.Expr.Kind == "call" &&
		strings.EqualFold(s.Expr.Text, "array_pop") && len(s.Expr.Args) == 1 {
		s.Kind = "array_pop"
		s.Target = s.Expr.Args[0]
		s.Expr.Text = "array_pop"
		s.Expr.Args = nil
	}
	return s
}

// preg_match_local has a deliberately fixed source signature:
// preg_match('/.../s', $subject, $out, PREG_OFFSET_CAPTURE, $offset).
// Keeping the writable output outside Expr prevents it being mistaken for an
// ordinary value argument downstream. The profile flag is deliberately not
// represented as generic call metadata: this statement kind is the signature.
func (p *parser) startsPregMatchLocal() bool {
	return p.pos+1 < len(p.ts) && p.peek().kind == "id" &&
		strings.EqualFold(p.peek().text, "preg_match") && p.ts[p.pos+1].text == "("
}

func (p *parser) pregMatchLocal(assign ir.Stmt) ir.Stmt {
	call := p.pop()
	p.need("(")
	pattern := p.preg83Pattern()
	p.need(",")
	subject := p.preg83Subject()
	p.need(",")
	output := p.pop()
	if output.kind != "var" {
		p.fail("preg_match output requires a bare variable")
	}
	p.need(",")
	flag := p.pop()
	if flag.kind != "id" || flag.text != "PREG_OFFSET_CAPTURE" {
		p.fail("preg_match requires PREG_OFFSET_CAPTURE")
	}
	p.need(",")
	offset := p.preg83Offset()
	p.need(")")
	return ir.Stmt{
		Kind:   "preg_match_local",
		Name:   assign.Name,
		Line:   assign.Line,
		Target: ir.Expr{Kind: "var", Text: output.text, Line: output.line},
		Expr: ir.Expr{Kind: "call", Text: "preg_match", Line: call.line,
			Args: []ir.Expr{pattern, subject, offset}},
	}
}

func (p *parser) preg83Pattern() ir.Expr {
	t := p.pop()
	if t.kind != "string" || len(t.text) < 3 || t.text[0] != '/' || !strings.HasSuffix(t.text, "/s") {
		p.fail("preg_match requires a literal /.../s pattern")
	}
	return ir.Expr{Kind: "string", Text: t.text, Line: t.line}
}

func (p *parser) preg83Subject() ir.Expr {
	t := p.pop()
	if t.kind != "var" && t.kind != "string" {
		p.fail("preg_match subject requires a string literal, typed local or named property read")
	}
	return p.preg83PropertyInput(t)
}

func (p *parser) preg83Offset() ir.Expr {
	t := p.pop()
	if t.kind == "sym" && t.text == "-" {
		return p.negativeInteger(t)
	}
	if t.kind != "var" && t.kind != "int" {
		p.fail("preg_match offset requires an integer literal, typed local or named property read")
	}
	return p.preg83PropertyInput(t)
}

// Do not use the general expression parser here: regex inputs admit only
// named reads rooted in a bare local, not calls, indices or other effects.
func (p *parser) preg83PropertyInput(t token) ir.Expr {
	e := ir.Expr{Kind: t.kind, Text: t.text, Line: t.line}
	if t.kind == "var" {
		for p.accept("->") {
			e = ir.Expr{Kind: "property", Text: p.name("id"), Line: t.line, Args: []ir.Expr{e}}
		}
	}
	return e
}

// Keep assignment parsing (including [] append targets) unchanged. Standalone
// calls begin with a direct name, callable variable, or bare-variable method.
// Semantic validation rejects non-call roots and chained receivers.
func (p *parser) startsCallStatement() bool {
	t := p.peek()
	if t.kind != "id" && t.kind != "var" {
		return false
	}
	if p.pos+1 < len(p.ts) && p.ts[p.pos+1].kind == "sym" && p.ts[p.pos+1].text == "(" {
		return true
	}
	return t.kind == "var" && p.pos+3 < len(p.ts) &&
		p.ts[p.pos+1].text == "->" && p.ts[p.pos+2].kind == "id" && p.ts[p.pos+3].text == "("
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

func isRelationalComparison(t token) bool {
	return t.kind == "sym" && (t.text == "<" || t.text == ">")
}

func (p *parser) prec(t token) int {
	if t.kind == "sym" && t.text == "." {
		return 6
	}
	if t.kind == "sym" && (t.text == "+" || t.text == "-") {
		return 7
	}
	if isRelationalComparison(t) {
		return 5
	}
	if isStrictComparison(t) {
		return 4
	}
	if t.kind == "sym" && t.text == "&&" {
		return 3
	}
	if t.kind == "sym" && t.text == "||" {
		return 2
	}
	if t.kind == "sym" && t.text == "??" {
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
		// Constants are case-sensitive and have a separate namespace from
		// functions. Keep the name symbolic until target-aware generation.
		if (t.text == "PHP_MAJOR_VERSION" || t.text == "PHP_MINOR_VERSION") && p.peek().text != "(" {
			e.Kind = "version_constant"
			break
		}
		if strings.EqualFold(t.text, "fn") {
			p.need("(")
			e = ir.Expr{Kind: "arrow", Line: t.line}
			if !p.accept(")") {
				for {
					e.Params = append(e.Params, p.name("var"))
					if p.accept(")") {
						break
					}
					p.need(",")
				}
			}
			p.need("=>")
			e.Args = []ir.Expr{p.expr(0)}
			break
		}
		if strings.EqualFold(t.text, "empty") {
			// This subset tests a defined array/string variable, never a suppressed
			// undefined/index/property read or an arbitrary expression.
			p.need("(")
			base := p.pop()
			if base.kind != "var" {
				p.fail("empty requires a single defined variable")
			}
			p.need(")")
			e = ir.Expr{Kind: "empty_var", Line: t.line, Args: []ir.Expr{
				{Kind: "var", Text: base.text, Line: base.line},
			}}
			break
		}
		if strings.EqualFold(t.text, "isset") {
			p.need("(")
			var operands []ir.Expr
			for {
				base := p.pop()
				if base.kind != "var" {
					p.fail("isset requires a variable or indexed variable")
				}
				address := ir.Expr{Kind: "isset_var", Line: t.line, Args: []ir.Expr{{Kind: "var", Text: base.text, Line: base.line}}}
				if p.accept("->") {
					address.Kind = "isset_property"
					for {
						property := p.peek()
						name := p.name("id")
						address.Args = append(address.Args, ir.Expr{Kind: "string", Text: name, Line: property.line})
						if !p.accept("->") {
							break
						}
					}
				}
				for address.Kind != "isset_property" && p.accept("[") {
					key := p.pop()
					keyExpr := ir.Expr{Kind: key.kind, Text: key.text, Line: key.line}
					if key.kind == "sym" && key.text == "-" {
						keyExpr = p.negativeInteger(key)
					} else if key.kind != "int" && key.kind != "string" && key.kind != "var" {
						p.fail("isset key must be an integer/string literal or variable")
					}
					p.need("]")
					address.Kind = "isset_index"
					address.Args = append(address.Args, keyExpr)
				}
				if address.Kind != "isset_property" && len(address.Args) > 2 {
					address.Kind = "isset_path"
				}
				operands = append(operands, address)
				if p.accept(")") {
					break
				}
				p.need(",")
			}
			e = operands[0]
			if len(operands) > 1 {
				e = ir.Expr{Kind: "isset_multi", Line: t.line, Args: operands}
			}
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
			e.Args = p.args(")")
		default:
			e.Kind = "call"
			p.need("(")
			e.Args = p.args(")")
		}
	case "sym":
		switch t.text {
		case "!":
			e = ir.Expr{Kind: "not", Args: []ir.Expr{p.expr(8)}, Line: t.line}
		case "-":
			e = p.negativeInteger(t)
		case "(":
			// The supported cast spelling is deliberately narrow. Looking ahead
			// keeps ordinary parenthesized expressions intact.
			if p.peek().kind == "id" && (strings.EqualFold(p.peek().text, "bool") || strings.EqualFold(p.peek().text, "array")) && p.pos+1 < len(p.ts) && p.ts[p.pos+1].text == ")" {
				kind := "bool_cast"
				if strings.EqualFold(p.peek().text, "array") {
					kind = "object_array_cast"
				}
				p.pop()
				p.need(")")
				e = ir.Expr{Kind: kind, Args: []ir.Expr{p.expr(8)}, Line: t.line}
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
		if p.accept("(") {
			if e.Kind != "var" {
				p.fail("dynamic invocation requires a callable variable")
			}
			e = ir.Expr{Kind: "invoke", Line: t.line, Args: append([]ir.Expr{e}, p.args(")")...)}
			continue
		}
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
		rightMin := q + 1
		if op.text == "??" {
			rightMin = q // Right-associative, above full ternary and below ||.
		}
		right := p.expr(rightMin)
		if op.text == "??" {
			e = ir.Expr{Kind: "coalesce", Args: []ir.Expr{e, right}, Line: op.line}
		} else {
			e = ir.Expr{Kind: "binary", Text: op.text, Args: []ir.Expr{e, right}, Line: op.line}
		}
		if isStrictComparison(op) && isStrictComparison(p.peek()) {
			p.fail("strict comparison is non-associative; use parentheses")
		}
		if isRelationalComparison(op) && isRelationalComparison(p.peek()) {
			p.fail("relational comparison is non-associative; use parentheses")
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
