# Optional PCRE2 engine bridge (B1)

`internal/pcre2` is an experimental **engine-only** Go/C bridge. It does not
implement PHP `preg_match`, output arguments, capture arrays, `preg_last_error`,
compiler admission or PHP-version compatibility. The default build still uses
`CGO_ENABLED=0`. There is no automatic fallback or dependency installation.

This implements B1 of the consumer's [staged contract](https://github.com/katsuakikureyama/Gossip/blob/f6236ac365fbc09e8491e351c46ba013720a6de4/docs/pcre2-bridge-contract.md).
B2 (PHP values and per-Execution error state) and B3 (validated writable local
output IR/storage) remain separate, unfinished work. No storage/Frame APIs,
generated runtime imports, compiler behavior or Gossip pins change here.

## Engine contract

`Match(Request)` accepts pattern **body** bytes, whole subject bytes, a
nonnegative byte offset, `ByteDotAll`, and explicit positive engine limits.
PHP delimiters/modifiers and negative-offset normalization are not this API.
Pattern and subject are each bounded to 4,096 bytes; there are at most 64
capturing groups (plus group zero). Limit ceilings are match 100,000, depth
10,000 and heap 131,072 KiB. These are engine limits, not a process memory limit.

The fixed profile is 8-bit PCRE2, DOTALL, LF newline, BSR_UNICODE and built-in
C character tables. UTF/UCP/JIT are disabled. Inline configuration that changes
the fixed engine profile is rejected; ordinary regex syntax is still interpreted
by PCRE2, not converted to RE2. NUL/invalid UTF-8 remain bytes, and offsets can
start within a UTF-8 encoding. A successful build is not a promise of arbitrary
PCRE syntax, PHP behavior or safe public arbitrary-pattern execution.

The result preserves the raw return code and compiled capture count, native
lexically ordered name table, and numbered captures in numeric order. Only
groups through the engine's returned count are present. An unmatched middle
group has `Present=false` and start/end -1; a matched empty group has
`Present=true` and equal nonnegative offsets. Trailing unmatched groups are
omitted. Capture bytes and names are Go-owned copies.

No-match (-1) and match-limit (-47) are explicit raw outcomes. Unknown negative
engine outcomes produce `EngineError` with the raw code; malformed patterns
produce `CompileError`, invalid arguments/profile produce `ErrInvalidInput`,
and disabled builds return typed `ErrUnavailable`. None are translated to
PHP false/error state. An error must never be treated as an ordinary no-match.

Each call owns its C input copies, compiled pattern, compile/match contexts and
match data, freeing them before returning. Empty inputs have valid C buffers.
There is no retained Go pointer, compiled-pattern cache, callback into Go or
global last-error state. Calls are synchronous: a Go context cancellation
cannot interrupt an arbitrary C match. Tests use a separate process timeout;
it is not an emulated PHP error or runtime cancellation implementation.

## Build and bounded validation

Both `-tags gohp_pcre2` and `CGO_ENABLED=1` are required. Without either, only
the unavailable implementation is built and no PCRE headers/archive are needed.
The enabled build requires explicitly supplied header/library paths; no
`pkg-config`, Homebrew lookup, downloading or silent dynamic-library fallback
is embedded in the package. For a prepared, reviewed installation:

```sh
# Set PCRE2_PREFIX to the verified installation, not a floating/unreviewed one.
CGO_ENABLED=0 go test ./internal/pcre2
CGO_ENABLED=0 go test -tags gohp_pcre2 ./internal/pcre2
CGO_ENABLED=1 \
  CGO_CFLAGS="-I${PCRE2_PREFIX}/include" \
  CGO_LDFLAGS="${PCRE2_PREFIX}/lib/libpcre2-8.a" \
  GOHP_PCRE2_REFERENCE_DIR="${GOSSIP_CHECKOUT}/docs/evidence" \
  go test -count=1 -timeout=30s -tags gohp_pcre2 -v ./internal/pcre2
```

`TestRetainedConsumerReference` is explicitly opt-in; without the environment
variable it skips and establishes **no** consumer comparison. It verifies both
retained evidence file hashes, reads the pattern from Gossip (without copying
WordPress source into GoHP), removes only `/` and `/s`, and compares all 24 raw
engine results, names, presence, bytes and byte spans against the independent
C-probe record. The recorded PHP8.3.36/PCRE10.42 reference is reused transitively,
not re-executed or relabeled as a new PHP-vs-generated-Go run. No B2 PHP mapper
is implemented in this test or production code.

The default pure-Go CI lane is unchanged; it checks the disabled package along
with existing tests. These engine tests are not entries in the PHP compatibility
catalog. A Linux enabled-backend CI lane needs a separately pinned development
image, compiler/library identity and measurements before claiming support.
The first local candidate is darwin/arm64, PCRE2 10.47, linked to an explicit
static archive; its [actual measurements](evidence/pcre2-bridge-2026-10-27.md)
record 24/24 retained raw comparisons and the enabled/disabled checks. This is
bounded bridge evidence, not a PHP `preg_*` support entry.

## Distribution and capability boundaries

No PCRE source/header/archive is vendored and no executable is published. The
installed dependency's `BSD-3-Clause WITH PCRE2-exception` notice and the
project's [undecided license](licensing.md) remain separate distribution gates;
this internal experiment is not release approval or a licensing conclusion.

UI/WebMCP/MCP capability: **excluded** — an internal developer-only library
and local test operation, with no browser interface, service route, resource
mutation, network call, account, auth change or production deployment target.
