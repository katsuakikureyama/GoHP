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

At clean code `5156b663854dfa3cba8cbeb370b3cd6bc8389ba4`, tree
`eb71d4b20c305a163b7b0a6b8803f18e580e996b`, affected CGO0 checks pass:
codegen0.200s, sourceclosure0.754s, catalog0.559s, CLI1.226s, runtime0.928s.
Frontend compiled0.223s with no matching direct tests; semantic has no direct
test files and is exercised through codegen. Command:

```sh
CGO_ENABLED=0 go test -count=1 ./internal/frontend ./internal/semantic ./internal/codegen ./internal/sourceclosure ./internal/compat ./cmd/gohp ./runtime -run 'Preg83|Match1|InspectGraphTargetProfile|Catalog'
```

The initial enabled-runtime test fails0.196s because it expected a literal match
to return1 with budget1; the actual measured return is false/error2. This is an
owned test assumption, not a production-runtime change. The reset control now
uses a completed no-match on an empty subject (expected0/error0), while the
small default-budget control remains no-match0. Only this enabled test is
rerun; previously passed CGO0 checks are reused. Eight new default tooling catalog
checks cover profile dispatch/source maps, source/IR/target rejection, CLI
publication, closure metadata and fail-closed runtime inputs/unavailability.
The enabled runtime test uses an owned tiny regex, not copied GPL WordPress
source, and separately observes fixed-budget/error-state behavior.

Corrected enabled test at clean `6f8f954007342a3aef8c067c1acd815c840e6a6a`,
tree `5af142e77efdf752d6c54f6c3ec7c0e26372c097`, PASS0.202s. Command:

```sh
CGO_ENABLED=1 go test -count=1 -tags gohp_pcre2 ./runtime -run '^TestPreg83LocalMatch1UsesProfileLimitAndResetsState$'
```

Environment: cached Go1.23.0Darwinarm64, GOTOOLCHAIN=local, GOPROXY/GOSUMDB=off,
GOWORK=off, existing shared GOCACHE, verified clang/SDK15.5,
CGO_CFLAGS=`-I<PCRE2_PREFIX>/include -O2 -Wall -Wextra -Werror`,
CGO_LDFLAGS=`<PCRE2_PREFIX>/lib/libpcre2-8.a`, existing PCRE10.47_1.
Match1 false/error2, completed no-match0/error0 reset, and unchanged default
no-match0/error0 are measured separately. Duplicate archive linker warning is
not runtime stderr or native-PHP evidence. The corrected file is build-tagged
out of CGO0, whose earlier PASS is reused. No production code changed after
the first local pass. The merge of reviewed139/mainc8d21776 into this branch
preserves the exact same5af142e tree; no test rerun for that ancestry change.
Required remote CI and generated-original comparison remain separate pending
results, not inferred from these local checks.

Use cached Go1.23.0/offline/shared GOCACHE for affected packages. The enabled
test uses existing verified clang/SDK15.5/staticPCRE10.47_1 and `gohp_pcre2`;
no daemon/image/install/pull/new CI resource or persistent binary. Existing
default tooling/CI versus enabled native measurements remain distinct.

The inherited attribute-rejection repair6804605 is owned by BOOT scope6522006068;
this branch preserves it and does not alter hash-comment/attribute lexing.
CLI capability `gohp.cli.generate` retains WebMCP exclusion; its input schema
now includes the optional closed profile enum, with fresh report-last output
and generated/unverified result. Auth/routes/UI/deployment unchanged.
