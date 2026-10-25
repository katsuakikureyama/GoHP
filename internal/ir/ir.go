// Package ir defines the versioned, frontend-independent GoHP subset IR.
package ir

const Version = "0.1"

type Expr struct {
	Kind, Text string
	Line       int
	Args       []Expr
}
type Stmt struct {
	Kind, Name string
	Line       int
	Expr       Expr
	Target     Expr // assign_index only: index(variable, literal-or-variable key)
	Body       []Stmt
	Key, Value string
}
type Function struct {
	Name   string
	Params []string
	Body   []Stmt
	Line   int
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
