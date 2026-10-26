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
and PHP 8.3 operator precedence. The `legacy_arrays.php` fixture
also covers `array(...)` syntax, empty returns, mixed nesting with `[]`, trailing
commas and case-insensitive `array` spelling within the same closed array subset.
Automatic literal keys now follow PHP 8.3's negative-integer progression, also
after string-only entries; see the [array index boundary](docs/array-index-boundary.md).
Negative decimal constants such as `-5` can now be written directly; general
numeric negation and float-promoting tokens remain outside the [bounded syntax](docs/negative-integer-boundary.md).
Unsupported syntax must fail
with a diagnostic; a separate runtime-negative test must fail on a missing key.
Compilation alone never proves PHP compatibility.
Declared public properties now accept bounded [simple assignment](docs/public-property-write.md)
through a statically known generated-class variable (including known aliases
and method `$this`). Arrays detach at write/read boundaries; object identifiers
retain identity. This adds neither dynamic properties nor unknown receiver writes.
Simple data classes also support [fixed public constructors and known-object
array snapshots](docs/block-data-constructor.md), including implicit-null public
properties and trailing untyped `= null` constructor defaults. No inheritance,
typed/non-null-default constructors, ordinary method defaults or general array casts.
Bounded [`while (condition) { body }`](docs/while-boundary.md) re-evaluates its
condition before each iteration. Bare [`continue;`](docs/foreach-continue-boundary.md)
and [`break;`](docs/foreach-break-boundary.md) target the nearest enclosing
`foreach` or `while`, including inside braced conditionals. Continue rechecks
the while condition; break exits without rechecking it. Numeric levels,
unbraced/alternative syntax and for/do/switch are excluded.
Supported functions/builtins, bare-variable methods and generated arrows may
also be used as [standalone calls](docs/call-statement-boundary.md), preserving
effects and guards while discarding the return value. General expression
statements and chained call receivers remain excluded.
The [form-feed escape](docs/form-feed-boundary.md) `"\f"` decodes to byte 0x0C;
single-quoted `'\f'` stays literal. Other previously unsupported escapes and
interpolation remain excluded.
Together with bounded
[trailing untyped null defaults and defined bare/multiple isset](docs/array-get-prerequisites.md),
this resolves the syntax blockers measured for Gossip's `_wp_array_get`.
The subsequent [literal-null identity fix](docs/literal-null-identity.md) and
Gossip's separately pinned comparison are recorded in the [integration contract](docs/integration.md).
Builtins currently include `count(array)`, bounded `strlen(string)` (byte
length only, no implicit non-string coercion), `array_keys(array)` (one argument,
ordered keys, no value filtering) and `array_values(array)` (ordered values,
zero-based indexes, no flattening), literal-mask `rtrim`, `is_bool`, `is_string`, `is_int`, `is_null`, `is_array`
and ASCII `strtolower(string)`. [`is_object(value)`](docs/is-object-boundary.md)
recognizes generated class instances and generated arrows (PHP Closures), not
arrays/scalars or arbitrary foreign Go values. String-only `strcasecmp(string, string)` adds
binary-safe ASCII case-insensitive comparison with an integer result; no
coercion, Unicode folding or locale collation. See the
[comparison boundary](docs/strcasecmp-boundary.md).
[`str_contains(string, string): bool`](docs/str-contains-boundary.md) checks
case-sensitive byte substring membership, including an empty needle. Actual
strings only; no coercion, Unicode folding or token-boundary interpretation.
[`str_starts_with(string, string): bool`](docs/str-starts-with-boundary.md) checks
the byte prefix instead: matches only in the middle or at the end are false.
It has the same actual-string and empty-needle boundaries.
Bounded [`strtok`](docs/strtok-boundary.md) supports string initialization and
continuation with byte delimiters. Its cursor belongs to an explicit execution,
shared across nested calls but isolated across separate host entries. Empty
delimiter sets are allowed; uninitialized continuation and coercion are excluded.
This does not certify the complete WordPress render/HTML path.
[`array_key_first(array)`](docs/array-key-first-boundary.md) and
[`array_key_last(array)`](docs/array-key-last-boundary.md) return the first/last
insertion-order int/string key, or null for an empty array, without modifying it.
[`array_reverse(array, bool preserve_keys = false)`](docs/array-reverse-boundary.md)
reverses outer order without modifying the input, retaining string keys and
optionally integer keys; the optional flag accepts actual booleans only.
[`array_merge(array ...arrays)`](docs/array-merge-boundary.md) joins explicitly
supplied arrays, reindexing integer keys and replacing duplicate string-keyed
values in place; zero arguments return an empty array. No argument unpacking.
[`array_slice(array, int offset, ?int length = null, bool preserve_keys = false)`](docs/array-slice-boundary.md)
selects by position with positive/negative bounds, keeping string keys and
optionally integer keys. Actual integer/bool arguments only, without coercion.
`array_key_exists(int|string, array)` checks
first-level key presence, including null-valued entries, without other key
coercions. `array_is_list(array)` checks zero-based consecutive integer keys in
insertion order (empty arrays are lists). See the compatibility contract and the linked
string/boolean boundaries; these are partial implementations.
Prefix logical negation `!` (including `!!`) uses the existing scalar/array
truthiness rules and returns bool; object truthiness and loose `!=` remain excluded.
Strict `!==` compares only null, bool, int64 and string by both type and value,
returning the opposite of scalar `===` without coercion. Unparenthesized chains
mixing `===` and `!==` are rejected. A literal `null` operand additionally lowers
to a once-only null type probe, including against arrays/objects/callables;
general array/object comparison remains excluded.
Logical `||` uses the same truthiness domain and returns bool, skipping its right
operand when the left is true. It binds below `&&` and above full ternary;
keyword `or` and bitwise `|` remain unsupported. See the [boolean boundary](docs/boolean-boundary.md).
Braced `if` accepts `elseif` / `else if` chains and an optional braced `else`,
including nested/empty bodies and early returns. Reached conditions run once in
order; the first truthy branch prevents evaluation of later conditions/bodies.
Each arm is validated independently; variables introduced
inside branches do not become available after the if, even if both arms assign
them. Unbraced bodies, alternative syntax and path-sensitive inference remain excluded.
The integrated string subset also includes `sprintf('%0Ns', string|int64)` for
canonical widths 0–4096 and `str_replace(string, string, string)` only. General
formatting, array substitutions, coercions and by-reference count are excluded.
The callback subset admits `array_map('declared_function', array)` with one
array and one literal user function whose required arity is one (trailing
untyped null defaults are allowed), including finite recursion.
String-only `addslashes` is byte quoting, not SQL/HTML/shell sanitization.
One-level `$array[$key] = $value` and `$array[] = $value` statements update
existing array variables with detached value copies. Explicit keys must be
int/string literals or variables; append uses PHP 8.3 integer-index progression.
Nested targets, auto-vivification and references remain excluded.
Bounded `empty($arrayVariable)` tests only a defined actual array's outer entry
count, without inspecting values; undefined/scalar/index/property operands remain
excluded. See the [empty boundary](docs/array-empty-boundary.md).
One-level `isset($array[$key])` retains its strict array/key guard without
reading a missing element. Defined bare variables and a comma-separated list of
defined bare/indexed addresses are evaluated left-to-right with short-circuiting.
The separate bounded nested-path and `??` form accepts only a defined variable
root plus literal/defined-variable keys, keeps normal reads strict, and does not
support properties or undefined suppression. See the [array isset boundary](docs/array-isset-boundary.md)
and [nested array-probe boundary](docs/nested-array-probes.md).
Named global functions may declare `: array` or `: bool`; explicit returns and fallthrough
are checked rather than silently dropping the declaration. Bool returns require
actual bool values, without weak scalar coercion. Fixed positional `array`
parameters are checked before the body and receive detached value copies. Named
global functions may additionally use trailing untyped `= null` defaults; typed
or method defaults remain excluded. See the
[array parameter boundary](docs/array-parameter-boundary.md). Capture-free
`fn($value, $key) => expression` arrows can be passed through global `callable`
parameters and invoked through variables; [this bounded contract](docs/arrow-callable-boundary.md)
does not include string/method callables, captures or typed arrow signatures.
Other parameter and return types and typed methods remain excluded. See the
[array return boundary](docs/array-return-boundary.md) and
[bool return boundary](docs/bool-return-boundary.md).
Strict scalar `in_array(needle, haystack, true)` is also available: the needle
and all array values must be null/bool/int64/string. Loose comparison and
array/object values are excluded; see the [membership boundary](docs/in-array-strict-boundary.md).

The predefined `true`, `false` and `null` literals accept uppercase and mixed
case (for example `TRUE`, `FaLsE`, `NuLl`). Variable/property names and string
contents remain case-sensitive; this does not imply general keyword support.

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
- [Legacy array syntax boundary](docs/legacy-array-syntax.md)
- [Literal-mask rtrim boundary](docs/rtrim-boundary.md)
- [Bounded `sprintf` zero-padding boundary](docs/sprintf-boundary.md)
- [Bounded `str_replace` string boundary](docs/str-replace-boundary.md)
- [Full ternary and string-only `stripslashes` boundary](docs/ternary-stripslashes-boundary.md)
- [Combined conditional/array compiler boundary](docs/gossip-compiler-integration.md)
- [Literal callback mapping and string quoting boundary](docs/array-map-boundary.md)
- [One-level array assignment and lvalue boundary](docs/array-write-boundary.md)
- [Nested array probes and null coalescing boundary](docs/nested-array-probes.md)
- [Boolean validation subset](docs/boolean-boundary.md)
- [String-only ASCII comparison](docs/strcasecmp-boundary.md)
- [Combined charset and typed-array compiler](docs/charset-array-integration.md)
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
配列は `[]` と `array(...)` を同じ限定処理へ変換し、空配列や入れ子を比較対象にします。
`rtrim` は文字列入力＋明示的な固定文字一覧（範囲指定なし）の2引数だけを対象にします。
`sprintf` は `%0Ns`（N は 0〜4096 の正規十進幅）の文字列／整数値だけを、byte 幅の左ゼロ埋めとして対象にします。
`str_replace` は3つの文字列だけを、byte 単位・case-sensitive・non-overlap の置換として対象にします。
full ternary `condition ? yes : no` と `stripslashes` の文字列1引数も限定的に対象にします。
`array_map`は宣言済み1引数関数のliteral名＋単一配列のみ、`addslashes`は実文字列のみです。
一般callback・自動再帰・SQL/HTML/shell向けサニタイズの対応ではありません。
波括弧付きの `if` / `elseif` / `else if` / `else`、短絡 `&&` / `||`、論理否定 `!`、scalar の `===` / `!==`、限定型の `(bool)` と型判定・ASCII小文字化も比較対象にします。
参照、クロージャ、継承、Composer や PHP 拡張、WordPress 全体は未対応です。
未対応を互換性成功とは扱いません。正式ライセンスは監査と所有者判断待ちです。
初期の互換性対象は PHP **8.3 系のみ**で、native 比較には **8.3 系の 8.3.36 以上**を使います。
CLI の `-php-version 8.3` はパッチ番号ではなく言語系列の指定です。8.4 以降は未対応で、
追加時には系列ごとの最低版・固定oracleイメージdigest・専用fixture・検証証拠を別途そろえます。
最低版の指定だけで、コンパイラや生成物のセキュリティを保証するものではありません。
組込み関数は `count(array)`、限定的な `strlen(string)`・`array_keys(array)`・`array_values(array)`、
固定maskの `rtrim`、`is_bool`・`is_string`・`is_int`・`is_null`・`is_array`・ASCIIの `strtolower(string)` を対象にします。
`strlen` はバイト数を返し、数値・nullなどからの暗黙変換は対象外です。
`array_keys` は1引数で順序を保ったキー一覧を返します。値による絞り込みは未対応です。
`array_values` は順序を保った値一覧を0始まりに振り直し、入れ子の配列は平坦化しません。
`array_key_exists` は整数/文字列キーの存在を判定します。値がnullでも存在と判定し、
入れ子の再帰検索やbool/nullキーの暗黙変換はしません。
`array_is_list` は外側のキーが0始まりの連続整数かを順序も含めて判定します。
空配列はtrue、値や入れ子のキーは判定に影響せず、配列を変更しません。
生成レポートには Go の行から PHP の文開始行への対応を含みます。
列単位の位置情報やエラー表示の自動変換、本格パーサへの切替は未対応です。
関数・機能の台帳と実テスト結果から、PHP バージョン別の対応状況を JSON／Markdown
に自動集計します。CI のサマリーと成果物から確認でき、未実行や途中失敗は成功扱いにしません。
保存済みの複数実行結果は `gohp-compat-view` で並べて表示し、PHP バージョン・種別・
結果で絞り込めます。過去の PASS と新しい失敗は別々に保持します。
現行 revision への適用判断・複数バージョンをまとめた対応認定・プロジェクト全体の互換性判定は未対応です。
`gohp-inventory` は `composer.json` / `composer.lock` から依存・PHP条件・拡張を
読み取ります。インストールやフック実行はせず、読取成功でも互換性は未確認のままです。
