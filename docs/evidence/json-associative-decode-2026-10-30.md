# Associative JSON and first complete original generation/build — 2026-10-30

Scope6517354075 was posted then immediately reread, and reread again before
production integration; no duplicate JSON claim. Base main
`af3e9ada36728cb317096b35af6ca0f22e3492c4` and Gossip67/master7b3acd6.
Parent owns signature/state/profile design, compiler integration, important
review and verification; bounded workers supplied two runtime files and two
owned fixture/host files only, without tests/commits. Parent corrected a nested
fixture write to the admitted one-level form and added invalid-UTF8 bytes via
existing byte substr, without adding a builtin or touching original source.

The [closed contract](../json-associative-decode.md) admits exact two-argument
literal-true associative decoding and per-Execution code/message readers.
Ordered arrays/strings/bool/null/int64 only; fraction/exponent/out-of-int64
tokens fail explicitly on valid JSON, never silently round/stringify/null.
Invalid syntax retains its JSON null/error path. No flags/custom depth/object
mode/constants, new dependency/storage lane/IR version or PHP interpreter.

## Local verification and correction

Initial clean code `ffdd4dfbe3eb7acd292e7563e485c098ac93f6ec`, tree
`2aaca214d099a556dfc0185e69211d0c15761224`. Cached Go1.23.0 Darwinarm64,
CGO0; cached Go on child PATH, GOTOOLCHAIN=local/GOPROXY=off/GOSUMDB=off/
GOWORK=off and existing shared GOCACHE, no dedicated cache or persistent binary:

```sh
go test -count=1 -run '^Test(JSONAssociative|StrtokLoweringAndSourceMap|ContextualCallableUsesCallerExecution)' ./internal/codegen ./runtime ./tests
go test -count=1 ./internal/compat
```

Codegen0.268s/catalog0.304s PASS. Runtime and generated-host build FAILED:
combining a UTF16 pair performed an intermediate operation in uint16, so
constant65536 overflowed at compile time. Widen each surrogate contribution to
rune before combining; no parser/profile/error contract weakened. Initial
failure retained, not relabeled PASS. Fixture lowering itself had no failure.

Clean corrected code `35524a636073f4abbc8bb69b9e36263e71ac62d8`, tree
`2bec82d98e3ca2c36031d5aa5bd3f5e46c0f2a88`, same environment:

```sh
go test -count=1 -run '^Test(JSONAssociative|ContextualCallableUsesCallerExecution)' ./runtime
go test -count=1 -run '^TestJSONAssociativeHostEntry$' ./tests
```

PASS0.216/0.627s once; earlier successful compiler/catalog/fixture-lowering
ranges reused, no full manual rerun. Ten catalog checks cover compiler/maps,
source/IR negatives, four runtime value/error/depth/state checks, real fixture
lowering, external generated ABI and independent native differential comparison.
New owned fixture SHA256
`d4efaa01dbe151eabf146e89a5fe54e96134fda7fb8d9ede5e1a618b8b2f9600`.
No native local PHP, and the existing Docker daemon was unavailable; no daemon
started, image downloaded or local native result claimed. Existing pinned
PHP8.3/8.4 automatic CI is the independent native gate for the exact candidate.

## One complete unchanged-original probe, then one build

WordPress/WordPress `54106f6c545fc8b881f6b39b7963002bece6c28c`, same
complete Parser/Block/Frame and notice/provenance bytes. Existing closure
manifest SHA256 `89f897c64146705559eae41bb50d49aff57abad0a11ee02d16d27fe78d64adaa`.
One invocation at clean35524a6, same environment (personal paths replaced):

```sh
go run ./cmd/gohp -manifest GOSSIP_ROOT/gossip/tests/gohp/parser-source-closure.json \
  -source-root GOSSIP_ROOT -entry wp-includes/class-wp-block-parser.php \
  -php-version 8.3 -experimental-pcre2 -out FRESH/parser.go -report FRESH/report.json
go build -o FRESH/parser-build FRESH/parser.go
```

Both exits0. Generation stdout/stderr0bytes; v0.2 Go/report files published,
without rewriting/extracting/concatenating the original PHP. All three original
sources retain their fixed identities. Generated Go SHA256
`55c66c931ed9da4be7a0a263129242b169a68127e83bb526fbc1f6d247f402dd`,
report SHA256 `bbf61c482f333b6509cff72b8e27cd2c7833238bf81a4148dda3e4f08104a1df`.
The one CGO0 Darwinarm64 build produced a temporary Mach-O executable, then
only that generated binary was removed. Go/report/logs are retained outside
source worktrees; no persistent binary or new cache.

This establishes **complete fixed three-file generation and Go build**, not
original parser invocation, enabled-native-PCRE execution,12-case/24-call
comparison or G83-PARSER acceptance. Empty generated main was not run as a
fake parser success. The next measured integration is the actual original
parse calls/driver and native-engine configuration, not another claimed
compile blocker. Existing enabled-PCRE observations are not remeasured here.

Original sources/all manifests/pins/auth/routes/deployment unchanged. Local
gohp.cli.generate remains WebMCP-excluded with unchanged source/report/fresh
publication contracts. Devboard unidentified/not synchronized, production
destination not configured. Existing137/older successes are reused.
