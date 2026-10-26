# Compatibility registry and verification ledger (design v0.4)

## Decision and delivery status

The [PHP 8.x direction](php8-versioned-compatibility.md) governs this registry:
feature/case-by-target evidence must distinguish deliberate native version
differences from transpiler mismatches. Current schemas implement only the
documented subset of that design; planned targets do not inherit 8.3 results.

Maintain a **function/feature catalog** separately from an **immutable test-run
ledger**. A catalog describes supported contracts and limits; a successful run
proves specific cases at a revision and environment, not all PHP behaviour.
Parsing, emitting Go, compiling and native-PHP behaviour comparison are distinct.

The initial [`compatibility/catalog.json`](../compatibility/catalog.json) now has
an experimental executable implementation: validated catalog, stable case IDs,
single-run test evidence, per-target JSON/Markdown and CI artifacts. See the
[ledger operation contract](compatibility-ledger.md). The remaining sections
preserve the design direction; read-only multi-run views and target/kind/status
filters are implemented in `gohp-compat-view`. The separate
`gohp-compat-compare` reads two compatible archives to diagnose bounded
hash-only stdout observations; it is not a combined support verdict. Historical
applicability, combined support verdicts, PHP callsite inventory and service APIs
are still follow-ups.
A read-only [Composer metadata inventory](compatibility.md#composer-metadata-inventory)
exists without resolving requirements or certifying packages. No dashboard, automatic
compatibility promotion or deployment is added.

## Catalog: what do we implement?

| Field | Contract |
|---|---|
| `id` | Stable identity such as `php.builtin.count`; never display text |
| `kind`, `name` | Builtin, syntax, operator, type, extension, application or tooling; displayed name |
| `implementation` | `unsupported`, `partial`, or `implemented` within the stated contract; not a test verdict |
| `contract` | Exact supported input/operation domain, e.g. one-argument `count(array)` |
| `php_targets` | Intended target minors, not evidence of PASS |
| `checks` | Real test file/name, kind, optional fixture and measured assertions |
| `limitations` | Unsupported inputs/modes and unverified behaviour |

Use separate entries or case IDs for distinct signatures/modes when implemented.
Recursive `count`, scalar input and default array counting must not inherit each
other's PASS. Rejecting `count(1)` proves a compiler boundary, not PHP warning or
exception compatibility. Go helpers `rt.Count`, `rt.Clone`, `rt.Add` are not
additional PHP builtins. Every unlisted builtin remains unsupported until
deliberately added; the seed is not an exhaustive list of PHP functions. Unknown
extension/package requirements remain unresolved, never implicitly supported.

Fixtures aggregate features, but rejection cases now have named subtests and
catalog checks have stable case IDs. Link only assertions actually executed.
Appearing in a fixture does not prove every aspect of a feature.

## Run ledger: what was checked and passed?

CI publishes a distinct `gohp.compatibility.run/v0.4` JSON artifact for each
active target job: PHP 8.3 and PHP 8.4. The same bounded corpus is executed
independently for each target. A later series receives its own policy/matrix
entry and artifact only after it is explicitly enabled.
Keep it outside the source commit: embedding the commit's own SHA in a tracked
report would require a different commit. Record:

- Full GoHP commit, source-tree/dirty state, catalog hash and schema.
- Unique run/attempt/job ID, immutable evidence URL, timestamp and command.
- Exact Go version, OS/architecture, flags, runtime revision, PHP target and
  actual oracle version, plus reference image content digest.
- A bounded, pre-suite PHP CLI startup baseline and canonical JSON SHA-256:
  SAPI, integer size, default timezone, an explicit allowlist of INI values,
  available selected locale categories, and sorted regular/Zend extension names
  with nullable versions. It is captured by a separate `php -r` process with
  the suite's inherited environment; a probe error or malformed value rejects
  the run before tests. The hash catches accidental metadata mismatch, not
  producer tampering.
- Case ID, linked feature IDs, fixture hashes and generated-source hashes where
  applicable. Changed fixtures must not silently reuse older results.
- Separate `parse`, `codegen`, `go-build`, `go-execute`, `differential` phases,
  each `not-run`, `pass`, `fail`, `skip` or `error`. Unit/compiler/negative checks
  retain their own kind; they must not be relabeled differential.
- Measured comparisons: each observed PHP/Go process has hash-only
  `{stdout_bytes, stdout_sha256, exit_code}` data (no raw stdout), alongside
  stdout/exit verdicts and a bounded
  per-positive-case empty-stderr observation. The `quiet-success/v1` profile
  starts the native fixture process with
  `-d error_reporting=-1 -d display_errors=stderr -d display_startup_errors=1
  -d log_errors=0 -d html_errors=0`; it records only each process's stderr byte
  count and SHA-256, never raw diagnostic text. `phase.quiet_stderr` is pass
  only when both streams are empty. General stderr/warnings, exception identity
  and filesystem/DB side effects remain `not-checked`.
  Expected rejection/failure has an explicit expectation and diagnostic.
- Logs or log hashes, both oracle/generated outcomes and the aggregate job
  result. Infrastructure errors are neither semantic failures nor success.

The current differential harness measures stdout equality and successful exit,
and for positive fixture processes the `quiet-success/v1` empty-stderr check.
The profile is deliberately narrower than diagnostic equivalence: matching
non-empty stderr does not pass, and it does not compare warning, error or
exception text. The explicit PHP arguments apply only to the fixture process,
including the booleans wrapper which still establishes `LC_CTYPE=C`; they do not
alter the startup baseline/version probe, Go build or generated Go execution.
A fixture may override or suppress warnings itself, so the record is not an
effective per-fixture configuration capture. The missing-key check measures
explicit Go failure and source-map location, not equivalent PHP errors.
Source-map checks are compiler tooling tests, not PHP language compatibility
evidence.

The PHP baseline is intentionally not a full environment, `phpinfo`, INI dump
or per-fixture effective-state capture. A fixture may change its own process
state (for example the booleans fixture sets `LC_CTYPE=C`), and its source hash
rather than the startup baseline binds that override. The ledger does not use
configuration metadata to classify native-version changes automatically or to
claim diagnostic equivalence. The selected values use PHP's documented
[`phpversion`](https://www.php.net/manual/en/function.phpversion.php),
[`ini_get`](https://www.php.net/manual/en/function.ini-get.php),
[`get_loaded_extensions`](https://www.php.net/manual/en/function.get-loaded-extensions.php),
[`setlocale`](https://www.php.net/manual/en/function.setlocale.php) and
[`date_default_timezone_get`](https://www.php.net/manual/en/function.date-default-timezone-get.php)
interfaces. Readers strictly retain v0.1/v0.2/v0.3 evidence as its historical
scope: they invent neither a missing baseline, quiet-success observation nor
process-output observation. Their generated view is v0.4.

The separate read-only two-archive output comparison is a derived diagnostic,
not a support matrix or a target-admission rule. It requires the same clean
candidate/corpus identity (commit, tree, source/catalog hashes,
toolchain/platform/image, flags, command and case/fixture identity), and reports PHP
version/image/baseline differences. It labels only matching stdout in both runs,
an observed native stdout difference, an observed local transpiler stdout
mismatch, or `not-comparable`. The native label does not establish why native
output differs and never proves an intended PHP version or specification change.
Unrelated failed/incomplete, legacy/output-missing, dirty or changed-input records are not
comparable for native cross-run classification. A measured target-local mismatch
can still be reported in its failed run without granting it success.
It does not compare nonzero exits, exceptions, diagnostics or side
effects.

### Derived per-target view

Show one row per entry and target, with separate columns for implementation,
latest applicable result, verified contract, limitations and evidence.

| Display status | Meaning |
|---|---|
| `unsupported` | No supported lowering/runtime contract |
| `unverified` | Implementation exists but applicable complete evidence is absent |
| `compile-pass` | Required parse/codegen/build checks passed, without differential proof |
| `differential-pass (case-limited)` | Required native comparisons passed for stated cases and measured dimensions |
| `failed` | A required case failed at this revision/target |
| `incomplete` | A required case skipped, errored or did not run |

Keep historical PHP 7.4.33 and PHP 8.3.30 evidence separate from active PHP
8.3.36-or-newer and PHP 8.4.27-or-newer evidence. A failed/incomplete current
run must not be hidden by an older PASS. Historical PASS stays visible with its
revision but does not automatically apply now. Reuse only exact matching
target/revision/configuration results, or an explicitly reviewed unaffected-scope
decision with provenance. Missing evidence never becomes green. Do not average
mixed targets into a universal compatibility percentage. Tooling entries show
`compiler-check-pass`, never differential PASS from Go-only unit tests.

There is no blanket `supported`/Done state in v0.4. Future support claims need
an owner-reviewed bounded contract, all required cases for every claimed target,
explicit negative boundaries and all claimed observable dimensions. Arbitrary
WordPress/Composer application compatibility cannot be derived by counting green
builtin rows.

## Historical baseline evidence

Historical record — this predates the active PHP 8.3.36-or-newer policy and is
not a current support claim: commit `a09893dee9503aafee3117bde02e9cb36e3e80eb` has two passing pinned matrix
jobs in [PR #4 run 40102827264](https://github.com/katsuakikureyama/GoHP/actions/runs/40102827264).
They ran `scripts/check.sh` with Go 1.23.12 and PHP 7.4.33 / 8.3.30. Positive
fixtures cover functions/classes/ordered arrays/`count`, nested arrays and scalar
conversions, maximum integer keys, and version-sensitive precedence. The
[inventory](compatibility.md) and catalog link these cases to bounded features,
without claiming complete coverage. Source-map checks postdate that commit and
must not inherit its PASS. Current-change CI evidence is recorded in its PR and
workspace handoff, not retroactively added to this historical baseline.

## Incremental implementation plan

1. Review the seed and freeze its schema; add stable case IDs and required checks
   per target. Validate unique IDs, real fixture/test references, allowed values
   and explicit unsupported/untested boundaries in CI.
2. Collect case-level events from the existing harness without rerunning passing
   suites merely to fill the ledger. Write immutable per-job manifests. Missing
   PHP oracle remains a failure, never a skip or PASS.
3. Generate Markdown/JSON function-and-feature views from catalog + run artifacts,
   including historical/current evidence and filters for target/status/kind.
   Derived views and two-archive stdout diagnostics must not become an
   independently edited authority or a combined support verdict.
4. Add read-only project inventory from parser callsites and Composer metadata
   as data. Dynamic calls, unresolved names and absent extension contracts are
   unknown. Never execute Composer hooks or applications to inventory them.
5. Define a service query/report API only afterwards. The local run and view
   operations have explicit WebMCP exclusions in the ledger operation contract;
   `gohp.cli.generate` retains its local-tool exclusion. Any future
   browser/service/MCP surface needs its own capability/auth contract.

Add/update the contract, target-specific positive cases, negative boundaries and
catalog links in the same function/feature change. Test only the affected scope
locally and use the existing pinned CI matrix once per revision/configuration.
Green tests are evidence to record, not a reason for another verification cycle.

For a new PHP series, update the version-policy registry's series label and
minimum oracle version, the matrix's immutable image digest, and the catalog's
status and required corpus together. Do not infer a support verdict from an
existing target run or from a shared fixture. The current active policy is 8.3
with a minimum native oracle of 8.3.36 and 8.4 with a minimum native oracle of
8.4.27. They retain separate target rows and evidence; a later series remains
disabled until its own requirements are reviewed.
