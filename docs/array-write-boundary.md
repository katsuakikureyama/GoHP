# One-level array element assignment

## Lvalue design

`$local[key] = value;` and `$local[] = value;` are admitted only when the root
variable already exists and contains an array. An explicit key is an
integer/string literal or a variable with an actual int64/string value.
No effectful key expressions are admitted.
An explicit `assign_index` statement carries `Target` as an index expression
whose base is one variable and whose key is one literal/variable; `Expr` is the
right-hand value. This additive IR field/kind keeps version 0.1, not a public
Go API guarantee. Malformed target shapes are rejected before generation.

The root binding is the only writable location. Generated code evaluates its
base, key and RHS once and stores the functional `ArraySet` result back into
that binding. `ArraySet` clones the array and inserted value before replacing
or inserting an entry. There are no reference cells, in-place nested lvalues,
or object/property writes. Existing assignment/argument/return copies remain
independent; an RHS reading the same root observes its pre-write value.
Foreach continues to iterate its entry snapshot, not a live mutable array.

Existing normalized keys are replaced without moving their position; new keys
are inserted at the end. Integer/string normalization is the existing `Key`
contract. Static invalid domains fail before generation, unknown values are
guarded at runtime. Failure behavior is not PHP warning/exception equivalence.
Integer-index bookkeeping is shared with literal construction, including the
[PHP 8.3 negative-key boundary](array-index-boundary.md). Copies preserve this
state.

## Append statements

`$local[] = value;` has its own `assign_append` statement kind with a variable
`Target` and an RHS `Expr`; it is not an index-read expression with a missing key.
The semantic validator checks the target shape, definitely-existing root and
array-or-unknown type before code generation. Generated code calls
`ArrayAppend(base, value)` once and writes its result back to that root.
It uses the same copy semantics as `ArraySet`; `$a[] = $a` appends a finite
pre-write value snapshot, not a self-reference.

The next integer key follows PHP 8.3, including negative keys and string-only
prefixes. Copy/map preserve that index while keys/values reindex it. The last
available int64 key may be appended once; another append fails explicitly.
Unknown non-array roots are rejected at runtime rather than auto-vivified.
`TestDifferential/array_append` observes keys/types, copies, RHS evaluation and
foreach snapshots; unit/codegen/runtime-negative checks cover isolation,
malformed IR, static exclusions and dynamic domain/overflow failures.

## Excluded forms

No `$a[x][y]`, `$a[x][]` or `$a[][y]` target, undefined/null/scalar auto-vivification,
references, compound assignment, assignment expressions, unset, or PHP's
bool/null/float key coercions. Object alias semantics are not extended.
Empty-bracket reads remain invalid; append is a statement-only operation.
Key calls/operators are rejected rather than promising general evaluation-order
compatibility. Future nested writes/references require a separate lvalue design.

Native fixtures use PHP 8.3 >=8.3.36 and compare stdout/normal exits. The
[PHP array reference](https://www.php.net/manual/en/language.types.array.php)
describes the broader native behavior; this contract is deliberately narrower.
No original WordPress helper or Gossip compiler pin is changed or validated.

This remains the WebMCP-excluded local `gohp.cli.generate` capability: simulated
PHP value writes introduce no service resource operation, authentication route,
external write, deployment or automatic submission.
