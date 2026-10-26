# Bounded array_slice contract

`array_slice(array, int offset, ?int length = null, bool preserve_keys = false)`
accepts two to four positional arguments. Direct names are case-insensitive;
each supplied argument is evaluated once. The signature describes this builtin,
not new support for nullable parameter declarations in user functions.

Offset selects an insertion-order position, not a key. Negative offsets count
from the end and clamp at the beginning. Offsets beyond the end select nothing.
Omitted/null length selects the remainder; positive length selects at most that
many entries; zero selects nothing; negative length excludes that many trailing
positions. Signed int64 extremes are handled without overflow.

String keys always remain. Integer keys are renumbered from zero by default or
with false; true preserves them. Canonical integer-string keys are normalized
by existing array construction. Result outer/nested array storage is detached,
with append metadata rebuilt from selected keys, not inherited from excluded
entries. A selected preserved MaxInt64 key still blocks subsequent append;
excluding it does not. Empty/string-only results append at zero. Negative integer
keys retain the PHP 8.3 negative-index append behavior.

Actual arrays, int64 offsets, null/int64 lengths and bool flags are required.
Known invalid domains fail before output; dynamic domains fail at runtime even
for empty/out-of-range selections. No weak coercion, named arguments, unpacking,
builtin callable lookup, references or PHP exception/message equivalence is added.
The result has array semantic kind. Only PHP target 8.3 is validated, using the
native oracle >= 8.3.36; later targets remain separate future work.

The seven checks in catalog entry `php.builtin.array_slice` cover native PHP
differential output, compiler/source-map/defaults, source rejection, runtime
bounds/copy/append behavior and generated runtime guards. Status stays `partial`.
No Gossip pin or original-source certification follows from these checks.

Capability: existing developer-local `gohp.cli.generate`, WebMCP `excluded`.
No new browser operation, route, authentication, deployment or automatic submit.

Reference: [PHP array_slice manual](https://www.php.net/manual/en/function.array-slice.php).
