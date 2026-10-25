# One-level array element assignment

## Lvalue design

Only `$local[key] = value;` is admitted, where the root variable already exists
and contains an array. The key is an integer/string literal or a variable with
an actual int64/string value. No effectful key expressions are admitted.
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

## Excluded forms

No `$a[]` append, `$a[x][y]` target, undefined/null/scalar auto-vivification,
references, compound assignment, assignment expressions, unset, or PHP's
bool/null/float key coercions. Object alias semantics are not extended.
Key calls/operators are rejected rather than promising general evaluation-order
compatibility. Future nested writes/references require a separate lvalue design.

Native fixtures use PHP 8.3 >=8.3.36 and compare stdout/normal exits. The
[PHP array reference](https://www.php.net/manual/en/language.types.array.php)
describes the broader native behavior; this contract is deliberately narrower.
No original WordPress helper or Gossip compiler pin is changed or validated.

This remains the WebMCP-excluded local `gohp.cli.generate` capability: simulated
PHP value writes introduce no service resource operation, authentication route,
external write, deployment or automatic submission.
