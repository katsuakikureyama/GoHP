# Capture-free arrows and generated callable invocation

This partial boundary admits `fn($value, $key) => expression` (case-insensitive
`fn`, zero or more fixed untyped positional parameters), global function
`callable` parameters and `$callback(arguments)` expressions. Only generated
arrow values satisfy that callable contract. It is not PHP's general callable
resolver. Standalone expression statements remain outside the existing grammar.

Arrows have exactly one expression and a fresh parameter-only scope. Outer
variables, including `$this`, are rejected even in dead arms; captures are not
silently omitted. Nested arrows are permitted only when each independently has
no free variables. Declared global functions can be called normally from the
body, including forward references. Parameter/return annotations, defaults,
references, variadics, static arrows and anonymous `function` syntax are excluded.

IR 0.1 adds arrow-only `Expr.Params`; `arrow.Args` has one body, and `invoke.Args`
has a variable receiver followed by arguments. `Function.ParamTypes` additionally
admits canonical `callable`. Malformed shapes/types and known wrong entry values
fail before output. Codegen rejects free variables against its fresh scope.
Runtime `CallableParameter` accepts only the opaque generated wrapper before a
named function body executes. Strings (even valid function names), arrays,
invokable objects, raw Go functions and null are rejected. Arrows may be held in
variables/arrays or returned untyped; that does not permit arbitrary values to
be invoked. Direct expression/index/method invocation is not added.

`Invoke` checks callable identity and exact argument count before running the
arrow body. Arguments are evaluated once in order; nested arrays are copied at
invocation and results are copied on return. Wrapper identity is immutable and
preserved by Clone. Missing and extra arguments both fail in this subset even
where PHP accepts extras. Guards occur after argument expression evaluation;
PHP error evaluation order/message/TypeError/catchability are not claimed.
Callable values remain outside supported truthiness, scalar conversion, strict
comparison and arithmetic domains, with existing static/runtime rejection.
`array_map` continues to accept only its literal named one-argument callback form.

PHP's [arrow reference](https://www.php.net/manual/en/functions.arrow.php) describes
automatic by-value capture and richer signatures; those are deliberately not
implemented here. Its [callable reference](https://www.php.net/manual/en/language.types.callable.php)
also includes string, method and invokable-object representations, which are
outside this generated-arrow-only contract.

The existing stable `php.syntax.closure` catalog entry becomes partial and keeps
the anonymous-function rejection check. Owned `arrow_callable.php` observes
value/key order, sparse/negative/string keys, empty arrays, short-circuit markers,
mixed array/callable/untyped arguments, forwarding/reuse, zero/nested arrows and
caller preservation. Native comparison covers stdout/normal exit, not general
stderr/exceptions/external side effects. Separate tests cover malformed IR,
capture/domain rejection, arity, copy isolation and pre-body entry guards.

## Integration and adoption

This branch incorporates #51 (`918c4eb47b4d7070ed24f26bcfd5b290030bb4e2`) and #50
(`58943797038a1c87215acef265fb10951d94ec7e`) preserving both histories. The fixture
list conflict retains array_parameter, bool_return and bool_return_composition;
parent PASS reports are not evidence for the new merged configuration.

The owned all-match fixture is not a replacement for WordPress's original
helper/polyfill. Gossip must separately pin this complete compiler revision,
extract unchanged source and compare its native reference corpus before adoption.
Its current pins, 17-original/149-case successful integration and rejected-candidate
evidence are unchanged. No main merge, parent PR closure, deployment or release.

Capability `gohp.cli.generate` remains explicitly WebMCP-excluded: internal local
developer translation, no browser/API operation, server resource, authentication
route, external write or auto-submit. No dependency or PHP target/oracle change.
