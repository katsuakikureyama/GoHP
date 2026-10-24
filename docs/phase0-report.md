# Phase 0 implementation report

## Working development capabilities

The frontend, versioned IR, Go generator and compatibility runtime are separate
packages. Generated programs contain ordinary Go functions, structs, loops and
runtime helper calls. They do not contain a PHP evaluator or opcode VM. The CLI
emits source and a versioned provenance report; Gossip can consume that boundary
without WordPress-specific compiler semantics.

Native PHP 8.3.30 and PHP 7.4.33 reference runs use immutable official image
content digests alongside Go 1.23.12. Both Docker Compose services execute with
network disabled, a read-only filesystem, temporary build/cache storage and no
Linux capabilities. The Go module has no external Go dependencies.

## Validation

For each PHP reference, `sh scripts/check.sh` runs `go vet ./...`, all tests,
CLI compilation, source generation, native binary compilation and an independent
byte comparison with PHP. Three positive differential fixtures run per version:
functions/classes/ordered arrays; nested arrays/key replacement/scalars; and
version-specific operator precedence. The generated binary is executed with an
empty PATH so it cannot discover a PHP executable.

Three runtime unit tests cover array ordering/replacement, independent nested
copies and explicit overflow failure. Thirteen unsupported-source cases must
produce coded diagnostics. A missing-key runtime case must exit nonzero; it does
not claim equivalence with PHP warning behavior. The CLI must reject the closure
fixture without creating a new Go output file.

The `1 . 2 + 10` fixture matches `22` on PHP 7.4 and `112` on PHP 8.3, proving the
version selection affects lowering. The main fixture prints `total=10`, the three
ordered key/value pairs, and `3`. See `compatibility.md` for precisely bounded
claims. CI defines the same two container jobs; local execution is evidence of
readiness, not evidence that GitHub Actions has already run.

## Open work and risks

References, array mutation/COW optimization, closures, exceptions, inheritance,
include/autoload, Composer package resolution, most builtins and all extension
adapters remain unsupported. No WordPress application compatibility, DB mutation
compatibility, source maps, production sandbox or performance guarantee exists.
Some valid PHP expressions compile but fail explicitly when they reach an
unsupported runtime value operation; compilation is not functional PASS.

Licensing is a proposal awaiting audit and owner decision. All implementation
work is on `feat/phase0-transpiler`; the initial `main` contains only an empty
review baseline commit. No application code is merged, deployed or released.

Gossip integration requests and the CLI/report contract are recorded in
`integration.md`; no Gossip implementation is changed. The next PRs should first
strengthen the semantic/type-validation boundary and parser adapter, then add
references/array writes with real differential cases before broader PHP features.
