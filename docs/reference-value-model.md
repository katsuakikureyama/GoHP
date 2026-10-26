# Reference, value and assignment foundation

Status: **design + native-oracle specification; scalar frames, shared scalar containers, owned array-valued Slots, synchronous mutations, value temporaries and an opt-in runtime value adapter implemented**.
Baseline: main `565e65e6c9d8e211b06779739bfba76b07c4d92d` (#81).
Target: PHP 8.3, oracle floor 8.3.36. Later series need separate policy and
evidence; they must not inherit these results. No compiler/runtime/IR changes
were introduced by the original design or its PHP-only specification. The first
implementation is [scalar slot infrastructure](scalar-slot-foundation.md), kept
with [explicit scalar Frame lifecycle](storage-frame-lifecycle.md) and
[bounded generated integration](generated-scalar-frames.md) for nullary or all-exact-scalar-parameter
functions; PHP reference syntax remains unsupported. The native specification below is
still PHP 8.3-only even though the compiler now has an independent PHP 8.4 target.
[Shared scalar containers](shared-scalar-array-storage.md) now exercise internal
separation/copy timing; runtime arrays and generated array code are not migrated.
[Owned Slot array values](owned-array-slot-values.md) now have explicit retain/read/release
and synchronous one-level mutations without a Read/copy/writeback temporary;
Frame and element bindings remain scalar-only, including through aliases.
Internal value temporaries now capture, transfer and release owned snapshots.
An opt-in runtime ValueSlot/ValueTemporary facade converts canonical flat values
without exposing storage types; codegen does not emit these APIs yet.

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
conditional `strtok` propagation is not a complete request-state design. When
callbacks or other stateful operations are enabled, propagate context by their
actual effects/signatures; no process-global PHP variables, static state or
tokenizer. A generic callable registry may follow this foundation. Do not
substitute a custom Go hook engine for untested original `WP_Hook` behavior.

## Delivery gates and status

| Stage | Deliverable / exit condition | Status |
|---|---|---|
| R0 | Design and executable PHP-only reference contracts; record exact native run separately | This change; GoHP reference support remains unsupported |
| R1 | Internal slot/value ownership primitives and shared-array separation design implemented behind closed boundaries; existing no-reference behavior preserved | Partial: scalar frames/codegen, shared scalar containers/key-index policy, owned Slots, synchronous mutations, value temporaries and opt-in runtime value adapter implemented; persistent locations, generated temporary/frame integration and nested ownership pending |
| R2 | Bounded variable reference assignment/unset, explicit address IR and binding lifetime; source + malformed-IR negatives and PHP/Go comparisons | Not implemented |
| R3 | Reference-bearing arrays, element/declared-property locations, copy-timing and lifetime cases; key order/next-index and existing copy isolation retained | Not implemented |
| R4 | By-reference parameters/returns, reference foreach and closure captures, each with explicit bounded contracts and call-path evidence | Not implemented |
| R5 | Generic dynamic callable forms and original application acceptance (full block parser, then hooks) | Not implemented; not implied by R0–R4 |

Next bounded implementation: expression-level ownership planning and a proven
closed-scope selection for R1's generated temporaries, then operation-specific locations
across calls. Build on shared scalar containers, internal value temporaries,
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
