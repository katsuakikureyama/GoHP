# Exact scalar global-function parameters

Named global functions admit case-insensitive `int`, `string` and `bool`
parameter annotations, including mixtures with existing array, callable and
untyped parameters. Existing trailing **untyped** `= null` defaults remain
available. IR 0.1 reuses aligned, canonical `Function.ParamTypes`; no new shape
or parameter-passing mode is introduced.

The admitted values are actual int64, string and bool only. Known direct-call
mismatches reject before generation, including forward calls and dead branches.
Unknown arguments and exported Go host entries use `rt.ScalarParameter` before
the body. Existing arity and Execution guards remain. Literal `array_map`
callbacks use the same entry guard for each reached element; an empty input
does not call the callback. There is no callback-element type inference.

Annotations constrain entry, not the local variable's later assignments.
Scalar values pass by value; local reassignment leaves the caller unchanged.
Static type facts update on assignment and unknown later values retain their
consumer's guards. Array/callable mixtures keep their existing copy/guard rules.

## Owned scalar frames

A global function with **all** parameters annotated int/string/bool and an
entirely closed scalar body now qualifies for the existing
[generated ScalarFrame](generated-scalar-frames.md). Every argument is checked
before its first Write. The deferred Close is installed before initialization,
so a later parameter failure also releases already initialized locals. Nullary
selection remains valid. Mixed/untyped parameters, calls, array/object producers,
methods and arrows keep legacy maps; no caller-based or runtime guessing selects
storage. General frame/shared-array ownership and PHP references remain unfinished.

## Evidence and exclusions

`php.syntax.scalar-function-parameters` is partial. Its six catalog checks cover
native/generated success, canonical lowering/source maps, source/IR rejection,
generated runtime guards and an actual external Go module calling the host ABI.
`scalar_function_parameters.php` observes exact values, integer bounds, strings,
false/zero, local/caller isolation, loops/early returns, an untyped null suffix,
array copies, forwarding and a literal callback. PHP 8.3 and 8.4 must each provide
their own evidence; neither target inherits the other's PASS. The normal
comparison covers stdout, successful exit and empty stderr (`quiet-success/v1`),
not arbitrary diagnostic equality or reference/destructor behavior.

[PHP type declarations](https://www.php.net/manual/en/language.types.declarations.php)
permit scalar coercion by default. This deliberately narrower exact-value subset
does **not** implement that coercion or `strict_types` syntax. Typed defaults,
nullable/union/intersection/float/class types, typed ordinary methods/arrows,
references, variadics and named arguments remain excluded. Runtime failure is
not PHP TypeError text, catchability or invalid-call side-effect equivalence.
Existing array/bool return declarations and the separate
[constructor boundary](scalar-constructor-boundary.md) remain intact; no scalar
int/string return declaration is added.

The existing local `gohp.cli.generate` capability stays WebMCP-excluded. No new
CLI option, route, authentication, target policy, oracle environment, deployment
or Gossip adoption is included. New generated output requires the matching
runtime revision; previous valid IR and generated Go retain their contracts.
