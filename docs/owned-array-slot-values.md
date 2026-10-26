# Owned scalar-array values in internal Slots

General `internal/storage.Slot` bindings can now own null/bool/int64/string or
an open `*ScalarArray`. This extends the **canonical** Slot/cell engine; there is
no second array-slot reference model. A bounded value-only runtime adapter now
uses this storage explicitly. This is R1 infrastructure, not PHP reference
syntax or a migration of generated array operations.

## Value ownership

- `NewSlot(array)` and `Write(array)` retain a fresh array handle by `Copy`.
  The caller still owns its input. Retain happens before replacing/releasing
  the old payload, including writes through a shared outer reference cell.
- `Read()` dereferences the outer binding. Scalar results need no cleanup;
  an array result is a **new caller-owned handle**, never the stored handle.
  Close each returned array on every exit path, normally using `defer`.
- Value assignment copies container ownership, not outer reference ownership.
  Array elements still follow the existing separation/reference-copy rules.
  A read handle can outlive the Slot, and closing it cannot close the Slot's
  owned handle. Index metadata is preserved by the existing container model.
- `Bind` shares the outer cell, which owns its payload once irrespective of
  how many bindings refer to that cell. Promoting a direct value to a cell
  moves ownership without an extra array retain. Rebinding releases only the
  destination's previous binding. Retain the new cell before releasing the old binding.
- `Unset` releases an ordinary payload, or one reference binding. The last
  reference binding releases its cell's payload. Replacement closes the prior
  owned array handle once. Escaped element aliases remain independently owned.

Nil/closed array handles and foreign values reject before replacement or
release. A typed nil `*ScalarArray` is not PHP null. Go struct/pointer copies,
GC and finalizers are not ownership operations. Slots and arrays remain
pointer-only, non-concurrent handles; explicit cleanup is the caller's duty.

`Read` is a value operation, **not a writable array location**. Mutating a read
copy does not update ordinary elements in the stored value; explicit value
assignment back is a separate operation. Do not lower PHP root/element writes
as an unconditional Read/mutate/Write sequence: that can separate too early.

## Owned value temporaries

Internal `Temporary` owns a scalar or scalar-element array **value snapshot**,
not a source binding, reference argument or writable location. It reuses Slot's
canonical retain/release rules; there is no second cell or reference engine.

- `NewTemporary(value)` retains once; the caller still owns its input. A zero
  Temporary is inactive; use `NewTemporary(nil)` for a live present-null value.
- `Slot.Snapshot()` returns `(nil, false)` for absence, without initializing the
  Slot. A present result retains its value once, without a `Read` intermediate
  or outer alias. Later source replacement/rebinding/unset cannot change that
  captured value. Array element references still follow the canonical rules.
- `MoveTo(destination)` writes through the destination's current binding and
  consumes the temporary only after successful assignment. Write retains before
  release, including self-root assignment. Nil destinations or scalar-only
  domain failures leave both owners intact for retry or cleanup.
- `Take()` consumes the temporary and transfers its existing value owner to the
  caller, without another retain. A returned array requires caller `Close`.
  Neither method exposes a borrowed payload while the temporary is live.
- `Close()` releases a remaining value once and is safe after consumption or on
  an inactive zero value. Nil receivers reject. `Take`/`MoveTo` reject inactive
  values; consumed temporaries cannot be reused or resurrected.

Install `defer temporary.Close()` immediately after acquisition as a Go failure
guard, including when a later operand panics. Also consume/close at the selected
operation's lifetime boundary: postponing every release to function exit is
**not** a valid universal lowering. A held array snapshot can keep an earlier
container's element reference alive, changing later separation. The tests
observe that releasing such a snapshot before versus after a subsequent copy
write produces independent versus shared element values. This is internal
lifetime evidence, not a new PHP evaluation-order or exception contract.

These pointer-only, non-concurrent storage types are never directly exposed to
external consumers. The opt-in runtime facade below wraps them; codegen does not
emit it yet. Generated arrays still use eager-copy `runtime.Array`, and selected
`ScalarFrame` functions exclude them. Expression-level ownership planning
(evaluation, result transfer, cleanup and lazy branches) remains necessary
before a generated scope can actually own/release shared-array temporaries.

## Opt-in runtime boundary

`runtime.ValueSlot` and `runtime.ValueTemporary` provide an additive, value-only
adapter for external Go consumers. They accept exact null/bool/int64/string or
a **canonical flat scalar `runtime.Array`**, not every existing runtime value.
They expose neither internal handles nor Bind/Reference/element aliases.
Both are pointer-only, non-concurrent, and require explicit cleanup.

| API | Ownership and result |
|---|---|
| `NewValueSlot(value)` / `Write(value)` | Import a detached value before replacing the owned payload; invalid input preserves the destination |
| `Read()` | Export an independent legacy value; exported Array.Entries never shares a backing slice with storage or another export |
| `Snapshot()` / `NewValueTemporary(value)` | Retain an independently owned value temporary, not the source binding |
| `ArrayWrite(key, value)` / `ArrayAppend(value)` | Mutate the current owned array via canonical storage; scalar elements only, existing bounded key/index rules |
| `ValueTemporary.MoveTo(slot)` | Assign through the destination's current value binding, consume only on success; nil/closed destination leaves the temporary usable |
| `ValueTemporary.Take()` | Consume and export a detached legacy value; release internal ownership before returning, so the result needs no Close |
| `Close()` | Release remaining ownership once; safe after consumption; nil receivers reject |

A zero ValueSlot is open but undefined: Read/Snapshot and array mutation reject,
while Write initializes it. Close is terminal. A zero ValueTemporary is inactive;
only Close is valid until constructed. Neither wrapper can resurrect after Close
or consumption. A taken result survives producer `defer Close` and slot teardown.

Array import preflights **all** entries before creating owned storage: exact,
already-normalized int64/string keys, no duplicates, scalar-only values, and
private index state equal to replaying the ordered integer keys through the
shared `arraykey.Index` policy. The last check matters because hosts can edit
public `Array.Entries` without updating hidden next-index/overflow history.
Such non-reconstructible arrays are **rejected by this new opt-in boundary**,
not silently repaired or accepted with changed append behavior. No index setter
or alternate key policy is introduced. Write/append without deletion/reordering
preserve the reconstructible invariant, so export through NewArray retains key
order, the next automatic key and overflow behavior. Host input/output mutations
do not affect storage; snapshots and MoveTo keep ordinary value-copy isolation.

Nested arrays, objects, callables, foreign host values, noncanonical/duplicate
keys and stale index metadata remain outside this adapter. Existing Array APIs
and their nested-value behavior are unchanged; this is not a drop-in replacement
for every Array or function parameter. Runtime unit checks and a real external
module test cover this boundary, not generated-code or native PHP equivalence.
New callers must pin a runtime revision providing these APIs; old generated Go
and IR 0.1 are unchanged. No new CLI, service operation, authentication or WebMCP
surface is introduced; the existing local generation exclusion still applies.

## Synchronous writes through the owning Slot

`ArrayWrite(key, value)`, `ArrayAppend(value)`, `ArrayBind(key, source)` and
`ArrayReference(key)` operate directly on the **current** array-valued binding.
They borrow its owned handle privately for the duration of one synchronous
call, without creating an array read copy, promoting the outer Slot to a
reference, or exposing the handle/element Slot. The existing ScalarArray engine
still owns key normalization, index allocation, validation and separation.

- Outer aliases share the mutation; ordinary array value copies separate on
  mutation. Element references retain the existing early/late copy-timing rules.
- Nil/undefined/null/scalar roots reject without autovivification. Unsupported
  keys, values, sources and exhausted append indices reject before separation;
  a failed operation must not change subsequent alias behavior.
- `ArrayBind` changes only the selected element binding. An undefined valid
  source becomes null only after validation. `ArrayReference` returns a fresh
  caller-owned alias and creates a missing null element after separation. The
  caller must `Unset` that alias; rebinding it does not rebind the element.
- The borrow cannot escape, invoke user code or span a root replacement,
  rebinding or unset. No extra temporary owner or release obligation is created
  (apart from the explicitly returned reference). Existing `Slot.Unset` and
  replacement/rebinding continue to release the payload.
- Arguments are already evaluated when each method resolves the current root.
  If an argument's evaluation replaces/rebinds that Slot, mutation targets its
  new array, not an earlier snapshot. This is an internal API contract, **not**
  a universal PHP LHS/RHS evaluation rule or a location held across callbacks.

These operations cover one-level scalar elements only. They do not provide
persistent address handles, nested paths, element deletion, general frames or
generated temporary lifetime. Future lowering must choose an operation-specific
address/lifetime contract and verify it against each admitted native target.

## Keep bounded scalar domains closed

Frame-owned bindings and ScalarArray-owned elements remain scalar-only.
Their private slots have an immutable scalar-only domain, retained across
Unset/reuse/rebinding. A reference cell counts both total bindings and live
scalar-only bindings. While any scalar-only binding exists, **all** aliases
reject array writes, including otherwise unrestricted general Slots.

Frame/ScalarArray Bind first validates the source value without creating an
owned read temporary. An array-valued source is rejected before creating a
missing binding or separating an array. Joining a scalar source adds the
destination's constraint to its canonical cell; it cannot be bypassed by
writing through an earlier or later alias.

Container Copy retains no extra element binding/constraint. Separation adds
actual element bindings where references remain shared; closing the last
container releases them. A constraint is removed only when that constrained
binding is released/rebound. Once every constrained binding is gone, surviving
unrestricted aliases may own arrays. Exported aliases are not permanently
constrained merely because a former frame/element owner was scalar-only.

These rules prevent an array from containing an array through direct writes,
Bind or later alias writes. In particular they block self/indirect reference
cycles. They are an internal admission boundary, not full PHP restrictions.
Nested arrays, objects, cycles, deletion and destructor order are not supported
by this storage. The existing value runtime still handles its earlier subset.

## Evidence and next gate

Internal unit checks cover owned copies, outer aliases, replacement/final
release, scalar-domain propagation, synchronous mutations/current-root lookup,
invalid mutations without separation, and temporary capture/transfer/release.
They remain under the **unsupported** reference feature in the catalog. The
opt-in runtime adapter has six separate unit/consumer checks under
`gohp.tooling.owned-value-adapter`, not a generated PHP feature claim. Existing
scalar Frame ABI checks and both active target CI corpora must remain green.
PHP 8.3.36 native-only observations in #82 are unchanged; they are not new
generated-reference or PHP 8.4 native-reference evidence.

R1 remains partial: synchronous Slot-rooted writes, internal value temporaries
and the opt-in runtime adapter are implemented, but persistent locations, general
generated-frame/temporary integration, nested ownership and cyclic reclamation
are unfinished. The next
gate is an expression-level ownership plan and a proven closed scope for
generated array temporaries, plus operation-specific locations across calls,
before enabling array reference syntax or replacing runtime storage. There is
no performance or application-acceptance claim.

The runtime APIs are additive; no compiler/parser/IR, CLI/capability/auth/WebMCP,
target/oracle/CI, Gossip pin or deployment change. The local CLI's existing exclusion remains.
