// Package semantic validates the closed GoHP subset before code generation.
package semantic

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/katsuakikureyama/GoHP/internal/ir"
	"github.com/katsuakikureyama/GoHP/internal/phpversion"
	"github.com/katsuakikureyama/GoHP/internal/sprintfsubset"
)

type kind uint8

const (
	unknownKind kind = iota
	intKind
	stringKind
	boolKind
	nullKind
	arrayKind
	objectKind
	callableKind
)

type valueType struct {
	kind  kind
	class string
}

type classInfo struct {
	properties map[string]valueType
	methods    map[string]int
}

type validator struct {
	classes    map[string]classInfo
	functions  map[string]int
	parameters map[string][]string
}

// Validate rejects malformed IR and operations that are statically outside the
// supported PHP subset. Unknown dynamic values remain guarded by the runtime.
func Validate(p ir.Program) error {
	if err := phpversion.ValidateTarget(p.PHPVersion); err != nil {
		return fmt.Errorf("PHP003: %w", err)
	}
	v := validator{classes: make(map[string]classInfo, len(p.Classes)), functions: make(map[string]int, len(p.Functions)), parameters: make(map[string][]string, len(p.Functions))}
	// Collect before validating bodies so literal callbacks can refer forward
	// or back to their own declaration without dynamic callable resolution.
	for _, f := range p.Functions {
		if err := validateParameterTypes(f, false); err != nil {
			return err
		}
		v.functions[strings.ToLower(f.Name)] = len(f.Params)
		v.parameters[strings.ToLower(f.Name)] = f.ParamTypes
	}
	for _, c := range p.Classes {
		if !validName(c.Name) {
			return malformed(1, "invalid class name")
		}
		info := classInfo{properties: map[string]valueType{}, methods: map[string]int{}}
		for _, property := range c.Properties {
			if !validName(property.Name) {
				return malformed(property.Value.Line, "invalid property name")
			}
			info.properties[property.Name] = literalType(property.Value)
		}
		for _, method := range c.Methods {
			if err := validateParameterTypes(method, true); err != nil {
				return err
			}
			if !validName(method.Name) {
				return malformed(method.Line, "invalid method name")
			}
			info.methods[strings.ToLower(method.Name)] = len(method.Params)
		}
		v.classes[strings.ToLower(c.Name)] = info
	}
	for _, f := range p.Functions {
		if err := v.function(f, ""); err != nil {
			return err
		}
	}
	for _, c := range p.Classes {
		for _, property := range c.Properties {
			if property.Value.Kind != "int" && property.Value.Kind != "string" && property.Value.Kind != "bool" && property.Value.Kind != "null" {
				return unsupported(property.Value.Line, "property defaults must be scalar literals")
			}
			t, err := v.expr(property.Value, map[string]valueType{})
			if err != nil {
				return err
			}
			if !isScalar(t) {
				return unsupported(property.Value.Line, "property defaults must be scalar literals")
			}
		}
		for _, method := range c.Methods {
			if err := v.function(method, strings.ToLower(c.Name)); err != nil {
				return err
			}
		}
	}
	return v.stmts(p.Body, map[string]valueType{}, "", false)
}

func validateParameterTypes(f ir.Function, method bool) error {
	if len(f.ParamTypes) != 0 && len(f.ParamTypes) != len(f.Params) {
		return malformed(f.Line, "parameter types must align with parameter names")
	}
	for _, annotation := range f.ParamTypes {
		if annotation != "" && annotation != "array" && annotation != "callable" {
			return malformed(f.Line, "unsupported IR parameter type")
		}
		if method && annotation != "" {
			return unsupported(f.Line, "typed method parameters are unsupported")
		}
	}
	return nil
}

func (v validator) function(f ir.Function, class string) error {
	if strings.EqualFold(f.Name, "fn") {
		return unsupported(f.Line, "fn is reserved for arrow expressions")
	}
	if f.ReturnType != "" && f.ReturnType != "array" && f.ReturnType != "bool" {
		return malformed(f.Line, "unsupported IR return type")
	}
	if class != "" && f.ReturnType != "" {
		return unsupported(f.Line, "typed method returns are unsupported")
	}
	if !validName(f.Name) {
		return malformed(f.Line, "invalid function name")
	}
	scope := map[string]valueType{}
	if class != "" {
		scope["this"] = valueType{kind: objectKind, class: class}
	}
	for i, param := range f.Params {
		if !validName(param) {
			return malformed(f.Line, "invalid parameter name")
		}
		scope[param] = valueType{kind: unknownKind}
		if len(f.ParamTypes) != 0 && f.ParamTypes[i] == "array" {
			scope[param] = valueType{kind: arrayKind}
		}
		if len(f.ParamTypes) != 0 && f.ParamTypes[i] == "callable" {
			scope[param] = valueType{kind: callableKind}
		}
	}
	return v.stmts(f.Body, scope, f.ReturnType, false)
}

func (v validator) stmts(stmts []ir.Stmt, scope map[string]valueType, returnType string, inForeach bool) error {
	for _, stmt := range stmts {
		if stmt.Kind != "if" && stmt.ElseBody != nil {
			return malformed(stmt.Line, "else body requires an if statement")
		}
		switch stmt.Kind {
		case "continue":
			if stmt.Name != "" || stmt.Key != "" || stmt.Value != "" || stmt.Body != nil || !emptyExpr(stmt.Expr) || !emptyExpr(stmt.Target) {
				return malformed(stmt.Line, "continue must have no payload")
			}
			if !inForeach {
				return unsupported(stmt.Line, "continue requires an enclosing foreach")
			}
			// Still validate dead statements, without letting dead assignments
			// overwrite the type state that reaches the next loop iteration.
			scope = cloneScope(scope)
		case "assign":
			if !validName(stmt.Name) {
				return malformed(stmt.Line, "invalid assignment target")
			}
			t, err := v.expr(stmt.Expr, scope)
			if err != nil {
				return err
			}
			scope[stmt.Name] = t
		case "assign_index":
			target := stmt.Target
			if target.Kind != "index" || len(target.Args) != 2 || target.Args[0].Kind != "var" || stmt.Name != "" {
				return malformed(stmt.Line, "array assignment requires an indexed variable target")
			}
			base, key := target.Args[0], target.Args[1]
			if key.Kind != "int" && key.Kind != "string" && key.Kind != "var" {
				return unsupported(stmt.Line, "array assignment key must be an integer/string literal or variable")
			}
			if _, ok := scope[base.Text]; !ok {
				return unsupported(stmt.Line, "array assignment requires an existing variable")
			}
			if _, err := v.expr(target, scope); err != nil {
				return err
			}
			if _, err := v.expr(stmt.Expr, scope); err != nil {
				return err
			}
			scope[base.Text] = valueType{kind: arrayKind}
		case "assign_append":
			base := stmt.Target
			if base.Kind != "var" || len(base.Args) != 0 || !validName(base.Text) || stmt.Name != "" {
				return malformed(stmt.Line, "array append requires a variable target")
			}
			if _, ok := scope[base.Text]; !ok {
				return unsupported(stmt.Line, "array append requires an existing variable")
			}
			t, err := v.expr(base, scope)
			if err != nil {
				return err
			}
			if t.kind != arrayKind && t.kind != unknownKind {
				return unsupported(stmt.Line, "array append requires an array")
			}
			if _, err := v.expr(stmt.Expr, scope); err != nil {
				return err
			}
			scope[base.Text] = valueType{kind: arrayKind}
		case "echo":
			t, err := v.expr(stmt.Expr, scope)
			if err != nil {
				return err
			}
			if !isScalarOrUnknown(t) {
				return unsupported(stmt.Line, "echo requires a supported scalar")
			}
		case "return":
			t, err := v.expr(stmt.Expr, scope)
			if err != nil {
				return err
			}
			if returnType == "array" && t.kind != arrayKind && t.kind != unknownKind {
				return unsupported(stmt.Line, "array return requires an array value")
			}
			if returnType == "bool" && t.kind != boolKind && t.kind != unknownKind {
				return unsupported(stmt.Line, "bool return requires a bool value")
			}
		case "foreach":
			if !validName(stmt.Value) || stmt.Key != "" && !validName(stmt.Key) {
				return malformed(stmt.Line, "invalid foreach binding")
			}
			t, err := v.expr(stmt.Expr, scope)
			if err != nil {
				return err
			}
			if t.kind != arrayKind && t.kind != unknownKind {
				return unsupported(stmt.Line, "foreach requires an array")
			}
			inner := cloneScope(scope)
			inner[stmt.Value] = valueType{kind: unknownKind}
			if stmt.Key != "" {
				inner[stmt.Key] = valueType{kind: unknownKind}
			}
			if err := v.stmts(stmt.Body, inner, returnType, true); err != nil {
				return err
			}
			// A foreach body may not execute. Existing variables changed inside it
			// therefore become unknown after the loop instead of acquiring a false
			// definite type.
			for name, before := range scope {
				if after := inner[name]; after != before {
					scope[name] = valueType{kind: unknownKind}
				}
			}
		case "if":
			condition, err := v.expr(stmt.Expr, scope)
			if err != nil {
				return err
			}
			if !isBoolDomain(condition) {
				return unsupported(stmt.Line, "if condition requires a supported value")
			}
			inner := cloneScope(scope)
			if err := v.stmts(stmt.Body, inner, returnType, inForeach); err != nil {
				return err
			}
			other := cloneScope(scope)
			if err := v.stmts(stmt.ElseBody, other, returnType, inForeach); err != nil {
				return err
			}
			// Validate both arms from the pre-if scope. Branch-only bindings do
			// not escape, even if both arms define them. A changed outer type
			// stays unknown; no path-sensitive or return-sensitive inference.
			for name, before := range scope {
				if inner[name] != before || other[name] != before {
					scope[name] = valueType{kind: unknownKind}
				}
			}
		default:
			return malformed(stmt.Line, "unsupported IR statement "+stmt.Kind)
		}
	}
	return nil
}

func emptyExpr(e ir.Expr) bool {
	return e.Kind == "" && e.Text == "" && e.Line == 0 && e.Args == nil && e.Params == nil
}

func (v validator) expr(expr ir.Expr, scope map[string]valueType) (valueType, error) {
	if expr.Kind != "arrow" && expr.Params != nil {
		return valueType{}, malformed(expr.Line, "only arrow expressions may declare parameters")
	}
	require := func(n int) error {
		if len(expr.Args) != n {
			return malformed(expr.Line, fmt.Sprintf("%s expression requires %d operands", expr.Kind, n))
		}
		return nil
	}
	switch expr.Kind {
	case "arrow":
		if err := require(1); err != nil {
			return valueType{}, err
		}
		if expr.Text != "" {
			return valueType{}, malformed(expr.Line, "arrow expression must not have a name or annotation")
		}
		inner := map[string]valueType{}
		for _, param := range expr.Params {
			if !validName(param) || param == "this" || param == "GLOBALS" || strings.HasPrefix(param, "_") {
				return valueType{}, unsupported(expr.Line, "invalid or reserved arrow parameter")
			}
			if _, duplicate := inner[param]; duplicate {
				return valueType{}, unsupported(expr.Line, "duplicate arrow parameter")
			}
			inner[param] = valueType{kind: unknownKind}
		}
		// Do not import the outer scope. Codegen rejects any free variable,
		// including one in a nested arrow: captures are outside this contract.
		if _, err := v.expr(expr.Args[0], inner); err != nil {
			return valueType{}, err
		}
		return valueType{kind: callableKind}, nil
	case "invoke":
		if expr.Text != "" || len(expr.Args) == 0 || expr.Args[0].Kind != "var" {
			return valueType{}, malformed(expr.Line, "invocation requires a callable variable receiver")
		}
		args, err := v.exprs(expr.Args, scope)
		if err != nil {
			return valueType{}, err
		}
		if args[0].kind != callableKind && args[0].kind != unknownKind {
			return valueType{}, unsupported(expr.Line, "invocation requires a generated arrow callable")
		}
		return valueType{kind: unknownKind}, nil
	case "int":
		if err := require(0); err != nil {
			return valueType{}, err
		}
		if _, err := strconv.ParseInt(expr.Text, 10, 64); err != nil {
			return valueType{}, malformed(expr.Line, "invalid integer literal")
		}
		return valueType{kind: intKind}, nil
	case "string":
		if err := require(0); err != nil {
			return valueType{}, err
		}
		return valueType{kind: stringKind}, nil
	case "bool":
		if err := require(0); err != nil {
			return valueType{}, err
		}
		if expr.Text != "true" && expr.Text != "false" {
			return valueType{}, malformed(expr.Line, "invalid boolean literal")
		}
		return valueType{kind: boolKind}, nil
	case "null":
		if err := require(0); err != nil {
			return valueType{}, err
		}
		if expr.Text != "null" {
			return valueType{}, malformed(expr.Line, "invalid null literal")
		}
		return valueType{kind: nullKind}, nil
	case "var":
		if err := require(0); err != nil {
			return valueType{}, err
		}
		if !validName(expr.Text) {
			return valueType{}, malformed(expr.Line, "invalid variable name")
		}
		if t, ok := scope[expr.Text]; ok {
			return t, nil
		}
		// Code generation owns the established undefined-variable diagnostic.
		return valueType{kind: unknownKind}, nil
	case "ternary":
		if err := require(3); err != nil {
			return valueType{}, err
		}
		// Validate even the unselected arm: laziness is a runtime property,
		// not permission to accept source outside the closed subset.
		operands, err := v.exprs(expr.Args, scope)
		if err != nil {
			return valueType{}, err
		}
		if !isBoolDomain(operands[0]) {
			return valueType{}, unsupported(expr.Line, "ternary condition requires a supported value")
		}
		if operands[1] == operands[2] {
			return operands[1], nil
		}
		return valueType{kind: unknownKind}, nil
	case "binary":
		if err := require(2); err != nil {
			return valueType{}, err
		}
		left, err := v.expr(expr.Args[0], scope)
		if err != nil {
			return valueType{}, err
		}
		right, err := v.expr(expr.Args[1], scope)
		if err != nil {
			return valueType{}, err
		}
		switch expr.Text {
		case "+":
			if !isNumericOrUnknown(left) || !isNumericOrUnknown(right) {
				return valueType{}, unsupported(expr.Line, "addition requires supported numeric scalars")
			}
			return valueType{kind: intKind}, nil
		case ".":
			if !isScalarOrUnknown(left) || !isScalarOrUnknown(right) {
				return valueType{}, unsupported(expr.Line, "concatenation requires supported scalars")
			}
			return valueType{kind: stringKind}, nil
		case "===", "!==":
			if !isScalarOrUnknown(left) || !isScalarOrUnknown(right) {
				return valueType{}, unsupported(expr.Line, "strict comparison requires supported scalars")
			}
			return valueType{kind: boolKind}, nil
		case "&&", "||":
			if !isBoolDomain(left) || !isBoolDomain(right) {
				return valueType{}, unsupported(expr.Line, "logical "+expr.Text+" requires supported values")
			}
			return valueType{kind: boolKind}, nil
		default:
			return valueType{}, malformed(expr.Line, "unsupported IR operator "+expr.Text)
		}
	case "array":
		for _, pair := range expr.Args {
			if pair.Kind != "pair" || len(pair.Args) != 2 || pair.Params != nil {
				return valueType{}, malformed(pair.Line, "array entries must be key/value pairs")
			}
			key := pair.Args[0]
			if key.Kind == "auto" {
				if len(key.Args) != 0 || key.Params != nil {
					return valueType{}, malformed(key.Line, "automatic array key has operands")
				}
			} else {
				keyType, err := v.expr(key, scope)
				if err != nil {
					return valueType{}, err
				}
				if keyType.kind != intKind && keyType.kind != stringKind {
					return valueType{}, unsupported(key.Line, "array keys must be integer/string literals")
				}
			}
			if _, err := v.expr(pair.Args[1], scope); err != nil {
				return valueType{}, err
			}
		}
		return valueType{kind: arrayKind}, nil
	case "bool_cast", "not":
		if err := require(1); err != nil {
			return valueType{}, err
		}
		operand, err := v.expr(expr.Args[0], scope)
		if err != nil {
			return valueType{}, err
		}
		if !isBoolDomain(operand) {
			if expr.Kind == "not" {
				return valueType{}, unsupported(expr.Line, "logical not requires a supported value")
			}
			return valueType{}, unsupported(expr.Line, "boolean cast requires a supported value")
		}
		return valueType{kind: boolKind}, nil
	case "empty_array":
		if err := require(1); err != nil {
			return valueType{}, err
		}
		base := expr.Args[0]
		if expr.Text != "" || base.Kind != "var" {
			return valueType{}, malformed(expr.Line, "empty requires a single array variable")
		}
		if _, ok := scope[base.Text]; !ok {
			return valueType{}, unsupported(expr.Line, "empty requires an existing array variable")
		}
		baseType, err := v.expr(base, scope)
		if err != nil {
			return valueType{}, err
		}
		if baseType.kind != arrayKind && baseType.kind != unknownKind {
			return valueType{}, unsupported(expr.Line, "empty requires an array")
		}
		return valueType{kind: boolKind}, nil
	case "isset_index":
		if err := require(2); err != nil {
			return valueType{}, err
		}
		base, key := expr.Args[0], expr.Args[1]
		if expr.Text != "" || base.Kind != "var" {
			return valueType{}, malformed(expr.Line, "isset requires an indexed variable")
		}
		if key.Kind != "int" && key.Kind != "string" && key.Kind != "var" {
			return valueType{}, unsupported(expr.Line, "isset key must be an integer/string literal or variable")
		}
		if _, ok := scope[base.Text]; !ok {
			return valueType{}, unsupported(expr.Line, "isset requires an existing array variable")
		}
		if key.Kind == "var" {
			if _, ok := scope[key.Text]; !ok {
				return valueType{}, unsupported(expr.Line, "isset requires an existing key variable")
			}
		}
		baseType, err := v.expr(base, scope)
		if err != nil {
			return valueType{}, err
		}
		keyType, err := v.expr(key, scope)
		if err != nil {
			return valueType{}, err
		}
		if baseType.kind != arrayKind && baseType.kind != unknownKind {
			return valueType{}, unsupported(expr.Line, "isset requires an array")
		}
		if keyType.kind != intKind && keyType.kind != stringKind && keyType.kind != unknownKind {
			return valueType{}, unsupported(expr.Line, "isset key must be an integer or string")
		}
		return valueType{kind: boolKind}, nil
	case "index":
		if err := require(2); err != nil {
			return valueType{}, err
		}
		base, err := v.expr(expr.Args[0], scope)
		if err != nil {
			return valueType{}, err
		}
		key, err := v.expr(expr.Args[1], scope)
		if err != nil {
			return valueType{}, err
		}
		if base.kind != arrayKind && base.kind != unknownKind {
			return valueType{}, unsupported(expr.Line, "indexing requires an array")
		}
		if key.kind != intKind && key.kind != stringKind && key.kind != unknownKind {
			return valueType{}, unsupported(expr.Line, "array index must be an integer or string")
		}
		return valueType{kind: unknownKind}, nil
	case "call":
		if strings.EqualFold(expr.Text, "empty") {
			return valueType{}, malformed(expr.Line, "empty must use the empty_array language construct")
		}
		if strings.EqualFold(expr.Text, "isset") {
			return valueType{}, malformed(expr.Line, "isset must use the isset_index language construct")
		}
		if !validName(expr.Text) {
			return valueType{}, malformed(expr.Line, "invalid function name")
		}
		args, err := v.exprs(expr.Args, scope)
		if err != nil {
			return valueType{}, err
		}
		if strings.EqualFold(expr.Text, "in_array") {
			if len(args) != 3 || expr.Args[2].Kind != "bool" || expr.Args[2].Text != "true" {
				return valueType{}, unsupported(expr.Line, "in_array expects three arguments with literal true strictness")
			}
			if !isScalarOrUnknown(args[0]) {
				return valueType{}, unsupported(expr.Line, "in_array requires a scalar needle")
			}
			if args[1].kind != arrayKind && args[1].kind != unknownKind {
				return valueType{}, unsupported(expr.Line, "in_array requires an array haystack")
			}
			if expr.Args[1].Kind == "array" {
				for _, pair := range expr.Args[1].Args {
					value, err := v.expr(pair.Args[1], scope)
					if err != nil {
						return valueType{}, err
					}
					if !isScalarOrUnknown(value) {
						return valueType{}, unsupported(expr.Line, "in_array requires scalar haystack values")
					}
				}
			}
			return valueType{kind: boolKind}, nil
		}
		if strings.EqualFold(expr.Text, "array_map") {
			if len(expr.Args) != 2 || expr.Args[0].Kind != "string" || !validName(expr.Args[0].Text) {
				return valueType{}, unsupported(expr.Line, "array_map expects a literal named callback and one array")
			}
			if arity, ok := v.functions[strings.ToLower(expr.Args[0].Text)]; !ok || arity != 1 {
				return valueType{}, unsupported(expr.Line, "array_map callback must be a declared one-argument function")
			}
			if args[1].kind != arrayKind && args[1].kind != unknownKind {
				return valueType{}, unsupported(expr.Line, "array_map requires an array")
			}
			return valueType{kind: arrayKind}, nil
		}
		if strings.EqualFold(expr.Text, "addslashes") {
			if len(args) != 1 {
				return valueType{}, unsupported(expr.Line, "addslashes expects one argument")
			}
			if args[0].kind != stringKind && args[0].kind != unknownKind {
				return valueType{}, unsupported(expr.Line, "addslashes requires a string argument")
			}
			return valueType{kind: stringKind}, nil
		}
		if strings.EqualFold(expr.Text, "stripslashes") {
			if len(expr.Args) != 1 {
				return valueType{}, unsupported(expr.Line, "stripslashes expects one argument")
			}
			if args[0].kind != stringKind && args[0].kind != unknownKind {
				return valueType{}, unsupported(expr.Line, "stripslashes requires a string argument")
			}
			return valueType{kind: stringKind}, nil
		}
		if strings.EqualFold(expr.Text, "str_replace") {
			if len(expr.Args) != 3 {
				return valueType{}, unsupported(expr.Line, "str_replace expects three arguments")
			}
			for _, arg := range args {
				if arg.kind != stringKind && arg.kind != unknownKind {
					return valueType{}, unsupported(expr.Line, "str_replace requires three string arguments")
				}
			}
			return valueType{kind: stringKind}, nil
		}
		if strings.EqualFold(expr.Text, "sprintf") {
			if len(expr.Args) != 2 {
				return valueType{}, unsupported(expr.Line, "sprintf expects two arguments")
			}
			if args[0].kind != stringKind && args[0].kind != unknownKind {
				return valueType{}, unsupported(expr.Line, "sprintf format requires string")
			}
			if expr.Args[0].Kind == "string" {
				if _, ok := sprintfsubset.Width(expr.Args[0].Text); !ok {
					return valueType{}, unsupported(expr.Line, "sprintf format outside supported %0Ns subset (width 0..4096)")
				}
			}
			if args[1].kind != stringKind && args[1].kind != intKind && args[1].kind != unknownKind {
				return valueType{}, unsupported(expr.Line, "sprintf value requires string or integer")
			}
			return valueType{kind: stringKind}, nil
		}
		if strings.EqualFold(expr.Text, "count") {
			if len(args) != 1 {
				return valueType{}, unsupported(expr.Line, "count expects one argument")
			}
			if args[0].kind != arrayKind && args[0].kind != unknownKind {
				return valueType{}, unsupported(expr.Line, "count requires an array")
			}
			return valueType{kind: intKind}, nil
		}
		if strings.EqualFold(expr.Text, "strlen") {
			if len(args) != 1 {
				return valueType{}, unsupported(expr.Line, "strlen expects one argument")
			}
			if args[0].kind != stringKind && args[0].kind != unknownKind {
				return valueType{}, unsupported(expr.Line, "strlen requires a string")
			}
			return valueType{kind: intKind}, nil
		}
		if strings.EqualFold(expr.Text, "array_keys") {
			if len(args) != 1 {
				return valueType{}, unsupported(expr.Line, "array_keys expects one argument")
			}
			if args[0].kind != arrayKind && args[0].kind != unknownKind {
				return valueType{}, unsupported(expr.Line, "array_keys requires an array")
			}
			return valueType{kind: arrayKind}, nil
		}
		if strings.EqualFold(expr.Text, "array_key_first") || strings.EqualFold(expr.Text, "array_key_last") {
			name := strings.ToLower(expr.Text)
			if len(args) != 1 {
				return valueType{}, unsupported(expr.Line, name+" expects one argument")
			}
			if args[0].kind != arrayKind && args[0].kind != unknownKind {
				return valueType{}, unsupported(expr.Line, name+" requires an array")
			}
			// int|string|null has no dedicated union in the current type model.
			// Preserve downstream runtime guards instead of promising an int key.
			return valueType{kind: unknownKind}, nil
		}
		if strings.EqualFold(expr.Text, "array_merge") {
			for _, arg := range args {
				if arg.kind != arrayKind && arg.kind != unknownKind {
					return valueType{}, unsupported(expr.Line, "array_merge requires arrays")
				}
			}
			return valueType{kind: arrayKind}, nil
		}
		if strings.EqualFold(expr.Text, "array_reverse") {
			if len(args) < 1 || len(args) > 2 {
				return valueType{}, unsupported(expr.Line, "array_reverse expects one or two arguments")
			}
			if args[0].kind != arrayKind && args[0].kind != unknownKind {
				return valueType{}, unsupported(expr.Line, "array_reverse requires an array")
			}
			if len(args) == 2 && args[1].kind != boolKind && args[1].kind != unknownKind {
				return valueType{}, unsupported(expr.Line, "array_reverse preserve_keys requires bool")
			}
			return valueType{kind: arrayKind}, nil
		}
		if strings.EqualFold(expr.Text, "array_values") {
			if len(args) != 1 {
				return valueType{}, unsupported(expr.Line, "array_values expects one argument")
			}
			if args[0].kind != arrayKind && args[0].kind != unknownKind {
				return valueType{}, unsupported(expr.Line, "array_values requires an array")
			}
			return valueType{kind: arrayKind}, nil
		}
		if strings.EqualFold(expr.Text, "array_key_exists") {
			if len(args) != 2 {
				return valueType{}, unsupported(expr.Line, "array_key_exists expects two arguments")
			}
			if args[0].kind != intKind && args[0].kind != stringKind && args[0].kind != unknownKind {
				return valueType{}, unsupported(expr.Line, "array_key_exists requires an integer or string key")
			}
			if args[1].kind != arrayKind && args[1].kind != unknownKind {
				return valueType{}, unsupported(expr.Line, "array_key_exists requires an array")
			}
			return valueType{kind: boolKind}, nil
		}
		if strings.EqualFold(expr.Text, "array_is_list") {
			if len(args) != 1 {
				return valueType{}, unsupported(expr.Line, "array_is_list expects one argument")
			}
			if args[0].kind != arrayKind && args[0].kind != unknownKind {
				return valueType{}, unsupported(expr.Line, "array_is_list requires an array")
			}
			return valueType{kind: boolKind}, nil
		}
		if strings.EqualFold(expr.Text, "rtrim") {
			if len(expr.Args) != 2 {
				return valueType{}, unsupported(expr.Line, "rtrim expects two arguments")
			}
			if args[0].kind != stringKind && args[0].kind != unknownKind {
				return valueType{}, unsupported(expr.Line, "rtrim requires a string first argument")
			}
			if expr.Args[1].Kind != "string" || args[1].kind != stringKind {
				return valueType{}, unsupported(expr.Line, "rtrim requires a string literal mask")
			}
			if strings.Contains(expr.Args[1].Text, "..") {
				return valueType{}, unsupported(expr.Line, "rtrim character ranges unsupported")
			}
			return valueType{kind: stringKind}, nil
		}
		if strings.EqualFold(expr.Text, "is_null") || strings.EqualFold(expr.Text, "is_int") || strings.EqualFold(expr.Text, "is_bool") || strings.EqualFold(expr.Text, "is_string") || strings.EqualFold(expr.Text, "is_array") {
			if len(expr.Args) != 1 {
				return valueType{}, unsupported(expr.Line, strings.ToLower(expr.Text)+" expects one argument")
			}
			return valueType{kind: boolKind}, nil
		}
		if strings.EqualFold(expr.Text, "strcasecmp") {
			if len(args) != 2 {
				return valueType{}, unsupported(expr.Line, "strcasecmp expects two arguments")
			}
			for _, arg := range args {
				if arg.kind != stringKind && arg.kind != unknownKind {
					return valueType{}, unsupported(expr.Line, "strcasecmp requires two strings")
				}
			}
			return valueType{kind: intKind}, nil
		}
		if strings.EqualFold(expr.Text, "strtolower") {
			if len(expr.Args) != 1 {
				return valueType{}, unsupported(expr.Line, "strtolower expects one argument")
			}
			if args[0].kind != stringKind && args[0].kind != unknownKind {
				return valueType{}, unsupported(expr.Line, "strtolower requires a string argument")
			}
			return valueType{kind: stringKind}, nil
		}
		if arity, ok := v.functions[strings.ToLower(expr.Text)]; ok {
			if len(args) != arity {
				return valueType{}, unsupported(expr.Line, "unsupported argument count for "+expr.Text)
			}
			for i, annotation := range v.parameters[strings.ToLower(expr.Text)] {
				if annotation == "array" && args[i].kind != arrayKind && args[i].kind != unknownKind {
					return valueType{}, unsupported(expr.Line, "array parameter requires an array argument")
				}
				if annotation == "callable" && args[i].kind != callableKind && args[i].kind != unknownKind {
					return valueType{}, unsupported(expr.Line, "callable parameter requires a generated arrow")
				}
			}
		}
		return valueType{kind: unknownKind}, nil
	case "new":
		if err := require(0); err != nil {
			return valueType{}, err
		}
		if !validName(expr.Text) {
			return valueType{}, malformed(expr.Line, "invalid class name")
		}
		return valueType{kind: objectKind, class: strings.ToLower(expr.Text)}, nil
	case "property":
		if err := require(1); err != nil {
			return valueType{}, err
		}
		if !validName(expr.Text) {
			return valueType{}, malformed(expr.Line, "invalid property name")
		}
		receiver, err := v.expr(expr.Args[0], scope)
		if err != nil {
			return valueType{}, err
		}
		if receiver.kind != objectKind && receiver.kind != unknownKind {
			return valueType{}, unsupported(expr.Line, "property access requires an object")
		}
		if receiver.kind == objectKind {
			if info, ok := v.classes[receiver.class]; ok {
				property, ok := info.properties[expr.Text]
				if !ok {
					return valueType{}, unsupported(expr.Line, "unknown property "+receiver.class+"::$"+expr.Text)
				}
				return property, nil
			}
		}
		return valueType{kind: unknownKind}, nil
	case "method":
		if len(expr.Args) == 0 {
			return valueType{}, malformed(expr.Line, "method expression requires a receiver")
		}
		if !validName(expr.Text) {
			return valueType{}, malformed(expr.Line, "invalid method name")
		}
		receiver, err := v.expr(expr.Args[0], scope)
		if err != nil {
			return valueType{}, err
		}
		if _, err := v.exprs(expr.Args[1:], scope); err != nil {
			return valueType{}, err
		}
		if receiver.kind != objectKind && receiver.kind != unknownKind {
			return valueType{}, unsupported(expr.Line, "method call requires an object")
		}
		if receiver.kind == objectKind {
			if info, ok := v.classes[receiver.class]; ok {
				arity, ok := info.methods[strings.ToLower(expr.Text)]
				if !ok {
					return valueType{}, unsupported(expr.Line, "unknown method "+receiver.class+"::"+expr.Text)
				}
				if len(expr.Args)-1 != arity {
					return valueType{}, unsupported(expr.Line, "unsupported argument count for "+receiver.class+"::"+expr.Text)
				}
			}
		}
		return valueType{kind: unknownKind}, nil
	default:
		return valueType{}, malformed(expr.Line, "unsupported IR expression "+expr.Kind)
	}
}

func (v validator) exprs(exprs []ir.Expr, scope map[string]valueType) ([]valueType, error) {
	types := make([]valueType, 0, len(exprs))
	for _, expr := range exprs {
		t, err := v.expr(expr, scope)
		if err != nil {
			return nil, err
		}
		types = append(types, t)
	}
	return types, nil
}

func literalType(expr ir.Expr) valueType {
	switch expr.Kind {
	case "int":
		return valueType{kind: intKind}
	case "string":
		return valueType{kind: stringKind}
	case "bool":
		return valueType{kind: boolKind}
	case "null":
		return valueType{kind: nullKind}
	default:
		return valueType{kind: unknownKind}
	}
}

func isScalar(t valueType) bool {
	return t.kind == intKind || t.kind == stringKind || t.kind == boolKind || t.kind == nullKind
}

func isScalarOrUnknown(t valueType) bool { return isScalar(t) || t.kind == unknownKind }

// PHP truthiness is defined for the supported scalar and array domains. A
// known object is intentionally excluded until the runtime has an explicit
// object truthiness contract; unknown values stay dynamically guarded.
func isBoolDomain(t valueType) bool {
	return t.kind == nullKind || t.kind == boolKind || t.kind == intKind || t.kind == stringKind || t.kind == arrayKind || t.kind == unknownKind
}

func isNumericOrUnknown(t valueType) bool {
	return isScalar(t) || t.kind == unknownKind
}

func cloneScope(scope map[string]valueType) map[string]valueType {
	clone := make(map[string]valueType, len(scope))
	for name, t := range scope {
		clone[name] = t
	}
	return clone
}

func validName(name string) bool {
	if name == "" || !nameStart(name[0]) {
		return false
	}
	for i := 1; i < len(name); i++ {
		if !namePart(name[i]) {
			return false
		}
	}
	return true
}

func nameStart(b byte) bool { return b == '_' || b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' }
func namePart(b byte) bool  { return nameStart(b) || b >= '0' && b <= '9' }

func malformed(line int, message string) error {
	return fmt.Errorf("IR002 line %d: %s", line, message)
}

func unsupported(line int, message string) error {
	return fmt.Errorf("PHP004 line %d: %s", line, message)
}
