# Literal-only negative decimal integers

The frontend accepts one `-` followed by a supported decimal integer token,
for example `-5`, `- 5` or `- /* comment */ 5`. It folds that form into the
existing `int` IR; `-0` becomes integer zero. No evaluator, new runtime operation
or IR version is introduced. The source position belongs to the minus token.

The magnitude must already fit the positive int64 token range: 0 through
9223372036854775807. The supported negative constants therefore stop at
-9223372036854775807. PHP's [integer rules](https://www.php.net/manual/en/language.types.integer.php)
describe negation separately from integer tokens and promote out-of-range
numbers to float. The positive token in `-9223372036854775808` exceeds int64;
GoHP rejects it before sign folding rather than changing that PHP float into an
integer. `-9223372036854775807 + -1` can still reach MinInt64 through existing
integer addition. Tests use a 64-bit PHP oracle; no 32-bit compatibility is claimed.

Negative constants are admitted wherever the existing subset admits integer
literals: values, arguments/returns, public scalar property defaults, literal
array keys, index reads, one-level writes and bounded array-element `isset`
probes. The isset key grammar still excludes parentheses and general expressions,
even where ordinary expression contexts allow them. They compose with existing `+`,
concatenation, strict equality and ternary expressions. Short/legacy arrays and
append retain their existing ordering, normalization and PHP 8.3 index rules.

This is not general unary negation: `-$value`, `-function_call()`, `-(1 + 2)`,
`-'5'` and repeated signs are excluded. Subtraction, decrement, unary plus,
nondecimal bases, numeric separators, floats and new coercions remain excluded.
Parentheses around the complete supported constant, as in `(-5)`, remain valid.

`TestDifferential/negative_integers` compares native stdout and successful exits.
Frontend/codegen checks cover IR values/source lines and generated statement
mapping; negative checks prove rejection, not PHP exception equivalence.
No stderr, warnings or general side-effect parity is claimed.

The existing local `gohp.cli.generate` capability remains WebMCP-excluded under
the [integration contract](integration.md). This syntax change introduces no
browser/server route, authentication, external resource write or deployment.
