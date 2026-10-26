# PHP 8.x version-specific compatibility

Owner direction, 2026-10-26: design GoHP as a transpiler that can verify
compatibility separately for each PHP 8.x version.

This is the product and architecture direction for the general-purpose PHP-to-Go
compiler. Gossip is an application consumer and acceptance corpus; its needs
help discover language gaps but do not define the compiler's language semantics.

## Target and evidence identity

The selected PHP major/minor defines the language semantics to preserve. The
exact native patch release identifies the oracle used to test those semantics.
Model PHP 8.0, 8.1, 8.2, 8.3, 8.4, 8.5 and subsequent 8.x series as independent
targets as they are admitted. This list is a design scope, not enabled support.

Keep these identities separate:

- Target series: explicit compiler input, propagated through frontend, IR,
  semantic validation, generated code and relevant runtime operations.
- Oracle: actual PHP_VERSION, immutable image digest, extensions and versions,
  php.ini settings, error reporting, locale, timezone and platform details.
- Candidate: compiler/runtime commit, generated source hash, Go toolchain,
  platform/build flags and dependency identities.
- Corpus: stable case IDs, exact source/input hashes, expected outcome per
  target and the dimensions actually observed.

The current minimum oracle patch is an execution admission rule. It is not
evidence for every patch in that series. Patch-specific differences require
separate records; adding a newer oracle must preserve earlier results.

## Semantic architecture

Use the existing `internal/phpversion` registry as the target admission boundary.
Centralize version-dependent rules in explicit semantic profiles as needed;
shared behavior can share implementation. Every difference must name its
applicable targets and supporting cases. Avoid scattered guesses based on the
installed PHP or a process-global current version.

Keep `Program.PHPVersion` meaningful through lowering. A version-dependent
runtime operation must receive or be generated against the selected profile.
Two generated targets hosted in one Go process must not change each other's
semantics. The storage/reference foundation must retain room for these profiles
without assuming that an 8.3 observation applies to every PHP 8.x target.

Preserve the selected target's behavior, including intentional differences
between targets. Cross-version migration or rewriting behavior is a separate
future operation and must never happen implicitly during transpilation.

## Verification and reporting

For each case and admitted target, execute the original source with that
target's pinned native PHP, then execute Go generated for that same target.
Compare the observed behavior within the declared contract. Also retain native
outcomes across targets so an intentional PHP version difference can be
distinguished from a transpiler regression.

Use common case IDs for common contracts and explicit target expectations for
differences. A syntax form rejected by a particular PHP target has an explicit
negative expectation; it is not silently skipped. Unsupported compiler syntax,
native behavior mismatch, unavailable oracle and unmeasured behavior must remain
distinguishable. Native-only observations never become generated-Go PASS.

The compatibility view is a feature/case-by-target matrix backed by immutable
per-run evidence. Keep implementation state separate from verification state.
Never inherit a PASS from another minor, hide a failure behind an older run, or
report an undifferentiated PHP 8.x compatibility percentage.

Required comparison dimensions depend on the claimed contract: typed results
and bytes, exit behavior, warnings/deprecations, exception identity, state
mutation and external side effects where relevant. Encode unmeasured dimensions
explicitly. Nondeterministic fields may be normalized only by documented,
case-specific rules that preserve meaningful differences.

The current v0.1 harness measures stdout and normal exits; it does not yet
provide the full diagnostics/state/configuration evidence above. Evolve the
evidence schema and readers explicitly when adding these fields. Historical
archives keep their original scope and do not gain new guarantees.

## Delivery and current boundary

Currently only target **8.3** is enabled, with a native oracle admission floor of
**8.3.36**, and only the documented language subset is implemented. Other PHP
8.x targets are planned/unverified. This document enables no new targets.

Implement the direction in bounded increments:

1. Extend target/configuration evidence and per-target expectations where the
   current registry/report schema is insufficient.
2. Add one additional minor with its own pinned oracle, relevant semantic
   differences, positive/negative cases and CI job. Keep 8.3 as a regression row.
3. Publish target-by-target results, then expand targets and observable behavior
   using the same admission process. No specific new minor is selected here.

An enabled target requires a coherent policy, frontend/runtime behavior,
required corpus, pinned oracle and reporting path. A target being enabled is
still not a claim that every PHP feature or application works on it. Application
acceptance records additionally pin the application version, dependencies and
configuration, as Gossip already does for its bounded source selections.

This decision changes documentation only. Existing CLI capabilities and their
WebMCP exclusions are unchanged; no service operation is introduced.
