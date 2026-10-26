# PHP 8.4 bounded array_any

For target `8.4`, `array_any($array, 'declared_name')` accepts exactly one
actual array and a literal name of a declared global function. The callback
must have exactly two required, untyped value parameters and an explicit
`: bool` return declaration. Names are case-insensitive and forward declarations
are resolved statically. Unknown array inputs are checked at runtime.

The [PHP contract](https://www.php.net/manual/en/function.array-any.php) and
[implemented RFC](https://wiki.php.net/rfc/array_find#array_any) pass value then
key, stop at the first true callback result and return false for no match.
This subset visits stored insertion order and normalized int/string keys,
clones each callback array value through the existing value-copy boundary,
and accepts actual bool results only. Objects retain existing identity semantics;
this is not a deep object-copy or general side-effect guarantee. Empty arrays
invoke no callback, but invalid callback declarations are still rejected.

The generated adapter calls the selected function directly. It preserves the
request-local execution context when `strtok` is used. The builtin is not added
to the pure scalar-frame allowlist. It introduces neither a callback registry
nor an interpreter. Target `8.3` retains ordinary user-function lookup for this
name; absent declarations still fail. Target `8.4` continues to reject user
redeclarations of its builtin name.

Dynamic callback values, arrows, method/static callbacks, builtin callbacks,
optional/typed/reference parameters, one-parameter callbacks, missing bool
return declarations and truthiness coercion are deliberately outside this
bounded contract, even when some are valid native PHP. Existing bool-return
guards reject unknown non-bool values and fallthrough; their panic is not a PHP
TypeError-equivalence claim. Other new array search functions remain unsupported.

The `php84_array_any` fixture is required only for 8.4. The separate 8.3-only
user-declaration fixture stays intact. Each target needs its own fixed-oracle
evidence; an excluded fixture is not-measured, not a borrowed PASS. Differential
evidence covers stdout, normal exits and quiet stderr, not arbitrary diagnostics,
exceptions or side effects. Callback visit markers are bounded stdout evidence.

Existing capability `gohp.cli.generate` remains WebMCP **excluded**: local
developer source generation with the existing input/output contract, no browser
surface, remote operation, authentication route, auto-submit or publication.
Target/catalog configuration changes no authorization boundary. No Gossip pin,
deployment, release or additional PHP target is introduced.
