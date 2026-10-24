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
docker compose --profile legacy build php74
docker compose --profile legacy run --rm php74
```

Or use Go **1.23.12** and PHP **8.3** or **7.4** locally:

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

## Scope and documentation

- [Architecture and decisions](docs/architecture.md)
- [Compatibility and version matrix](docs/compatibility.md)
- [Upstream assessment](docs/upstream.md)
- [Gossip integration contract](docs/integration.md)
- [Development and roadmap](CONTRIBUTING.md)
- [Security boundaries](SECURITY.md)
- [Licensing proposal](docs/licensing.md) — no final project license yet

## 日本語概要

PHP を読みやすい Go ソースへ変換する初期 PoC です。関数・配列・クラスを
含む fixture を実際にビルドし、PHP との標準出力の一致を確認します。
参照、クロージャ、継承、Composer や PHP 拡張、WordPress 全体は未対応です。
未対応を互換性成功とは扱いません。正式ライセンスは監査と所有者判断待ちです。
