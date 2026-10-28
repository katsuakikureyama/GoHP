# Named regex inputs — 2026-10-29

Scope: the [PHP8.3-only local-output experiment](../preg83-local-output.md).
This is not default regex support, PHP8.4 regex, a whole WordPress parser result
or general output/reference/error compatibility.

## Source and checks

All executable checks below used clean GoHP commit
`8403235e0e14fb460354bb73c9400b7e94708004`, tree
`f0b982ac33665bdcb46f71ac9de9bbffb71d393e`, based on maina30709e.
Go1.23.0 Darwin/ARM64; offline cached modules and the existing build cache.
Each successful configuration was run once. Later evidence-only changes reuse
these results; no successful suite or source probe was repeated.

With `CGO_ENABLED=0 GOCACHE=CACHE GOPROXY=off GOTOOLCHAIN=go1.23.0`:

```sh
go test -count=1 -run '^TestPreg83(Compiler|PropertyInputs|CLI)' ./internal/codegen ./cmd/gohp ./tests
go test -count=1 ./internal/compat
go vet ./internal/frontend ./internal/semantic ./internal/codegen ./tests
```

PASS: codegen0.459s, CLI2.296s, generated guards4.169s, catalog0.896s;
vet exit0. Named property nodes/root metadata, destination-root aliases, known
class declarations and source maps are checked. Generated CGO0 binaries reject
nonstring subjects, noninteger offsets and null/scalar intermediate objects
before subsequent output. These tooling checks do not execute a native engine.

## Enabled native comparison

First setup attempt stopped **before test execution**: the Go launcher tried
toolchain module verification while `GOSUMDB=off`. It was not a PHP/codegen
comparison failure. The correction directly selects the already-cached
Go1.23.0 binary and puts its bin directory first on PATH with
`GOTOOLCHAIN=local`; no dependency installation or source change.

Successful invocation used the same existing enabled configuration as B3:
CGO1, tag `gohp_pcre2`, Apple clang17, explicit macOS15.5 SDK, Homebrew PCRE2
10.47_1 static archive, CFLAGS `-I<PCRE_PREFIX>/include -O2 -Wall -Wextra -Werror`,
LDFLAGS `<PCRE_PREFIX>/lib/libpcre2-8.a`, and offline `GOPROXY/GOSUMDB/GOWORK`.
Header SHA256 `058462dc030e845ee627e19ac36347f561f5cc44d89eaf1c0010f2b363ab68d5`;
archive `ce40246eba1dfdd99e43142cfc33264610d7c965fab3062f386c68a6636a4a54`.
No pkg-config or dynamic-library fallback is used.

```sh
GOHP_PREG83_GENERATED_PHP_IMAGE=php:8.3.36-cli@sha256:98d555ff97b3865140ad31032508c3ca089fa97a8944b14ca9dc03ea05d788cd \
CGO_ENABLED=1 go test -v -tags gohp_pcre2 -count=1 \
  -run '^TestPreg83GeneratedNativeComparison$' ./tests
```

PASS6.02s (package6.341s): **15 ordered typed observations** from actual generated
Go versus fresh PHP8.3.36. Four new property rows cover `$this` subject/offset,
nested UTF-8 byte input and negative offset, changed properties between calls,
capture-copy independence and unchanged input properties. The shared-execution
row now additionally resets a measured backtracking error through a successful
property-input call. The remaining rows retain the earlier literal/local input
regression cases. Runtime/host Execution-isolation assertions also pass.

The oracle is the cached immutable image above, 64-bit Linux ARM64, PCRE10.42,
C LC_CTYPE, JIT0, backtrack100000, recursion10000 and memory128M. The harness
checks version/profile and uses pull-never, network-none, read-only, nonroot,
cap-drop-ALL, no-new-privileges, CPU1 and pids64; only the generated PHP input is
mounted read-only. Both children require exit0 and empty stderr.

Comparison preserves types, ordered array entries/key types and base64 string
bytes. It is **structural JSON equality, not serialized JSON byte equality**;
serializer key order, escapes and final newline differ. General side-effect,
warning/exception text and resource-error equivalence remain unmeasured.

| SHA256 identity | Value |
|---|---|
| Common PHP fixture after its existing single NUL substitution | `ed05e03fe9d576676e2883e48a2cca9e4e6989418b7435bd053e02549d28669c` |
| Generated Go | `30b079b777e3ef613609d8e1f0ea3f8e275dc46723840162fbd6edc279044064` |
| Go observer | `27e25a117425a844445bd8dda0c809bc96fb7732165be36398670962646d78f2` |
| PHP observer | `2a4225e1ba6383724512506bc5fc9905f8b8a2a89d5416ee0c1671e5e55de986` |
| Generated stdout | `9be47033150e76a8dca619f69831e554f699a364153208f9d383ef1805fc7e68` |
| [Native stdout](preg83-property-inputs-2026-10-29-native.json) | `7332f6904519ee656e5987786f3123f1d319cd4c041e57344ddc4efe0c9f68a9` |

## Unchanged full-source admission probe

One invocation at the same clean source commit:

```sh
CGO_ENABLED=0 GOCACHE=CACHE GOPROXY=off GOTOOLCHAIN=go1.23.0 \
go run ./cmd/gohp -in GOSSIP_ROOT/wp-includes/class-wp-block-parser.php \
  -php-version 8.3 -experimental-pcre2
```

Exit1, with no output/report destination requested:

```text
PHP002 line 349: unsupported parameter declaration (at "block")
exit status 1
```

The last line is the go-run wrapper. Whole original WordPress54106f6 source
SHA256 `8563621090de5e307431b1d29b78a9f75fd9ee238a0137b563a7db7b2e1f8d3b`
is unchanged: no extraction, concatenation, rewriting or replacement algorithm.
Regex property inputs at252/255 now parse; named class parameter349 is the next
frontend rejection. Complete semantic validation/generation/build/full parser
execution and application acceptance are still blocked. Remaining object/stack,
JSON and loading contracts are not silently certified by this probe.

Default Linux CGO0 CI is a separate regression gate; enabled Linux and PHP8.4
regex remain unmeasured. No route/auth/deployment/consumer-pin change; existing
developer-local CLI UI/WebMCP/MCP exclusion remains.
