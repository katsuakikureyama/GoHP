# Internal scalar slots — first R1 slice

This is internal storage machinery, **not PHP reference syntax support**.
`runtime/scalar_slot.go` is unexported and not connected to generated frames,
codegen, the parser, IR or public host APIs. Existing `Clone`, arrays, object
identity and execution context behavior are unchanged. R1 is still partial:
shared array storage and automatic frame/temporary ownership are not implemented.

It follows the [reference/value design](reference-value-model.md) and GoHP's
[version-specific PHP 8.x verification policy](php-version-compatibility.md).
The initial behavioral reference is PHP 8.3.36; these unit checks do not establish
compatibility with every PHP 8.x release or add another active compiler target.

## Closed internal contract

| Operation | Internal contract |
|---|---|
| `newScalarSlot` | Create one independent, undefined binding; undefined is not present null |
| `isDefined` / `read` | Probe presence without mutation; reading undefined fails; read returns a scalar payload, never a reference cell |
| `assign` | Accept only exact Go `int64`, `string`, `bool`, or untyped nil; write through an existing alias; reject everything else before mutation |
| `bind` | Source must already be defined; rebind only the destination to the source cell, leaving the destination's former aliases intact |
| `unset` | Release just this binding; repeated unset is harmless; assigning afterwards creates independent contents |

Go copies of `scalarSlot` are handles to the **same** binding. They do not create
a PHP value copy or another reference owner. Create a distinct slot, then use
`assign(source.read())` for a scalar value copy or `bind(source)` for a distinct
alias. The zero Go handle is invalid; every operation diagnoses it explicitly.
Self-binding and binding two slots already in the same alias set are no-ops.

Reference cells count distinct live bindings. Binding another slot increments
that count; rebind/unset releases the old binding; final release clears the
payload. Go GC and finalizers are not used to decide logical ownership. Owners
must explicitly release slots; frame cleanup, temporary lifetime and escapes
are later integration work. Operations are synchronous, not concurrency-safe.

Arrays (including empty ones), objects, callables, floats, Go `int`, typed nil,
foreign values, slots and cells are rejected as payloads. No general coercion,
address lookup, undefined-source creation, element/property access, cycle or
destructor behavior is introduced. This prevents reference-bearing arrays from
accidentally using the existing eager `Clone` semantics.

## Evidence and status

Five runtime unit checks cover scalar copies versus shared writes, destination
rebinding and unset/recreation, retained bindings and parameter-like local
rebinding, Go handle identity/owner counts, and failure without state mutation.
The parameter-like sequence is explicit slot operations, **not generated
function-call or by-reference-return evidence**.

The catalog attaches these `unit.scalar_slot.*` checks to the existing
`php.syntax.array-write-and-reference` entry, which stays `unsupported`. The
existing source-rejection checks remain. Unit PASS is not PHP differential
PASS, and the historical 15-case native-only oracle from [PR #82](https://github.com/katsuakikureyama/GoHP/pull/82)
is not rerun or relabeled as generated-Go compatibility evidence.

Next: explicit frame/temporary lifetime and reference-aware array ownership /
separation in bounded internal increments, then migrate codegen locations and
calls with target-specific differential evidence. Keep `&` and `unset` closed
until that integration's own bounded contract and negative checks pass.
No new CLI, service/user operation, WebMCP capability or authentication route.
