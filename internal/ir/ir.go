// Package ir defines the versioned, frontend-independent GoHP subset IR.
package ir

const Version = "0.1"

// Preg83Profile identifies the deliberately narrow experimental PCRE2 source
// profile. It is a program-level admission choice, not an engine-selection
// fallback or a claim that arbitrary PHP PCRE calls are supported.
const Preg83Profile = "pcre2-php83-v1"

type Expr struct {
	// version_constant retains the exact predefined name in Text, no Args or
	// Params; its integer value is resolved against Program.PHPVersion.
	Kind, Text string
	Line       int
	Args       []Expr
	Params     []string // Only arrow expressions: fixed, untyped, capture-free parameters.
}
type Stmt struct {
	// Bare break targets the nearest loop/switch; continue requires a loop.
	// Both transfers have no payload or explicit level.
	// while uses Expr (condition) and Body only.
	// switch uses Expr (selector) and Body of case/default clauses. A case has
	// one literal string Expr; default has none. Both have statement Body only.
	// call_stmt uses Expr only: direct call or bare-variable method/invoke.
	// assign_nested_property[_append/_concat] uses Target=2+ named property
	// nodes rooted in a bare local, Expr=RHS. This is not a PHP reference cell.
	// list_assign uses Expr (RHS) and Target=list with nonempty bare-var Args.
	// Targets are positional locals, not general addresses or references.
	// preg_match_local: Name is the return local, Target the bare output local,
	// Expr a canonical preg_match call with pattern/subject/offset only. This
	// statement encodes fixed PREG_OFFSET_CAPTURE; it is not a reference API.
	Kind, Name string
	Line       int
	Expr       Expr
	Target     Expr // assign_index: index(variable,key); assign_append: variable; assign_property/assign_property_append: property(variable)
	Body       []Stmt
	ElseBody   []Stmt // if only; nil means absent, non-nil may be empty
	Key, Value string
}
type Function struct {
	Name               string
	Params             []string
	ParamTypes         []string // Aligned with Params: global array/callable/int/string/bool; constructor int/string/bool; "" untyped.
	OptionalNullParams int      // Number of trailing, untyped parameters defaulting to null.
	Body               []Stmt
	Line               int
	ReturnType         string // Empty (untyped), "array" or "bool" on a named global function.
}
type Property struct {
	Name  string
	Value Expr
}
type Class struct {
	Name       string
	Properties []Property
	Methods    []Function
}
type Program struct {
	Version, PHPVersion string
	// RegexProfile is empty unless the frontend was explicitly asked to admit
	// the experimental PHP 8.3 local-output preg_match form.
	RegexProfile string
	Functions    []Function
	Classes      []Class
	Body         []Stmt
}
