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

## Values and semantics

Values currently use Go `any`: int64, byte strings, bool, null, ordered arrays and
native generated class pointers. PHP arrays have insertion order, numeric-string
key normalization, last-value replacement without reordering and independent
nested copies on assignments, arguments and returns. This eager copy models value
semantics for the supported immutable-array operations; it does not implement
PHP references or production copy-on-write optimization. Object identity remains
pointer identity. Class methods are real generated Go methods; interface assertions
provide method dispatch, not an AST/opcode evaluator. Only public scalar property
defaults, default construction and property reads are implemented.

Runtime arithmetic is restricted to integral coercions. Overflow, missing keys,
invalid receiver types and unimplemented coercions terminate nonzero. They are
not PHP-compatible exception/warning emulation. Generated lines have no complete
source map yet; compile-time diagnostics include original PHP line numbers.

## Version model

Only `7.4` and `8.3` target labels are accepted. The main grammar is their shared
subset. Concatenation/addition precedence changes with the target. These labels
are not full-language support claims. PHP 5.6 and other 7.x/8.x minors require
separate semantic tables and native reference runners before acceptance.

## Next decisions

Adopt a full version-aware parser adapter; extend validation as the type model
grows. Design explicit lvalue/reference cells before array writes or
closures. Add scalar types and conversions with warning/exception differential
fixtures; then inheritance and include/autoload boundaries. Source maps, structured
runtime diagnostics and persistent compatibility reports remain open work.
