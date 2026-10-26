# Actual-integer binary subtraction

Binary minus is left-associative at the same precedence as addition, above PHP8
string concatenation. It accepts actual int64 operands only; unknown values are
guarded before subtracting. Overflow fails closed instead of promoting to float.
Safe bounds do not negate MinInt64. Existing negative-decimal-literal support
is unchanged; no general unary negation, decrement, compound assignment, weak
coercion or general PHP error equivalence. Decrement tokens are kept intact and
rejected, not split into binary minus and a negative literal.

This closes a newly observed blocker in unchanged add_freeform: strlen(document)
minus offset in its lazy coalesce fallback. An earlier Gossip attempt stopped
at code generation with PHP002 expected semicolon at minus, before build/run/
PHP comparison; it is not recorded as consumer PASS. Compiler fixtures compare
association, precedence, negatives, unknown forwarding, lazy fallback and
property reads on admitted PHP8.3.36. Runtime/source negatives cover actual
domains, extreme bounds and excluded syntax. Original adoption is separate.

Capability gohp.cli.generate remains WebMCP-excluded (local developer compiler).
No route/auth/dependency/CI/deployment changes or production target creation.
