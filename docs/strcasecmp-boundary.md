# String-only ASCII case-insensitive comparison

`strcasecmp(left, right)` accepts exactly two actual strings and returns an
integer. Function names are case-insensitive; redeclaring the builtin is
rejected. Known non-string arguments (including dead branches) fail before
generation. Unknown parameters/call results are admitted only with the runtime
two-string guard. There is no scalar coercion or object-to-string conversion.

ASCII A–Z bytes fold to a–z; all other bytes, including NUL and UTF-8 bytes,
remain unchanged. Comparison continues through NUL and uses neither locale
nor Unicode case folding. For PHP 8.3 the result is the first unequal folded
byte difference, or -1/0/1 for unequal/equal lengths after an equal prefix.
This matches the selected PHP 8.3 implementation, not a guarantee that a
particular nonzero magnitude is portable to every PHP version.

References: the [PHP manual](https://www.php.net/manual/en/function.strcasecmp.php)
defines binary-safe ASCII-only comparison and warns that only the sign of a
nonzero result is portable. The selected
[PHP 8.3.36 implementation](https://github.com/php/php-src/blob/php-8.3.36/Zend/zend_operators.c)
(`zend_binary_strcasecmp`) supplies the bounded numeric behavior above.

Generated calls evaluate each argument once. Owned native fixtures observe
numeric results, strict integer comparisons, byte/prefix distinctions,
function/index arguments, operand markers and predicate/short-circuit
composition. Runtime unit tests also exercise NUL, high bytes and both operand
guards; generated-binary negative tests require explicit guard failure.
The native harness observes stdout/normal exit, not general stderr, PHP
warning/exception/TypeError text, catchability or external side effects.

This branch is based on published logical/strict-comparison stack #38
(`58105a31293772be3b08208ebeffcfa1fe76b98b`), not array-return stack #36.
The latter's isset and return-declaration additions are not integrated here.
Existing Gossip manifests and previous evidence remain pinned separately.
The target original helper is WordPress `_is_utf8_charset` (compat.php102–111),
whose unchanged source also needs the existing `!`, `||`, `is_string` and
scalar `===`. This compiler change alone does not establish that original
helper's compatibility: Gossip needs its own pinned original-source run.
`is_utf8_charset` also depends on defaults/coalescing/get_option, and neither
it nor `_canonical_charset`, encoding conversion or full WordPress is claimed.

`gohp.cli.generate` remains WebMCP-excluded local developer tooling. No new
browser/service/MCP operation, auth route, cookie forwarding, automatic
submission, external delivery, deployment, release or license decision.
Devboard card remains unidentified/unsynchronized, not Done.
