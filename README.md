# GoHP — Go Hypertext Processor

GoHP is an experimental, general-purpose PHP-to-Go **source transpiler**. The
Phase 0 compiler emits readable Go functions and native structs, builds with Go,
and runs without PHP. It does not embed an opcode interpreter. It is not ready
to run WordPress, arbitrary Composer packages, or customer applications.

## Reproduce the proof of concept

Prerequisites: Docker Engine with Docker Compose v2. No credentials, external
services or PHP packages are required. Images are downloaded during build; test
execution has network access disabled.

```sh
docker compose build test
docker compose run --rm test
```

Or use Go **1.23.12** and native PHP **8.3.36 or newer within the 8.3 series** locally:

```sh
sh scripts/check.sh
go run ./cmd/gohp -in tests/fixtures/core.php -out build/example.go \
  -php-version 8.3 -report build/example.report.json
go build -o build/example build/example.go
./build/example
```

Generated Go imports this module's `runtime` package; build inside this module,
or add a dependency on a pinned GoHP commit in the consumer's `go.mod`. Never
ship an unpinned local checkout replacement as a release dependency.

Use fresh `-out` and `-report` paths for each invocation: existing destinations
are rejected. A report requires a file output. Builders must wait for successful
exit and verify report hashes before using either artifact; see the
[artifact publication contract](docs/integration.md).

`core.php` exercises a user function, integer sum, mixed ordered array,
`foreach`, public class property, method call and `count`. Its expected output:

```text
total=10
0:2
bonus:5
1:3
3
```

The test harness executes native PHP, generates Go, compiles a standalone binary,
and compares stdout byte for byte, requiring successful exits from both. Other
fixtures check duplicate/numeric array keys, nested arrays, scalar conversions
and PHP-version-dependent operator precedence. Unsupported syntax must fail
with a diagnostic; a separate runtime-negative test must fail on a missing key.
Compilation alone never proves PHP compatibility.
Builtins currently include `count(array)`, bounded `strlen(string)` (byte
length only, no implicit non-string coercion) and `array_keys(array)` (one argument,
ordered keys, no value filtering). See the compatibility contract.

### Active PHP target

GoHP's initial compatibility target is the PHP **8.3** language series only.
Native differential runs require PHP **8.3.36 or newer within 8.3**; the actual
oracle version is recorded in run evidence, along with the pinned image digest
in CI (host-PHP runs need no container). The CLI deliberately takes the language-series
label `-php-version 8.3`, rather than a patch number.

PHP 8.4 and later are not enabled or claimed. Adding a series requires a
separate policy/matrix entry with its own minimum release, pinned oracle image
digest, target-specific corpus and passing evidence; it is not a consequence
of accepting the same CLI flag. The minimum release floor is a maintenance
rule, not a security guarantee for the compiler, generated binaries or input.
See the [compatibility version policy](docs/compatibility.md#active-php-version-policy).

## Scope and documentation

- [Architecture and decisions](docs/architecture.md)
- [Compatibility and version matrix](docs/compatibility.md)
- [Function/feature catalog and verification-ledger design](docs/compatibility-registry.md)
- [Generate and browse function/feature evidence](docs/compatibility-ledger.md)
- [Read Composer requirements without executing code](docs/compatibility.md#composer-metadata-inventory)
- [Upstream assessment](docs/upstream.md)
- [Parser adapter assessment](docs/parser-adapter-assessment.md)
- [Statement-level source maps](docs/source-maps.md)
- [Gossip integration contract](docs/integration.md)
- [Development and roadmap](CONTRIBUTING.md)
- [Security boundaries](SECURITY.md)
- [Licensing proposal](docs/licensing.md) — no final project license yet

## 日本語概要

PHP を読みやすい Go ソースへ変換する初期 PoC です。関数・配列・クラスを
含む fixture を実際にビルドし、PHP との標準出力の一致を確認します。
参照、クロージャ、継承、Composer や PHP 拡張、WordPress 全体は未対応です。
未対応を互換性成功とは扱いません。正式ライセンスは監査と所有者判断待ちです。
初期の互換性対象は PHP **8.3 系のみ**で、native 比較には **8.3 系の 8.3.36 以上**を使います。
CLI の `-php-version 8.3` はパッチ番号ではなく言語系列の指定です。8.4 以降は未対応で、
追加時には系列ごとの最低版・固定oracleイメージdigest・専用fixture・検証証拠を別途そろえます。
最低版の指定だけで、コンパイラや生成物のセキュリティを保証するものではありません。
組込み関数は `count(array)` と限定的な `strlen(string)`、`array_keys(array)` に対応します。
`strlen` はバイト数を返し、数値・nullなどからの暗黙変換は対象外です。
`array_keys` は1引数で順序を保ったキー一覧を返します。値による絞り込みは未対応です。
生成レポートには Go の行から PHP の文開始行への対応を含みます。
列単位の位置情報やエラー表示の自動変換、本格パーサへの切替は未対応です。
関数・機能の台帳と実テスト結果から、PHP バージョン別の対応状況を JSON／Markdown
に自動集計します。CI のサマリーと成果物から確認でき、未実行や途中失敗は成功扱いにしません。
保存済みの複数実行結果は `gohp-compat-view` で並べて表示し、PHP バージョン・種別・
結果で絞り込めます。過去の PASS と新しい失敗は別々に保持します。
現行 revision への適用判断・複数バージョンをまとめた対応認定・プロジェクト全体の互換性判定は未対応です。
`gohp-inventory` は `composer.json` / `composer.lock` から依存・PHP条件・拡張を
読み取ります。インストールやフック実行はせず、読取成功でも互換性は未確認のままです。
