package frontend

import (
	"crypto/sha256"
	"fmt"
	"path"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/katsuakikureyama/GoHP/internal/ir"
	"github.com/katsuakikureyama/GoHP/internal/phpversion"
)

// ModuleDeclaration identifies a complete, syntax-checked declaration. Its
// line belongs to ModuleInspection.SourceID, not a synthetic combined source.
type ModuleDeclaration struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
	Line int    `json:"line"`
}

type ModuleRequire struct {
	Line   int    `json:"line"`
	Suffix string `json:"suffix"`
}

// ModuleInspection deliberately exposes no ir.Program: load directives cannot
// disappear into a seemingly executable program through inspection alone.
// Method/function bodies are parsed, but semantic admission is not established.
type ModuleInspection struct {
	SourceID     string              `json:"source_id"`
	PHPVersion   string              `json:"php_target"`
	RegexProfile string              `json:"regex_profile,omitempty"`
	Declarations []ModuleDeclaration `json:"declarations"`
	Requires     []ModuleRequire     `json:"requires"`
}

// InspectModule parses one declaration-only module with trailing static
// require_once directives. It never reads files, executes PHP or returns IR.
// Ordinary Parse/ParseWithOptions retain their existing single-file grammar.
func InspectModule(sourceID, src, version string, options Options) (out ModuleInspection, err error) {
	out, _, err = parseStaticModule(sourceID, src, version, options, false)
	return out, err
}

// StaticRequireOnce is an unresolved frontend-only statement. Semantic/codegen
// admission deliberately rejects it. Only the checked closure resolver may
// replace these top-level edges after verifying the complete source graph.
const StaticRequireOnce = "static_require_once"

type ParsedModule struct {
	Inspection ModuleInspection
	Program    ir.Program
}

// ParseStaticModule retains original-file IR and explicit unresolved load nodes.
// entry permits a driver body after its leading load section; dependencies are
// declaration-only. This filesystem-free adapter alone is not a loaded program.
func ParseStaticModule(sourceID, src, version string, options Options, entry bool) (ParsedModule, error) {
	inspection, program, err := parseStaticModule(sourceID, src, version, options, entry)
	if err != nil {
		return ParsedModule{}, err
	}
	return ParsedModule{Inspection: inspection, Program: program}, nil
}

func parseStaticModule(sourceID, src, version string, options Options, entry bool) (out ModuleInspection, program ir.Program, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			if e, ok := recovered.(error); ok {
				err = e
			} else {
				panic(recovered)
			}
		}
		if err != nil {
			out = ModuleInspection{}
			program = ir.Program{}
			err = ir.WithSource(err, sourceID)
		}
	}()
	if !modulePath(sourceID) {
		return out, program, fmt.Errorf("PHP003: module source ID must be a canonical relative slash path")
	}
	if err := phpversion.ValidateTarget(version); err != nil {
		return out, program, fmt.Errorf("PHP003: %w", err)
	}
	profile, err := options.RegexProfile(version)
	if err != nil {
		return out, program, err
	}
	ts, err := lex(src)
	if err != nil {
		return out, program, err
	}
	p := &parser{ts: ts, version: version, experimentalPCRE2: options.ExperimentalPCRE2}
	out = ModuleInspection{SourceID: sourceID, PHPVersion: version,
		Declarations: []ModuleDeclaration{}, Requires: []ModuleRequire{}}
	out.RegexProfile = profile
	program = ir.Program{Version: ir.Version, PHPVersion: version, RegexProfile: out.RegexProfile,
		Sources: []ir.SourceFile{{ID: sourceID, Bytes: int64(len(src)), SHA256: fmt.Sprintf("%x", sha256.Sum256([]byte(src)))}}}
	loadsStarted := false
	bodyStarted := false
	for p.peek().kind != "eof" {
		line := p.peek().line
		switch {
		case p.accept("require_once"):
			if bodyStarted {
				p.fail("static require_once must precede executable entry statements")
			}
			loadsStarted = true
			p.need("__DIR__")
			p.need(".")
			suffix := p.pop()
			if suffix.kind != "string" || !strings.HasPrefix(suffix.text, "/") || !modulePath(strings.TrimPrefix(suffix.text, "/")) {
				p.fail("require_once needs one canonical /literal/path suffix")
			}
			p.need(";")
			out.Requires = append(out.Requires, ModuleRequire{Line: line, Suffix: suffix.text})
			program.Body = append(program.Body, ir.Stmt{Kind: StaticRequireOnce, Name: suffix.text, Line: line, SourceID: sourceID})
		case p.accept("function"):
			if loadsStarted || bodyStarted {
				p.fail("module declarations must precede trailing require_once directives")
			}
			f := p.function(false)
			if err := rejectLocalModuleLoads(f.Body); err != nil {
				return out, program, err
			}
			f.SourceID = sourceID
			program.Functions = append(program.Functions, f)
			out.Declarations = append(out.Declarations, ModuleDeclaration{Kind: "function", Name: f.Name, Line: line})
		case p.accept("class"):
			if loadsStarted || bodyStarted {
				p.fail("module declarations must precede trailing require_once directives")
			}
			c := p.class()
			for _, method := range c.Methods {
				if err := rejectLocalModuleLoads(method.Body); err != nil {
					return out, program, err
				}
			}
			c.SourceID = sourceID
			program.Classes = append(program.Classes, c)
			out.Declarations = append(out.Declarations, ModuleDeclaration{Kind: "class", Name: c.Name, Line: line})
		default:
			if !entry {
				p.fail("static modules admit only declarations and trailing require_once directives")
			}
			bodyStarted = true
			statement := p.stmt()
			if err := rejectLocalModuleLoads([]ir.Stmt{statement}); err != nil {
				return out, program, err
			}
			statement.SourceID = sourceID
			program.Body = append(program.Body, statement)
		}
	}
	return out, program, nil
}

// The existing expression parser can represent parenthesized include/require
// syntax as a direct call. Graph inspection has no semantic pass, so it must
// reject those constructs explicitly rather than miss a function-local edge.
func rejectLocalModuleLoads(body []ir.Stmt) error {
	var expression func(ir.Expr) error
	expression = func(e ir.Expr) error {
		if e.Kind == "call" {
			switch strings.ToLower(e.Text) {
			case "require", "require_once", "include", "include_once":
				return fmt.Errorf("PHP002 line %d: local source loading is outside the static module profile", e.Line)
			}
		}
		for _, argument := range e.Args {
			if err := expression(argument); err != nil {
				return err
			}
		}
		return nil
	}
	for _, statement := range body {
		if err := expression(statement.Expr); err != nil {
			return err
		}
		if err := expression(statement.Target); err != nil {
			return err
		}
		if err := rejectLocalModuleLoads(statement.Body); err != nil {
			return err
		}
		if err := rejectLocalModuleLoads(statement.ElseBody); err != nil {
			return err
		}
	}
	return nil
}

// This deliberately small identity grammar has no host/CWD/volume semantics.
// It excludes aliases rather than normalizing them into a different identity.
func modulePath(value string) bool {
	if value == "" || !utf8.ValidString(value) || strings.HasPrefix(value, "/") || strings.ContainsAny(value, "\\:") || path.Clean(value) != value {
		return false
	}
	for _, r := range value {
		if r == 0 || unicode.IsControl(r) {
			return false
		}
	}
	for _, part := range strings.Split(value, "/") {
		if part == "" || part == "." || part == ".." {
			return false
		}
	}
	return true
}
