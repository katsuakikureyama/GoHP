# Shared array key/index policy (PHP 8.3 and 8.4 targets)

Both supported literal forms (`[]` and `array(...)`) assign an omitted key
after the highest previous integer key, including when that key is negative.
For example, `['-5' => 'a', 'b']` has integer keys `-5`, `-4`. String-only
prefix entries do not force the first integer key or its successor to zero.
An implicit entry with no previous integer key still starts at zero.

This is the PHP 8.3 behavior described in the
[PHP array manual](https://www.php.net/manual/en/language.types.array.php).
The [PHP 8.3.36 hash implementation](https://github.com/php/php-src/blob/php-8.3.36/Zend/zend_hash.c)
also distinguishes an uninitialized next index from a real zero next index.
GoHP represents that distinction explicitly; it does not embed Zend code.

`internal/arraykey` is the shared dependency-leaf policy for `runtime.Array` and
[internal ScalarArray](shared-scalar-array-storage.md). `Normalize` accepts only
exact int64/string keys; decimal strings convert only when they round-trip to
the same int64 text. Leading plus/zero, negative zero, whitespace, non-decimal
forms and out-of-range integers remain string keys. This preserves the existing
runtime subset; it does not add PHP's null/bool/float key coercions.

`Index` is copyable value state: `Next` reads the candidate or rejects overflow,
while `Observe` records an inserted integer. Reading a candidate does not
consume it. The container owns this state, independently of reference cells.
ScalarArray validates both values and available indices before separation, so
failed operations cannot alter later alias/copy behavior.

`NewArray` and one-level `ArraySet` share integer-index bookkeeping. Lower
keys and replacement of existing keys never rewind it. `Clone` and the bounded
single-array `ArrayMap` preserve this state. `ArrayKeys` and `ArrayValues`
start a fresh zero-based sequence, or an uninitialized index for empty output.
The int64 minimum key can advance safely; the existing explicit maximum key
is accepted, but a subsequent automatic key fails explicitly instead of wrapping.
That rejection is not a promise of PHP exception/message equivalence.
Slice/reverse/merge record their resulting integer keys through the same policy.
ScalarArray also tracks Write/Bind/Reference insertion and Append, copying index
state when shared storage separates; it is not connected to generated arrays.

`TestDifferential/negative_array_keys` observes key type, order and value using
the supported PHP syntax, linked to `php.syntax.array-literal` in the catalog.
That fixture retains canonical negative numeric strings to exercise key
normalization. Direct negative constants are now covered by the separate
[literal-only negation contract](negative-integer-boundary.md) and native fixture.
Runtime regression tests also check write/copy/map/
reindex bookkeeping. The separate `TestDifferential/array_append` now observes
that state through supported one-level append statements.
Native runs compare stdout and normal exits only; stderr, exceptions and side
effects are not checked by this fixture.

One-level `$a[] = ...` on an existing array is covered by the
[assignment boundary](array-write-boundary.md). Unset, references, array union,
dynamic literal keys and additional key coercions remain excluded.
The active PHP 8.3 and 8.4 jobs measure existing array fixtures independently.
Sharing this implementation is not a blanket PHP 8.x compatibility claim;
later targets still need their own admission policy and evidence. The original
negative-index behavior change is documented in the
[PHP 8.3 migration guide](https://www.php.net/manual/en/migration83.incompatible.php).
This is a correction inside the existing local `gohp.cli.generate` capability,
still WebMCP-excluded under the [integration contract](integration.md); no new
route, authentication, external resource write or deployment is introduced.
