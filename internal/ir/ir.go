// Package ir defines the versioned, frontend-independent GoHP subset IR.
package ir

const Version = "0.1"

// Preg83Profile identifies the deliberately narrow experimental PCRE2 source
// profile. It is a program-level admission choice, not an engine-selection
// fallback or a claim that arbitrary PHP PCRE calls are supported.
const Preg83Profile = "pcre2-php83-v1"

type Expr struct {
	// empty_var: one defined bare var, no Text/Params; actual array or string.
	// Legacy empty_array remains array-only; generic empty calls are invalid.
	// isset_property: Args are a defined bare root variable followed by one or
	// more literal identifier strings, not ordinary eager property-read nodes.
	// version_constant retains the exact predefined name in Text, no Args or
	// Params; its integer value is resolved against Program.PHPVersion.
	Kind, Text string
	Line       int
	Args       []Expr
	Params     []string // Only arrow expressions: fixed, untyped, capture-free parameters.
}
type Stmt struct {
	SourceID string // Optional nested inheritance; required on located top-level statements.
	// Bare break targets the nearest loop/switch; continue requires a loop.
	// Both transfers have no payload or explicit level.
	// while uses Expr (condition) and Body only.
	// switch uses Expr (selector) and Body of case/default clauses. A case has
	// one literal string Expr; default has none. Both have statement Body only.
	// call_stmt uses Expr only: direct call or bare-variable method/invoke.
	// array_push: Target is an existing local or direct property(variable).
	// Expr is canonical call/array_push with only the value arguments (possibly
	// none); Name is an optional bare result local. Not a general reference API
	// or value expression; the target address precedes payload evaluation.
	// array_pop: same bounded Target/optional Name, with canonical call/array_pop
	// Expr and nil Args/Params. Mixed element/null result is assigned last.
	// assign_nested_property[_append/_concat] uses Target=2+ named property
	// nodes rooted in a bare local, Expr=RHS. This is not a PHP reference cell.
	// list_assign uses Expr (RHS) and Target=list with nonempty bare-var Args.
	// Targets are positional locals, not general addresses or references.
	// preg_match_local: Name is the return local, Target the bare output local,
	// Expr a canonical preg_match call with pattern/subject/offset only. This
	// statement encodes fixed PREG_OFFSET_CAPTURE; it is not a reference API.
	// Subject/offset may be literal, typed bare local or named property chain
	// rooted in a bare local. Property input types are checked at runtime.
	Kind, Name string
	Line       int
	Expr       Expr
	Target     Expr // assign_index: index(variable,key); assign_append: variable; assign_property/assign_property_append: property(variable)
	Body       []Stmt
	ElseBody   []Stmt // if only; nil means absent, non-nil may be empty
	Key, Value string
}
type Function struct {
	SourceID           string // Original source identity; methods may inherit their class identity.
	Name               string
	Params             []string
	ParamTypes         []string // Aligned: global array/callable/int/string/bool; constructor scalars; class:lowername in any function/method; "" untyped.
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
	SourceID   string
	Line       int
	Name       string
	Properties []Property
	Methods    []Function
}
type Program struct {
	Sources             []SourceFile // Empty for legacy line-only IR; metadata is not checked-buffer authority.
	Version, PHPVersion string
	// RegexProfile is empty unless the frontend was explicitly asked to admit
	// the experimental PHP 8.3 local-output preg_match form.
	RegexProfile string
	Functions    []Function
	Classes      []Class
	Body         []Stmt
}
