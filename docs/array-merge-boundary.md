# Bounded array_merge contract

`array_merge(array ...arrays)` accepts zero or more explicitly supplied positional
array arguments. Each argument expression is evaluated once; direct names are
case-insensitive. Zero arguments return an empty array. A single argument is
still copied and its integer keys are renumbered.

Inputs are traversed in argument and insertion order. Integer-keyed entries are
appended under consecutive keys starting at zero, even when input integer keys
repeat, are negative or reach int64 limits. Stored string keys are preserved;
later duplicate string keys overwrite the value at its original position.
Canonical integer-string keys have already been normalized by array construction;
noncanonical keys such as `'02'` remain strings.

The result uses detached outer/nested array storage and does not mutate input
contents or append metadata. Later nested values replace earlier values as a
whole; this is not recursive merging or the array-union `+` operator. Result
append metadata is rebuilt from result keys rather than inherited from inputs.
String-only and empty outputs start appending at integer key zero.

Known nonarray arguments are rejected before output, including later positions
and dead branches. Dynamic arguments are checked at runtime. The result has
array semantic kind, so scalar echo is rejected. Named arguments, argument
unpacking, builtin callable lookup, general variadic PHP function declarations,
new value/reference types and PHP exception/message equivalence are not added.
Only the existing PHP 8.3 target is validated (native oracle >= 8.3.36).

`tests/fixtures/array_merge.php` observes the bounded behavior against native PHP.
Compiler/source-map, runtime copy/key/append and generated rejection tests are
linked in catalog entry `php.builtin.array_merge`; status remains `partial`.
Gossip source compatibility needs separate pin/evidence and is not inferred.

This remains inside developer-local `gohp.cli.generate`, WebMCP `excluded`:
no browser operation, service route, authentication change or automatic submit.

Reference: [PHP array_merge manual](https://www.php.net/manual/en/function.array-merge.php).
