# Bounded array-variable empty

`empty($arrayVariable)` returns an actual bool: true only when the defined
array variable has zero outer entries. A singleton containing null, false, zero,
an empty string or another empty array is not empty. Element values are not
inspected or converted. The array and its next-index metadata are unchanged.

This is a deliberately limited subset of the
[PHP empty language construct](https://www.php.net/manual/en/function.empty.php),
not a general function or a replacement for PHP's undefined-address suppression.
Only one unparenthesized variable operand is admitted; spelling is case-insensitive.
The parser emits `empty_array` with one variable operand and empty Text. Semantic
validation requires an existing binding and an array or unknown type, and returns
bool. Unknown arguments are guarded by `rt.ArrayEmpty`; nonarrays fail explicitly.
Direct generic-call IR, malformed operands and global redeclaration are rejected.

Undefined variables, branch-only bindings, scalar/object operands, index/property
probes, literals, function calls, named arguments and callable lookup are outside
this subset. PHP can accept many of these forms; their GoHP rejection is intentional,
not PHP diagnostic equivalence. General warning/exception behavior is unimplemented.
Dead branches remain statically validated; valid short-circuit expressions stay lazy.

The owned `array_empty.php` fixture compares stdout and normal exits with PHP 8.3
>=8.3.36. Compiler checks cover IR/source maps and excluded forms; runtime tests
cover preservation and generated-code guards. The catalog entry remains partial.
This compiler prerequisite does not certify Gossip's unchanged WordPress functions:
new compiler pins and original-source evidence are separate. Bounded
[`array_key_last`](array-key-last-boundary.md) is a separate contract; guard the
empty case before using its result as an index. PHP 8.4+ needs separate validation.

Capability `gohp.cli.generate` remains WebMCP-excluded local developer tooling.
No new server operation, route, authentication, dependency, release or deployment.
