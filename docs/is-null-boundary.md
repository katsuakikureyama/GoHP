# Null type predicate

`is_null(value)` accepts exactly one evaluated supported value and returns an
actual bool. Direct calls are case-insensitive, and redeclaration is rejected.
Only the runtime's canonical nil value returns true. False, zero, empty strings,
arrays, class instances and generated arrows return false without conversion,
traversal or mutation. The argument expression is evaluated once.

This implements the [PHP is_null value contract](https://www.php.net/manual/en/function.is-null.php)
within the existing GoHP subset. It is not `isset`: the operand is an ordinary
expression, not a suppressed address probe. Undefined variables remain rejected
before generation; missing array-element reads retain their existing runtime
failure. PHP warning-plus-null behavior for undefined reads is not implemented.
Use existing `array_key_exists` or bounded `isset` to check presence without
reading an absent element. Present-null, present-false and missing are distinct.

Semantic validation records a bool result and rejects invalid arity, reserved
name reuse and known bool-as-array misuse. Existing conditional and short-circuit
rules apply; no branch-sensitive type narrowing is added. No new source value,
float/resource support, named arguments or builtin callable lookup is introduced.
Runtime-only tests distinguish canonical nil from Go typed nil pointers, slices
and maps; these Go values do not become new PHP source types.

The owned `is_null.php` fixture observes strict bool results for null/non-null
values, dynamic forwarding, single evaluation, lazy evaluation, present-null
versus missing keys and generated callback composition. Its original evidence
compares stdout and normal exits with PHP 8.3 >=8.3.36, not general
error/exception behavior. The current compiler targets are PHP 8.3 and 8.4
under the [active PHP version policy](compatibility.md#active-php-version-policy);
this historical 8.3 fixture description neither supplies nor claims separate
8.4 evidence. The feature remains partial in the compatibility catalog.

Gossip's original-source pins, corpora and evidence are unchanged. This fixture
does not establish additional original-function compatibility or automatically
adopt this compiler revision. Any Gossip adoption still requires its own pinned
verification; it does not borrow evidence between the active PHP targets.
Capability `gohp.cli.generate` stays WebMCP-excluded: local developer translation,
no browser/API operation, authentication change, dependency or deployment.
