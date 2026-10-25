# Bounded array_reverse contract

`array_reverse(array, bool preserve_keys = false)` accepts one or two positional
arguments, each evaluated once. Direct function names are case-insensitive.
The result is an array in reversed outer insertion order. Default/false replaces
integer keys with consecutive indexes from zero; true preserves their stored
values. String keys are always preserved, including noncanonical numeric strings.
Existing array construction normalizes canonical integer-string keys beforehand.

The input is unchanged. Nested arrays retain their own order and keys, and use
the existing detached value-copy model; this is not recursive reversal. Result
append metadata is rebuilt from the resulting keys. Preserving a maximum int64
key retains the append-overflow guard; reindexing allows subsequent append.
Negative-key append follows the active PHP 8.3 contract.

Known nonarrays and nonboolean flags, wrong arity, undefined reads, named args,
builtin redeclarations and scalar echo of the result fail before output. Dynamic
arguments are checked at runtime, including a flag passed with an empty array.
Unlike PHP weak coercion, integers/strings/null are not accepted as flags.
No new value/reference domain, internal pointer API, builtin callable lookup,
PHP exception/message equivalence or PHP 8.4 support is claimed.

`tests/fixtures/array_reverse.php` supplies the native PHP 8.3.36+ comparison.
Compiler/source-map, runtime copy/key/append and generated-binary rejection tests
are linked by `php.builtin.array_reverse` in the compatibility catalog. A fixture
PASS is bounded evidence, not full PHP compatibility or Gossip source adoption.

This extends developer-local `gohp.cli.generate`, which remains WebMCP `excluded`:
no browser operation, service route, authentication change or automatic submit.
Gossip pins/source evidence and deployment configuration are unchanged.

Reference: [PHP array_reverse manual](https://www.php.net/manual/en/function.array-reverse.php).
