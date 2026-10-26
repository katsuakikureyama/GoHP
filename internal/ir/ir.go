// Package ir defines the versioned, frontend-independent GoHP subset IR.
package ir

const Version = "0.1"

type Expr struct {
	// version_constant retains the exact predefined name in Text, no Args or
	// Params; its integer value is resolved against Program.PHPVersion.
	Kind, Text string
	Line       int
	Args       []Expr
	Params     []string // Only arrow expressions: fixed, untyped, capture-free parameters.
}
type Stmt struct {
	// continue and break target the nearest foreach/while and have no payload.
	// while uses Expr (condition) and Body only.
	// call_stmt uses Expr only: direct call or bare-variable method/invoke.
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
	Functions           []Function
	Classes             []Class
	Body                []Stmt
}
