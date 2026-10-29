# Guarded inferred-unknown public-object snapshots — 2026-10-29

The existing `object_array_cast` / `runtime.ObjectToArray` path now admits
inferred-unknown operands. It still requires an actual nonnil public-property
object at runtime; this is not PHP's general scalar/array/null/Closure cast.
Known nonobjects and undefined reads reject before generation. No new IR kind,
runtime operation, class inference, report version or storage lane is added.
See the [central snapshot contract](../block-data-constructor.md).

## Focused local checks and limited corrections

Initial clean code `dbbcdc225247b0b826267ce0546814d9d7cf6e0a`, tree
`cc78cd70c8de172658a6be97937561c207127b3a`, based on main `efcdbc47`.
Cached Go1.23.0 DarwinARM64/CGO0, offline modules, cached Go also on child PATH,
shared existing `GOCACHE=/private/tmp/gohp-go-cache`:

```sh
go test -count=1 ./internal/codegen ./internal/compat
go test -count=1 ./runtime ./tests -run '^(TestObjectToArray.*|TestUnknownObjectArray(Fixture|HostEntry))$'
```

- Codegen0.386s, catalog0.601s and runtime0.421s PASS once.
- The two new `tests` checks initially FAIL (package0.249s). The fixture used
  unsupported multi-level array assignment at89, and the generated external
  module needed its Go directive aligned with the root module.
- Only test/fixture data changed: nested payload updates now read/copy/replace
  through the existing one-level array operations. The test module declares
  Go1.23.0 and uses the same offline `-mod=mod` child convention as existing
  generated-consumer tests. Production semantics were not weakened.
- Corrected clean code `af2cf093f71a311ad83a87c1a41ae5c687f87e21`, tree
  `7bcf1453ede92d9ae515a6f2dd69d55e9142adb8`:

```sh
go test -count=1 ./tests -run '^TestUnknownObjectArray(Fixture|HostEntry)$'
```

PASS0.771s. Already-passing compiler/catalog/runtime checks and the unchanged
production/source probe are reused, not manually rerun. The four old negative
source forms now admitted by the new contract are replaced by the new positive
tests; undefined and known-nonobject rejection remain covered.

Six new catalog checks at stable feature `php.syntax.public-data-class` cover
native differential, compiler lowering/maps, source/IR boundaries, generated
external ABI guards and the committed fixture. The fixture SHA256 is
`b15f5915e3ebc3ea410758305e894ccce1e3b10edf04b2daeb202e47e6ab37ec`.
It observes untyped parameter/return, indexed/extracted values and nested named
properties, once-only effects, null/falsy properties, declaration order, empty
classes, detached nested arrays and shared child identifiers. Its extracted
local is repository-owned, not a rewrite of the original algorithm.

Generated external ABI checks observe exactly one operand before the snapshot,
exactly one snapshot and no following statement after rejected actual domains.
They include nil/typed-nil generated objects, scalars, arrays, arrows, partial
protocols and false markers. A complete structural host explicitly opts in;
the existing marker is not unforgeable provenance or authorization.
Native equivalence requires separate automatic PHP8.3/8.4 candidate CI evidence;
these local checks alone are not native PHP PASS. Exact CI/run/archive outcomes
belong in the PR result, independently for each target. General stderr,
exception and side-effect equivalence remains unmeasured.

## One unchanged complete-original closure probe

At clean initial code `dbbcdc2` above, WordPress/WordPress fixed source
`54106f6c545fc8b881f6b39b7963002bece6c28c`, version7.2-beta1-64693, complete
Parser/Block/Frame plus existing notices/provenance, unchanged
[Gossip closure manifest](https://github.com/katsuakikureyama/Gossip/blob/master/gossip/tests/gohp/parser-source-closure.json)
SHA256 `89f897c64146705559eae41bb50d49aff57abad0a11ee02d16d27fe78d64adaa`.
One invocation using the same cached/offline/CGO0 configuration:

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

The last line comes from go run. The former semantic rejection of the unknown
snapshot at394 passes; the new rejection is in **codegen** at165, the original
`array_push($this->stack, new WP_Block_Parser_Frame(...))` call. A lower source
line here is a later compiler stage, not a claim of runtime progress through
the parser. No generated Go/report/staging artifacts; empty temp directory
removed. All original methods passed the current semantic checker, but that
does not prove their runtime behavior. Original generation/build/execution,
12-case/24-call comparison and G83-PARSER remain blocked. Remaining stack,
tokenizer/JSON and other full-unit contracts still need evidence.

No original/manifest/pin changes, extraction, concatenation, hand-written
parser, interpreter fallback or unrelated builtin implementation. Earlier
successful native/consumer evidence stays at its own pins. The same local
`gohp.cli.generate` capability remains WebMCP-excluded; no new resource/schema,
route, auth mode, auto-submit, deployment target, dedicated cache or persistent
binary. Devboard unidentified/not synchronized.
