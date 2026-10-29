# Guarded inferred-unknown object snapshots — 2026-10-29

The existing `object_array_cast` now admits inferred-unknown expressions to
the unchanged `runtime.ObjectToArray` guard. Known scalar/null/array/Closure
operands still reject statically; actual unsupported domains reject at runtime.
No new IR kind, reflection, runtime API, storage lane, builtin or inference.
The [snapshot contract](../block-data-constructor.md) requires declaration order,
null retention, detached nested arrays and shared object identifiers. Explicit
full host-protocol opt-in is not generated provenance.

## Focused local checks and one correction

Base main `efcdbc47ffeb506d2bec1f9e6d84f0c67c37332c`.
Clean initial code `04eca52b2160f7c6b21bdd04375e503a8c4cd4a6`, tree
`5ae08b696a1d9d019d03692ebe9794c40af9a836`. Cached Go1.23.0 Darwinarm64,
CGO0, GOTOOLCHAIN=local/GOPROXY=off/GOSUMDB=off/GOWORK=off, cached Go on child
PATH and existing shared GOCACHE (no dedicated cache or persistent binary):

```sh
go test -count=1 -run '^Test(UnknownObjectSnapshot|BlockDataConstructor|BlockConstructorAndArrayCast|ObjectToArray)' ./internal/codegen ./runtime ./tests
go test -count=1 ./internal/compat
```

Runtime0.192s, generated host/fixture/old guards2.571s and catalog0.324s PASS
once. Codegen0.383s FAIL only in the new lowering/map assertion: expected Go
interface text omitted gofmt's surrounding spaces. Production behavior was
correct; no domain guard was relaxed. Other selected compiler checks passed.
The test-only correction was committed at
`8cc58c4af410c1e845d3840ac701c17d4eebf46e`, tree
`1e555a845956eb1a18af0f468f3231c7edb0afa2`, then only the failed test rerun:

```sh
go test -count=1 -run '^TestUnknownObjectSnapshotLoweringAndSourceMap$' ./internal/codegen
```

PASS0.208s; earlier successful ranges reused, not manually rerun.
The initial six catalog checks cover owned native differential comparison, unknown
parameter/property/return/arrow lowering with maps, source negatives, strict
raw IR shape, committed fixture lowering and external generated-entry ABI.
The host check observes one effectful receiver/getter/snapshot, generated
objects and explicit full-protocol hosts, rejecting nil/typed nil/partial or
invalid protocols/scalars/arrays/callables before requesting a snapshot.

Owned fixture SHA256
`1c7e42c3855e4818b92fcce1d1ad8829e6a6f49952352a81bf5291b3cf699584`
uses an **indexed** frame, untyped global return and object-property snapshot.
It does not use or certify unimplemented `array_pop`. The first fixture draft's
array_pop was removed before any test execution; this is an owned isolated
feature fixture, not a rewrite or substitute of the original parser. General
exceptions, side effects and stderr equivalence remain unmeasured. Independent
native PHP8.3/8.4 evidence belongs to the exact candidate CI, not local lowering.

### Initial automatic CI failure and limited correction

Head `4d3514a1bbee137bee5831fee2bda091d7780901`, tree
`e363f0795ea475f72f23c1ffb87c8c4718dd8482`, push40887469673 and
PR40887478076 attempt1 failed both targets. Saved push artifacts12902958815
(8.3)/12904195113(8.4) read once. Each raw go-test stream has only three leaf
failures, all old `TestParserOutputUnsupportedSource` expectations:
cast_nonsimple_factory, cast_foreign_receiver and cast_passthrough. Those
inferred-unknown method results now correctly reach the same snapshot guard.
The aggregate compatibility run is failed/incomplete, not a success even where
individual observations passed. No infrastructure rerun or failure relabeling.

Move these three cases to affirmative `TestUnknownObjectSnapshotMethodResults`
and register a seventh catalog check, leaving all unrelated negatives intact.
Production/fixture/runtime/CLI settings and original-probe inputs are unchanged.
Only the corrected old source-negative group, new positive routing test and
changed catalog need local revalidation; prior PASS/probe are reused. The exact
correction commit/result and new automatic CI are recorded with PR134.

Clean correction `7581f6e6560d84b1e2d643bf61ac84bf95a579e9`, tree
`fbf20d4f6f7ab5da2a28f21926b56885fca938e1`, same cached/offline settings:

```sh
go test -count=1 -run '^(TestParserOutputUnsupportedSource|TestUnknownObjectSnapshotMethodResults)$' ./internal/codegen
go test -count=1 ./internal/compat
```

PASS0.236/0.278s once. Old source-negative group and new positive routing group
are consistent with the extended contract. No production/fixture change,
successful broader suite or original probe rerun.

## One unchanged complete-original closure probe

At clean `8cc58c4` with the same cached/offline configuration, use unchanged
WordPress/WordPress `54106f6c545fc8b881f6b39b7963002bece6c28c` complete
Parser/Block/Frame and original notice/provenance. Existing closure manifest
SHA256 `89f897c64146705559eae41bb50d49aff57abad0a11ee02d16d27fe78d64adaa`;
Parser SHA256 `8563621090de5e307431b1d29b78a9f75fd9ee238a0137b563a7db7b2e1f8d3b`.
One invocation (personal paths replaced by stable placeholders):

```sh
go run ./cmd/gohp -manifest GOSSIP_ROOT/gossip/tests/gohp/parser-source-closure.json \
  -source-root GOSSIP_ROOT -entry wp-includes/class-wp-block-parser.php \
  -php-version 8.3 -experimental-pcre2 -out FRESH/parser.go -report FRESH/report.json
```

Exit1, stdout0bytes, stderr102bytes, SHA256
`9db57cfa3d6c3e0411ef3e4725b12b1d18e38b5d6258c605bc15abb53d689c3e`:

```text
wp-includes/class-wp-block-parser.php: PHP004 line 165: unsupported function array_push
exit status 1
```

The last line is the go-run wrapper. No generated Go/report/staging artifact;
raw logs retained outside source worktrees. The prior semantic cast gate394
passes and the complete closure reaches **code generation**, which rejects the
earlier source-line `array_push`165. This is a later pipeline stage, not a
regression or a source-line progress percentage. The earlier semantic pass did
not resolve generic function calls. `array_pop`, remaining stack/tokenizer/JSON
contracts, complete Go generation/build/execution,12-case/24-call comparison and
G83-PARSER remain open. No original rewrite, extraction, concatenation or fallback.

Source/manifest/pins/auth/routes/deployment remain unchanged. Existing old
successes and enabled-PCRE evidence are not manually remeasured. Local
`gohp.cli.generate` stays WebMCP-excluded with unchanged input/publication/result
contracts. Devboard unidentified/not synchronized; no whole-application Done.
