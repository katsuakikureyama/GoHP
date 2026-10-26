# Running the compatibility ledger

The catalog now drives an executable, **run-scoped** evidence ledger. It produces
one PHP-target report at a time; it does not claim arbitrary PHP compatibility or
merge historical results into a current support verdict.

From a trusted GoHP checkout with the required Go and native PHP installed:

```sh
go run ./cmd/gohp-compat -out build/compatibility-run-001
```

Use a fresh output directory on every invocation. This command validates the
catalog and runs `go test -count=1 -json ./cmd/... ./internal/... ./runtime ./tests`
**once**, covering all current test packages but excluding generated build files.
New test package roots must be added to this command. It is also used by
`scripts/check.sh`, replacing that script's previous test invocation, not adding
another test pass. The script still runs vet and its CLI/native-output smoke
checks; the ledger alone does not certify those later steps or the whole CI job.
Require a successful CI job before using its artifacts as release evidence.

## Outputs and failure behaviour

- `catalog.json`: the exact validated catalog used for this run, hash-bound to the report.
- `go-test.jsonl`: exact Go test event stream, including failures and phase data.
- `go-test.stderr`: test process stderr (infrastructure/build failures included).
- `run.json`: `gohp.compatibility.run/v0.2` metadata, case evidence and feature rows.
- `summary.md`: human-readable function/feature table for the selected target,
  with implementation state, bounded result, case IDs and limitations.

The command exits nonzero on a failed test process or incomplete/rejected
evidence. It retains failed-run logs and, when evaluation is possible, a failed
or incomplete report. Invalid catalog or metadata can fail before creating the
directory. Truncated/malformed evidence may leave raw logs without a report;
missing `run.json` is never success. Files are created exclusively with mode
0600 in a mode-0700 directory. Existing files/directories and symlink aliases are
not overwritten. An interrupted run is not resumed or cleaned up automatically.
Use a private workspace without concurrent mutation. This is not a multi-file
crash transaction or protection against a hostile process changing directories.

The tool executes repository tests, including child processes. Only run it on
trusted code. CI retains its no-network, read-only container root and reduced
capabilities; the narrowly mounted evidence directory is writable. This does not
make the host CLI a general untrusted-code sandbox.

## Evidence and provenance

The catalog's `case_id` is independent of display text. Its `file` and exact Go
`test` name identify the event that must run and terminate successfully. The
loader rejects invalid schema/enums, duplicate feature IDs, conflicting case IDs,
missing files/functions and paths escaping the repository. Nested subtest names
are checked against actual execution, not guessed from syntax. The 18 rejection
cases now have individual names instead of sharing one aggregate PASS.

Differential cases emit phase evidence from the **existing execution**, not a
second run: parse, Go generation, Go build, Go execution and comparison. A
differential PASS requires the matching case/target/oracle version, current
fixture hash, generated-source hash, successful phases, equal stdout and both
successful exits. Stderr, warnings/exception identity and side effects remain
`not-checked`; negative and tooling tests never become PHP-equivalence proof.
Parent/package/process failure, absent events and skipped cases cannot produce
a green feature verdict. Unsupported entries remain unsupported even when their
expected-rejection checks pass.

Reports record the catalog/event hashes, actual toolchain and PHP version,
platform, build flags, source fingerprint and a bounded PHP CLI startup baseline.
Before the test suite, the runner starts one separate `php -r` probe with the
same inherited environment as the suite. It records `scope: cli-startup-baseline`,
`php_version`, SAPI, integer size, `date_default_timezone_get()`, the selected
INI values, the selected locale categories, and lexically sorted regular and
Zend extension `{name, version}` lists. Configuration and locale values are
nullable where PHP cannot report them. The INI allowlist is exactly `precision`,
`serialize_precision`, `default_charset`, `date.timezone`, `error_reporting`,
`display_errors`, `display_startup_errors`, `log_errors`, `zend.assertions` and
`assert.exception`; locale categories are exactly `LC_CTYPE`, `LC_NUMERIC`,
`LC_TIME`, `LC_COLLATE`, `LC_MONETARY` and `LC_MESSAGES` when available.
The collected baseline is canonical-JSON SHA-256 bound by
`metadata.php_baseline_sha256`; the hash detects accidental metadata mismatch,
not authenticity. A failed or malformed probe rejects the run before tests.
It deliberately does not save a full `phpinfo`, environment, INI dump or
filesystem paths.

This is a CLI startup baseline, not the effective configuration of every
fixture process and not a reproducibility or compatibility guarantee. In
particular, the booleans fixture sets `LC_CTYPE=C` inside its own PHP process;
that harness code is source-fingerprint-bound and is distinct from the pre-suite
baseline. The runner does not infer a native-version difference versus a
compiler mismatch from this metadata, and it does not expand diagnostic
comparison. The probe uses PHP's documented [`phpversion`](https://www.php.net/manual/en/function.phpversion.php),
[`ini_get`](https://www.php.net/manual/en/function.ini-get.php),
[`get_loaded_extensions`](https://www.php.net/manual/en/function.get-loaded-extensions.php),
[`setlocale`](https://www.php.net/manual/en/function.setlocale.php) and
[`date_default_timezone_get`](https://www.php.net/manual/en/function.date-default-timezone-get.php)
interfaces.

The source fingerprint is SHA-256
of a lexically ordered JSON list of `{path, sha256}` for Go/PHP files, `go.mod`,
`go.sum` when present, `compatibility/catalog.json` and the inventory's
`internal/inventory/testdata/**/composer.json` / `composer.lock` fixtures, excluding `.git`, `build`
and `.cache` trees. It does not hash documentation or container configuration;
image identities and flags are recorded separately. Source must not change during
the run. A hash binds bytes; it is not a signature or a guarantee of trusted code.

Local Git metadata is detected automatically. CI passes the actual checkout
commit/tree/dirty state (including a PR merge checkout, not falsely labeled as
the branch head), run/attempt identity, URL and the image digests used to build.
Those fields are CI/caller assertions, not signed attestations. Image checkouts
without `.git` or injected metadata are marked `unbound-checkout`; dirty or
unbound runs are useful local evidence but must not be represented as proof for
a clean release commit. No automatic historical applicability or cache reuse is
implemented.

CI uses a separate artifact per PHP job/run/attempt, publishes the Markdown in
the job summary and retains artifacts for **30 days**, including failed-run
logs. Download and preserve evidence separately before relying on it for a
long-lived release; expired artifacts are not permanent audit storage. Upload
uses a commit-pinned [official artifact action](https://github.com/actions/upload-artifact/blob/cf430e030ddbb5b0abf93d22962f4752f3646cd9/action.yml)
without overwrite. Missing artifacts fail the upload step, not silently pass.

## Local capability boundary

Stable capability ID: `gohp.compat.run`; service/resource: local GoHP development
tool / caller-owned test-evidence directory. Input schema:

```json
{"type":"object","properties":{"out":{"type":"string","minLength":1}},"required":["out"],"additionalProperties":false}
```

`inspect`: validate catalog and metadata. `prepare`: allocate a fresh directory.
`save`: execute tests once and save logs, JSON and Markdown. `commit`: none; no
production promotion, upload, publish, payment or external message occurs in the
CLI. Results are `pass`, `failed` or `incomplete` within the run, or a nonzero
preflight/evaluation error; artifact presence alone is not a success signal.
Permission/validation authority is local filesystem access plus the catalog
validator; the run report and logs are its audit trail. Retries use new paths,
not overwritten evidence. The CLI default is `build/compatibility` when omitted.

WebMCP policy: **excluded**, because this is an internal development command that
executes trusted repository tests and writes local artifacts. No browser/server
route, MCP protocol, session forwarding, auto-submit or new authentication path
is introduced. CI artifact upload is an explicit workflow step using existing
GitHub Actions authorization, not a hidden action of the CLI.
The v0.2 PHP baseline and its hash are runner-produced audit metadata, not
caller-supplied capability inputs or a configuration API; this change does not
add an inspect/search/state/link operation or alter the exclusion.

## Browse saved runs without rerunning tests

Download and extract the desired CI artifacts into separate caller-owned folders,
then use this read-only consumer (all flags precede the directory arguments):

```sh
go run ./cmd/gohp-compat-view -kind builtin build/php74-run build/php83-run
go run ./cmd/gohp-compat-view -target 8.3 -status unsupported -format json build/php83-run
go run ./cmd/gohp-compat-view -target 8.4 -format json build/php84-run
```

The CLI does not fetch artifacts, invoke tests/PHP/Git, change evidence or write
output files; it prints Markdown (default) or `gohp.compatibility.view/v0.2` JSON
to stdout. `go run` itself compiles the trusted Go tool and may populate the Go
build cache; a prebuilt binary needs neither Go nor the source checkout.

Each directory must contain `run.json`, the exact `catalog.json` snapshot and
`go-test.jsonl`. The loader checks their schema and byte hashes, reconstructs
verdicts with the same event evaluator using **recorded** fixture hashes, then
requires the entire saved report to match. It does not open the catalog's
source paths in the consumer checkout. Missing/altered/truncated evidence,
invented verdicts and symlinks escaping the archive are errors, not green rows.
Archives with source-read failures remain incomplete. Fixed size limits are
16 MiB for the run, 4 MiB for the catalog and 128 MiB for events. Use private
directories without concurrent mutation; this is not an untrusted archive extractor.
`summary.md` and `go-test.stderr` are not consumed or validated by the view.

This is **internal consistency**, not authenticity or fresh verification:
source/fixture/generated bytes are not included in archives and cannot be
rehash-checked here. v0.2 archives additionally bind a bounded pre-suite PHP
CLI baseline, but that still is not an effective per-fixture configuration or a
complete environment capture. A malicious producer can fabricate a consistent bundle.
Metadata, environment identities and exit codes are producer assertions, not
signed attestations. CI job status and subsequent smoke checks are not proven
by the archived test report. Consult the trusted CI job before release use.

Runs remain separate in caller-provided order, including duplicate inputs,
different revisions/configurations/catalogs and failed/incomplete runs. No run
is labeled latest, overwritten, ranked or selected as the best result. Filters
are exact AND matches on feature-row target, kind and status. Target filtering
does not select runs: an 8.3 run's 8.4 row remains `not-measured` when filtering
for 8.4. All supplied run headers, original overall results and all case details
remain visible, including runs with zero matching rows. There is no combined
compatibility percentage, promotion or current-revision applicability decision.
The live runner admits PHP 8.3 only when its actual oracle is at least 8.3.36,
and PHP 8.4 only when its actual oracle is at least 8.4.27. CI supplies a
separate pinned image identity for each target; local host-PHP runs need no
container and do not imply a verified image digest. The same bounded corpus is
run separately for each target. A report has no 8.4 differential verdict until
the 8.4 execution completes; it cannot inherit the 8.3 result. The archive
reader may display legacy 7.4 and older 8.3 runs, but they are historical
evidence rather than active compatibility gates.

Accepted kinds: `syntax`, `operator`, `type`, `builtin`, `extension`, `tooling`,
`application`. Accepted statuses: `unsupported`, `unverified`,
`compiler-check-pass`, `differential-pass (case-limited)`, `failed`, `incomplete`,
`not-measured`. Empty filters mean all. A successful read (exit 0) can contain
failed runs or no matching features; it is not a compatibility gate. Usage
errors exit 2; rejected archives or output failures exit 1. All inputs are
validated before stdout starts; a broken output stream may still be partial.

### Local view capability

Stable ID: `gohp.compat.view`; owner/resource: local GoHP development tool /
caller-readable archived test evidence. Input schema:

```json
{"type":"object","properties":{"archives":{"type":"array","minItems":1,"items":{"type":"string","minLength":1}},"target":{"enum":["","7.4","8.3","8.4"]},"kind":{"enum":["","syntax","operator","type","builtin","extension","tooling","application"]},"status":{"enum":["","unsupported","unverified","compiler-check-pass","differential-pass (case-limited)","failed","incomplete","not-measured"]},"format":{"enum":["markdown","json"]}},"required":["archives"],"additionalProperties":false}
```

The `7.4` filter remains solely for inspecting archived legacy evidence; it
does not add 7.4 to the active compiler or validation policy. The `8.4` filter
selects stored 8.4 rows only; it never turns an 8.3 result into 8.4 evidence.

`inspect`: read/validate archives and filters. `prepare`: derive a view in memory.
`save` / `commit`: none. Result: a versioned view with filter, notice and per-run
metadata, hashes, original overall, total row count, selected features and all
case evidence; or a usage/archive/output error. No authorization beyond local
filesystem read permission; no server operation is introduced. Output includes
source run/attempt/URL and hashes for audit, but is not a new evidence authority.
WebMCP: **excluded**, local development evidence reader with no browser surface,
MCP route, credentials, auto-submit, publication or remote authentication.
Its catalog checks are Go tooling tests, never proof of PHP semantics.
The v0.2 view always emits its own schema while accepting historical v0.1 and
current v0.2 runs; it does not synthesize a baseline for v0.1 evidence. Reading
the additional metadata adds no callable operation or authority and remains
within this capability's exclusion.

## Remaining work

Historical applicability decisions, combined-target support verdicts, remote
query APIs, PHP callsite inventory and stronger semantic comparisons remain
unimplemented. A bounded [Composer metadata reader](compatibility.md#composer-metadata-inventory)
and read-only per-run views/filtering are implemented; dependency resolution and
project compatibility decisions are not. The catalog
is deliberately incomplete and the ledger's per-case verdict is not a blanket
`supported` state. Add contracts, cases and negative boundaries together when
extending a function or feature; do not mark untested modes green.
