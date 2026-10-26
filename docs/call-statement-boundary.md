# Bounded standalone call statements

Direct calls `name(args);`, bare-variable method calls `$object->name(args);`
and generated-arrow variable calls `$callback(args);` evaluate the supported
call once and discard only its result. No echo, new variable binding, extra
clone or boolean conversion is introduced. Argument, callee, return-type and
runtime domain checks still apply even when nobody uses the result.

Named functions and builtins retain their existing arity and value domains.
Method/callable receivers must be defined bare variables, with the existing
generated-class/arrow validation and runtime guards. This does not admit foreign
objects, string callables or dynamic method names. Calls work in function,
constructor, method, braced branch, foreach and while bodies; execution-aware
lowering propagates the same strtok context. Dead calls after break/continue
are validated but not emitted, just like other dead statements.

IR v0.1 adds `call_stmt` containing only `Expr` (call, method or invoke).
Names, assignment targets, bindings and bodies are forbidden. Codegen emits
`_ = <existing call expression>` with the PHP statement source map. Existing
assignment/index/append/property statement parsing is preserved.

This is not general expression-statement support. Literal/variable reads,
binary/ternary expressions, standalone construction, parenthesized calls,
chained receivers, nullsafe/static calls, assignments inside arguments and
unbraced conditionals remain outside this increment. Anonymous closures,
references, destructors, warning/exception equivalence and typed ordinary methods are
not added. Discarding an object result does not add a PHP destructor model.

## Why this increment

The unchanged WordPress `wp_render_elements_class_name` caller uses
`$tags->add_class($matched_class_name);` at elements.php:303. This is a generic
statement-shape dependency, not a reason to special-case that class or create
a substitute HTML implementation. The [render dependency audit](render-dependencies.md)
keeps this syntax separate from actual HTML behavior and consumer adoption.

Catalog `php.syntax.call-statement` remains partial with differential,
compiler/source-map, source/IR rejection and generated runtime-guard checks.
Owned effect/recording fixtures are not unchanged WordPress execution. Only PHP
target 8.3 is validated (oracle >=8.3.36), not later PHP targets.

Capability: developer-local `gohp.cli.generate`, WebMCP `excluded`. No new
browser/server/MCP operation, route/authentication, external write or deployment.

Reference: [PHP expressions and statements](https://www.php.net/manual/en/language.expressions.php).
