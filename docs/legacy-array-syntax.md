# Legacy array syntax boundary

`array(...)` is parsed as a language construct, not a callable builtin. Its
items use the same parser and `array`/`pair` IR as `[...]`. Runtime and codegen
are unchanged. This follows PHP's [array syntax](https://www.php.net/manual/en/language.types.array.php)
and [array construct](https://www.php.net/manual/en/function.array.php).

Included within the existing subset:

- Empty arrays, values with automatic keys, literal integer/string keys.
- Existing duplicate-key behavior, reads, count and ordered foreach.
- Nested `array(...)` and `[...]`, trailing commas and case-insensitive spelling
  of the `array` keyword. This is not generalized case-insensitive keyword support.
- Array literals returned by named, untyped functions in the closed subset.

This literal syntax introduces no new runtime semantics. References, unpacking,
dynamic construction keys, boolean keys, union, typed declarations and arbitrary PHP syntax
remain unsupported. Malformed delimiters, missing elements and a bare `array`
identifier must fail. Unknown ordinary function calls still fail closed.
One-level element writes are now covered separately by the
[existing-array assignment contract](array-write-boundary.md); nested/append
targets and reference semantics remain excluded.
Both literal forms share the corrected
[PHP 8.3 automatic-index behavior](array-index-boundary.md) for negative keys.

`tests/fixtures/legacy_arrays.php` is repository-owned differential input;
it does not vendor or rewrite WordPress. `TestDifferential/legacy_arrays` is
linked to `php.syntax.array-literal` in the catalog, with explicit negative
cases. Current native runs use PHP 8.3.36 or newer within 8.3; previous 7.4 and
older 8.3 results are historical evidence only. The CI ledger records the actual
oracle version. A negative-case PASS means rejection, not PHP compatibility.
Output/exit comparison does not prove warning, exception or side-effect parity.

The motivation is Gossip's unchanged WordPress `__return_empty_array()` unit,
which the prior GoHP pin rejected as an unknown function. Gossip owns that
original-source fixture and must independently pin this candidate, check report
hashes, build against the same runtime and compare its actual return value/type.
No WordPress source, dependency pin or full-core support claim is changed here.

This remains the local `gohp.cli.generate` capability, explicitly excluded from
WebMCP under the [integration contract](integration.md). No dependency, browser,
S2S, MCP, cookie, deployment or license change is part of this addition.
