# Bounded defined-variable empty: arrays and strings

`empty($arrayVariable)` returns an actual bool: true only when the defined
array variable has zero outer entries. A singleton containing null, false, zero,
an empty string or another empty array is not empty. Element values are not
inspected or converted. The array and its next-index metadata are unchanged.

The same defined-local construct also accepts actual strings. Exactly the empty
string and the one-byte ASCII string `"0"` are empty; `"00"`, `"0.0"`, `"-0"`,
whitespace, NUL-containing strings and all other byte strings are nonempty.
There is no trimming, numeric parsing, Unicode normalization or scalar coercion.
This follows the [official boolean conversion rules](https://www.php.net/manual/en/language.types.boolean.php#language.types.boolean.casting).

This is a deliberately limited subset of the
[PHP empty language construct](https://www.php.net/manual/en/function.empty.php),
not a general function or a replacement for PHP's undefined-address suppression.
Only one unparenthesized variable operand is admitted; spelling is case-insensitive.
The parser emits additive neutral `empty_var` IR with one variable operand and
empty Text/Params. Semantic validation requires an existing binding with array,
string or unknown type, and returns bool. Unknown arguments use
`rt.VariableEmpty`; every non-array/non-string value fails explicitly. The
legacy raw `empty_array` IR and `rt.ArrayEmpty` remain strictly array-only,
including rejecting actual strings. IR version remains0.1; no report schema,
source-map precision or storage optimization is expanded.
Direct generic-call IR, malformed operands and global redeclaration are rejected.

Undefined variables, branch-only bindings, non-string scalar/object operands, index/property
probes, literals, function calls, named arguments and callable lookup are outside
this subset. PHP can accept many of these forms; their GoHP rejection is intentional,
not PHP diagnostic equivalence. General warning/exception behavior is unimplemented.
Dead branches remain statically validated; valid short-circuit expressions stay lazy.

The owned `array_empty.php` fixture retains the array regression. The new
`string_empty.php` adds PHP's zero-string distinction, substring-derived locals,
typed/untyped/arrow helpers, reassignments, lazy guards and unchanged inputs.
Each active PHP8.3/8.4 target needs its independent pinned native result; one
target's PASS is not another's evidence. Compiler checks cover additive/legacy
IR, source maps and excluded forms; runtime tests
cover preservation and generated-code guards. The catalog entry remains partial.
This compiler prerequisite does not certify Gossip's unchanged WordPress functions:
new compiler pins and original-source evidence are separate. Bounded
[`array_key_last`](array-key-last-boundary.md) is a separate contract; guard the
empty case before using its result as an index. Later minors stay deferred.

Gossip's measured complete-original blocker before this increment is354,
`if (!empty($html))` with `$html` returned by substr. A compiler fixture or
isolated source-admission advance does not establish full parser generation,
execution or G83-PARSER acceptance; other stack/snapshot/tokenizer/JSON domains
still need their own complete-unit evidence.

Capability `gohp.cli.generate` remains WebMCP-excluded local developer tooling.
No new server operation, route, authentication, dependency, release or deployment.
