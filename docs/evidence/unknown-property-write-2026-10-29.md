# Guarded unknown-local property writes — 2026-10-29

Defined unknown locals now admit named direct property assignment/array append
under the [declared-property contract](../public-property-write.md). Known bad
types and known absent class slots still reject before generation. One RHS runs
before the explicit nonnil Object/Get/Set and declared-slot check; array/object
copy/identity and current-slot append behavior remain. Fully opted-in structural
PropertyObject hosts are supported like nested writes, not an unforgeable
generated-class provenance claim. No dynamic properties, direct concat,
references, coercing append or general PHP exceptions are admitted.

## Local checks

Cached Go1.23.0 Darwinarm64, CGO0, offline modules, existing shared
`GOCACHE=/private/tmp/gohp-go-cache`, cached Go also on child-process PATH.
At clean implementation43789e8322156fc59b7d00a28824df18e39e06b0,
tree d19e156536169ef954f1b0244b23d0df0c1e53b1:

```sh
go test -count=1 ./internal/codegen ./runtime ./internal/compat
go test -count=1 ./tests -run '^TestUnknownPropertyWriteHostEntry$'
```

Runtime0.440s, compat0.689s, generated external ABI1.020s PASS. The latter also
checks the committed positive fixture lowers without requiring local PHP.
Six legacy compiler tests failed: four constructor/class-guard/mapping checks
expected the previous direct setter spelling; list/switch source negatives
expected unknown receivers to reject statically. The new contract instead
requires generated runtime rejection of actual bad values; production was not
weakened to satisfy test expectations.

At clean corrected tests0f448b6609f15b90d6853be6dc2db4565edb7940,
tree8e9403adea224a88e18e42c60b709673bfdd50e9:

```sh
go test -count=1 ./internal/codegen \
  -run '^(TestBlockDataConstructorLoweringAndSourceMap|TestClassParameterLowering|TestConstructorNullDefaults|TestListAssignmentUnsupportedSource|TestScalarConstructorLowering|TestStringSwitchUnsupportedSource)$'
go test -count=1 ./tests ./internal/compat \
  -run '^(TestUnknownPropertyWriteControlFlowRuntimeGuards|TestCatalogTargetScopes)$'
```

PASS compiler0.246s, new actual list/loop/switch guard execution1.659s and updated
catalog target scopes0.346s. Previously passing checks are reused. No unchanged
runtime/ABI package or successful old suite was manually rerun.

The new owned fixture SHA256 is
`d66cd5f8ae71b8c84fffa2bc6cefc9d77eda77842a6b147253cc03237033aa76`.
It observes array-extracted objects, untyped helpers, aliases/distinct instances,
falsy slots, detached nested arrays, branch/loop-widened locals, existing $this
updates and same-slot RHS-reset append. Eight new catalog checks cover source,
IR, runtime locations, generated external ABI and a native differential case.
Native results are per candidate/target automatic CI, not local PHP evidence.

## Complete original closure: next semantic blocker

Same unchanged WordPress/WordPress54106f6c545fc8b881f6b39b7963002bece6c28c
Parser/Block/Frame, license/version selection and existing separate
[Gossip closure manifest](https://github.com/katsuakikureyama/Gossip/blob/master/gossip/tests/gohp/parser-source-closure.json),
SHA25689f897c64146705559eae41bb50d49aff57abad0a11ee02d16d27fe78d64adaa.
One probe at clean code0f448b6, with the same offline/CGO0 configuration:

```sh
go run ./cmd/gohp -manifest GOSSIP_ROOT/gossip/tests/gohp/parser-source-closure.json \
  -source-root GOSSIP_ROOT -entry wp-includes/class-wp-block-parser.php \
  -php-version 8.3 -experimental-pcre2 -out FRESH/parser.go -report FRESH/report.json
```

Exit1, stdout0bytes, stderr94bytes:

```text
wp-includes/class-wp-block-parser.php: PHP004 line 354: empty requires an array
exit status 1
```

Stderr SHA256 `f55bef7a8cb235bbac2fa946289eaaacfd440bb3b520db67ec66ddee3e488fd8`;
the last line comes from go run. No output/report/staging artifacts; empty probe
directory removed. The prior unknown receiver write at209 now passes admission.
The original add_inner_block method uses `!empty($html)` at354 on a substring;
the existing empty subset accepts only arrays. Reaching354 does not establish
all earlier method semantics or native behavior. Full original generation/build/
execution, full12-case/24-call comparison and G83-PARSER remain blocked. Stack,
snapshot, tokenizer/JSON and remaining value-domain dependencies stay open.

No original source/edge/manifest/pin changes, extraction, concatenation or
interpreter fallback. Existing130 loading/native evidence is reused unchanged.
Local CLI capability/excluded scope is updated in the same change set; no
browser/server/MCP route, auth/deployment or new dependency. Devboard unidentified
and not synchronized.
