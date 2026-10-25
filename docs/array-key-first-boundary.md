# First insertion-order array key

`array_key_first(array)` accepts exactly one evaluated actual array, using
case-insensitive direct calls. It returns the first outer entry's stored int64
or string key, or canonical null if empty. It does not return the lowest key,
the first value, or the last key. Zero and empty-string keys remain distinct
from null. See [PHP's contract](https://www.php.net/manual/en/function.array-key-first.php).

The input, values and next-index metadata are unchanged. Replacement of an
existing key preserves its position; appending to a nonempty array preserves
the first key. Numeric-string normalization uses the existing array rules.
No values are traversed or converted.

The argument/domain checks and int|string|null result model match
[`array_key_last`](array-key-last-boundary.md): known nonarrays fail statically,
dynamic nonarrays fail at runtime, and the result keeps unknown semantic kind
with downstream guards. Guard the empty case before using the key as an index.
No union/narrowing inference, internal array-pointer API, named arguments,
builtin callable lookup or PHP warning/exception equivalence is introduced.

The owned native fixture compares stdout and normal exits with PHP 8.3 >=8.3.36,
including first versus last, replacement/append, actual key types, callback
composition and guarded first-value access. Unit tests also cover integer
extremes, binary keys and unchanged metadata. The catalog remains partial.
This is not the `array_first` value-returning polyfill; prior Gossip evidence is
not repinned or reused as evidence for this revision. PHP 8.4+ remains separate.

`gohp.cli.generate` remains WebMCP-excluded local developer tooling. No new
server operation, authentication route, dependency or deployment is introduced.
