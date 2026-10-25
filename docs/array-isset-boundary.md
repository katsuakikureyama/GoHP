# One-level array-element isset

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

The parser preserves the address in `isset_index` IR with exactly two operands
(root variable, key) and an empty Text field. It does not lower this to a normal
function call or eagerly read the indexed value. Semantic validation rejects
malformed IR, undeclared root/key bindings and known-invalid domains before Go
generation. Unknown function arguments are guarded by `rt.ArrayIsset` at runtime.
The probe neither changes the binding nor copies or modifies its array entries.
Boolean use in `if`, `&&`, ternary and function arguments uses existing semantics.

## Deliberate exclusions

This is not general PHP isset: no simple-variable probe, multiple operands,
undefined/null/scalar roots, nested indexes, object properties, string offsets,
effectful root/key expressions, parenthesized operands/keys, references or
bool/null/float key coercions. Direct IR calls named isset and global function
redeclarations are rejected. Variable/call negation, repeated signs, arithmetic
key expressions and parenthesized keys remain unsupported. The out-of-range
positive token in `-9223372036854775808` is rejected before sign folding;
existing int64 variables and numeric string normalization retain their behavior.
Dynamic domain rejection is a Go guard, not a claim of PHP
warning/exception equivalence. Broader behavior is described in the
[PHP isset reference](https://www.php.net/manual/en/function.isset.php).

## Evidence and next integration

Runtime tests cover presence/null/falsy distinctions, normalization, invalid
domains and input preservation. Compiler tests cover lowering and static/IR
rejections. Generated-code negative tests cover unknown argument domain guards.
`TestDifferential/array_isset` compares a synthetic fixture with native PHP 8.3
>=8.3.36, including a key-selection loop using existing indexed writes. The
native harness checks stdout and normal exits only; general stderr, exceptions
and external side effects are not compared.

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
