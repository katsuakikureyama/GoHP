# Static declaration closure — 2026-10-29

The explicit manifest/root/entry CLI now connects checked original-file modules
to located IR, Go generation and fresh v0.2 reports. General PHP include/autoload
and executable dependency initialization remain excluded. See the
[contract](../static-source-loading.md).

## Local checks

Cached Go1.23.0 Darwin/arm64, CGO0, offline modules, shared existing
`GOCACHE=/private/tmp/gohp-go-cache`; commands use that cached Go on PATH.

At clean implementation `a1b29da`, `go test -count=1 ./internal/frontend
./internal/sourceclosure ./cmd/gohp ./internal/compat` passed frontend0.237s,
sourceclosure0.647s and compat0.880s. CLI failed two new expectations: the
existing lowercase generated class name and exact size-mismatch diagnostic.
`go test -count=1 ./tests -run '^TestStaticClosureFixture$'` failed because the
new owned manifest did not use the corpus schema's literal `unresolved`
dependency field. These were test/data defects, not relaxed production checks.

At clean corrected code `532e16c043d488d91269026626037b2c7d210e46`, tree
`f30f55406ca24b891d47768dbff4bf01be4d6827`:

```sh
go test -count=1 ./cmd/gohp ./tests \
  -run '^(TestCLIBundleFreshReport|TestCLIBundleFailures|TestStaticClosureFixture)$'
```

PASS CLI0.872s / fixture compiler0.369s. Previously passing packages and artifact
preservation checks are reused; only the failed/changed scope was rerun. The
owned fixture retains real entry/dependency files, a nested diamond and repeat
edge, shared-object mutation, falsy property presence and array-copy separation.
Its fixture identity is the complete checked manifest SHA. Native comparisons
for8.3/8.4 belong to candidate-specific automatic CI; no local native PASS claim.

## Complete original three-file admission

WordPress/WordPress `54106f6c545fc8b881f6b39b7963002bece6c28c`, version
7.2-beta1-64693; entry `wp-includes/class-wp-block-parser.php`, with complete
Block/Frame dependencies. The new separate
[Gossip manifest](https://github.com/katsuakikureyama/Gossip/blob/master/gossip/tests/gohp/parser-source-closure.json)
selects those original sources plus license.txt/version.php notice/provenance.
All five hashes/sizes match the existing native baseline; no source, edge,
algorithm, old manifest or consumer pin changes.

At clean code532e16c, with fresh output/report paths:

```sh
go run ./cmd/gohp -manifest GOSSIP_ROOT/gossip/tests/gohp/parser-source-closure.json \
  -source-root GOSSIP_ROOT -entry wp-includes/class-wp-block-parser.php \
  -php-version 8.3 -experimental-pcre2 -out FRESH/parser.go -report FRESH/report.json
```

Final manifest SHA256
`89f897c64146705559eae41bb50d49aff57abad0a11ee02d16d27fe78d64adaa`.
Exit1, stdout0bytes, no output/report/staging artifacts:

```text
wp-includes/class-wp-block-parser.php: PHP004 line 209: property assignment requires a known generated class variable
exit status 1
```

Stderr132bytes, SHA256
`f0d26b1d56accceea51bdbd3ec6ce3f2e3fd5036187fb9abe357f530f196238b`;
the last line comes from `go run`. The empty temporary directory was removed.
An initial probe had incorrectly labeled the upstream repository
WordPress/wordpress-develop (manifest SHA1f0d46fb66a519921697c82dd20f6f2e6330046930501bc680d6aacc8e9320d5).
After correcting provenance to WordPress/WordPress, one final-manifest probe
produced the identical rejection. This is not two successful compatibility runs.

Both original requires at403/410 now resolve and all three original files parse.
Semantic admission rejects `$stack_top->prev_offset = ...` after `array_pop`
because the receiver's generated class is unknown. Reaching this semantic error
does not establish earlier method/body semantics. Generation/build/execution,
full parse comparison and G83-PARSER remain blocked. P2 unknown-receiver writes,
stack mutation and object snapshots, plus tokenizer/JSON integration, stay open.
No extraction, concatenation, interpreter fallback or handwritten parser is used.

Nine new catalog checks are registered, including one real-file differential
case. No old successful native/consumer suite or enabled-PCRE run was repeated.
Local CLI capability and WebMCP exclusion are updated in the same change set;
no route/auth/deployment change. Devboard is unidentified/not synchronized.
