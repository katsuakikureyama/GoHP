# Bounded nested array probes and coalescing

This contract adds a deliberately small nested-array address subset. It is
intended to model the pure array/boolean path checks in WordPress
`wp_should_add_elements_class_name` (WordPress `wp-includes/block-supports/elements.php`, lines
33–97), not to make WordPress, its external resolvers, or class-based block
settings runnable. Gossip subsequently verified that unchanged caller at its
exact GoHP #67 pin; see the [consumer evidence](integration.md). This does not
certify the surrounding render hook/CSS/HTML pipeline or later compiler revisions.

## Nested `isset` address

The admitted nested form has a defined variable root and **two or more** keys:

```php
isset($block['attrs']['style'])
isset($block[$first][$second])
```

Each key is either an integer/string literal or an already-defined variable
whose runtime value is an integer/string. The IR is `isset_path`, with an empty
`Text`, flat operands of the root followed by two-or-more keys, and no embedded
normal-read tree. A malformed shape, undefined root/key, and statically known
bad key are rejected before Go emission; validation also visits dead operands.

At runtime every intermediate must be an array or null. A missing key or null
at any point returns false without visiting a later key. A non-array
intermediate (including string, bool, or object) is an explicit runtime failure,
not PHP warning/exception equivalence. The final value is set whenever it is
non-null, including false, zero, empty string, and an empty array. The runtime
path probe returns a detached clone through the established `Clone` boundary;
`ArrayPathIsset` wraps that probe for presence only.

The original one-level `isset($array[$key])` remains exactly the existing strict
`ArrayIsset` contract. It is neither widened nor reimplemented by this feature.
Multiple `isset` operands keep left-to-right short circuiting: an earlier false
presence result prevents a later nested runtime guard from running.

## Bounded `??`

The same address language admits a null-coalescing expression only as:

```php
$block['attrs']['style'] ?? $fallback
$value ?? $fallback
```

The left side is a defined bare variable or a defined variable root followed by
one-or-more admitted index keys. It is represented as `coalesce` IR with two
arguments: the normal address tree and the fallback expression. The compiler
does not lower the left side through a normal eager `Get` read. The path probe
is evaluated once; missing or null selects the fallback, while a non-null value
is returned as a detached clone. The fallback is lazy and evaluated once only
when selected. Its expression, and every static address key, are still
validated even if a test case makes them unreachable.

`??` is right-associative. In this subset it binds below `||` and above full
ternary, matching the targeted PHP operator ordering. It does not imply support
for `??=`, assignment expressions, a general expression left side, property
addresses, calls/effectful keys, undefined-variable suppression, or key
coercion. Normal array reads retain their existing strict missing-key behavior.

## Explicit exclusions

The extension excludes properties, objects/classes, external resolvers,
references, string offsets, effectful/arithmetic keys,
undefined roots/keys, dynamic function names, and array auto-vivification.
It accepts no `??=` assignment, no expression left side, and no general PHP
error/warning behavior. It does not add control-flow narrowing from either
presence or coalescing results.

`isset` retains its direct-key parser grammar; coalescing uses the normal index
expression tree, so parentheses around an otherwise admitted variable/literal
key do not change its address shape or semantics.

## Evidence and integration boundary

The catalog, compiler/runtime tests, and a pinned native-PHP corpus must record
the exact supported path shapes, static rejections, clone/non-mutation behavior,
short-circuiting, laziness, and precedence before this contract can be reported
as a measured compatibility result. A successful compiler test is not a
WordPress consumer adoption decision.

This remains the developer-local `gohp.cli.generate` capability, explicitly
WebMCP-excluded. It adds no browser/server/MCP operation, route, authentication
change, dependency, deployment, release, external write, or automatic submit.

References: [PHP `isset`](https://www.php.net/manual/en/function.isset.php),
[PHP comparison / null coalescing](https://www.php.net/manual/en/language.operators.comparison.php#language.operators.comparison.coalesce),
[PHP operator precedence](https://www.php.net/manual/en/language.operators.precedence.php),
and [PHP assignment operators](https://www.php.net/manual/en/language.operators.assignment.php).
