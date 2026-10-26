# B1 PCRE2 bridge measurement — 2026-10-27 JST

Status: **bounded engine bridge PASS**, not PHP builtin/compiler compatibility.
All commands ran in an isolated GoHP feature worktree. No new dependency install,
PHP execution, binary publication, Gossip repin or production deployment occurred.

## Executed identity

- Base: `5a958a2606832510cdd63cbf6ae2de02b5f0db68` (#108).
- Tested clean code commit: `f083c50c4a9f6d6dede49934c031924127d1a98e`.
- Tested tree: `c9070446489425eca27291389180a0dab80e5763`.
- Local Go: `go1.23.0 darwin/arm64` (module minimum; **not** CI's Go1.23.12).
- C compiler: Apple clang17.0.0 (`clang-1700.0.13.5`), target
  `arm64-apple-darwin24.5.0`; explicit Xcode toolchain binary, not launcher.
- SDKROOT: Xcode MacOSX SDK, `xcrun --show-sdk-version` reported 15.5.
  That later metadata command printed sandbox FSEvents/cache diagnostics;
  it was not part of the successful compile or match process.
- Installed dependency: Homebrew Cellar `pcre2/10.47_1`; runtime version
  `10.47 2025-10-21`, 8-bit API. Header/archive hashes checked before and after.
- `otool -L` on the executed test binary listed only libresolv.9,
  CoreFoundation and libSystem.B: no PCRE dynamic library. The explicit static
  archive was linked. Linker warned about the same archive supplied twice by
  cgo: `ignoring duplicate libraries`; build exited 0. No fallback library.
- `go version -m` on that test binary confirmed Go1.23.0 (test binaries do not
  provide a full module build-settings record here); flags below were explicit.

| Artifact | SHA-256 |
| --- | --- |
| Installed `include/pcre2.h` | `058462dc030e845ee627e19ac36347f561f5cc44d89eaf1c0010f2b363ab68d5` |
| Installed `lib/libpcre2-8.a` | `ce40246eba1dfdd99e43142cfc33264610d7c965fab3062f386c68a6636a4a54` |
| Actual clang executable | `4c458256bcdf913774de1bb4a37768244d3b41f32ad55afb2dfec3432f46f4fe` |
| Installed `LICENCE.md` | `197d8a73ffee0d6b09adba2f9c677b5f5aede24edf89258a68e48248d010d811` |
| Executed `pcre2.test` | `bca7e148af86ac2a1a6873c51adf1d036dcf66f0e9a63baa32890bef361567d1` |
| [Enabled run log](pcre2-bridge-2026-10-27.log) | `e44cdda0c9a2590571ffb8000a7e2186eeacf19cc16b771c6672e79b6bbd9a36` |
| `internal/pcre2/bridge.c` | `84fe528d37117be6e5776759a0c53295022e651d76753a6c50fff95a71d5a5d8` |
| `internal/pcre2/bridge.go` | `f6be7ac9b2d50ca24f011e2f6e2e3b00659462f162917cdb4e1210ea8075ce7c` |
| `internal/pcre2/bridge.h` | `cc901fce71928028f1dcf3fddd6fb21f79ec5e5175a34af2578ed740a3dc2372` |
| `internal/pcre2/bridge_disabled.go` | `f4319590f11a604712f7cc1d55351ad2d905e91804af54bbc0e7219a83f39ee0` |
| `internal/pcre2/bridge_disabled_test.go` | `3d1b2e5ebe7d3e54aec681bdfc02716333216fda0b2ecad94fd27ff2210f6b41` |
| `internal/pcre2/bridge_enabled.go` | `35b3910d0a82753297de576718c3eace65901643cf2c9d7be03b5a64ce37f2dd` |
| `internal/pcre2/bridge_enabled_test.go` | `704b2ce9b00bbbd676c7a89e9f42ae3c2300a237aec55ab5c4b82d89f0166be1` |
| `internal/pcre2/reference_test.go` | `014c3065585ffab4f900fd7db8dd364fb21a52b757efdc9882a1207b56bdc530` |

The installed bottle/source identity is recorded by the archive/header, not a
claim that PCRE upstream source was rebuilt or audited in this run. Toolchain
hashes do not guarantee byte-reproducible binaries on another host.

## Commands and outcomes

Path placeholders below avoid publishing personal absolute paths. `GO123` was
the cached Go1.23.0 binary; `PCRE2_PREFIX` the exact Cellar installation above;
`CC` the Xcode toolchain clang; `SDKROOT` its MacOSX SDK. `RUN_DIR` was a fresh
temporary directory, and `GOSSIP_CHECKOUT` the clean #49 feature checkout at
`ad6555a1d835ed3cc83992fec82770f4f5aefc2e`, tree
`2e66dfa7656beb268c9e58d44c570d2199de2ffe` (same tree as merged master
`f6236ac365fbc09e8491e351c46ba013720a6de4`).
All Go commands used `GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOWORK=off`
and `GOCACHE=$RUN_DIR/cache`. No successful test configuration was rerun.

```sh
# Enabled compile, explicit C compiler and SDK environment inherited:
CGO_ENABLED=1 CGO_CFLAGS="-I${PCRE2_PREFIX}/include -O2 -Wall -Wextra -Werror" \
  CGO_LDFLAGS="${PCRE2_PREFIX}/lib/libpcre2-8.a" \
  "$GO123" test -c -tags gohp_pcre2 -o "$RUN_DIR/pcre2.test" ./internal/pcre2
GOHP_PCRE2_REFERENCE_DIR="$GOSSIP_CHECKOUT/docs/evidence" \
  "$RUN_DIR/pcre2.test" -test.v -test.timeout=30s

# Three distinct disabled configurations, each executed once successfully:
CGO_ENABLED=0 "$GO123" test -count=1 -v ./internal/pcre2
CGO_ENABLED=0 "$GO123" test -count=1 -tags gohp_pcre2 -v ./internal/pcre2
CGO_ENABLED=1 "$GO123" test -count=1 -v ./internal/pcre2

# Affected-package static analysis, same enabled compile settings:
CGO_ENABLED=1 CGO_CFLAGS="-I${PCRE2_PREFIX}/include -O2 -Wall -Wextra -Werror" \
  CGO_LDFLAGS="${PCRE2_PREFIX}/lib/libpcre2-8.a" \
  "$GO123" vet -tags gohp_pcre2 ./internal/pcre2
```

Enabled build/run/vet all exited 0. Eight generic top-level tests passed,
including 12 invalid-input subcases. The retained consumer test passed all
**24/24** raw cases (17 matches, 6 no-matches, 1 match-limit outcome), with the
same raw codes, names, capture presence, bytes and spans as the C-probe record.
The three disabled package tests exited 0 (0.243s, 0.225s, 0.220s respectively).
Format/diff checks passed. No whole manual PHP suite was run for this internal
package change; separately triggered default CI must be identified by its own
run/commit and does not exercise the enabled PCRE backend.

Before that success, initial commit
`c6752ebb129c46f9c832f6070210bd0fc37a5fad` had two build-stage failures:

1. Enabled compile without SDKROOT could not find `stdlib.h` in runtime/cgo.
   Explicit SDKROOT fixed the environment; no match executed in the failed build.
2. Disabled test's `errors.As` target used a value rather than pointer error
   type; vet rejected it before test execution. The two-line test fix was
   committed as `f083c50`; affected disabled checks were then run successfully.

## Retained references and exclusions

Both external files are hash-checked before replay:

- [Gossip PHP reference](https://github.com/katsuakikureyama/Gossip/blob/f6236ac365fbc09e8491e351c46ba013720a6de4/docs/evidence/parser-token-reference-2026-10-27.json):
  `5cbf560247afcd2c5f1462cab2f95f16bb822ac2d4d1756be16221cffdc6dc60`.
- [Gossip C-probe reference](https://github.com/katsuakikureyama/Gossip/blob/f6236ac365fbc09e8491e351c46ba013720a6de4/docs/evidence/parser-pcre2-comparison-2026-10-27.json):
  `5390d6c560ca7b0be9ab6363ddc65f12e0908ba4b68cffaa2d0998c1130d021d`.
- Exact pattern-body SHA:
  `b25801e7cea22f1635132d088e0292dcaeab141e9b036dbcdee6969166503473`.

This is actual Go→C execution compared with stored independent C results.
PHP8.3.36/PCRE10.42 evidence is reused through that record, not freshly executed.
No PHP8.4 bridge comparison, PHP wrapper, generated preg call, full next_token,
parser, Linux enabled backend, benchmark, allocator-fault injection, sanitizer
run, release/license decision or production deployment is claimed. C allocation
failure cleanup was reviewed, not fault-injected. See the [boundary](../optional-pcre2-bridge.md).
