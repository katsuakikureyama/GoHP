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

The container owns one Slot per insertion-ordered key. Before Write, Bind or
Reference, a shared container is separated:

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

Read returns value/presence without creating an absent key or separating the
container. Keys returns a detached ordered key list. Updating a key preserves
its position. Close is idempotent on a valid handle; the last container owner
releases every element once. Escaped aliases can outlive all array handles.
Closed/nil operations reject, including nil Close. There is no concurrency or
PHP destructor-order contract.

Invalid keys, unsupported values and nil Bind sources reject **before** storage
separation. A premature separation could change later reference behavior even
if the failing operation changed no visible value, so non-mutation includes
the sharing relationship, not just a snapshot of current values.

## Deliberately bounded domain

Elements are exact null/bool/int64/string, including byte strings. Slot and
Frame still reject array handles as values. Nested arrays, objects, callables,
host pointers and cyclic values cannot enter this storage.

Keys are exact int64 or string **raw identities**. In this internal API the
string `"1"` and integer `1` are distinct; this is not PHP key conversion.
No append, next-index/overflow state or element deletion is implemented. Before
runtime integration, a shared normalization/index policy must preserve the
existing runtime's key and append contracts without duplicating or importing
runtime helpers into storage (which would create an import cycle).

No pointer into a slice of entries is exposed. Temporary/host ownership,
nested-value retain/release, array-valued slots, address IR and operation-specific
evaluation order still need explicit integration contracts. Do not silently
wrap a runtime.Array or eager Clone and treat it as reference-capable storage.

## Evidence and next gate

Six `unit.storage_array.*` checks cover value/copy isolation, copy timing,
reference extraction, rebinding, last-owner release and invalid/closed use.
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
these internal checks separately. R1 remains partial. The next gate is shared
key/index policy and explicit value/temporary ownership before replacing any
runtime array path. No performance, warning/exception, cycle or application
acceptance claim is made.

No CLI/route/authentication/capability, target policy, oracle environment, CI
configuration, Gossip pin or deployment changes. The existing local
`gohp.cli.generate` WebMCP exclusion remains applicable.
