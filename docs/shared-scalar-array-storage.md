# Shared scalar array storage (bounded R1 foundation)

`internal/storage.ScalarArray` introduces explicitly owned, shared containers of
the canonical scalar Slot bindings. This is **internal infrastructure only**:
the existing `runtime.Array`, eager Clone, generated frames, parser and IR do
not use these containers. PHP references remain unsupported.

## Ownership and separation

An array handle is pointer-only and must not be copied as a Go struct. Its zero
value is an open empty array. `Copy()` returns a new owned handle to the same
container, increasing the container owner count but **not** each element's
reference count. Call `Close()` for every handle on every exit path; GC and Go
pointer copies cannot supply these logical ownership operations.

The container owns one Slot per insertion-ordered key and its automatic-index
state. Before Write, Bind, Reference or Append, a shared container is separated:

- An ordinary element is copied by scalar value.
- A reference element with only its container binding left becomes an ordinary
  value in the new container.
- A reference with other live bindings retains its canonical Slot reference
  cell; the new container adds one element owner.

Write then assigns through the selected element binding. Bind replaces only
that element's binding with an alias of an independently owned source Slot.
Reference returns a fresh caller-owned alias, never the container's Slot; an
absent key becomes present null only in the separated destination. The caller
must Unset that alias. Rebinding/unsetting it cannot replace/remove the element.
Append inserts an ordinary scalar at the next automatic key and returns that
key. Separation copies the index state by value, so later insertions in one
container cannot advance another's index, even when some element cells alias.

Read returns value/presence without creating an absent key or separating the
container. Keys returns a detached ordered key list. Updating a key preserves
its position. Close is idempotent on a valid handle; the last container owner
releases every element once. Escaped aliases can outlive all array handles.
Closed/nil operations reject, including nil Close. There is no concurrency or
PHP destructor-order contract.

Invalid keys, unsupported values, nil Bind sources and append overflow reject
**before** storage separation. A premature separation could change later
reference behavior even if the failing operation changed no visible value, so non-mutation includes
the sharing relationship, not just a snapshot of current values.

## Deliberately bounded domain

Elements are exact null/bool/int64/string, including byte strings. General
[Slots now own array values](owned-array-slot-values.md), but Frame/element
bindings stay scalar-only. Their live constraints follow the canonical cell,
so external aliases cannot introduce nested arrays by writing later.
Nested arrays, objects, callables, host pointers and cycles remain excluded.

Keys are exact int64 or string, normalized by `internal/arraykey`, the same
dependency-leaf policy used by `runtime.Array`. Canonical in-range decimal
strings such as `"1"` and `"-5"` become integer keys. Noncanonical strings such
as `"01"`, `"+1"`, `"-0"`, whitespace and out-of-range integers remain strings.
All Read/Write/Bind/Reference operations use this normalization; duplicate keys
replace or bind in place. This supersedes #96's temporary raw-key contract.
Null/bool/float key coercion and element deletion remain excluded.

Empty/string-only containers append at zero. The first inserted integer key,
even if negative, establishes its successor. Lower/replacement keys do not
rewind it. Explicit maximum int64 keys are valid, but subsequent Append rejects;
this is not PHP exception equivalence. Bind and missing-key Reference also
record inserted integer keys. A failed operation does not consume an index.
See the [shared key/index contract](array-index-boundary.md).

The policy preserves the existing runtime behavior for the active 8.3/8.4
targets. It does not admit new targets or imply identical behavior across all
PHP 8.x versions. Neither runtime nor storage imports the other for this policy.

No pointer into a slice of entries is exposed. Internal value temporaries now
have explicit retain/transfer/release; generated/host integration,
nested-value retain/release, address IR and operation-specific
evaluation order still need explicit integration contracts. Do not silently
wrap a runtime.Array or eager Clone and treat it as reference-capable storage.

## Evidence and next gate

Nine `unit.storage_array.*` checks cover value/copy isolation, copy timing,
reference extraction, rebinding, last-owner release, invalid/closed use,
normalized keys, index/copy/append behavior and failure atomicity.
They remain under the **unsupported** reference feature in the catalog, rather
than claiming PHP-to-Go array-reference compatibility.

The two copy-timing sequences specified in
[the reference design](reference-value-model.md#copy-timing-is-part-of-correctness)
are reproduced internally: releasing an external alias before separation yields
independent values, while separating with the alias still live preserves shared
elements after that external alias is released. PHP 8.3.36 native-only evidence
was recorded separately in PR #82 and is reused as the unchanged behavioral
reference; it is not a new generated-Go comparison or PHP 8.4 native claim.
The singleton-reference rule is informed by the pinned
[PHP 8.3.36 array duplication implementation](https://github.com/php/php-src/blob/php-8.3.36/Zend/zend_hash.c).
Its self-array exception is outside this scalar-only domain.

Both active target CI jobs must preserve the existing compiler corpus and record
these internal checks separately. Existing array/append differential fixtures
exercise the runtime policy per target; they are not evidence for generated
reference arrays. R1 remains partial. Explicit Slot array ownership, synchronous
mutations and value temporaries now have an opt-in value-only runtime adapter
for canonical flat values (see the owned Slot contract). Persistent locations
and general generated temporary ownership are still needed for broader migration;
the first nullary straight-line generated scope now uses the value-only adapter
with operation-scoped temporaries, without reference operations.
No performance, warning/exception, cycle or application acceptance claim is made.

No CLI/route/authentication/capability, target policy, oracle environment, CI
configuration, Gossip pin or deployment changes. The existing local
`gohp.cli.generate` WebMCP exclusion remains applicable.
