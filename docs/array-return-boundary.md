# Bounded array return declarations

Named global functions may declare `: array` (case-insensitive in source).
The declaration is retained as canonical `Function.ReturnType = "array"` in
the additive IR 0.1 field; the empty string preserves existing untyped behavior.
Unknown IR type names are rejected rather than ignored. Typed class methods,
typed parameters, nullable/union/intersection types and other return types are
not part of this contract.

Every explicit return in a typed function, including returns inside `if`,
`elseif`, `else if`, `else` and `foreach`, is lowered through
`rt.ArrayReturn(value)`. It evaluates the value
once, requires an actual runtime Array, then clones its supported nested values
and preserves key/index metadata. The returned array does not alias caller
array storage under the existing copy semantics. Unknown parameter/call values
are checked at runtime; known non-array returns, including bare/null returns
and statically invalid dead branches, fail before generation.

Falling off a typed function is also guarded via `rt.ArrayReturn(nil)`, so it
cannot silently return null. This does not prove path completeness statically.
No branch-sensitive narrowing or inferred return type at call sites is added.
Untyped functions and methods retain their previous return behavior.

The [PHP type-declaration reference](https://www.php.net/manual/en/language.types.declarations.php)
requires a declared return type to be enforced. GoHP's rejection/panic is not
PHP TypeError text, stack, catchability or exception equivalence. No scalar
coercions, general type system, references, object alias or cyclic-array
semantics are added.

Runtime and codegen tests cover the guard, copy and rejected boundaries.
`TestDifferential/array_return` observes actual typed returns, empty/associative/
nested arrays, early returns, forward calls, one-time evaluation and caller
copy isolation against PHP 8.3 >=8.3.36. Native harness evidence measures stdout
and normal exits; generated runtime-negative tests separately require explicit
failure on unknown non-arrays or fallthrough. Neither is whole-PHP support.

The control-flow integration fixture additionally covers every selected arm,
typed callbacks, `isset`/`strcasecmp` composition, negative-key append metadata,
and caller/copy isolation. Invalid known returns in dead `else`/`elseif` arms
are still rejected statically; unknown returns in selected arms and empty-else
fallthrough are runtime-guarded. Each arm retains its independent input scope.

This unlocks the syntax needed by WordPress `wp_array_slice_assoc`; adopting
it in Gossip still requires an immutable compiler pin and original-source
test. Existing Gossip pins/evidence are not automatically upgraded.

The local `gohp.cli.generate` capability remains explicitly excluded from
WebMCP: developer-owned source translation, no browser/service action,
authentication route, external write or automatic submission.
