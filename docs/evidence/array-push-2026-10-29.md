# Bounded array_push locations — 2026-10-29

This increment addresses the fixed original parser's measured codegen blocker
at165, not unrelated builtin breadth. Only a standalone call or a whole RHS
assigned to a bare local is admitted. The existing local/direct named property
address is separate from value arguments in additive IR0.1. The property owner
is resolved first; values snapshot once left-to-right, the current actual array
is read afterward, each successful append publishes, and the int count result
is bound last. No general references, embedded mutable expressions, array_pop,
new storage lane or report schema. See the
[central contract and official sources](../array-write-boundary.md#bounded-array_push-calls).

## One native design diagnostic

Before implementation, the already-cached exact image
`php:8.3.36-cli@sha256:98d555ff97b3865140ad31032508c3ca089fa97a8944b14ca9dc03ea05d788cd`
was used once, without networking or persistent container/image creation:

```sh
docker run --rm --network none \
  php:8.3.36-cli@sha256:98d555ff97b3865140ad31032508c3ca089fa97a8944b14ca9dc03ea05d788cd \
  php -d error_reporting=-1 -d display_errors=stderr -d log_errors=0 -r \
  'class Box { public $items = ["old"]; } function reset_box($b) { $b->items = ["reset"]; echo "rhs;"; return "added"; } $b = new Box(); echo PHP_VERSION . ";"; $n = array_push($b->items, $b->items, reset_box($b)); echo json_encode([$n, $b->items]) . "\n"; $b->items = "before"; $n = array_push($b->items, reset_box($b)); echo json_encode([$n, $b->items]) . "\n";'
```

Exit0; observed stdout:

```text
8.3.36;rhs;[3,["reset",["old"],"added"]]
rhs;[2,["reset","added"]]
```

This establishes the native reset/repair and old-payload observation used in
the design. It is not a Go/PHP fixture comparison or PHP8.4 evidence. PHP's
broader array defaults/json_decode usage here is diagnostic-only, not added
to the compiler. The cached development image `gohp-test` was unavailable;
no new development image or dedicated cache was built.

## Focused checks and a corrected static target guard

Initial clean code `2d01aa04b09d926d48299baa4a2210065af435e8`, tree
`eff55f7dcdaba035382c04adb7a909f7335e2280`, base main10086b9.
Cached Go1.23.0 DarwinARM64/CGO0, cached Go also on child PATH, offline modules
(`GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOWORK=off`), shared existing
`GOCACHE=/private/tmp/gohp-go-cache`:

```sh
go test -count=1 ./internal/frontend ./internal/codegen ./internal/compat
go test -count=1 ./runtime ./tests -run '^TestArrayPush'
```

- Frontend0.295s, catalog1.942s, runtime0.407s and generated host/fixture2.491s
  PASS once.
- Codegen1.454s FAIL: source/rawIR undefined targets were generated instead of
  statically rejected (two assertions), and an excluded nested-property target
  used malformed-IR rather than unsupported-source classification.
- The missing scope-existence check was added before validating the address;
  non-bare property receivers now reject as unsupported, while malformed
  metadata still uses IR002. No runtime/value/evaluation implementation changed.
- Corrected clean code `69126e5acc20b70161115607417e495ecb79f645`, tree
  `66863866d8b22d07ddc3fe9d15ae446798141853`:

```sh
go test -count=1 ./internal/codegen -run '^TestArrayPush'
```

The three directly affected compiler checks PASS0.262s. Unaffected earlier
frontend/catalog/runtime/host/fixture results are reused; no full successful
suite rerun. The original probe below has a declared `$this` root and is not
affected by this undefined-address correction, so it was not repeated.

Nine new checks at `php.builtin.array-push` include owned native differential,
lowering/maps, source/IR boundaries, three runtime contracts and generated
fixture/host ABI checks. Fixture SHA256 is
`5f6bd7cd6e5613a5aaf164b2c4191067ea5efa1e34a1a50d8d4cdd8df7ebc663`.
It observes zero/multiple counts, negative/string/sparse keys, alias/self
snapshots, detached nested arrays, shared child identifiers, known/unknown
property targets, same-property replacement/repair, ordered effects and
`$this->stack` pushes of fresh generated frames. Runtime/host checks retain
array guards even for zero values, owner-before-payload/current-array-after-
payload ordering, result-alias and loop behavior, and successful prefixes on
overflow. Complete structural hosts explicitly opt in, not as provenance/auth.

Native equivalence belongs to each exact-candidate automatic PHP8.3/8.4 run,
not these local lowering/ABI tests. The PR result records exact commits, runs,
archives, fixture output and quiet-stderr observations. General exception,
stderr and side-effect equivalence remains unmeasured.

## One unchanged complete-original closure probe

At initial clean code `2d01aa0` above, fixed WordPress/WordPress
`54106f6c545fc8b881f6b39b7963002bece6c28c`, version7.2-beta1-64693, complete
Parser/Block/Frame plus existing notice/provenance, unchanged
[Gossip closure manifest](https://github.com/katsuakikureyama/Gossip/blob/master/gossip/tests/gohp/parser-source-closure.json)
SHA256 `89f897c64146705559eae41bb50d49aff57abad0a11ee02d16d27fe78d64adaa`:

```sh
go run ./cmd/gohp -manifest GOSSIP_ROOT/gossip/tests/gohp/parser-source-closure.json \
  -source-root GOSSIP_ROOT -entry wp-includes/class-wp-block-parser.php \
  -php-version 8.3 -experimental-pcre2 -out FRESH/parser.go -report FRESH/report.json
```

One invocation, same cached/offline/CGO0 settings. Exit1, stdout0bytes,
stderr101bytes, SHA256
`246e5052200024fe11f7578fb32e1d9d2bd9215f7449d6c06bb020a0c0be9b00`:

```text
wp-includes/class-wp-block-parser.php: PHP004 line 205: unsupported function array_pop
exit status 1
```

The last line comes from go run. The former array_push165 codegen rejection
passes. No original Go/report/staging artifacts; empty temporary directory
removed. Next is the original `array_pop($this->stack)` at205. This is still
compiler admission, not original runtime progress. Full original generation,
build/run,12-case/24-call comparison and G83-PARSER remain blocked; remaining
stack, tokenizer/JSON and other full-unit contracts need evidence.

No original/manifest/consumer-pin changes, source extraction/concatenation,
handwritten parser, interpreter fallback or unrelated builtin. Earlier native
consumer evidence remains at its own pins. Local `gohp.cli.generate` stays
WebMCP-excluded with fresh/report-last publication and generated/unverified
result; no resource/schema/auth/route/auto-submit/deployment changes, dedicated
cache or persistent binary. Devboard unidentified/not synchronized.
