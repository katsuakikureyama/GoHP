// Package semantic validates the closed GoHP subset before code generation.
package semantic

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/katsuakikureyama/GoHP/internal/ir"
	"github.com/katsuakikureyama/GoHP/internal/phpversion"
	"github.com/katsuakikureyama/GoHP/internal/sprintfsubset"
	"github.com/katsuakikureyama/GoHP/internal/switchstring"
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
	kind       kind
	class      string
	directNull bool // A preceding standalone `$local = null`, not merely a null-typed expression.
}

type classInfo struct {
	properties       map[string]valueType
	methods          map[string]int
	methodMinimum    map[string]int
	methodParameters map[string][]string
	methodNewReturn  map[string]string // Only a sole "return new DeclaredClass(...)" body.
	constructorArity int               // Zero also describes a class without a constructor.
	constructorMin   int
	constructorTypes []string
}

type validator struct {
	target       string
	regexProfile string
	classes      map[string]classInfo
	functions    map[string]int
	minimum      map[string]int
	parameters   map[string][]string
	returns      map[string]string
}

// Validate rejects malformed IR and operations that are statically outside the
// supported PHP subset. Unknown dynamic values remain guarded by the runtime.
func Validate(p ir.Program) error {
	if err := phpversion.ValidateTarget(p.PHPVersion); err != nil {
		return fmt.Errorf("PHP003: %w", err)
	}
	if p.RegexProfile != "" && p.RegexProfile != ir.Preg83Profile {
		return malformed(1, "unknown regex profile")
	}
	if p.RegexProfile != "" && p.PHPVersion != "8.3" {
		return unsupported(1, "regex profile is only admitted for PHP 8.3")
	}
	v := validator{target: p.PHPVersion, regexProfile: p.RegexProfile, classes: make(map[string]classInfo, len(p.Classes)), functions: make(map[string]int, len(p.Functions)), minimum: map[string]int{}, parameters: make(map[string][]string, len(p.Functions)), returns: map[string]string{}}
	if err := v.preg83Placement(p); err != nil {
		return err
	}
	// Collect before validating bodies so literal callbacks can refer forward
	// or back to their own declaration without dynamic callable resolution.
	for _, f := range p.Functions {
		if phpversion.ReservedFunction(p.PHPVersion, f.Name) {
			return unsupported(f.Line, "function name reserved in PHP "+p.PHPVersion+": "+f.Name)
		}
		if err := validateParameterTypes(f, false); err != nil {
			return err
		}
		v.functions[strings.ToLower(f.Name)] = len(f.Params)
		v.minimum[strings.ToLower(f.Name)] = len(f.Params) - f.OptionalNullParams
		v.parameters[strings.ToLower(f.Name)] = f.ParamTypes
		v.returns[strings.ToLower(f.Name)] = f.ReturnType
	}
	for _, c := range p.Classes {
		if phpversion.DeprecatedClassName(p.PHPVersion, c.Name) {
			return unsupported(1, "deprecated class name outside PHP "+p.PHPVersion+" subset: "+c.Name)
		}
		if !validName(c.Name) {
			return malformed(1, "invalid class name")
		}
		info := classInfo{properties: map[string]valueType{}, methods: map[string]int{}, methodMinimum: map[string]int{}, methodParameters: map[string][]string{}, methodNewReturn: map[string]string{}}
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
			info.methodMinimum[strings.ToLower(method.Name)] = len(method.Params) - method.OptionalNullParams
			info.methodParameters[strings.ToLower(method.Name)] = method.ParamTypes
			if len(method.Body) == 1 && method.Body[0].Kind == "return" && method.Body[0].Expr.Kind == "new" {
				info.methodNewReturn[strings.ToLower(method.Name)] = strings.ToLower(method.Body[0].Expr.Text)
			}
			if strings.EqualFold(method.Name, "__construct") {
				info.constructorArity = len(method.Params)
				info.constructorMin = len(method.Params) - method.OptionalNullParams
				info.constructorTypes = method.ParamTypes
			}
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
	return v.stmts(p.Body, map[string]valueType{}, "", "")
}

// preg_match_local deliberately has no control-flow join contract. It may live
// in a top-level function/method body, but not in a nested statement body.
// Scan before ordinary validation so malformed nested raw IR also fails closed.
func (v validator) preg83Placement(p ir.Program) error {
	var scan func([]ir.Stmt, bool) error
	scan = func(stmts []ir.Stmt, nested bool) error {
		for _, stmt := range stmts {
			if stmt.Kind == "preg_match_local" && nested {
				return unsupported(stmt.Line, "preg_match local output requires a straight-line scope")
			}
			if err := scan(stmt.Body, true); err != nil {
				return err
			}
			if err := scan(stmt.ElseBody, true); err != nil {
				return err
			}
		}
		return nil
	}
	if err := scan(p.Body, false); err != nil {
		return err
	}
	for _, f := range p.Functions {
		if err := scan(f.Body, false); err != nil {
			return err
		}
	}
	for _, c := range p.Classes {
		for _, f := range c.Methods {
			if err := scan(f.Body, false); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateParameterTypes(f ir.Function, method bool) error {
	if f.OptionalNullParams < 0 || f.OptionalNullParams > len(f.Params) {
		return malformed(f.Line, "invalid optional-null parameter count")
	}
	if len(f.ParamTypes) != 0 && len(f.ParamTypes) != len(f.Params) {
		return malformed(f.Line, "parameter types must align with parameter names")
	}
	constructor := method && strings.EqualFold(f.Name, "__construct")
	for _, annotation := range f.ParamTypes {
		scalar := annotation == "int" || annotation == "string" || annotation == "bool"
		_, classType := classParameterType(annotation)
		if annotation != "" && annotation != "array" && annotation != "callable" && !scalar && !classType {
			return malformed(f.Line, "unsupported IR parameter type")
		}
		if annotation != "" && method && !classType && !(constructor && scalar) {
			return unsupported(f.Line, "scalar parameter types require a public constructor; other typed methods are unsupported")
		}
	}
	for i := len(f.Params) - f.OptionalNullParams; i < len(f.Params); i++ {
		if len(f.ParamTypes) != 0 && f.ParamTypes[i] != "" {
			return unsupported(f.Line, "null defaults require untyped parameters")
		}
	}
	return nil
}

func (v validator) function(f ir.Function, class string) error {
	if class != "" && strings.EqualFold(f.Name, "__construct") && constructorHasReturn(f.Body) {
		return unsupported(f.Line, "constructor return statements are unsupported")
	}
	if strings.EqualFold(f.Name, "fn") {
		return unsupported(f.Line, "fn is reserved for arrow expressions")
	}
	if class == "" && strings.EqualFold(f.Name, "list") {
		return unsupported(f.Line, "list is a reserved language construct")
	}
	if class == "" && (strings.EqualFold(f.Name, "switch") || strings.EqualFold(f.Name, "case") || strings.EqualFold(f.Name, "default")) {
		return unsupported(f.Line, "reserved switch language construct")
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
		if len(f.ParamTypes) != 0 {
			if name, ok := classParameterType(f.ParamTypes[i]); ok {
				if _, declared := v.classes[name]; !declared {
					return unsupported(f.Line, "class parameter requires a declared generated class: "+name)
				}
				scope[param] = valueType{kind: objectKind, class: name}
			}
			switch f.ParamTypes[i] {
			case "int":
				scope[param] = valueType{kind: intKind}
			case "string":
				scope[param] = valueType{kind: stringKind}
			case "bool":
				scope[param] = valueType{kind: boolKind}
			}
		}
	}
	return v.stmts(f.Body, scope, f.ReturnType, "")
}

// control is the nearest breakable construct: empty, loop or switch. PHP's
// switch-target continue emits a warning, outside the quiet supported lane.
func (v validator) stmts(stmts []ir.Stmt, scope map[string]valueType, returnType string, control string) error {
	for _, stmt := range stmts {
		if stmt.Kind != "if" && stmt.ElseBody != nil {
			return malformed(stmt.Line, "else body requires an if statement")
		}
		switch stmt.Kind {
		case "assign_nested_property", "assign_nested_property_append", "assign_nested_property_concat":
			if err := v.nestedProperty(stmt, scope); err != nil {
				return err
			}
		case "switch":
			if stmt.Name != "" || stmt.Key != "" || stmt.Value != "" || !emptyExpr(stmt.Target) {
				return malformed(stmt.Line, "switch requires only selector and clauses")
			}
			t, err := v.expr(stmt.Expr, scope)
			if err != nil {
				return err
			}
			if t.kind != stringKind && t.kind != unknownKind || (stmt.Expr.Kind == "string" && !switchstring.Admitted(stmt.Expr.Text)) {
				return unsupported(stmt.Line, "switch requires a nonnumeric string")
			}
			entry := cloneScope(scope)
			// Any earlier clause can fall through. Lose entry type facts for
			// existing locals written anywhere, including dead tails/loops.
			widenWhileWrites(stmt.Body, entry)
			seenDefault := false
			for _, clause := range stmt.Body {
				if clause.Name != "" || clause.Key != "" || clause.Value != "" || !emptyExpr(clause.Target) || clause.ElseBody != nil {
					return malformed(clause.Line, "switch clause contains hidden payload")
				}
				switch clause.Kind {
				case "case":
					if clause.Expr.Kind != "string" || clause.Expr.Args != nil || clause.Expr.Params != nil || !switchstring.Admitted(clause.Expr.Text) {
						return unsupported(clause.Line, "switch case requires a nonnumeric literal string")
					}
				case "default":
					if seenDefault || !emptyExpr(clause.Expr) {
						return malformed(clause.Line, "switch requires at most one payload-free default")
					}
					seenDefault = true
				default:
					return malformed(clause.Line, "switch body requires case/default clauses")
				}
				if err := v.stmts(clause.Body, cloneScope(entry), returnType, "switch"); err != nil {
					return err
				}
			}
			for local := range scope {
				scope[local] = entry[local]
			}
		case "list_assign":
			target := stmt.Target
			if stmt.Name != "" || stmt.Key != "" || stmt.Value != "" || stmt.Body != nil ||
				target.Kind != "list" || target.Text != "" || target.Params != nil || len(target.Args) == 0 {
				return malformed(stmt.Line, "list assignment requires positional local targets")
			}
			for _, local := range target.Args {
				if local.Kind != "var" || !validName(local.Text) || local.Args != nil || local.Params != nil ||
					local.Text == "this" || local.Text == "GLOBALS" || strings.HasPrefix(local.Text, "_") {
					return unsupported(stmt.Line, "list assignment requires nonreserved bare locals")
				}
			}
			// Read in the original scope: targets cannot initialize their RHS.
			t, err := v.expr(stmt.Expr, scope)
			if err != nil {
				return err
			}
			if t.kind != arrayKind && t.kind != unknownKind {
				return unsupported(stmt.Line, "list assignment requires array")
			}
			for _, local := range target.Args {
				// Also invalidate prior class knowledge and direct-null facts.
				scope[local.Text] = valueType{kind: unknownKind}
			}
		case "call_stmt":
			if stmt.Name != "" || stmt.Key != "" || stmt.Value != "" || stmt.Body != nil || !emptyExpr(stmt.Target) {
				return malformed(stmt.Line, "call statement requires only a call expression")
			}
			switch stmt.Expr.Kind {
			case "call":
			case "method", "invoke":
				if len(stmt.Expr.Args) == 0 {
					return malformed(stmt.Line, "standalone method/callable requires a bare variable receiver")
				}
				if stmt.Expr.Args[0].Kind != "var" {
					return unsupported(stmt.Line, "standalone method/callable requires a bare variable receiver")
				}
			default:
				return unsupported(stmt.Line, "standalone expression must be a supported call")
			}
			if _, err := v.expr(stmt.Expr, scope); err != nil {
				return err
			}
		case "continue", "break":
			if stmt.Name != "" || stmt.Key != "" || stmt.Value != "" || stmt.Body != nil || !emptyExpr(stmt.Expr) || !emptyExpr(stmt.Target) {
				return malformed(stmt.Line, stmt.Kind+" must have no payload")
			}
			if control == "" {
				return unsupported(stmt.Line, stmt.Kind+" requires an enclosing foreach or while")
			}
			if stmt.Kind == "continue" && control == "switch" {
				return unsupported(stmt.Line, "continue targeting switch is outside quiet subset")
			}
			// Still validate dead statements, without letting dead assignments
			// overwrite the type state at the reachable loop transfer.
			scope = cloneScope(scope)
		case "assign":
			if !validName(stmt.Name) {
				return malformed(stmt.Line, "invalid assignment target")
			}
			t, err := v.expr(stmt.Expr, scope)
			if err != nil {
				return err
			}
			// The preg83 output boundary distinguishes an explicit local null
			// initialization from a merely null-typed derived expression.
			t.directNull = stmt.Expr.Kind == "null"
			scope[stmt.Name] = t
		case "preg_match_local":
			if err := v.pregMatchLocal(stmt, scope); err != nil {
				return err
			}
		case "assign_property", "assign_property_append":
			target := stmt.Target
			if stmt.Name != "" || stmt.Key != "" || stmt.Value != "" || stmt.Body != nil ||
				target.Kind != "property" || !validName(target.Text) || target.Params != nil || len(target.Args) != 1 {
				return malformed(stmt.Line, "property assignment requires a single named property target")
			}
			base := target.Args[0]
			if base.Kind != "var" || !validName(base.Text) || base.Args != nil || base.Params != nil {
				return malformed(stmt.Line, "property assignment requires a bare variable receiver")
			}
			receiver, exists := scope[base.Text]
			if !exists || receiver.kind != objectKind {
				return unsupported(stmt.Line, "property assignment requires a known generated class variable")
			}
			info, exists := v.classes[receiver.class]
			if !exists {
				return unsupported(stmt.Line, "property assignment requires a declared generated class")
			}
			if _, exists := info.properties[target.Text]; !exists {
				return unsupported(stmt.Line, "property assignment requires a declared public property")
			}
			if _, err := v.expr(base, scope); err != nil {
				return err
			}
			if _, err := v.expr(stmt.Expr, scope); err != nil {
				return err
			}
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
		case "while":
			if stmt.Name != "" || stmt.Key != "" || stmt.Value != "" || !emptyExpr(stmt.Target) {
				return malformed(stmt.Line, "while requires only a condition and body")
			}
			// Reject a known-invalid first condition before widening loop-carried
			// variables. Later evaluations and the body must not assume entry types.
			condition, err := v.expr(stmt.Expr, scope)
			if err != nil {
				return err
			}
			if !isBoolDomain(condition) {
				return unsupported(stmt.Line, "while condition requires a supported value")
			}
			loop := cloneScope(scope)
			widenWhileWrites(stmt.Body, loop)
			condition, err = v.expr(stmt.Expr, loop)
			if err != nil {
				return err
			}
			if !isBoolDomain(condition) {
				return unsupported(stmt.Line, "while condition requires a supported value")
			}
			if err := v.stmts(stmt.Body, cloneScope(loop), returnType, "loop"); err != nil {
				return err
			}
			// Zero iterations, continue and early break are possible. Do not use
			// the last textual assignment as a definite post-loop type.
			for name := range scope {
				scope[name] = loop[name]
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
			if err := v.stmts(stmt.Body, inner, returnType, "loop"); err != nil {
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
			if err := v.stmts(stmt.Body, inner, returnType, control); err != nil {
				return err
			}
			other := cloneScope(scope)
			if err := v.stmts(stmt.ElseBody, other, returnType, control); err != nil {
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

// Conservatively widen existing variable roots written anywhere in a while
// body, including nested loops and dead tails. Property writes do not replace
// their receiver; mutable property reads already have unknown value type.
// Check shapes defensively here: normal statement validation follows this scan.
func widenWhileWrites(stmts []ir.Stmt, scope map[string]valueType) {
	mark := func(name string) {
		if _, exists := scope[name]; exists {
			scope[name] = valueType{kind: unknownKind}
		}
	}
	for _, stmt := range stmts {
		switch stmt.Kind {
		case "list_assign":
			for _, local := range stmt.Target.Args {
				if local.Kind == "var" {
					mark(local.Text)
				}
			}
		case "assign":
			mark(stmt.Name)
		case "preg_match_local":
			mark(stmt.Name)
			if stmt.Target.Kind == "var" {
				mark(stmt.Target.Text)
			}
		case "assign_index":
			if len(stmt.Target.Args) > 0 && stmt.Target.Args[0].Kind == "var" {
				mark(stmt.Target.Args[0].Text)
			}
		case "assign_append":
			if stmt.Target.Kind == "var" {
				mark(stmt.Target.Text)
			}
		case "foreach":
			mark(stmt.Key)
			mark(stmt.Value)
		}
		widenWhileWrites(stmt.Body, scope)
		widenWhileWrites(stmt.ElseBody, scope)
	}
}

// pregMatchLocal validates the fixed experimental statement representation.
// It is intentionally stricter than the B2 Go API: source inputs are literal
// or known typed bare locals or guarded named property reads. The writable
// output is a pre-existing null local and no input root aliases a destination.
func (v validator) pregMatchLocal(stmt ir.Stmt, scope map[string]valueType) error {
	if v.target != "8.3" || v.regexProfile != ir.Preg83Profile {
		return unsupported(stmt.Line, "preg_match local output requires the PHP 8.3 experimental PCRE2 profile")
	}
	target := stmt.Target
	if !preg83LocalName(stmt.Name) || stmt.Key != "" || stmt.Value != "" || stmt.Body != nil ||
		target.Kind != "var" || !preg83LocalName(target.Text) || target.Args != nil || target.Params != nil {
		return malformed(stmt.Line, "preg_match local output requires distinct bare local variables")
	}
	if stmt.Name == target.Text {
		return unsupported(stmt.Line, "preg_match result and output locals must be distinct")
	}
	if stmt.Expr.Kind != "call" || stmt.Expr.Text != "preg_match" || stmt.Expr.Params != nil || len(stmt.Expr.Args) != 3 {
		return malformed(stmt.Line, "preg_match local output requires fixed call metadata")
	}
	pattern, subject, offset := stmt.Expr.Args[0], stmt.Expr.Args[1], stmt.Expr.Args[2]
	for _, input := range []ir.Expr{pattern, subject, offset} {
		if err := preg83InputMetadata(input); err != nil {
			return err
		}
	}
	if !preg83LiteralPattern(pattern) {
		return unsupported(stmt.Line, "preg_match requires a literal /.../s pattern")
	}
	if !preg83InputForm(subject, "string") || !preg83InputForm(offset, "int") {
		return unsupported(stmt.Line, "preg_match requires literal, typed-local or named-property subject and offset values")
	}
	for _, input := range []ir.Expr{subject, offset} {
		for input.Kind == "property" {
			input = input.Args[0]
		}
		if input.Kind == "var" && (input.Text == stmt.Name || input.Text == target.Text) {
			return unsupported(stmt.Line, "preg_match inputs must not alias result or output locals")
		}
	}
	if _, exists := scope[target.Text]; !exists || scope[target.Text].kind != nullKind || !scope[target.Text].directNull {
		return unsupported(stmt.Line, "preg_match output requires a directly initialized null local")
	}
	subjectType, err := v.expr(subject, scope)
	if err != nil {
		return err
	}
	if subjectType.kind != stringKind && !(subject.Kind == "property" && subjectType.kind == unknownKind) {
		return unsupported(stmt.Line, "preg_match subject requires a string value; only property reads may have an unknown domain")
	}
	offsetType, err := v.expr(offset, scope)
	if err != nil {
		return err
	}
	if offsetType.kind != intKind && !(offset.Kind == "property" && offsetType.kind == unknownKind) {
		return unsupported(stmt.Line, "preg_match offset requires an integer value; only property reads may have an unknown domain")
	}
	// The engine's return is int(0|1)|false. The existing type model has no
	// union representation, so preserve downstream runtime guards.
	scope[target.Text] = valueType{kind: arrayKind}
	scope[stmt.Name] = valueType{kind: unknownKind}
	return nil
}

func preg83LocalName(name string) bool {
	return validName(name) && name != "this" && name != "GLOBALS" && !strings.HasPrefix(name, "_")
}

func preg83LiteralPattern(expr ir.Expr) bool {
	return expr.Kind == "string" && expr.Args == nil && expr.Params == nil && len(expr.Text) >= 3 &&
		expr.Text[0] == '/' && strings.HasSuffix(expr.Text, "/s")
}

func preg83InputForm(expr ir.Expr, kind string) bool {
	if expr.Kind == "property" {
		for expr.Kind == "property" {
			expr = expr.Args[0] // Metadata was checked before input admission.
		}
		return expr.Kind == "var"
	}
	if expr.Args != nil || expr.Params != nil {
		return false
	}
	switch kind {
	case "string":
		return expr.Kind == "string" || expr.Kind == "var"
	case "int":
		return expr.Kind == "int" || expr.Kind == "var"
	default:
		return false
	}
}

func emptyExpr(e ir.Expr) bool {
	return e.Kind == "" && e.Text == "" && e.Line == 0 && e.Args == nil && e.Params == nil
}

func constructorHasReturn(body []ir.Stmt) bool {
	for _, stmt := range body {
		if stmt.Kind == "return" || constructorHasReturn(stmt.Body) || constructorHasReturn(stmt.ElseBody) {
			return true
		}
	}
	return false
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
	case "version_constant":
		if err := require(0); err != nil {
			return valueType{}, err
		}
		if _, err := phpversion.VersionConstant(v.target, expr.Text); err != nil {
			return valueType{}, malformed(expr.Line, err.Error())
		}
		return valueType{kind: intKind}, nil
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
	case "coalesce":
		if err := require(2); err != nil {
			return valueType{}, err
		}
		root, keys, ok := ir.ArrayAddressParts(expr.Args[0])
		if expr.Text != "" {
			return valueType{}, malformed(expr.Line, "coalesce requires a variable-rooted array address")
		}
		if !ok {
			return valueType{}, unsupported(expr.Line, "coalesce requires a variable-rooted array address")
		}
		if err := v.arrayPath(root, keys, scope); err != nil {
			return valueType{}, err
		}
		if _, err := v.expr(expr.Args[1], scope); err != nil {
			return valueType{}, err
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
		case "<", ">":
			for _, operand := range []valueType{left, right} {
				if operand.kind != intKind && operand.kind != nullKind && operand.kind != unknownKind {
					return valueType{}, unsupported(expr.Line, "relational comparison requires int or null operands")
				}
			}
			return valueType{kind: boolKind}, nil
		case "-":
			if (left.kind != intKind && left.kind != unknownKind) || (right.kind != intKind && right.kind != unknownKind) {
				return valueType{}, unsupported(expr.Line, "subtraction requires int operands")
			}
			return valueType{kind: intKind}, nil
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
			// Literal null identity is a type predicate, not array/object
			// identity. Both operands were validated above; the other value
			// must still be evaluated exactly once by code generation.
			if expr.Args[0].Kind == "null" || expr.Args[1].Kind == "null" {
				return valueType{kind: boolKind}, nil
			}
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
	case "isset_var":
		if err := require(1); err != nil {
			return valueType{}, err
		}
		base := expr.Args[0]
		if expr.Text != "" || base.Kind != "var" {
			return valueType{}, malformed(expr.Line, "isset requires a variable address")
		}
		if _, ok := scope[base.Text]; !ok {
			return valueType{}, unsupported(expr.Line, "isset requires a defined variable")
		}
		if _, err := v.expr(base, scope); err != nil {
			return valueType{}, err
		}
		return valueType{kind: boolKind}, nil
	case "isset_property":
		if expr.Text != "" || len(expr.Args) < 2 {
			return valueType{}, malformed(expr.Line, "property isset requires a root and named path")
		}
		base := expr.Args[0]
		if base.Kind != "var" || base.Args != nil || base.Params != nil || !validName(base.Text) {
			return valueType{}, malformed(expr.Line, "property isset requires a bare local root")
		}
		if _, defined := scope[base.Text]; !defined {
			return valueType{}, unsupported(expr.Line, "property isset requires a defined local root")
		}
		if _, err := v.expr(base, scope); err != nil {
			return valueType{}, err
		}
		if scope[base.Text].kind == callableKind {
			return valueType{}, unsupported(expr.Line, "property isset on callable values unsupported")
		}
		for _, property := range expr.Args[1:] {
			if property.Kind != "string" || !validName(property.Text) || property.Args != nil || property.Params != nil {
				return valueType{}, malformed(expr.Line, "property isset requires literal identifier names")
			}
		}
		// Do not validate through ordinary property reads: even a known missing
		// property must yield false without invoking its throwing Get boundary.
		return valueType{kind: boolKind}, nil
	case "isset_multi":
		if expr.Text != "" || len(expr.Args) < 2 {
			return valueType{}, malformed(expr.Line, "multi isset requires at least two addresses")
		}
		for _, operand := range expr.Args {
			if operand.Kind != "isset_var" && operand.Kind != "isset_index" && operand.Kind != "isset_path" && operand.Kind != "isset_property" {
				return valueType{}, malformed(expr.Line, "multi isset requires variable addresses")
			}
			if _, err := v.expr(operand, scope); err != nil {
				return valueType{}, err
			}
		}
		return valueType{kind: boolKind}, nil
	case "isset_path":
		if expr.Text != "" || len(expr.Args) < 3 {
			return valueType{}, malformed(expr.Line, "nested isset requires a root and at least two keys")
		}
		if err := v.arrayPath(expr.Args[0], expr.Args[1:], scope); err != nil {
			return valueType{}, err
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
			return valueType{}, malformed(expr.Line, "isset must use an address language construct")
		}
		if !validName(expr.Text) {
			return valueType{}, malformed(expr.Line, "invalid function name")
		}
		args, err := v.exprs(expr.Args, scope)
		if err != nil {
			return valueType{}, err
		}
		if strings.EqualFold(expr.Text, "preg_match") {
			return valueType{}, unsupported(expr.Line, "preg_match is only admitted as experimental local-output syntax")
		}
		if strings.EqualFold(expr.Text, "preg_last_error") {
			if v.target != "8.3" || v.regexProfile != ir.Preg83Profile {
				return valueType{}, unsupported(expr.Line, "preg_last_error requires the PHP 8.3 experimental PCRE2 profile")
			}
			if len(args) != 0 {
				return valueType{}, unsupported(expr.Line, "preg_last_error expects no arguments")
			}
			return valueType{kind: intKind}, nil
		}
		if strings.EqualFold(expr.Text, "strtok") {
			if len(args) != 1 && len(args) != 2 {
				return valueType{}, unsupported(expr.Line, "strtok requires one or two strings")
			}
			for _, arg := range args {
				if arg.kind != stringKind && arg.kind != unknownKind {
					return valueType{}, unsupported(expr.Line, "strtok requires actual strings")
				}
			}
			// string|false: preserve runtime guards instead of promising a string.
			return valueType{kind: unknownKind}, nil
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
		// In 8.3 these names remain ordinary user functions. Both 8.4
		// predicates share the same closed callback admission contract.
		if v.target == "8.4" && (strings.EqualFold(expr.Text, "array_any") || strings.EqualFold(expr.Text, "array_all")) {
			builtin := strings.ToLower(expr.Text)
			if len(expr.Args) != 2 || expr.Args[1].Kind != "string" || !validName(expr.Args[1].Text) {
				return valueType{}, unsupported(expr.Line, builtin+" expects one array and a literal named callback")
			}
			callback := strings.ToLower(expr.Args[1].Text)
			if v.functions[callback] != 2 || v.minimum[callback] != 2 || v.returns[callback] != "bool" {
				return valueType{}, unsupported(expr.Line, builtin+" callback requires two untyped value parameters and a bool return declaration")
			}
			for _, annotation := range v.parameters[callback] {
				if annotation != "" {
					return valueType{}, unsupported(expr.Line, builtin+" callback parameters must be untyped")
				}
			}
			if args[0].kind != arrayKind && args[0].kind != unknownKind {
				return valueType{}, unsupported(expr.Line, builtin+" requires an array")
			}
			return valueType{kind: boolKind}, nil
		}
		if strings.EqualFold(expr.Text, "array_map") {
			if len(expr.Args) != 2 || expr.Args[0].Kind != "string" || !validName(expr.Args[0].Text) {
				return valueType{}, unsupported(expr.Line, "array_map expects a literal named callback and one array")
			}
			if arity, ok := v.functions[strings.ToLower(expr.Args[0].Text)]; !ok || arity < 1 || v.minimum[strings.ToLower(expr.Args[0].Text)] > 1 {
				return valueType{}, unsupported(expr.Line, "array_map callback must be a declared one-argument function")
			}
			if args[1].kind != arrayKind && args[1].kind != unknownKind {
				return valueType{}, unsupported(expr.Line, "array_map requires an array")
			}
			return valueType{kind: arrayKind}, nil
		}
		if strings.EqualFold(expr.Text, "gettype") {
			if len(args) != 1 {
				return valueType{}, unsupported(expr.Line, "gettype expects one argument")
			}
			// All admitted PHP values have a type label. Unknown host values
			// remain subject to the runtime's closed-domain guard.
			return valueType{kind: stringKind}, nil
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
		if strings.EqualFold(expr.Text, "substr") {
			if len(args) < 2 || len(args) > 3 {
				return valueType{}, unsupported(expr.Line, "substr expects two or three arguments")
			}
			if args[0].kind != stringKind && args[0].kind != unknownKind {
				return valueType{}, unsupported(expr.Line, "substr requires a string")
			}
			if args[1].kind != intKind && args[1].kind != unknownKind {
				return valueType{}, unsupported(expr.Line, "substr offset requires int")
			}
			if len(args) == 3 && args[2].kind != intKind && args[2].kind != nullKind && args[2].kind != unknownKind {
				return valueType{}, unsupported(expr.Line, "substr length requires int or null")
			}
			return valueType{kind: stringKind}, nil
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
		if strings.EqualFold(expr.Text, "array_slice") {
			if len(args) < 2 || len(args) > 4 {
				return valueType{}, unsupported(expr.Line, "array_slice expects two to four arguments")
			}
			if args[0].kind != arrayKind && args[0].kind != unknownKind {
				return valueType{}, unsupported(expr.Line, "array_slice requires an array")
			}
			if args[1].kind != intKind && args[1].kind != unknownKind {
				return valueType{}, unsupported(expr.Line, "array_slice offset requires int")
			}
			if len(args) >= 3 && args[2].kind != intKind && args[2].kind != nullKind && args[2].kind != unknownKind {
				return valueType{}, unsupported(expr.Line, "array_slice length requires int or null")
			}
			if len(args) == 4 && args[3].kind != boolKind && args[3].kind != unknownKind {
				return valueType{}, unsupported(expr.Line, "array_slice preserve_keys requires bool")
			}
			return valueType{kind: arrayKind}, nil
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
		if strings.EqualFold(expr.Text, "is_object") || strings.EqualFold(expr.Text, "is_null") || strings.EqualFold(expr.Text, "is_int") || strings.EqualFold(expr.Text, "is_bool") || strings.EqualFold(expr.Text, "is_string") || strings.EqualFold(expr.Text, "is_array") {
			if len(expr.Args) != 1 {
				return valueType{}, unsupported(expr.Line, strings.ToLower(expr.Text)+" expects one argument")
			}
			return valueType{kind: boolKind}, nil
		}
		if strings.EqualFold(expr.Text, "str_starts_with") {
			if len(args) != 2 {
				return valueType{}, unsupported(expr.Line, "str_starts_with expects two arguments")
			}
			for _, arg := range args {
				if arg.kind != stringKind && arg.kind != unknownKind {
					return valueType{}, unsupported(expr.Line, "str_starts_with requires two strings")
				}
			}
			return valueType{kind: boolKind}, nil
		}
		if strings.EqualFold(expr.Text, "str_contains") {
			if len(args) != 2 {
				return valueType{}, unsupported(expr.Line, "str_contains expects two arguments")
			}
			for _, arg := range args {
				if arg.kind != stringKind && arg.kind != unknownKind {
					return valueType{}, unsupported(expr.Line, "str_contains requires two strings")
				}
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
			if len(args) > arity || len(args) < v.minimum[strings.ToLower(expr.Text)] {
				return valueType{}, unsupported(expr.Line, "unsupported argument count for "+expr.Text)
			}
			for i, annotation := range v.parameters[strings.ToLower(expr.Text)] {
				if i >= len(args) {
					break
				}
				if name, classType := classParameterType(annotation); classType && !classArgumentMatches(args[i], name) {
					return valueType{}, unsupported(expr.Line, "class parameter requires "+name+" object")
				}
				if annotation == "array" && args[i].kind != arrayKind && args[i].kind != unknownKind {
					return valueType{}, unsupported(expr.Line, "array parameter requires an array argument")
				}
				if annotation == "callable" && args[i].kind != callableKind && args[i].kind != unknownKind {
					return valueType{}, unsupported(expr.Line, "callable parameter requires a generated arrow")
				}
				if want, scalar := (map[string]kind{"int": intKind, "string": stringKind, "bool": boolKind})[annotation]; scalar && args[i].kind != unknownKind && args[i].kind != want {
					return valueType{}, unsupported(expr.Line, annotation+" parameter requires exact matching argument type")
				}
			}
		}
		return valueType{kind: unknownKind}, nil
	case "new":
		if !validName(expr.Text) {
			return valueType{}, malformed(expr.Line, "invalid class name")
		}
		info, exists := v.classes[strings.ToLower(expr.Text)]
		if !exists {
			return valueType{}, unsupported(expr.Line, "unknown class "+expr.Text)
		}
		args, err := v.exprs(expr.Args, scope)
		if err != nil {
			return valueType{}, err
		}
		if len(expr.Args) < info.constructorMin || len(expr.Args) > info.constructorArity {
			return valueType{}, unsupported(expr.Line, "unsupported constructor argument count for "+expr.Text)
		}
		for i, annotation := range info.constructorTypes {
			if i >= len(args) {
				break
			}
			if name, classType := classParameterType(annotation); classType {
				if !classArgumentMatches(args[i], name) {
					return valueType{}, unsupported(expr.Line, "class parameter requires "+name+" object")
				}
				continue
			}
			want := map[string]kind{"int": intKind, "string": stringKind, "bool": boolKind}[annotation]
			if annotation != "" && args[i].kind != unknownKind && args[i].kind != want {
				return valueType{}, unsupported(expr.Line, annotation+" constructor parameter requires exact matching argument type")
			}
		}
		return valueType{kind: objectKind, class: strings.ToLower(expr.Text)}, nil
	case "object_array_cast":
		if err := require(1); err != nil {
			return valueType{}, err
		}
		if expr.Text != "" {
			return valueType{}, malformed(expr.Line, "object array cast has unexpected metadata")
		}
		receiver, err := v.expr(expr.Args[0], scope)
		if err != nil {
			return valueType{}, err
		}
		if _, exists := v.classes[receiver.class]; receiver.kind != objectKind || !exists {
			return valueType{}, unsupported(expr.Line, "array cast requires a known generated class")
		}
		return valueType{kind: arrayKind}, nil
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
				_, ok := info.properties[expr.Text]
				if !ok {
					return valueType{}, unsupported(expr.Line, "unknown property "+receiver.class+"::$"+expr.Text)
				}
				// Public properties are untyped mutable slots. A default is not a
				// permanent type: aliases/methods may have changed the value.
				return valueType{kind: unknownKind}, nil
			}
		}
		return valueType{kind: unknownKind}, nil
	case "method":
		if strings.EqualFold(expr.Text, "__construct") {
			return valueType{}, unsupported(expr.Line, "direct constructor calls are unsupported")
		}
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
		args, err := v.exprs(expr.Args[1:], scope)
		if err != nil {
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
				if len(expr.Args)-1 > arity || len(expr.Args)-1 < info.methodMinimum[strings.ToLower(expr.Text)] {
					return valueType{}, unsupported(expr.Line, "unsupported argument count for "+receiver.class+"::"+expr.Text)
				}
				for i, annotation := range info.methodParameters[strings.ToLower(expr.Text)] {
					if i >= len(args) {
						break
					}
					if name, classType := classParameterType(annotation); classType && !classArgumentMatches(args[i], name) {
						return valueType{}, unsupported(expr.Line, "class parameter requires "+name+" object")
					}
				}
				if class, ok := info.methodNewReturn[strings.ToLower(expr.Text)]; ok {
					if _, declared := v.classes[class]; declared {
						return valueType{kind: objectKind, class: class}, nil
					}
				}
			}
		}
		return valueType{kind: unknownKind}, nil
	default:
		return valueType{}, malformed(expr.Line, "unsupported IR expression "+expr.Kind)
	}
}

// arrayPath validates addresses without performing an ordinary eager index read.
// Static validation still visits every defined key and the coalesce fallback.
func (v validator) arrayPath(root ir.Expr, keys []ir.Expr, scope map[string]valueType) error {
	if root.Kind != "var" {
		return malformed(root.Line, "array path requires a variable root")
	}
	if _, ok := scope[root.Text]; !ok {
		return unsupported(root.Line, "array path requires a defined variable")
	}
	base, err := v.expr(root, scope)
	if err != nil {
		return err
	}
	if len(keys) != 0 && base.kind != arrayKind && base.kind != nullKind && base.kind != unknownKind {
		return unsupported(root.Line, "array path requires array or null")
	}
	for _, key := range keys {
		if key.Kind != "int" && key.Kind != "string" && key.Kind != "var" {
			return unsupported(key.Line, "array path key must be an integer/string literal or variable")
		}
		if key.Kind == "var" {
			if _, ok := scope[key.Text]; !ok {
				return unsupported(key.Line, "array path requires a defined key variable")
			}
		}
		t, err := v.expr(key, scope)
		if err != nil {
			return err
		}
		if t.kind != intKind && t.kind != stringKind && t.kind != unknownKind {
			return unsupported(key.Line, "array path key must be an integer or string")
		}
	}
	return nil
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
