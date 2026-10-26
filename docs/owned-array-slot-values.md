# Owned scalar-array values in internal Slots

General `internal/storage.Slot` bindings can now own null/bool/int64/string or
an open `*ScalarArray`. This extends the **canonical** Slot/cell engine; there is
no second array-slot reference model. It is bounded R1 infrastructure, not PHP
reference syntax or a migration of `runtime.Array`/generated array operations.

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
Address operations and their evaluation/copy timing remain a later gate.

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

Six internal unit checks cover owned copies, outer aliases, replacement/final
release, scalar-domain propagation and invalid Bind without separation. They
remain under the **unsupported** reference feature in the catalog. Existing
scalar Frame ABI checks and both active target CI corpora must remain green.
PHP 8.3.36 native-only observations in #82 are unchanged; they are not new
generated-reference or PHP 8.4 native-reference evidence.

R1 remains partial: general generated-frame/temporary integration, writable
array locations, nested ownership and cyclic reclamation are unfinished. The
next gate is a bounded owned array location operation with explicit lifetime
and copy timing, before enabling new reference syntax or replacing runtime
storage. There is no performance or application-acceptance claim.

No compiler/parser/IR/runtime ABI, CLI/capability/auth/WebMCP, target/oracle/CI,
Gossip pin or deployment change. The local CLI's existing exclusion remains.
