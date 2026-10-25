# Array parameter declarations

Named global functions accept fixed positional `array $value` parameters, with
case-insensitive source spelling. Typed and untyped parameters may be mixed.
This is a partial language boundary, not general PHP type-declaration support.

The parser retains aligned canonical `Function.ParamTypes` metadata in IR 0.1.
Signature shape is validated before bodies, including forward calls. The entry
scope knows the argument is an array; known non-array direct arguments fail
before generation (even in dead code). Unknown values use `rt.ArrayParameter`
at function entry, after arity checking and before body statements. Actual
runtime arrays are cloned, including nested arrays and integer-key metadata;
null, scalar and object values are rejected, never converted to arrays.

Arguments retain existing left-to-right call evaluation. The entry annotation
does not constrain later assignments to the local variable. Caller arrays remain
unchanged by supported parameter writes. Existing array/bool/untyped returns
keep their separate contracts. Literal named one-argument `array_map` callbacks
use the same guard: each reached element must be an array, while empty mapping
does not invoke the callback. No new callable representation is introduced.

The [PHP argument reference](https://www.php.net/manual/en/functions.arguments.php)
describes argument evaluation and default by-value passing; the
[type reference](https://www.php.net/manual/en/language.types.declarations.php)
describes declaration checks. Runtime rejection here is a Go failure, not PHP
TypeError message, stack, catchability or error-side-effect equivalence.

Excluded: callable/scalar/nullable/union/intersection parameter types, defaults,
references, variadics, named arguments and typed methods. Return-type inference
and path-completeness proofs are not added. Direct IR accepts no noncanonical
annotations or misaligned metadata. Existing untyped IR remains valid.

`php.syntax.array-parameter-type` is partial. `array_parameter.php` compares
stdout and normal exit with the admitted PHP 8.3 oracle: empty/mixed/nested
arrays, forward calls, mixed parameter order, bool returns, append metadata,
copy isolation, single ordered argument evaluation, literal callbacks and local
reassignment. Unit/compiler/generated-binary negative checks cover rejection,
malformed IR and guards preceding body execution; PHP error equivalence and
external side effects remain unverified.

This change is based on bool-return PR #49, including its control-flow stack.
The independent PR #48 remains unchanged, not merged or closed here. Gossip
must separately pin and compare original WordPress code; callable parameters,
arrow callbacks and dynamic value/key invocation are still prerequisites.
Existing Gossip pins and its successful original-source count are unchanged.

Capability `gohp.cli.generate` remains explicitly WebMCP-excluded: local
developer source translation, no browser surface, service operation, auth route,
external write, auto-submit, deployment or automatic adoption.
