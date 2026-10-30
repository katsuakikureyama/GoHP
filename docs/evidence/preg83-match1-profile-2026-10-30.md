# Explicit PHP8.3 match1 compiler profile — 2026-10-30

## Contract

Scope6522495611 adds the separate compile-time profile
`pcre2-php83-match1-v1`, explicitly selected alongside `-experimental-pcre2`.
CLI/IR/closure/report identity and a distinct generated runtime entry preserve
the old `pcre2-php83-v1`/match100000 contract. The new entry fixes Match1,
Depth10000, Heap131072KiB and the existing subject/pattern4096-byte limits.
No arbitrary/mutable budget, PHP ini support, backend/error-mapping expansion,
new Execution state/storage/IR/report version or PHP8.4 admission.

The request comes from
[Gossip74's separately recorded native resource reference](https://github.com/katsuakikureyama/Gossip/pull/74):
same12 documents/24 actual original PHP calls,16error2 outcomes at match1.
That existing native record is reused; no PHP native rerun is part of this
change. Actual generated-original comparison belongs to its separate Gossip
fixture/pin/evidence. This compiler/tooling change alone does not establish
WordPress/Gossip or application acceptance.

## Verification status

Affected compiler/closure/CLI/default-runtime/catalog checks and the targeted
enabled-backend runtime test are pending. Eight new default tooling catalog
checks cover profile dispatch/source maps, source/IR/target rejection, CLI
publication, closure metadata and fail-closed runtime inputs/unavailability.
The enabled runtime test uses an owned tiny regex, not copied GPL WordPress
source, and separately observes fixed-budget/error-state behavior.

Use cached Go1.23.0/offline/shared GOCACHE for affected packages. The enabled
test uses existing verified clang/SDK15.5/staticPCRE10.47_1 and `gohp_pcre2`;
no daemon/image/install/pull/new CI resource or persistent binary. Existing
default tooling/CI versus enabled native measurements remain distinct.

The inherited attribute-rejection repair6804605 is owned by BOOT scope6522006068;
this branch preserves it and does not alter hash-comment/attribute lexing.
CLI capability `gohp.cli.generate` retains WebMCP exclusion; its input schema
now includes the optional closed profile enum, with fresh report-last output
and generated/unverified result. Auth/routes/UI/deployment unchanged.
