# Bounded array/string-variable empty

`empty($arrayVariable)` returns an actual bool: true only when the defined
array variable has zero outer entries. A singleton containing null, false, zero,
an empty string or another empty array is not empty. Element values are not
inspected or converted. The array and its next-index metadata are unchanged.

This is a deliberately limited subset of the
[PHP empty language construct](https://www.php.net/manual/en/function.empty.php),
not a general function or a replacement for PHP's undefined-address suppression.
Only one unparenthesized variable operand is admitted; spelling is case-insensitive.
The parser emits `empty_value` with one variable operand and empty Text. Semantic
validation requires an existing binding and array, string or unknown type, and
returns bool. Unknown arguments are guarded by `rt.EmptyValue`; other actual
domains fail explicitly. The legacy manually supplied `empty_array` IR and
`rt.ArrayEmpty` API remain strict array-only; neither is silently broadened.
Direct generic-call IR, malformed operands and global redeclaration are rejected.

## Measured string extension

For actual byte strings, only the empty string and exactly `"0"` are empty.
`"00"`, `"0.0"`, `" 0"`, `"0 "`, whitespace, embedded NUL, UTF-8 and arbitrary
other bytes are nonempty. There is no numeric parsing, trimming, scalar coercion
or Unicode normalization. This admits the original parser's substring-derived
`$html` local without changing its source. That original caller still requires
a separately pinned whole-closure probe; this fixture alone does not certify it.

The string/array domain may change between calls or local assignments; each
reached invocation checks the actual value. Existing logical short circuit is
preserved. Generated `empty_value` scopes remain in the legacy storage lane;
scalar-frame/shared-value allowlists are not expanded.

Undefined variables, branch-only bindings, nonstring scalar/object operands, index/property
probes, literals, function calls, named arguments and callable lookup are outside
this subset. PHP can accept many of these forms; their GoHP rejection is intentional,
not PHP diagnostic equivalence. General warning/exception behavior is unimplemented.
Dead branches remain statically validated; valid short-circuit expressions stay lazy.

The owned `array_empty.php` and `empty_string_values.php` fixtures are registered
for separate pinned PHP8.3.36/PHP8.4.27 comparisons. Compiler checks cover IR/file-aware
source maps and excluded forms; runtime tests
cover preservation and generated-code guards. The catalog entry remains partial.
This compiler prerequisite does not certify Gossip's unchanged WordPress functions:
new compiler pins and original-source evidence are separate. Bounded
[`array_key_last`](array-key-last-boundary.md) is a separate contract; guard the
empty case before using its result as an index. No evidence is copied across targets.

Capability `gohp.cli.generate` remains WebMCP-excluded local developer tooling.
No new server operation, route, authentication, dependency, release or deployment.
