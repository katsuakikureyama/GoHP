# Scalar slot foundation (R1, first increment)

`internal/storage` implements bindings for null, bool, int64 and byte strings.
It is internal compiler infrastructure, not a public runtime API or PHP syntax
support. Generated programs still use the existing value runtime. References
remain unsupported in the compatibility catalog.

A zero-value Slot is absent; a present null is different. Read returns a value
and presence without creating storage. Write changes the contents of an existing
reference group. Bind joins the source group and releases only the destination's
old binding; an undefined source becomes null. Unset removes a binding, leaving
other aliases intact. Recreating that binding is independent. Reading a scalar
and writing it elsewhere does not propagate the outer reference.

Each reference cell counts logical bindings. Bind retains its new cell before
releasing the old one; Unset deterministically releases an owner and clears the
contents on last release. These counts must not use Go GC/finalizers. Slots are
pointer-only, must not be copied, and are not concurrency-safe. Future generated
frames must explicitly release every binding on all exit paths. Such frame
integration and lifetime enforcement are not implemented by this increment.

Arrays, objects, callables, host pointers and cells passed as values are rejected
before any write. Supporting arrays here by calling the existing eager Clone
would violate the copy-timing contracts in [the design](reference-value-model.md).
Shared ArrayStorage, detach-on-write, temporary ownership, locations for array
elements/properties and cyclic values remain unfinished R1/R3 work.

Tests exercise alias writes, ordinary scalar copies, rebinding, unset/recreation,
missing versus null, self-binding, and the local-binding behavior of a simulated
by-reference parameter. These are Go storage contracts, not PHP-to-Go differential
tests. They do not increase the supported PHP feature count or prove generated
reference execution. Native PHP expectations were recorded separately in #82.

No service operation is added; the existing local `gohp.cli.generate` WebMCP
exclusion remains applicable. No Gossip compiler pin changes are required.

## Verification

At commit `01a28db` (Go 1.23.0, darwin/arm64), with
`GOCACHE=/tmp/gohp-go-cache GOPROXY=off GOTOOLCHAIN=go1.23.0`:

- `go test -count=1 ./internal/storage`: PASS (8 tests, 0.194s).
- `go vet ./internal/storage`: PASS.

The PHP-only specification from #82 is reused as the behavioral reference; no
new PHP execution or generated-reference comparison is claimed. Remaining R1
array/temporary/frame integration and R2–R5 are not complete.
