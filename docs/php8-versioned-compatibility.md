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
- Oracle: actual PHP_VERSION, immutable image digest, bounded CLI startup
  baseline (extensions and versions, selected php.ini settings, locale,
  timezone and platform details), and its canonical metadata hash.
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

The bounded `PHP_MAJOR_VERSION` and `PHP_MINOR_VERSION` expressions demonstrate
that rule without adding a process-global target: their symbolic `version_constant`
IR nodes are resolved at code generation from each program's `PHPVersion`.
For enabled targets, `8.3` yields `8`/`3` and `8.4` yields `8`/`4`, independently
of the compiler host and native-oracle patch. The [target version constant
contract](target-version-constants.md) defines its syntax and exclusions.

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

The current v0.4 harness measures stdout and normal exits, plus a bounded
startup baseline from one separate `php -r` probe before the suite. The probe
uses the same inherited environment and records only the documented allowlist:
CLI SAPI, integer size, default timezone, selected INI settings, selected
available locale categories and sorted regular/Zend extension lists. It fails
closed before tests if it cannot produce the required baseline, and a canonical
JSON SHA-256 detects accidental metadata mismatch. This is not a full
environment/`phpinfo`/INI capture, per-fixture effective state, authenticity
proof, reproducibility guarantee or automatic native-version-versus-compiler
classification. Positive differential fixtures also record `quiet-success/v1`:
the fixture's native PHP process is launched with explicit `-d` values that
expose all error reporting, send display output to stderr and suppress error
logging; generated Go's stderr is observed separately. A result passes this
profile only if both streams are empty. The evidence stores only byte counts and
SHA-256 values, not raw stderr. It is intentionally not warning/error/exception
text equality, and general stderr, exceptions and side effects remain
`not-checked`. The PHP override is not applied to the baseline/version probe,
Go build or generated Go execution. Fixture-local setup—such as the booleans
fixture's `LC_CTYPE=C`—remains source-hash-bound rather than inferred from the
baseline, and a fixture can still override or suppress warnings internally.
For each process actually observed, v0.4 stores only stdout byte count, SHA-256
and exit code (with `-1` for no portable numeric status), not raw stdout; absent
processes stay absent rather than becoming zero-valued results. Those fields are
event-hash-bound and replayed against the stdout/exit verdict. A separate
read-only two-archive comparison can call out a local PHP-versus-Go stdout
mismatch or a native stdout difference only when the candidate/corpus identity,
clean inputs and quiet successful processes are all compatible. A native
difference is an observation under the displayed oracle environments, never
proof that the PHP target caused it or that a PHP specification change intended
it. Nonzero exits, exceptions, diagnostics and side effects remain outside that
comparison.

The current harness does not yet provide the full diagnostics/state/configuration
evidence above. Historical v0.1/v0.2/v0.3 archives retain their original scope:
no baseline, quiet-success or process-output record is invented for them, while
the v0.4 reader can display them without granting new guarantees. PHP's [`ini_get`](https://www.php.net/manual/en/function.ini-get.php),
[`get_loaded_extensions`](https://www.php.net/manual/en/function.get-loaded-extensions.php),
[`setlocale`](https://www.php.net/manual/en/function.setlocale.php) and
[`date_default_timezone_get`](https://www.php.net/manual/en/function.date-default-timezone-get.php)
define the corresponding observations.

## Delivery and current boundary

Targets **8.3** and **8.4** are enabled for the documented language subset.
Their native oracle admission floors are **8.3.36** and **8.4.27** respectively.
The default remains 8.3. The target labels, floors, CI image identities and
target-specific profile rules are source configuration; each target still needs
its own successful differential evidence for a given revision. The same bounded
corpus is deliberately executed once per target. An 8.3 PASS is never an 8.4
PASS, and absence of an 8.4 report is `unverified`, not an inherited success.

PHP 8.4 has explicit target behavior: the target-resolved major/minor constants
share the documented cross-target boundary, while the four new core array
function names are reserved against declarations and class `_` is rejected
because its native deprecation is not modeled. The profile does not implement
the new functions or reproduce PHP diagnostics. Nullable typed defaults,
`exit`/`die`, property hooks and asymmetric property visibility remain
unsupported. See [the PHP 8.4 target contract](php84-target.md).

Implement the direction in bounded increments:

1. Extend target/configuration evidence and per-target expectations where the
   current registry/report schema is insufficient.
2. Extend 8.4 evidence from its independent CI row, then add another minor with
   its own pinned oracle, relevant semantic differences, positive/negative
   cases and CI job. Keep enabled targets as regression rows.
3. Publish target-by-target results, then expand targets and observable behavior
   using the same admission process. No specific new minor is selected here.

An enabled target requires a coherent policy, frontend/runtime behavior,
required corpus, pinned oracle and reporting path. A target being enabled is
still not a claim that every PHP feature or application works on it. Application
acceptance records additionally pin the application version, dependencies and
configuration, as Gossip already does for its bounded source selections.

The original 2026-10-26 direction was documentation-only. The subsequent bounded
target-version-constant increment implements the contract described here, while
leaving existing CLI capabilities and their WebMCP exclusions unchanged and
adding no service operation.
