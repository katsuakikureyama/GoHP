# Generated scalar frames (bounded R1 integration)

The canonical `internal/storage.Frame` now backs a closed subset of generated
functions through `runtime.ScalarFrame`. Generated Go still imports only the
public runtime: consumers can build outside the GoHP module without illegally
importing its internal storage package. This is a storage migration, not new
PHP syntax or reference support, and not a performance claim.

## Selection and lifetime

Only global functions with no parameters or **all exact int/string/bool**
parameters whose **entire body** stays in a proven scalar domain use the frame.
The conservative classifier runs after normal
semantic validation; it does not admit previously rejected source or IR.

- Expressions: int/string/bool/null literals, defined locals,
  [target major/minor constants](target-version-constants.md), the existing
  scalar `+`, `-`, `.`, `===`, `!==`, `&&`, `||`, boolean cast/not, ternary and
  defined-variable `isset` (including multiple such operands).
- Calls: only the explicit pure-builtin list `strlen`, `strtolower`, `strcasecmp`,
  `rtrim`, `substr`, `str_contains`, `str_starts_with`, `is_array`, `is_bool`,
  `is_int`, `is_null`, `is_string`, `is_object`, and [bounded `gettype`](gettype-boundary.md). Names are case-insensitive and
  **every operand recursively needs the same scalar proof**. A scalar result
  alone is insufficient: `is_array([])` and `strlen(user_function())` stay on
  the map path. No new accepted source, arity or type coercion is introduced.
- Statements: ordinary assignment, echo, return, if/else, while, break/continue,
  and result-discarding calls which satisfy the same builtin proof.
  Both branches and dead tails are inspected. Each operand must also qualify.
- All unlisted calls, untyped/mixed parameters (including optional-null ones), methods, arrows,
  arrays/objects, foreach, unknown forms and top-level main retain the existing
  value-map implementation. No runtime guessing, per-value fallback or
  interprocedural return-type assumption selects a frame.

The builtin list is deliberately explicit, not an inference from a PHP return
annotation or a promise to select every pure helper. Array consumers/producers,
user functions, callable/method calls and stateful `strtok` are excluded, as are
currently unlisted scalar helpers such as `sprintf` and `addslashes`. Normal
semantic validation and runtime guards still apply; mixed scalar values from
conditional expressions may pass the storage proof but fail a builtin's type
guard when reached. Lazy branches remain lazy. Updating a listed builtin to
return arrays/objects would require revisiting this proof before admission.

Selected functions create one `rt.NewScalarFrame()` after their arity/context
guards and immediately `defer v.Close()`. Typed parameters pass ScalarParameter
before their first Write; later-parameter failure also runs the deferred release.
See the separate [scalar parameter contract](scalar-function-parameters.md).
Variable reads and writes use Read
and Write; return guards, scalar copying, short-circuit behavior, Execution
propagation and statement source maps remain unchanged. Explicit/implicit
return and Go panic unwinding all run the deferred release. The scaffolding is
not attributed to a PHP source statement.

`ScalarFrame` is a thin pointer-only, non-copyable, non-concurrent facade over
canonical storage. Read of an absent local panics; present null remains nil.
Write retains storage's exact scalar guard. Close is idempotent and closed
reads/writes reject. No Bind, Reference, Slot or other alias operation is exposed
to generated code. No finalizers or GC-timed releases are introduced.

Existing untyped/array/callable/object behavior is untouched. This deliberately
does not solve array ownership by smuggling an eager-cloned array into scalar
storage. The shared-array separation design remains an independent R1/R3 gate.

## Evidence and boundaries

`gohp.tooling.generated-scalar-frames` records twelve checks independently for
PHP 8.3 and 8.4: the original seven frame contracts plus five builtin-integration
checks (native/generated comparison, selection, lowering/source maps, preserved
rejection and an external-module host consumer).
`tests/fixtures/scalar_frame.php` observes scalar copies, byte/UTF-8 strings,
boundary integers, conditional/loop exits, explicit/implicit return values,
repeated calls and coexistence with legacy array/arrow scopes. The normal
comparison measures stdout and successful exit plus the existing `quiet-success/v1`
requirement that both observed stderr streams are empty. This is not arbitrary
diagnostic equality or hidden destructor behavior. Internal storage's existing logical-owner release tests
remain separate; the fixture does not claim to observe reference release.

`scalar_frame_builtins.php` adds nested string helpers/predicates, local copies,
conditional/loop paths, discarded results and legacy array coexistence. The
builtin external consumer exercises runtime guard failure and recovery, host
types/arity/nil Execution and preservation of the caller's tokenizer state.
Scalar temporaries remain ordinary Go values. The separate internal
[owned value temporary](owned-array-slot-values.md#owned-value-temporaries)
can capture/transfer/release scalar-array snapshots through a separate opt-in
ValueSlot/ValueTemporary facade, but not this scalar frame or codegen. This does
not implement generated shared-array temporary lifetimes or a PHP exception model.

The external consumer checks the generated ABI, nil Execution and arity guards,
runtime overflow failure, subsequent calls, preserved invocation state and
legacy array copies without PHP. CI records the exact revision/runtime and
target-specific evidence. New generated files require a runtime revision that
provides ScalarFrame; existing generated Go remains compatible. IR stays at
0.1: the parameter extension reuses aligned ParamTypes; previous shapes and
accepted meanings are preserved.

R1 is partial. General frame migration, address IR, temporary and shared-array
ownership, PHP references, by-reference calls/returns, cycles and destructor
order remain unimplemented. There is no PHP warning/exception-equivalence or
Gossip/WordPress acceptance claim. The reference feature stays unsupported.

No user command/route/authentication/capability is added: the existing local
`gohp.cli.generate` WebMCP exclusion applies. No CLI options, target policies,
oracle environment, CI configuration, Gossip pins or deployment are changed.
