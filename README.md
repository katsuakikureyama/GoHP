# GoHP — Go Hypertext Processor

GoHP is an experimental **PHP-to-Go source transpiler with version-specific
PHP 8.x compatibility checks**. It emits readable Go functions and structs,
builds native binaries, and runs without an embedded PHP interpreter.
Phase 0 supports a documented closed subset on PHP **8.3 and 8.4**, verified
independently against pinned native PHP oracles.
Unsupported constructs fail with diagnostics; compilation alone is not compatibility.
**WordPress, arbitrary Composer packages and customer applications are not runnable.**
Gossip is a separate application-integration consumer, not special behavior in the compiler.

## What works today

Every entry below is bounded, not full PHP support. Exact restrictions and
function lists live in the linked tables; actual results are per-run evidence.

| Area | Current generated-code scope |
|---|---|
| [Syntax and control flow](docs/compatibility.md#syntax-and-control-flow) | Named functions, braced conditionals, foreach/while, bounded calls and array writes |
| [Built-in functions](docs/compatibility.md#built-in-functions) | Selected array/string/type helpers; bounded `array_any` / `array_all` are PHP 8.4-only |
| [Values and types](docs/compatibility.md#values-and-types) | int64, byte strings, bool, null and ordered arrays with detached value copies |
| [Objects and callables](docs/compatibility.md#objects-and-callables) | Simple public classes and capture-free arrows; no inheritance or captured closures |

PHP reference syntax, general include/autoload, PHP-compatible exceptions and extension
APIs remain unsupported. Internal shared-array/ownership primitives exist, but
**are not generated-array COW or PHP-reference support**.
See [implementation stages](docs/reference-value-model.md#delivery-gates-and-status).

## Active PHP targets

| Compiler label | Minimum native oracle in that series | Evidence |
|---|---|---|
| `8.3` (default) | 8.3.36 | Separate pinned CI job and archive |
| `8.4` | 8.4.27 | Separate pinned CI job and archive |

Other series are not active. A target label is not a patch selector; neither a
version floor nor passing fixtures guarantee security or whole-language support.
[Version policy](docs/compatibility.md#active-php-version-policy) and
[CI image pins](.github/workflows/ci.yml) define admission and reproducible runs.

## Reproduce the proof of concept

With Docker Engine and Compose v2, run the default 8.3 suite:

```sh
docker compose build test
docker compose run --rm test
```

Build downloads images; test execution has networking disabled and requires no
credentials or external services. Alternatively, with Go **1.23.12** and native
PHP meeting the selected target's floor:

```sh
# Use a fresh output workspace. Defaults to the 8.3 target.
sh scripts/check.sh
go run ./cmd/gohp -in tests/fixtures/core.php -out build/example.go \
  -php-version 8.3 -report build/example.report.json
go build -o build/example build/example.go
./build/example
```

For a local 8.4 suite, use native PHP 8.4.27+ and
`GOHP_PHP_TARGET=8.4 sh scripts/check.sh`; pass `-php-version 8.4` to the compiler.
CI builds a separate pinned image for each target; the default Compose image is 8.3.

Generated code imports this module's `runtime`: build inside this module, or
pin the same GoHP revision in the consumer's `go.mod`. Output/report destinations
must be fresh; use new paths for repeated generation. Wait for successful exit
and verify hashes before using artifacts. See the [publication contract](docs/integration.md).

## Progress and next milestones

[Public-object array snapshots](docs/block-data-constructor.md) accept known
classes or inferred-unknown expressions through the existing runtime guard.
This is not general `(array)` casting or complete original-parser execution.

[Static declaration closures](docs/static-source-loading.md) now connect checked
files and once-only edges to file-aware IR, diagnostics and Go generation.
Explicit manifest/root/entry mode is required; ordinary `-in` retains its subset.
General loading and the complete parser remain unsupported. New bundle
reports/maps are v0.2; legacy reports remain v0.1.

[Property `isset`](docs/property-isset.md) probes named paths without throwing
on missing/null slots or confusing nonnull falsy values with absence. Ordinary
property reads remain strict; dynamic/magic property semantics are not implied.

[Declared class parameters](docs/declared-class-parameters.md) add exact nonnull
object entry guards to functions, public methods and constructors. Object
identity is preserved; inheritance and nullable/union annotations remain excluded.

[Nonnumeric string switches](docs/string-switch.md) add token-name dispatch,
fallthrough and nearest-switch break for the fixed parser gate. Numeric/mixed
comparisons and the remaining whole-parser dependencies are still excluded.

The [positional list slice](docs/positional-list-assignment.md) adds bounded
multi-local array unpacking for the fixed parser gate. General destructuring,
missing-key warning semantics and the complete parser remain unsupported.

The [integer/null relational slice](docs/relational-int-null.md) addresses the
fixed original parser's first measured `<`/`>` syntax blocker. It does not yet
make the complete parser runnable; the application-first gate remains active.

The harness compares PHP and compiled-Go stdout byte for byte, checks normal
exits and a bounded quiet-stderr profile. It does not prove general warning,
exception or side-effect equivalence. Browse [CI runs](https://github.com/katsuakikureyama/GoHP/actions)
and the [function/feature evidence](docs/compatibility-ledger.md); test counts are
**not a PHP or WordPress compatibility percentage**.

The [roadmap](docs/roadmap.md) now prioritizes **one pinned Gossip application on
PHP 8.3**, beginning with the complete original WordPress block parser and its
dependencies. Implement the PHP semantics blocking that unit before unrelated
builtins, storage optimizations or new target expansion. Existing 8.4 checks
remain regression coverage; they are not a second application delivery target.
Gossip owns the [application acceptance gates](https://github.com/katsuakikureyama/Gossip/blob/master/docs/php83-application-acceptance.md).
Application coverage and performance comparisons are **not yet measured**.
The first [WordPress source manifest](docs/application-corpus.md) fixes whole-file
scope and supports [byte verification, lexical inventory and candidate name evidence](docs/application-corpus.md#lexical-inventory),
not application compatibility or resolved-call coverage.
The [dependency triage](docs/application-corpus.md#dependency-triage-separate-selection)
locates frequent candidates' definitions and records unresolved hook/translation/storage requirements.
Earlier [Gossip source comparisons](docs/integration.md) retain their original pins;
they do not certify current main or a complete WordPress application.

The optional [PHP 8.3 regex local-output experiment](docs/preg83-local-output.md)
connects a narrow `preg_match` statement to the PCRE2 adapter. Named property
subject/offset reads use actual string/integer guards. It needs explicit
compiler opt-in and a separately enabled native backend; ordinary CGO0 CI is
not evidence of PCRE compatibility. PHP 8.4 and general output references remain excluded.

[Single-quoted backslash handling](docs/single-quoted-backslashes.md) preserves
literal regex/text bytes without rewriting PHP source spelling, for both targets.
This is string syntax support, not full original WordPress parser execution.

## Scope and documentation

[Defined-variable `empty`](docs/array-empty-boundary.md) accepts arrays and
strings, including PHP's special `"0"` case; undefined addresses and other scalar
domains remain excluded.

[Named nested-property updates](docs/public-property-write.md#named-nested-property-updates)
support the parser's object-identity-preserving assignments, string concatenation
and array appends. [Direct property writes](docs/public-property-write.md) also
admit defined unknown locals using explicit runtime object/declared-slot guards.
Full parser execution still needs the remaining dependencies.

- [Architecture](docs/architecture.md) and [PHP 8.x versioned design](docs/php8-versioned-compatibility.md)
- [Compatibility tables and exclusions](docs/compatibility.md)
- [Roadmap and completion criteria](docs/roadmap.md)
- [Development](CONTRIBUTING.md), [security boundaries](SECURITY.md) and [licensing](docs/licensing.md)

No final project license has been selected. Parser adoption and licensing remain
separate decisions; this repository does not claim an OSS release license.

## 日本語概要

PHP 8.x の版別互換性を検証しながら対応範囲を広げる、Goソース生成型の実験的トランスパイラです。
現在は8.3／8.4の限定範囲を、それぞれ固定したPHPで独立検証します。
当面は固定版GossipのPHP 8.3動作を最優先とし、原本ブロックパーサー全体から進めます。
8.4の既存回帰検証は維持し、新版対応や性能最適化は後回しにします。
参照・COWは内部基盤と限定的な生成連携までで、PHP参照構文は未対応です。
WordPress全体はまだ動作せず、対応率・性能比較も未測定です。
[対応表](docs/compatibility.md)と[到達目標](docs/roadmap.md)で実装済み・未完了を区別します。
