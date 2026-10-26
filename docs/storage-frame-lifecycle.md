# Scalar frame lifecycle (R1, internal only)

`internal/storage.Frame` owns named scalar bindings on top of the existing
[Slot foundation](scalar-slot-foundation.md). A bounded
[generated scalar integration](generated-scalar-frames.md) now uses it through
a runtime facade; general `map[string]any` frames remain in place. It does not
admit PHP reference syntax or implement PHP exceptions.
Only exact null, bool, int64 and string values are admitted; the boundary is the
same for the independently tested PHP 8.3 and 8.4 targets.

## Ownership contract

- A zero Frame is open and empty. Read returns value/presence without creating
  storage; an absent name is distinct from a present null.
- Write changes the value through a binding. Bind joins an external Slot's
  reference group and releases only the destination's previous binding.
  Invalid values and nil sources reject before creating or mutating bindings.
- Reference returns a **new caller-owned alias**, never the frame-owned Slot.
  Referencing an absent name creates present null. The caller must explicitly
  Unset the returned Slot; retaining it extends the reference group's lifetime.
- Unset removes the selected name and its ownership only. Recreating that name
  is independent of an exported alias to its former binding.
- Close releases every frame-owned binding once and is idempotent. Exported
  references and bindings owned by other frames survive. The last owner clears
  the cell value. Read/write/bind/reference/unset reject after Close; closed
  frames cannot be reopened. Nil receivers reject, including nil Close.

Frames and Slots are pointer-only, must not be copied, and are not safe for
concurrent access. A frame owner must explicitly arrange `defer frame.Close()`
or equivalent release on every exit path. No Go finalizer or GC timing controls
logical ownership. Returned references need their own explicit release, for
example after transferring ownership into another frame with Bind.

No public API returns a pointer to a frame-owned Slot: such a pointer could be
reused after Close to resurrect an untracked local binding. An exported alias
instead has independent ownership. Closing a frame does not invalidate that
alias and does not rebind its caller.

## Evidence and unfinished work

Six Go contracts are registered as `unit.storage_frame.*` checks under the
existing `php.syntax.array-write-and-reference` entry. They exercise presence,
scalar copies, alias mutation/rebinding, unset/recreation, escaping references,
multiple local owners, explicit release on Go return/panic, closed/nil rejection
and non-mutating invalid inputs. That feature remains **unsupported** on both
targets. These are internal ownership checks, not PHP-to-Go differential proof.
Exact commit, commands and CI results are recorded in the PR/current-state.

The PHP-only reference specification from #82 remains separate and unchanged;
its PHP 8.3 observations are not inherited by PHP 8.4. This increment requires
no new native-only run because it neither admits nor generates new PHP syntax.

R1 is still partial: general generated-frame integration, address IR, temporary and
shared-array ownership, detach-on-write, cyclic values and PHP destructor order
are not implemented. Map iteration during Close has no destructor-order meaning
in the scalar-only domain. Existing array/object value semantics remain unchanged.
The separate scalar integration records generated-call and target-specific
native evidence; these internal ownership tests alone are not language support.

No CLI, route, service capability, authentication or WebMCP operation changes.
The existing local `gohp.cli.generate` exclusion applies. Gossip pins and corpora
are unchanged; no deployment destination or release is introduced.
