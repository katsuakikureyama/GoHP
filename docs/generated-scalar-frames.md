# Generated scalar frames (bounded R1 integration)

The canonical `internal/storage.Frame` now backs a closed subset of generated
functions through `runtime.ScalarFrame`. Generated Go still imports only the
public runtime: consumers can build outside the GoHP module without illegally
importing its internal storage package. This is a storage migration, not new
PHP syntax or reference support, and not a performance claim.

## Selection and lifetime

Only parameter-free global functions whose **entire body** stays in a proven
scalar domain use the new frame. The conservative classifier runs after normal
semantic validation; it does not admit previously rejected source or IR.

- Expressions: int/string/bool/null literals, defined locals, the existing
  scalar `+`, `-`, `.`, `===`, `!==`, `&&`, `||`, boolean cast/not, ternary and
  defined-variable `isset` (including multiple such operands).
- Statements: ordinary assignment, echo, return, if/else, while, break/continue.
  Both branches and dead tails are inspected. Each operand must also qualify.
- All calls, parameters (including optional-null ones), methods, arrows,
  arrays/objects, foreach, unknown forms and top-level main retain the existing
  value-map implementation. No runtime guessing, per-value fallback or
  interprocedural return-type assumption selects a frame.

Selected functions create one `rt.NewScalarFrame()` after their arity/context
guards and immediately `defer v.Close()`. Variable reads and writes use Read
and Write; return guards, scalar copying, short-circuit behavior, Execution
propagation and statement source maps remain unchanged. Explicit/implicit
return and Go panic unwinding all run the deferred release. The scaffolding is
not attributed to a PHP source statement.

`ScalarFrame` is a thin pointer-only, non-copyable, non-concurrent facade over
canonical storage. Read of an absent local panics; present null remains nil.
Write retains storage's exact scalar guard. Close is idempotent and closed
reads/writes reject. No Bind, Reference, Slot or other alias operation is exposed
to generated code. No finalizers or GC-timed releases are introduced.

Existing parameter/array/callable/object behavior is untouched. This deliberately
does not solve array ownership by smuggling an eager-cloned array into scalar
storage. The shared-array separation design remains an independent R1/R3 gate.

## Evidence and boundaries

`gohp.tooling.generated-scalar-frames` records seven checks independently for
PHP 8.3 and 8.4: a native/generated comparison, selection/source-map/rejection
contracts, an actual external-module consumer and two runtime-facade contracts.
`tests/fixtures/scalar_frame.php` observes scalar copies, byte/UTF-8 strings,
boundary integers, conditional/loop exits, explicit/implicit return values,
repeated calls and coexistence with legacy array/arrow scopes. The normal
comparison measures stdout and successful exit plus the existing `quiet-success/v1`
requirement that both observed stderr streams are empty. This is not arbitrary
diagnostic equality or hidden destructor behavior. Internal storage's existing logical-owner release tests
remain separate; the fixture does not claim to observe reference release.

The external consumer checks the generated ABI, nil Execution and arity guards,
runtime overflow failure, subsequent calls, preserved invocation state and
legacy array copies without PHP. CI records the exact revision/runtime and
target-specific evidence. New generated files require a runtime revision that
provides ScalarFrame; existing generated Go remains compatible. IR stays at
0.1 because its accepted shapes and meanings have not changed.

R1 is partial. General frame migration, address IR, temporary and shared-array
ownership, PHP references, by-reference calls/returns, cycles and destructor
order remain unimplemented. There is no PHP warning/exception-equivalence or
Gossip/WordPress acceptance claim. The reference feature stays unsupported.

No user command/route/authentication/capability is added: the existing local
`gohp.cli.generate` WebMCP exclusion applies. No CLI options, target policies,
oracle environment, CI configuration, Gossip pins or deployment are changed.
