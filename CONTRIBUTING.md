# Development

Work on a feature branch and open a PR linked to #1; use a draft while incomplete.
Keep one canonical change-set per branch and preserve concurrent changes.
Owner authorization (2026-10-26) permits ordinary reviewed/tested main integration
and normal release/deployment work without repeated approval at this early stage.
It does not authorize new deployment destinations, costs, destructive operations,
upstream forks or final licensing decisions. No production destination is configured.
English documents are authoritative; keep the Japanese README summary accurate.

Use Go 1.23.12 and a native oracle in the selected series: PHP 8.3.36+ for
`8.3` (default), or PHP 8.4.27+ for `8.4`. A compiler label is not a patch-version
selector. See [version policy](docs/compatibility.md#active-php-version-policy)
and [README](README.md#reproduce-the-proof-of-concept) for local/Compose commands.
Each active target has its own pinned CI image, corpus selection and run evidence;
do not substitute floating tags or transfer an 8.3 PASS to 8.4.
Missing/below-floor PHP fails rather than skips. Historical 7.4 evidence is not
an active gate. A later series needs an explicit policy/profile, minimum version,
reviewed immutable image, target-specific cases and independent passing evidence.

Format touched Go files. Run the smallest affected test/vet/build checks once
per code revision and effective configuration, recording commands, revision,
PASS/FAIL and unverified areas. Reuse unaffected evidence explicitly; do not
repeat or broaden passing checks without a relevant change or unresolved failure.
`sh scripts/check.sh` is the full differential suite; CI runs it independently
for both active targets. Documentation-only changes need content/link/diff
checks, not extra manual compiler reruns. Do not relabel older evidence as a new run.
Fixtures must be repository-owned and offline-safe. Do not add network or
production credentials to tests. Go packages currently have no external dependencies.

## Small follow-up roadmap

The [roadmap](docs/roadmap.md) is the canonical priority and completion checklist:
fixed application inventory and denominators, reference/value integration,
performance baseline, then broader application admission. Parser reuse has a
separate evaluation gate, not a promise that parsing provides semantic support.
Do not grow a second roadmap in this file or append a feature changelog to README.

Keep the [compatibility tables](docs/compatibility.md), catalog case links and
affected boundary documents aligned. README is an entry point; detailed
restrictions belong in those tables/contracts. Historical evidence retains its
original target and revision. New CLI/service operations must also update their
capability definition or explicit exclusion; documentation reorganization adds none.

A passing parser/build is not a compatibility claim. Each addition needs positive
native-PHP-vs-generated-Go tests and explicit negative boundary tests.
