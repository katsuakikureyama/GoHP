# Reference, value and assignment foundation

Status: **design + native-oracle specification; scalar frames, shared scalar containers, owned Slots/temporaries, runtime adapter, bounded scalar-entry generated array scopes (including if/while) and scalar request-global host bindings implemented**.
Original design baseline: main `565e65e6c9d8e211b06779739bfba76b07c4d92d` (#81).
Request/location design review: main `c8d21776d0ca46ed9b42fb66363d10765d9721d2`
(#139); design-only consolidation #141 did not add runtime or source admission.
Target: PHP 8.3, oracle floor 8.3.36. Later series need separate policy and
evidence; they must not inherit these results. No compiler/runtime/IR changes
were introduced by the original design or its PHP-only specification. The first
implementation is [scalar slot infrastructure](scalar-slot-foundation.md), kept
with [explicit scalar Frame lifecycle](storage-frame-lifecycle.md) and
[bounded generated integration](generated-scalar-frames.md) for nullary or all-exact-scalar-parameter
functions; PHP reference syntax remains unsupported. The native specification below is
still PHP 8.3-only even though the compiler now has an independent PHP 8.4 target.
[Shared scalar containers](shared-scalar-array-storage.md) now exercise internal
separation/copy timing; general runtime arrays and generated array code are not migrated.
[Owned Slot array values](owned-array-slot-values.md) now have explicit retain/read/release
and synchronous one-level mutations without a Read/copy/writeback temporary;
Frame and element bindings remain scalar-only, including through aliases.
Internal value temporaries now capture, transfer and release owned snapshots.
An opt-in runtime ValueSlot/ValueTemporary facade converts canonical flat values
without exposing storage types. Nullary/all-exact-scalar-parameter scopes with proved if joins
and domain-invariant while loops use operation-scoped temporaries, selected-arm snapshots and detached returns;
the [ownership plan](owned-array-slot-values.md#first-generated-scope-and-expression-ownership)
defines its proof and exclusions. This is not general generated-array migration.

The 2026-10-30 request/location design below does not implement `global`,
`$GLOBALS`, static storage, PHP references, interfaces or visibility. The later
[scalar request-global host ABI](#scalar-request-global-host-abi) adds guarded
scalar bindings and Close to `Execution`, alongside strtok/JSON/PCRE state.
It is not a general PHP symbol table or HTTP request manager. Existing
object-pointer identity is distinct from variable aliasing.

## Decision

Keep the source-to-source pipeline and generated Go functions. Introduce a
common storage model underneath them before enabling reference syntax. Separate
ordinary value assignment, binding a reference, and accessing a writable
location. Go object pointers already preserve object identity; they do not
implement PHP variable aliases. A blanket change from `any` to pointers, or
teaching the lexer about `&`, is insufficient.

Use a shared array-storage / detach-on-write model for reference-capable arrays.
The timing of separation can affect observable reference behavior, so it is
not merely a performance optimization. Keep the current eager-copy runtime
unchanged until its replacement has a measured migration path. The proposed
names below describe responsibilities, not a published Go API or final layout.

## Current implementation and required seams

| Current seam | Current contract | Required extension |
|---|---|---|
| `runtime.Clone`, `NewArray`, `Get`, `Entries` | Recursive detached arrays; object identifiers retained | Distinguish value reads/copies from array storage duplication and reference binding |
| `runtime.ArraySet`, `ArrayAppend` | Return a detached array for root replacement | Write through an owned location, separating the right array storage first |
| Generated `map[string]any` frames, parameter and return copies | Value-only locals/calls | Frame bindings with explicit location lifetime and parameter/return passing mode |
| Generated property Get/Set and array-property append | Copy values; preserve receiver identity and RHS effects | Shared location rules with operation-specific evaluation order |
| `runtime.Callable.InvokeWithExecution` | Fixed-arity generated arrows; clone arguments/results | Signature-aware value/reference arguments and results, shared with direct calls |
| IR `Stmt.Target`, `Function.Params` / `ParamTypes` | Bounded write targets; no reference metadata | Explicit address forms and bind/unset/by-reference metadata, not overloaded normal expressions |

Do not retain pointers into the current array's `[]Entry`: insertion or storage
separation can invalidate their meaning. An address needs owner/storage/key
identity and a defined lifetime. Unknown foreign Go values must not smuggle an
internal reference cell into the PHP value domain.

## Behavioral model

- **Value**: PHP-visible scalar/null, array handle or object identifier. Existing
  `any` payloads may remain internally; a full tagged-value rewrite is not a
  prerequisite. Undefined/missing storage is distinct from a present null.
- **Slot**: an owned variable, array element or property binding. It contains
  either an ordinary value or a link to a reference cell. Distinct ordinary
  variables remain distinct even when they hold the same object identifier.
- **RefCell**: shared mutable contents for explicit PHP aliases. Writing a value
  through an alias updates the cell. Rebinding one alias changes that slot's
  binding, not every other alias's binding.
- **ArrayStorage**: ordered keys, values/slots and next-index metadata. Ordinary
  array assignment shares storage logically; a write separates the applicable
  storage/path. Explicit references within elements are handled separately from
  sharing the array container itself.
- **Location**: an operation-specific address handle, never a copied result of
  `Get`. Read, write, bind, probe and unset have distinct rules for absent
  variables/elements. Creating null storage through a reference must not make
  ordinary reads or `isset` create storage.

| Operation | Required observable contract |
|---|---|
| Read/copy value | Dereference the outer slot; scalar copy or object-identifier copy does not alias the source variable. Arrays retain logical value semantics and eligible internal references |
| Assign value | Prepare the RHS once and write through the destination's existing binding; assignment does not break an existing alias |
| Bind reference (`=&`) | Bind the destination slot to the source's reference cell; release the old destination binding, without rebinding its other aliases |
| Unset | Remove only the selected binding; a recreated name/element is independent of surviving aliases |
| By-value call/return | Copy a value, not the caller's outer binding; an array's internal references still follow separation rules |
| By-reference call/return | Carry a location/cell according to the signature; local parameter rebinding is not caller rebinding; returned references outlive the callee frame when retained |
| Foreach | Keep value iteration and reference iteration distinct; reference iteration leaves its final binding alive until explicitly unset/rebound |

These contracts follow the official explanations of
[reference assignment and calls](https://www.php.net/manual/en/language.references.whatdo.php),
[unset](https://www.php.net/manual/en/language.references.unset.php),
[foreach references](https://www.php.net/manual/en/control-structures.foreach.php#control-structures.foreach.reference),
and [object identifiers](https://www.php.net/manual/en/language.oop5.references.php).
They do not make every address form or operation supported at once.

### Request ownership and common writable locations (design)

The next foundation is driven by original hooks, not an independent storage
rewrite. [Gossip75's immutable native baseline](https://github.com/katsuakikureyama/Gossip/blob/753c5cfbdb253d3657cc052234aedaa6e0d76971/docs/evidence/hooks-original-baseline-2026-10-30.md)
records eight actual PHP workflows: priority/insertion order, accepted arguments,
callback identity/removal, actions, reentry, mutation during iteration and
sentinel identity. Its GoHP probe rejects original `plugin.php:25` (`require`);
generated hook execution/comparison remains **not-run**. That native driver
clears globals between workflows; it does not prove request isolation.

Use the existing Slot/Frame foundation. The following owner roles and location
operations are a design contract, not general public Location types or admitted
syntax. The bounded scalar host implementation is documented separately below.
Internal Frame aliases can outlive their frame through separately owned scalar
cells; this is not a request-lifetime guard. Slot unset permits reuse, Frame
close rejects later frame operations, and the value-only runtime adapters have
their own narrower domain/lifecycle. Neither the scalar-only Frame ABI nor the
flat-value adapter exposes general PHP aliases or a persistent Location.

| State | Planned owner and lifetime |
|---|---|
| Generated functions, immutable class/declaration metadata and validated source identities | May be shared by the host process; no request-mutated PHP values stored here |
| PHP global bindings, function-static values, class-static values and evaluated-file bookkeeping | One logical PHP request's `Execution`; repeated/nested calls share it, a new request starts fresh |
| Local variable bindings and expression temporaries | Their explicit Frame/operation; retained aliases may keep a cell alive after a frame ends |
| Mutable objects, property storage and captured environments | The request's reachable value graph, not only the creating frame; no implicit retention in process-wide callback registries or pools |

This targets a traditional PHP request boundary. A long-running worker's
deliberately persistent PHP state would need a separate contract. Go HTTP
handlers must receive/create their own Execution and thread it through calls,
methods and callbacks; a package variable, thread-local or goroutine-local
"current request" is not the ownership mechanism. HTTP hosting is still
unimplemented. Existing comments asking hosts to use fresh contexts are not
proof that globals/statics/callbacks are isolated.

Fresh convenience wrappers represent separate logical executions. An eventual
request adapter must instead initialize its checked entry under the chosen
loading contract, invoke context-aware entries with the same Execution for
sibling/nested calls, and own final teardown. Allocating a fresh context for
each PHP function would incorrectly reset shared state. The global namespace
needs a request owner separate from any one callee's local Frame.

Required location semantics:

- Ordinary assignment copies a value according to its domain; it does not bind
  the destination variable to the source variable. An object identifier may be
  shared while either variable can still be rebound independently.
- `global $x` imports the request-global binding into the local frame by
  reference, not by copying its current value. Missing globals acquire null
  through this explicit import. Ordinary lookup/probe must not acquire that
  creating behavior. These rules follow PHP's
  [global binding contract](https://www.php.net/manual/en/language.variables.scope.php#language.variables.scope.global).
- A reference to existing contents and a location identified by owner/name/key
  are different handles. Unsetting a name removes that binding; an alias may
  still retain its old cell. Recreating the name must not reconnect old aliases
  accidentally. Do not retain pointers into an array's movable `[]Entry`.
- Direct `$GLOBALS['x']` access, when admitted, must reach the same global
  binding. It is not permission to expose the backing Go map as an ordinary
  mutable PHP array. PHP8.1+ has special whole-array write/reference restrictions
  and value-copy behavior; see the
  [official `$GLOBALS` contract](https://www.php.net/manual/en/reserved.variables.globals.php).
  Whole-table snapshots/mutations and dynamic names remain unsupported until
  separately specified and measured.
- Array-element and declared-property locations reuse read/write/bind/probe/
  unset responsibilities, while retaining their own owner, existence and
  evaluation rules. Sharing a protocol does not make every domain supported.
  Keep property visibility, typed/magic/dynamic access and interface dispatch
  explicit; do not treat every object as an unrestricted map.

Function/class-static storage needs stable declaration identities and separate
initialization state, not a local variable name or nil-as-uninitialized test.
Closure instances, inheritance/redeclaration, repeated calls and failed or
recursive initialization require PHP8.3 native contracts before admission;
this document does not settle those cases by choosing a Go map key.

Borrowed operation addresses are not persistent reference bindings. Each new
location must define owner retention, alias rebinding, unset/recreation and
release/invalidation. Request teardown must release owned roots and make stale
request-bound access fail without clearing another Execution. Existing host
property protocols do not establish request ownership: the future adapter must
define a checked transfer/ownership boundary before accepting cross-context
objects or callbacks. The scalar host ABI does not provide that guarantee. Pooling,
cross-request closures, cycles and PHP destructor order remain separate gates.

Keep operation-specific evaluation order from the existing property-append
regression: an RHS can replace the property before the append. A common Location
must not silently impose a universal "resolve the entire LHS, then evaluate
the RHS" rule. Loading also needs an explicit caller scope/Execution and entry
effects; accepting bare `require` by changing it to `require_once` is not a
bootstrap implementation.

Before source admission, require these observable checks within R2–R5:

| Check family | Required evidence | Current status |
|---|---|---|
| Global import versus value copy | Cross-call writes, missing import/null, local shadowing, alias rebinding and unset/recreated names agree with fixed PHP8.3 | Scalar host binding unit checks only; source admission and PHP/Go comparison still planned |
| Two request contexts | Interleave A/B with the same global/static/hook names; closing A cannot change B, and stale A-bound access cannot become B access | Scalar host globals/imports/Close have Go unit checks; static/hook ownership, explicit PHP request-boundary reference and application isolation still planned |
| Arrays and properties | Value-copy isolation, shared object identity, alias-bearing array separation timing and RHS replacement stay correct at admitted locations | Existing bounded value behavior only; persistent mixed-domain references not implemented |
| Original hooks | Reuse the immutable eight-workflow native record, then compare generated events, results, counters and private/public state at matching pins | Native reference recorded; generated comparison not-run |
| Request lifecycle | Define initialization, errors/cleanup and host entry/exit ownership before an HTTP acceptance claim | Planned; HTTP/bootstrap acceptance not established |

Do not add globals or generic references by bypassing the current domain
guards. A scalar-only first global/reference slice must reject array/object or
unknown-location cases explicitly; the actual array-valued WordPress hook table
would still be unsupported. Conversely, a hook-table implementation needs its
array/reference/callable contracts measured before being called compatible.
The existing eager-copy lane remains until that migration is proven. These
design rows do not add tests, change catalog implementation statuses or borrow
the PHP-only reference as generated-Go evidence.

### Scalar request-global host ABI

This first request-state implementation reuses `storage.Frame` and `Slot`.
It adds no source syntax, IR, compiler selection or generated entry behavior.
`global`, `$GLOBALS`, statics and general PHP references remain unsupported.
Only null, bool, int64 and byte strings are accepted; arrays, objects, callables
and foreign Go values reject before storage changes. This cannot hold the
array-valued WordPress hook table.

| Host operation | Contract |
|---|---|
| `Execution.ReadScalarGlobal(name)` | Returns `(value, present)`; missing and present null differ, and lookup never creates a binding |
| `Execution.WriteScalarGlobal(name, value)` | Assigns through the current global cell; invalid values do not allocate or replace storage |
| `Execution.UnsetScalarGlobal(name)` | Removes the global name; imported locals may retain the old cell, and recreating the name does not reconnect them |
| `ScalarFrame.ImportScalarGlobal(execution, name)` | Aliases the same-named global into the local frame, replacing its old local binding; a missing global becomes present null |
| `ScalarFrame.Unset(name)` | Removes only the local binding; a later ordinary Write is independent until another import |
| `Execution.Close()` | Idempotently closes its scalar global storage and imported frames, clears tokenizer/error state and rejects subsequent context operations |

Global names are the closed ASCII form `[A-Za-z][A-Za-z0-9_]*`, excluding
case-sensitive `this` and `GLOBALS`. Leading underscores, variable variables,
superglobals, Unicode names and whole-table access are not admitted by this
ABI. This is an implementation boundary, not PHP's full identifier domain.

Imports retain no exported Slot/address handles. A frame attaches to one
Execution for its remaining lifetime, including after local Unset or Close;
it cannot be rebound to a second request. Closed/nil frames and invalid
inputs reject before a missing global is created. A frame's Close releases
its local bindings without closing the request namespace. Conversely,
Execution.Close releases all still-imported frames, so stale accesses fail.
Closing a frame after its Execution remains safe and idempotent. Ordinary
nonimported frames keep their existing independent lifetime.

The host must use a fresh Execution per logical request, share it through
synchronous sibling/nested calls and defer its Close. Execution and frames
remain pointer-only, noncopyable and nonconcurrent. Generated convenience
wrappers are unchanged; no automatic source-global initialization or request
entry/exit adapter is present. General host objects/callbacks remain outside
this ownership guard, and no PHP destructor/exception ordering is promised.

`gohp.tooling.scalar-request-globals` records Go unit contracts for scalar
alias/copy/unset behavior, request A/B interleaving, teardown and guard failure
atomicity. They are not native-PHP differential evidence or full request
isolation. The PHP-only reference specification and original native hook
baseline are unchanged and are not relabeled by these tests.

Initial implementation checks at clean commit
`7ab0f8d8a64e4cdeafaebe798147ecbdd638654e`
(tree `9a434dd21cc4e7f745a8afeb13b75afbef65e82a`): cached Go1.23.0,
darwin/arm64, `CGO_ENABLED=0`, `GOHP_PHP_TARGET=8.3`, offline dependencies and
the existing shared Go cache. One affected runtime invocation passed (0.227s):

```sh
go test -count=1 ./runtime -run '^(TestScalarRequest.*|TestScalarFrame.*|TestStrtok.*|TestContextualCallable.*|TestJSONAssociativeExecutionIsolationAndGuards|TestPreg83RequiresExecutionWithoutOptionalBackend)$'
```

The four new unit checks and existing affected scalar/context guard checks
passed. No native PHP, Docker, generated program or tagged PCRE2 execution was
performed by this invocation. Five changed Markdown files/163 relative links,
the new anchor, catalog JSON/four test mappings and diff checks passed.
Later documentation-only recording does not change the tested runtime tree.

No new CLI operation, resource route or capability is exposed. These are
developer runtime primitives, excluded from WebMCP; the existing local
`gohp.cli.generate` exclusion continues to apply without input changes.

### Copy timing is part of correctness

The native specification distinguishes these sequences, starting with
`$x = 1; $a = ['ref' => &$x, 'plain' => 0]; $b = $a;`:

| Sequence after the copy | Expected final `[$a['ref'], $b['ref']]` |
|---|---|
| `unset($x); $b['ref'] = 2;` | `[1, 2]` |
| `$b['plain'] = 9; unset($x); $b['ref'] = 2;` | `[2, 2]` |

The second sequence separates array storage while the external alias still
exists. An eager deep-copy which always preserves internal reference cells at
assignment time would incorrectly make the first sequence share as well.
Conversely, recursively copying every cell loses the second sequence's alias.
This is a proposed contract until the pinned native run is recorded in the PR;
it is never evidence of generated-Go behavior.

PHP 8.3.36's
[`zend_array_dup_value`](https://github.com/php/php-src/blob/php-8.3.36/Zend/zend_hash.c)
unwraps a singly-owned element reference when duplicating array storage, with a
self-array exception. Its
[`SEPARATE_ARRAY`](https://github.com/php/php-src/blob/php-8.3.36/Zend/zend_types.h)
separates shared storage. These are implementation references, not code to copy
into GoHP and not a promise that an approximation covers all cases.

Go's garbage collector cannot supply the PHP-semantic ownership counts needed
here. The runtime must track logical owners/edges and release bindings,
temporary values and frames at defined points, independently of Go GC timing.
Host API entry/exit and Go value copies must not silently change those counts.
Self/cyclic references must not recurse indefinitely through `Clone`; bounded
cycle observations are specified, but general cycle reclamation, PHP destructor
order and `ReflectionReference` compatibility remain unresolved. Do not expose
cyclic reference support before its separate implementation and evidence.

### Evaluation and calls

Do not impose a universal “resolve full LHS, then RHS” transformation. Evaluate
receivers, keys and RHS only as specified by each admitted operation. In
particular, retain #80's property-append regression: an RHS that resets that
property must be followed by append to the post-RHS property. Later effectful
nested addresses, alias rebinding during RHS and by-reference arguments need
their own PHP comparisons before admission; source order is not inferred from
Go's argument order.

Direct functions, methods and dynamic callables must use the same signature
metadata for value/reference passing and return mode. Dynamic resolution must
retain argument address information until the passing mode is known; copying
all arguments into `[]any` first loses it. Builtins with reference parameters and
`call_user_func`-family adapters need individual contracts, not a universal
“pass every argument by reference” fallback.

Keep `Execution` explicitly owned by the logical invocation/request. Its current
effect-based strtok/JSON/PCRE propagation is not a complete request-state design. When
callbacks or other stateful operations are enabled, propagate context by their
actual effects/signatures; no process-global PHP variables, static state or
tokenizer. A generic callable registry may follow this foundation. Do not
substitute a custom Go hook engine for untested original `WP_Hook` behavior.

## Delivery gates and status

| Stage | Deliverable / exit condition | Status |
|---|---|---|
| R0 | Design and executable PHP-only reference contracts; record exact native run separately | This change; GoHP reference support remains unsupported |
| R1 | Internal slot/value ownership primitives and shared-array separation design implemented behind closed boundaries; existing no-reference behavior preserved | Partial: scalar frames/codegen, shared containers/key-index policy, owned Slots/mutations/temporaries, runtime adapter and scalar-entry generated scopes with bounded if/while implemented; cross-domain loops, array/untyped parameter/general-frame ownership, persistent locations and nested ownership pending |
| R2 | Bounded variable reference assignment/unset, explicit address IR and binding lifetime; source + malformed-IR negatives and PHP/Go comparisons | Not implemented |
| R3 | Reference-bearing arrays, element/declared-property locations, copy-timing and lifetime cases; key order/next-index and existing copy isolation retained | Not implemented |
| R4 | By-reference parameters/returns, reference foreach and closure captures, each with explicit bounded contracts and call-path evidence | Not implemented |
| R5 | Generic dynamic callable forms and original application acceptance (full block parser, then hooks) | Generic callables/hooks not implemented. The separate complete-parser case-limited Gossip evidence does not establish general references or whole-application acceptance |

Next bounded implementation is selected from the original hook/state blockers:
connect the guarded scalar request bindings to an explicitly checked source
entry/global-import contract, with native PHP8.3 comparison, before enabling
source `global`/references. Top-level initialization, scope/effect discovery and
all storage lanes must retain the same request owner. Array/untyped parameter entry
domains or broader loop invariants are justified only where that selected slice
needs them, not as a parallel storage-expansion queue. Build on
the first expression-scoped lowering, shared scalar containers, value temporaries,
synchronous Slot-rooted mutations and explicit Frame release. The synchronous
borrow is not a persistent address or a compiler evaluation-order policy. Keep parser acceptance unchanged;
exercise real alias/copy/lifetime contracts, then integrate the storage model
across codegen/runtime boundaries in reviewed increments. If a narrower scalar
R2 subset precedes array migration, reject arrays/objects/unsupported locations
explicitly instead of silently falling back to eager reference copying.

Each implementation change records its affected tests once for that revision
and configuration, reuses unaffected passing evidence explicitly, and adds
boundary cases only for actual contracts. Preserve the existing normal-value
corpus during migration. Extend the catalog with generated-Go comparisons only
when implemented; negative reference rejection is not positive support.
If the IR address/signature schema becomes incompatible, bump the IR version
and reject old shapes rather than silently changing `0.1`. Rebuild consumers
against the matching compiler/runtime commit; CLI/report compatibility and
source-map/error behavior need explicit checks at that migration boundary.

## Native-only specification and reproduction

[`tests/oracle/reference_semantics.php`](../tests/oracle/reference_semantics.php)
is a specification executable, **not a GoHP input fixture**. It uses exact typed
observations and explicit exceptions (not disableable `assert`), rejects other
PHP series/below-floor versions, and fails on PHP warnings/notices. Stable case
IDs and a final `native-oracle-only` JSON report distinguish this evidence from
the compiler compatibility ledger. It neither runs generated Go nor certifies
PHP-compatible warnings/errors. No network, packages or external state needed.

From the repository root, with the approved PHP 8.3 oracle:

```sh
php tests/oracle/reference_semantics.php
```

For reproducible container execution use the immutable PHP image selected in
`.github/workflows/ci.yml`, network disabled and a read-only source mount. Record
commit/tree, fixture SHA-256, actual PHP version, resolved image digest, exact
command, exit and stderr in the PR/current-state evidence. The script is **not
automatically run by `scripts/check.sh` or the compatibility CI** yet. Its manual
PASS cannot increase the ledger's feature/check counts. Future PHP versions and
generated-Go execution need independent evidence and versioned expectations.

This change does not add a service/user operation or change the existing
`gohp.cli.generate` WebMCP exclusion. No deployment or WordPress/Composer
compatibility claim is made.
