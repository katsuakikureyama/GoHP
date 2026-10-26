# Product policy: independently verified PHP 8.x compatibility

GoHP is a **transpiler designed to verify compatibility separately for PHP 8.x
version targets**, not a PHP 8.3-only compiler and not a claim that all PHP 8.x
programs already work. This is the owner's product direction. The first active
target remains the PHP 8.3 series with a native oracle floor of 8.3.36.

## Target, oracle and evidence are different

| Dimension | Meaning |
|---|---|
| Compiler target, e.g. `8.3` | Selected language/semantic series, fixed for the compilation and carried through IR/generated artifact provenance |
| Oracle release, e.g. `8.3.36` | Actual native PHP patch release used to measure behavior; must pass that target's explicit version policy |
| Oracle image digest | Immutable execution-image identity, recorded separately from the PHP version and source revision |
| Feature contract | Bounded inputs/behavior implemented for the selected target, not the entire PHP language |
| Run result | Evidence at one compiler/runtime revision, target, oracle and configuration; not a transferable support badge |

Minor-series targeting does not erase patch-level differences. Record the exact
patch and expectations; investigate a behavioral change before updating a floor
or digest. Passing on one release is not proof for every later patch. Add finer
semantic profiles/expectations if measured differences require them; do not
change results merely to make another oracle pass. A floor is not a general
security guarantee for the compiler, generated programs or deployments.

## Current delivery status

| Series | Status |
|---|---|
| PHP 8.3, native >= 8.3.36 | Only active compiler/differential target; closed subset, not full-language compatibility |
| PHP 8.4 and later 8.x | Intended independent additions; disabled until their own semantics, oracle policy and evidence are ready |
| PHP 8.0–8.2 | Not active targets; the PHP 8.x product direction does not silently expand the agreed initial scope |
| PHP 7.x | Not an active product target; retained 7.4 evidence is historical only |

Keep unsupported targets rejected at CLI, source frontend and direct IR
validation. Neither the host's installed PHP nor an image tagged `latest` may
select the compiler's behavior. Do not offer an ambiguous `8.x` or “latest”
compiler target that silently changes over time.

## Architecture seams

The current `internal/phpversion.Policy` registry owns enabled target labels and
oracle minimums. CLI/frontend/semantic validation use it, `ir.Program.PHPVersion`
carries the target, and generated-source/report provenance records it. This is
an admission boundary; **multiple-version parser/semantic profiles are not yet
implemented**. Merely adding another row to `policies` is insufficient.

When admitting another series, resolve an immutable semantic profile from the
explicit target at compilation. Keep common operations shared and put measured
version differences in named, testable frontend/validation/lowering/runtime
policies. Prefer target-specific lowering where possible. If runtime operations
need a profile, pass it explicitly with the generated program's invocation
context; never use mutable process-global “current PHP version” state. Multiple
programs targeting different versions must not interfere with each other.
These are future integration requirements, not claims about today's Execution API.

IR schema version and PHP target are separate axes. An incompatible IR shape or
meaning requires its own version change and rejection of old shapes; selecting
another PHP series does not itself authorize changing the IR ABI silently.
Generated code must use the matching pinned GoHP runtime revision.

The [reference/value foundation](reference-value-model.md) follows this policy.
Its scalar-slot unit tests establish internal contracts only. Reference-array
copy timing and other semantics require native evidence for every claimed target;
an internal helper without a version argument is not automatically proven common
to every PHP 8.x release.

## Independent verification and admission gate

For each additional target, deliver together:

1. Explicit series policy, approved minimum native release and pinned oracle
   image; retain the actual version/image identity in every run.
2. Reviewed syntax/semantic delta inventory. Separate shared cases from cases
   whose acceptance, results, diagnostics or side effects differ by target.
3. Implemented target-specific behavior and negative boundaries. A shared source
   case is reusable, but its PASS is not: execute it against each admitted target.
4. Target-scoped catalog contracts/expected outcomes and an independent native
   differential job/artifact. No averaging or combining results across targets.
5. Updated CLI help, view filters/input schemas, documentation and consumer pins;
   record unmeasured dimensions and unresolved features explicitly.

Current evidence tools already enforce one target per run and validate the
differential payload's target/oracle against that run. The archive viewer keeps
runs separate; filtering is not a combined compatibility verdict. Catalog rows
outside the measured target remain unmeasured. Keep these invariants when adding
profiles or a CI matrix entry. The existing
[ledger contract](compatibility-ledger.md) distinguishes unit/compiler/negative
evidence from native differential evidence and rejects missing/incomplete runs.

Do not promote a target by copying 8.3 results, enabling a broad CI version range,
or accepting syntax without behavioral comparison. A new target starts
unverified; only measured bounded contracts can be reported as passing. Current
CI measures stdout and successful exit for its differential cases; warnings,
exception identity and other side effects need separate observations before
they can be claimed equivalent.

This policy does not enable 8.4+, lower the 8.3.36 floor, change a runtime route,
add an interpreter, or claim WordPress/Composer application compatibility.
