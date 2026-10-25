# Literal null strict identity

Gossip's unchanged `_wp_array_get` generated and built on GoHP #63, but a real
array-valued path element reached `null === $path_element` and the scalar-only
identity runtime guard panicked. This change fixes that bounded contract without
rewriting the original or implementing general array identity.

For `===` or `!==`, when either validated operand is syntactically literal null,
code generation evaluates the other expression exactly once with `rt.IsNull`,
negated for inequality. Both operand orders, case-insensitive null and explicit
parentheses retain the same meaning. Arrays, class instances and callables are
non-null; no deep walk, conversion or mutation occurs. Missing variable/index
reads still fail through their ordinary boundaries, including in null probes.

Two variable operands do not gain general non-scalar identity, even if one
variable happens to be null at runtime. Non-null array/object comparisons,
loose equality, floats/resources, exception equivalence and general operand
order remain excluded. This uses the existing canonical nil runtime predicate;
no runtime API or dependencies change.

The catalog records partial operator support and the owned regression fixture.
Gossip must adopt a new exact pin and rerun its failed24-case configuration.
`gohp.cli.generate` stays explicitly WebMCP-excluded; route/auth unchanged.

Primary references: [strict comparisons](https://www.php.net/manual/en/language.operators.comparison.php),
[null predicate](https://www.php.net/manual/en/function.is-null.php).
