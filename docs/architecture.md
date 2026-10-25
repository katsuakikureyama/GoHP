# Architecture / ADR 0001 (proposed)

Decision: implement a deliberately small, independent source-to-source pipeline.
No upstream code is copied or forked. This creates a verifiable baseline, not a
replacement for a complete PHP frontend. Revisit frontend reuse before expanding
the grammar substantially; upstream evidence is in `upstream.md`.

```
PHP bytes -> frontend lexer/parser -> versioned semantic subset IR
          -> closed-subset validation + Go codegen -> Go source
          -> Go compiler -> native executable + compatibility runtime
```

`internal/frontend` owns tokens, source lines and PHP-version precedence.
`internal/ir` is independent of syntax tokens and contains expressions, statements,
functions and classes. IR version is `0.1`; reject incompatible IR versions.
`internal/semantic` validates IR shape and rejects statically known type/operation
mismatches before generation. `internal/codegen` performs the remaining
name/arity/scope checks and lowers IR into direct Go calls, loops and structs.
`runtime` guards values that remain dynamic and implements supported operations.
The CLI is the initial versioned adapter boundary, not a public Go API promise.
No WordPress hooks or special compiler behavior exist in the core.

The optional `Function.ReturnType` admits canonical empty/untyped, `array` or
`bool` on named global functions. Validation propagates that context through
all conditional and foreach bodies; code generation guards explicit returns
and fallthrough. The bool boundary admits actual bool only, not weak scalar
coercions. No call-site type inference, typed method/parameter support or
path-completeness analysis is implied. Existing IR version `0.1` is retained.

The additive internal `Stmt.ElseBody` field retains IR version `0.1`, with nil
for no else and a non-nil empty slice for an explicit empty else. Only if statements
may carry it. Both arms are validated from independent copies of the pre-if scope;
new bindings never escape, and any changed existing type becomes unknown.
Codegen uses Go if/else with one guarded condition evaluation. This is a closed
subset extension, not a public serialized IR compatibility promise; CLI/report
schemas and runtime interfaces are unchanged.
The frontend normalizes braced `elseif` and `else if` chains into these same
nested if/ElseBody nodes, retaining each condition's keyword source line. No
new IR fields, validator branches, codegen paths or runtime functions are needed;
the existing independent scopes and guarded lazy execution apply recursively.

The one-array `array_map` subset resolves a literal one-argument user function
at compile time and emits a direct function adapter. Both semantic and codegen
registries include all declarations before bodies, admitting forward references
and ordinary recursion without a dynamic callable engine. Runtime mapping
preserves ordered keys and the established nested-array value-copy boundary.

The additive `ternary` IR kind keeps IR version `0.1` and uses ordered `Args`
for condition, yes and no (no schema-field change). It lowers through a local
closure so the condition is evaluated once and the unselected arm is lazy;
semantic validation still checks all three operands and retains an unknown type
for mixed branch types. The string-only `stripslashes` runtime boundary is
byte-oriented and remains separate from WordPress recursive helpers.

## Values and semantics

Values currently use Go `any`: int64, byte strings, bool, null, ordered arrays and
native generated class pointers. PHP arrays have insertion order, numeric-string
key normalization, last-value replacement without reordering and independent
nested copies on assignments, arguments and returns. This eager copy models value
semantics for supported array operations, including functional root-binding
replacement for one-level writes; it does not implement
PHP references or production copy-on-write optimization. Object identity remains
pointer identity. Class methods are real generated Go methods; interface assertions
provide method dispatch, not an AST/opcode evaluator. Only public scalar property
defaults, default construction and property reads are implemented.

Runtime arithmetic is restricted to integral coercions. Overflow, missing keys,
invalid receiver types and unimplemented coercions terminate nonzero. They are
not PHP-compatible exception/warning emulation. Generated lines have no complete
source map yet; compile-time diagnostics include original PHP line numbers.

## Version model

Only the `8.3` target label is accepted by the active compiler. It denotes the
PHP 8.3 language series, while native differential execution enforces the
separate PHP 8.3.36-or-newer oracle floor. This is not a full-language support
claim. PHP 7.4 is retained only in historical evidence. PHP 8.4 and later need
their own policy entry, semantic table, native reference runner and corpus
before a target label may be accepted; they must not silently inherit 8.3
results.

## Next decisions

Adopt a full version-aware parser adapter; extend validation as the type model
grows. The [one-level assignment design](array-write-boundary.md) uses an explicit
root-binding target and functional array replacement, not reference cells.
Append uses a distinct `assign_append` statement with a variable-only target;
it shares runtime copy/index semantics without admitting empty-bracket reads.
Nested lvalues/references and closures require a further location/cell design.
Add scalar types and conversions with warning/exception differential
fixtures; then inheritance and include/autoload boundaries. Source maps, structured
runtime diagnostics and persistent compatibility reports remain open work.
