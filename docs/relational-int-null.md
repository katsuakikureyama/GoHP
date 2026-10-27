# Integer/null relational operators

The P1 slice of the fixed Gossip PHP8.3 parser gate needs `>` for offsets and
`<` for stack depth. This increment admits those two operators for actual
`int64` and `null` only. GoHP's existing PHP8.4 regression target checks the same
generic language contract independently; no new version or unrelated builtin
is introduced.

Two integers compare directly, without subtraction overflow or float precision
loss. When either operand is null, PHP compares both admitted operands as
booleans: null and zero compare equal; **null is less than every nonzero integer,
including negative integers**. Both runtime operands are checked first. Unknown
static values (including existing property reads) may reach that guard; known
strings, booleans, arrays and objects reject before generation. This deliberately
does not implement PHP's broader mixed-type ordering or exception diagnostics.

Precedence is arithmetic, concatenation, relational, strict identity, `&&`, `||`,
coalescing, then ternary. Prefix `!`/casts remain above the binary levels.
Unparenthesized relational chains reject; relation and identity may mix because
they occupy different precedence groups. Parentheses do not waive operand type
constraints. Each operand is evaluated once; no general PHP operand-order
guarantee is inferred from Go's lowering.

No `<=`, `>=`, `<=>`, `<>`, general equality or string/bool/float/array/object
ordering is admitted. Storage optimizations' allowlists are unchanged; scopes
using these new operators fall back to existing map-based local storage.
Original source and consumer pins are unchanged. This removes one admission
blocker, not the entire parser gate: list/switch, stack/object writes, regex
property inputs, JSON and original file loading still need separate work.

Catalog `php.syntax.relational-int-null` tracks compiler/negative/runtime checks
and the owned `relational.php` differential fixture. The native comparison
covers a 5×5 int/null matrix, int64 boundaries, precedence, count-controlled
loops, declared properties and single evaluation. CI separately records PHP
8.3.36 and 8.4.27 stdout/exit/quiet-stderr evidence. No PCRE backend, application
workflow or general exceptions/side-effect equivalence is certified here.

Capability: existing developer-local `gohp.cli.generate`; UI/WebMCP excluded.
No server operation, route, authentication, resource mutation or deployment.

References: [PHP comparisons](https://www.php.net/manual/en/language.operators.comparison.php)
and [operator precedence](https://www.php.net/manual/en/language.operators.precedence.php).
