# Bounded isset addresses

`isset($array[$key])` checks a single element of an already-defined array
variable. The key is an integer/string literal or an already-defined variable
whose actual value is int64/string. The construct name is case-insensitive.
Direct negative decimal literals such as `isset($array[-5])` use the same
[literal-only negation boundary](negative-integer-boundary.md) as reads/writes:
one minus, magnitude 0 through 9223372036854775807, optional whitespace/comments,
and `-0` normalized to integer zero. No general key-expression parsing is added.
Missing keys and stored null return false. All other supported stored values,
including false, zero, empty strings and empty arrays, return true. Existing
PHP integer/string key normalization is shared with array reads and writes.

The parser preserves the indexed address in `isset_index` IR with exactly two operands
(root variable, key) and an empty Text field. It does not lower this to a normal
function call or eagerly read the indexed value. Semantic validation rejects
malformed IR, undeclared root/key bindings and known-invalid domains before Go
generation. Unknown function arguments are guarded by `rt.ArrayIsset` at runtime.
The probe neither changes the binding nor copies or modifies its array entries.
Boolean use in `if`, `&&`, ternary and function arguments uses existing semantics.

In addition, `isset($value)` accepts one already-defined bare variable and
returns false only for null. `isset($first, $second, ...)` accepts two or more
defined bare, indexed or separately admitted named-property addresses. The generated `&&` chain preserves
left-to-right short-circuiting, so a later runtime guard is not evaluated when
an earlier address is null/missing. A bare-variable address is represented as
`isset_var`; two-or-more addresses are represented as `isset_multi`. This does
not change the indexed-address key or root rules above.

## Nested path extension

The separately bounded nested form is `isset($root[$first][$second]...)`, with
a defined variable root and two-or-more literal integer/string or defined
variable keys. It uses `isset_path` rather than chaining ordinary array reads:
missing/null intermediates return false, while a non-array intermediate fails
the explicit runtime path guard. Final false/zero/empty-string/empty-array
values remain set. The existing one-level `isset_index` / `ArrayIsset` behavior
above is unchanged. See the complete [nested probe and coalescing contract](nested-array-probes.md).

## Deliberate exclusions

Named-property paths have a separate [property isset contract](property-isset.md),
including quiet nonobject/missing/null behavior. They do not extend the array
root/key or mixed array/property address rules described here.

This is not general PHP isset: undefined addresses, nested indexes outside the
documented `isset_path` form, mixed array/property paths, string offsets,
effectful root/key expressions, parenthesized
operands/keys, references or bool/null/float key coercions remain excluded.
Indexed roots still must be actual arrays; bare probes do not make undefined
variables legal. Direct IR calls named isset and global function redeclarations
are rejected. Variable/call negation, repeated signs, arithmetic key expressions
and parenthesized keys remain unsupported. The out-of-range
positive token in `-9223372036854775808` is rejected before sign folding;
existing int64 variables and numeric string normalization retain their behavior.
Dynamic domain rejection is a Go guard, not a claim of PHP
warning/exception equivalence. Broader behavior is described in the
[PHP isset reference](https://www.php.net/manual/en/function.isset.php).

## Evidence and next integration

Runtime tests cover presence/null/falsy distinctions, normalization, invalid
domains and input preservation. Compiler tests cover lowering and static/IR
rejections. Generated-code negative tests cover unknown argument domain guards.
`TestDifferential/array_isset` compares the indexed form with native PHP 8.3
>=8.3.36, including a key-selection loop using existing indexed writes. The
native harness checks stdout and normal exits only; general stderr, exceptions
and external side effects are not compared.

`TestDifferential/isset_forms` separately covers defined bare variables,
multi-address left-to-right short-circuiting and the retained indexed behavior.
`TestDifferential/isset_negative_keys` covers direct negative keys, null/falsy
values, canonical zero, the admitted integer boundary, short-circuiting and
probe non-mutation. Compiler checks preserve int IR/source mapping and reject
out-of-bound syntax. The shared literal-folding helper adds no runtime or IR
operation; existing negative-literal checks continue to cover other contexts.

WordPress `wp_array_slice_assoc` is a candidate consumer, not verified by this
synthetic fixture. Its original `: array` return declaration now has a separate
[checked contract](array-return-boundary.md); accepting syntax alone is not an
integration. No Gossip source, compiler
pin, original excerpt, corpus or prior evidence is changed by this feature.

This remains the WebMCP-excluded local `gohp.cli.generate` capability: local
source translation only, no new service operation, authentication route,
browser action, external write or automatic submission.
