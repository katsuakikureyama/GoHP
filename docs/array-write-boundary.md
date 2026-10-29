# Bounded array writes and array_push

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
or inserting an entry. These local assignment forms do not add reference cells,
in-place nested lvalues or object/property writes. Existing assignment/argument/return copies remain
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

## Bounded array_push calls

`array_push($target, value, ...);` is admitted as a standalone call or as the
entire RHS of `$count = array_push(...);`. The target is an already-defined
bare local or one declared named public property of a defined bare local,
including `$this->stack`. Known classes retain static slot checks; inferred-
unknown receivers use the existing explicit PropertyObject protocol. Actual
targets must be arrays. There is no null/scalar autovivification. Calling with
only the target validates its actual array domain and returns its current count.

The frontend emits an additive `array_push` statement at IR0.1: Target is the
address, Name the optional bare result local, and Expr a canonical call named
`array_push` whose Args contain **only the value arguments** (possibly none).
The statement is not an ordinary value call or a general PHP reference API.
Generic call IR, hidden statement/call metadata and unsupported targets reject.

Generated code captures the property owner/name before evaluating payloads.
It snapshots each payload once, left to right, then reads the **current** target
array. Thus an argument can replace that same property before append, while
earlier array payloads retain their pre-replacement copies. Even a property
initially holding a scalar can be replaced by an array during argument evaluation;
the array-domain check is after all values. Existing local maps are stable for
the statement; no persistent reference or new alias binding is created.

ArrayPushLocal and NestedProperty.Push append via existing ArrayAppend, retaining
negative/string/sparse-key bookkeeping, detached nested arrays and shared child
object identifiers. Every successful append is published in order. A subsequent
overflow rejects without rolling back earlier appends; no PHP exception/message
equivalence is claimed. Zero-value calls do not publish a replacement. The int64
count is assigned last, including `$items = array_push($items, 'value');`, where
that result intentionally replaces the local array binding. Function arguments
still copy arrays; property-owning objects still retain identity.

The new statement stays outside scalar-frame/owned-value proofs and does not
add mutable expressions to their generic evaluation paths. Loop/switch widening
tracks both a local target and the result destination. Embedded calls in echo,
return, conditions, other arguments or arrow bodies remain rejected, as do
property/index result destinations, indexed/nested/dynamic/effectful targets,
undefined roots, named/unpacked/reference arguments and builtin redeclaration.

The [PHP array_push contract](https://www.php.net/manual/en/function.array-push.php)
and [left-to-right eager arguments](https://www.php.net/manual/en/functions.arguments.php)
define the bounded successful behavior. The fixed
[PHP8.3.36 implementation](https://github.com/php/php-src/blob/php-8.3.36/ext/standard/array.c)
also shows ordered insertion before count return. A network-disabled fixed-image
diagnostic confirmed property reset/repair and old payload snapshots before this
implementation. Exact commands, candidate-specific native fixtures and the
unchanged original-source probe belong in the
[array-push evidence](evidence/array-push-2026-10-29.md).
These observations do not certify complete original stack/parser execution.

## Other excluded forms

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
