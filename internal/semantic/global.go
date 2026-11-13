package semantic

import (
	"strings"

	"github.com/katsuakikureyama/GoHP/internal/ir"
)

// GlobalNames extracts the imported variable names from a sequence of
// leading global statements. It is a pure helper used by both semantic
// validation and code generation. Only contiguous leading "global" stmts
// are collected; the first non-global stops the scan.
func GlobalNames(stmts []ir.Stmt) []string {
	var names []string
	for _, s := range stmts {
		if s.Kind != "global" {
			break
		}
		for _, arg := range s.Expr.Args {
			names = append(names, arg.Text)
		}
	}
	return names
}

// validGlobalName checks the closed ASCII identifier form for scalar globals.
func validGlobalName(name string) bool {
	if name == "" || name == "this" || name == "GLOBALS" {
		return false
	}
	letter := func(b byte) bool { return b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' }
	if !letter(name[0]) {
		return false
	}
	for i := 1; i < len(name); i++ {
		b := name[i]
		if !letter(b) && !(b >= '0' && b <= '9') && b != '_' {
			return false
		}
	}
	return true
}

// validateGlobalStmt checks the shape of a single global statement.
func validateGlobalStmt(stmt ir.Stmt) error {
	if stmt.Name != "" || stmt.Key != "" || stmt.Value != "" || stmt.Body != nil || stmt.ElseBody != nil || !globalEmptyExpr(stmt.Target) {
		return malformed(stmt.Line, "global statement contains hidden payload")
	}
	if len(stmt.Expr.Args) == 0 {
		return malformed(stmt.Line, "global requires at least one variable name")
	}
	for _, arg := range stmt.Expr.Args {
		if arg.Kind != "var" || arg.Args != nil || arg.Params != nil {
			return malformed(stmt.Line, "global requires bare variable names")
		}
		if !validGlobalName(arg.Text) {
			return unsupported(stmt.Line, "global variable name unsupported: $"+arg.Text)
		}
	}
	return nil
}

// validateGlobalPlacement checks that global statements appear only in
// admitted positions: (a) leading prologue of a top-level non-method function
// body, (b) entry main top level (any position).
// It returns the set of global-bound names for the scope.
func validateGlobalPlacement(stmts []ir.Stmt, inFunc bool, inMethod bool) (map[string]bool, error) {
	globals := make(map[string]bool)

	if inMethod {
		// Methods cannot have global statements at all.
		for _, s := range stmts {
			if s.Kind == "global" {
				return nil, unsupported(s.Line, "global statement unsupported in methods")
			}
		}
		return globals, nil
	}

	if inFunc {
		// Function: global stmts must be the leading prologue (only globals
		// before the first other statement).
		seenNonGlobal := false
		for _, s := range stmts {
			if s.Kind == "global" {
				if seenNonGlobal {
					return nil, unsupported(s.Line, "global statement must appear before other statements in function")
				}
				if err := validateGlobalStmt(s); err != nil {
					return nil, err
				}
				for _, arg := range s.Expr.Args {
					globals[arg.Text] = true
				}
			} else {
				seenNonGlobal = true
			}
		}
	} else {
		// Main top level: global stmts allowed at any position.
		for _, s := range stmts {
			if s.Kind == "global" {
				if err := validateGlobalStmt(s); err != nil {
					return nil, err
				}
				for _, arg := range s.Expr.Args {
					globals[arg.Text] = true
				}
			}
		}
	}

	return globals, nil
}

// validateGlobalNested rejects global statements inside nested blocks
// (if/else/while/foreach/switch bodies). It only descends into nested
// statement bodies, not the top-level stmts slice.
func validateGlobalNested(stmts []ir.Stmt) error {
	var scanNested func([]ir.Stmt) error
	scanNested = func(ss []ir.Stmt) error {
		if ss == nil {
			return nil
		}
		for _, s := range ss {
			if s.Kind == "global" {
				return unsupported(s.Line, "global statement unsupported in nested block")
			}
			switch s.Kind {
			case "if":
				if err := scanNested(s.Body); err != nil {
					return err
				}
				if err := scanNested(s.ElseBody); err != nil {
					return err
				}
			case "while":
				if err := scanNested(s.Body); err != nil {
					return err
				}
			case "foreach":
				if err := scanNested(s.Body); err != nil {
					return err
				}
			case "switch":
				for _, clause := range s.Body {
					if err := scanNested(clause.Body); err != nil {
						return err
					}
				}
			}
		}
		return nil
	}
	// Only check nested bodies, not top-level statements.
	var scanTop func([]ir.Stmt) error
	scanTop = func(ss []ir.Stmt) error {
		for _, s := range ss {
			switch s.Kind {
			case "if":
				if err := scanNested(s.Body); err != nil {
					return err
				}
				if err := scanNested(s.ElseBody); err != nil {
					return err
				}
			case "while":
				if err := scanNested(s.Body); err != nil {
					return err
				}
			case "foreach":
				if err := scanNested(s.Body); err != nil {
					return err
				}
			case "switch":
				for _, clause := range s.Body {
					if err := scanNested(clause.Body); err != nil {
						return err
					}
				}
			default:
				// Recurse into nested bodies of other stmt kinds.
				if err := scanTop(s.Body); err != nil {
					return err
				}
				if err := scanTop(s.ElseBody); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return scanTop(stmts)
}

// validateGlobalUsage checks that within a scope, a name in G (global-bound)
// is only used as a plain variable read or plain assignment target.
// It rejects isset, unset, array index, property access, foreach binding,
// array_push/pop target, and compound assignment on global-bound names.
func validateGlobalUsage(stmts []ir.Stmt, globals map[string]bool) error {
	if len(globals) == 0 {
		return nil
	}
	var scan func([]ir.Stmt) error
	scan = func(ss []ir.Stmt) error {
		for _, s := range ss {
			switch s.Kind {
			case "assign":
				if globals[s.Name] {
					// Plain assignment to global name — check RHS doesn't use
					// the global name in a disallowed way (it's fine as var read).
					if err := validateGlobalExpr(s.Expr, globals); err != nil {
						return err
					}
				}
			case "assign_index":
				// Array index on a global-bound variable is admitted (one level).
				if len(s.Target.Args) > 0 && s.Target.Args[0].Kind == "var" {
					if globals[s.Target.Args[0].Text] {
						// Check RHS doesn't use the global in a disallowed way.
						if err := validateGlobalExpr(s.Expr, globals); err != nil {
							return err
						}
						// Check key expression is valid.
						if err := validateGlobalExpr(s.Target.Args[1], globals); err != nil {
							return err
						}
					}
				}
			case "assign_append":
				if s.Target.Kind == "var" && globals[s.Target.Text] {
					// Array append on a global-bound variable is admitted.
					if err := validateGlobalExpr(s.Expr, globals); err != nil {
						return err
					}
				}
			case "assign_property", "assign_property_append":
				if s.Target.Kind == "property" && len(s.Target.Args) > 0 && s.Target.Args[0].Kind == "var" {
					if globals[s.Target.Args[0].Text] {
						return unsupported(s.Line, "property access on global-bound variable is unsupported")
					}
				}
			case "assign_nested_property", "assign_nested_property_append", "assign_nested_property_concat":
				if s.Target.Kind == "property" && len(s.Target.Args) > 0 && s.Target.Args[0].Kind == "var" {
					if globals[s.Target.Args[0].Text] {
						return unsupported(s.Line, "nested property on global-bound variable is unsupported")
					}
				}
			case "echo":
				if err := validateGlobalExpr(s.Expr, globals); err != nil {
					return err
				}
			case "return":
				if err := validateGlobalExpr(s.Expr, globals); err != nil {
					return err
				}
			case "if", "while":
				if err := validateGlobalExpr(s.Expr, globals); err != nil {
					return err
				}
				if err := scan(s.Body); err != nil {
					return err
				}
				if err := scan(s.ElseBody); err != nil {
					return err
				}
			case "foreach":
				// Check if the foreach key/value binds a global name.
				if globals[s.Value] {
					return unsupported(s.Line, "foreach value binding on global-bound variable is unsupported")
				}
				if s.Key != "" && globals[s.Key] {
					return unsupported(s.Line, "foreach key binding on global-bound variable is unsupported")
				}
				if err := validateGlobalExpr(s.Expr, globals); err != nil {
					return err
				}
				if err := scan(s.Body); err != nil {
					return err
				}
			case "switch":
				if err := validateGlobalExpr(s.Expr, globals); err != nil {
					return err
				}
				for _, clause := range s.Body {
					if err := scan(clause.Body); err != nil {
						return err
					}
				}
			case "array_push":
				if s.Target.Kind == "var" && globals[s.Target.Text] {
					return unsupported(s.Line, "array_push on global-bound variable is unsupported")
				}
				if err := validateGlobalExpr(s.Expr, globals); err != nil {
					return err
				}
			case "array_pop":
				if s.Target.Kind == "var" && globals[s.Target.Text] {
					return unsupported(s.Line, "array_pop on global-bound variable is unsupported")
				}
			case "call_stmt":
				if err := validateGlobalExpr(s.Expr, globals); err != nil {
					return err
				}
			case "list_assign":
				if err := validateGlobalExpr(s.Expr, globals); err != nil {
					return err
				}
				for _, local := range s.Target.Args {
					if local.Kind == "var" && globals[local.Text] {
						return unsupported(s.Line, "list assignment to global-bound variable is unsupported")
					}
				}
			case "preg_match_local":
				if globals[s.Name] {
					return unsupported(s.Line, "preg_match return on global-bound variable is unsupported")
				}
				if s.Target.Kind == "var" && globals[s.Target.Text] {
					return unsupported(s.Line, "preg_match output on global-bound variable is unsupported")
				}
			case "continue", "break":
				// No expression to check.
			}
			if err := scan(s.Body); err != nil {
				return err
			}
			if err := scan(s.ElseBody); err != nil {
				return err
			}
		}
		return nil
	}
	return scan(stmts)
}

// validateGlobalExpr checks that an expression only uses global-bound names
// in admitted contexts: plain variable reads, one-level index reads,
// isset_var/isset_index (one level), empty_var, and count.
// It rejects nested index, property access on globals, isset_path,
// isset_property, and isset_multi containing disallowed forms.
func validateGlobalExpr(e ir.Expr, globals map[string]bool) error {
	switch e.Kind {
	case "var":
		// Plain variable read of a global name is fine.
		return nil
	case "isset_var":
		// isset($g) on a global-bound variable is admitted.
		if len(e.Args) > 0 && e.Args[0].Kind == "var" {
			if globals[e.Args[0].Text] {
				return nil
			}
		}
	case "isset_index":
		// isset($g[k]) on a global-bound variable is admitted (one level).
		if len(e.Args) > 0 && e.Args[0].Kind == "var" {
			if globals[e.Args[0].Text] {
				// Check the key argument is valid.
				if len(e.Args) > 1 {
					if err := validateGlobalExpr(e.Args[1], globals); err != nil {
						return err
					}
				}
				return nil
			}
		}
	case "isset_path", "isset_property":
		// Multi-level isset paths and property isset on globals are rejected.
		if len(e.Args) > 0 && e.Args[0].Kind == "var" {
			if globals[e.Args[0].Text] {
				return unsupported(e.Line, "nested isset path on global-bound variable is unsupported")
			}
		}
	case "isset_multi":
		for _, arg := range e.Args {
			if err := validateGlobalExpr(arg, globals); err != nil {
				return err
			}
		}
		return nil
	case "empty_var":
		// empty($g) on a global-bound variable is admitted.
		if len(e.Args) > 0 && e.Args[0].Kind == "var" {
			if globals[e.Args[0].Text] {
				return nil
			}
		}
	case "index":
		// Array index read — one level on a global-bound variable is admitted.
		// Nested index (index of index) on a global is rejected.
		if len(e.Args) > 0 && e.Args[0].Kind == "var" {
			if globals[e.Args[0].Text] {
				// Check the key argument is valid.
				if len(e.Args) > 1 {
					if err := validateGlobalExpr(e.Args[1], globals); err != nil {
						return err
					}
				}
				return nil
			}
		}
		// If the base is itself an index rooted in a global, reject nested access.
		if len(e.Args) > 0 && e.Args[0].Kind == "index" {
			if len(e.Args[0].Args) > 0 && e.Args[0].Args[0].Kind == "var" {
				if globals[e.Args[0].Args[0].Text] {
					return unsupported(e.Line, "nested array index on global-bound variable is unsupported")
				}
			}
		}
	case "property":
		// Property access — check if root is a global-bound variable.
		if len(e.Args) > 0 && e.Args[0].Kind == "var" {
			if globals[e.Args[0].Text] {
				return unsupported(e.Line, "property access on global-bound variable is unsupported")
			}
		}
	default:
		// For all other expression kinds, recurse into args.
	}
	for _, arg := range e.Args {
		if err := validateGlobalExpr(arg, globals); err != nil {
			return err
		}
	}
	return nil
}

// GlobalNamesInProgram collects all global-bound names across all functions
// and main body. Main's G set includes names from all functions (main's
// same-named variables must share the cell).
func GlobalNamesInProgram(p ir.Program) (mainGlobals map[string]bool, funcGlobals map[string]map[string]bool) {
	mainGlobals = make(map[string]bool)
	funcGlobals = make(map[string]map[string]bool)

	// Collect function globals.
	for _, f := range p.Functions {
		names := GlobalNames(f.Body)
		key := strings.ToLower(f.Name)
		funcGlobals[key] = make(map[string]bool)
		for _, n := range names {
			funcGlobals[key][n] = true
			mainGlobals[n] = true // Main shares all function global names.
		}
	}

	// Add main-level globals.
	for _, n := range GlobalNames(p.Body) {
		mainGlobals[n] = true
	}

	return mainGlobals, funcGlobals
}

func globalEmptyExpr(e ir.Expr) bool {
	return e.Kind == "" && e.Text == "" && len(e.Args) == 0 && len(e.Params) == 0
}
