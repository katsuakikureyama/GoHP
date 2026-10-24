# Compatibility inventory (Phase 0)

The [verification-ledger design](compatibility-registry.md) separates implemented
contracts from version-specific test evidence. Its initial
[function/feature catalog](../compatibility/catalog.json) is experimental data linked to
existing checks. The [experimental ledger](compatibility-ledger.md) now validates
it and generates a run-scoped PHP-target report from actual test events.
This table is the current curated summary, not a blanket support claim.

| Area | Current status | Evidence / gap |
|---|---|---|
| PHP 8.3 | Active closed subset; native oracle is >= 8.3.36 | Differential fixtures run against the approved 8.3 oracle |
| PHP 7.4, PHP 8.4+, other series | Not active targets | Historical evidence is retained; a separate policy/matrix is required before a series is enabled |
| Scalar values | int64, strings, bool, null | `arrays.php`; no float or interpolation |
| Operators | `+`, `.` and parentheses | Integral coercion only; float promotion unsupported |
| Boolean control | Braced one-arm `if`, short-circuit `&&`, scalar `===`, `(bool)` | Scalar/array truthiness only; no else/elseif, array/object identity or other comparison operators; [contract](boolean-boundary.md) |
| Functions | Named functions, fixed value arguments and returns | `core.php`; no optional/typed/by-reference/variadic args |
| Arrays | `[]` and `array(...)`; mixed ordered keys/values, reads, nested value copies, foreach | Literal int/string keys; no writes, references, unpacking or array union; see [syntax boundary](legacy-array-syntax.md) |
| Classes | Public scalar defaults, no-arg construction, property reads, methods | `core.php`; no inheritance, constructors, visibility or static members |
| Closures/callbacks/dynamic calls | Unsupported | Negative closure fixture; no claim of coverage |
| Exceptions/warnings | Unsupported PHP behavior | Nonzero runtime failures; never functional PASS |
| Built-ins | `count(array)`, restricted `rtrim`, `is_bool`, `is_string`, string-only ASCII `strtolower` | Exact arities and closed domains; see [rtrim](rtrim-boundary.md) and [boolean](boolean-boundary.md) boundaries. Other builtins fail closed |
| include/require/autoload/namespaces | Unsupported | Negative include test |
| Composer | No resolver or package support | Never execute Composer hooks to inspect metadata |
| mysqli / PDO / SQLite | Unsupported | No DB writes or credentials |
| mbstring / iconv / cURL / GD | Unsupported | Extension inventory and fixtures needed |
| XML / PCRE / serialization | Unsupported | Go library behavior must not be assumed equivalent |
| WordPress / plugins / themes | Not runnable | Independent adapter and pinned oracle required |

## Active PHP version policy

The sole active compatibility target is the PHP **8.3** language series. Native
differential tests must run against PHP **8.3.36 or newer** in that series. The
compiler target remains the series label `8.3`; a PHP patch version is an oracle
requirement and evidence attribute, not a CLI/compiler target. Each live evidence
run records the actual PHP version. CI also records its pinned image digest;
host-PHP runs have no container image identity unless explicitly supplied.

`internal/phpversion/policy.go` is the compiler-side registry of accepted
series labels and their minimum oracle version. `.github/workflows/ci.yml`,
`Dockerfile`, and `compose.yaml` select the fixed PHP 8.3.36 oracle image; its
resolved digest is asserted by the execution configuration and recorded in run
evidence. Documentation intentionally does not copy or invent that digest: the
resolved configuration and resulting run evidence are authoritative. The
[PHP 8.3.36 release](https://www.php.net/releases/8_3_36.php) and PHP's
[supported-version policy](https://www.php.net/supported-versions.php) are
review inputs for approved image-digest updates and EOL decisions.

PHP 7.4 and older 8.3 evidence remain readable historical records only; they
are not an active test gate or current support claim. PHP 8.4 and later are
currently disabled. To add a later series, introduce a separate explicit
policy/matrix entry with a series name, minimum security floor, resolved
immutable oracle digest, target-specific semantic table/corpus, native
differential job, and reviewed evidence. It must remain unclaimed until those
requirements pass. A version floor alone is not a security guarantee for
GoHP, generated binaries, input, dependencies, or deployment.

Before Go source generation, a semantic pass validates IR operand shape and
rejects statically known mismatches such as `count(1)`, scalar `foreach`/indexing,
and a known class receiver using an absent property, method, or wrong arity.
Values whose type genuinely depends on runtime input remain protected by explicit
runtime failures; validation success is not a differential compatibility result.

Composer inventory plan: read `composer.json` and `composer.lock` as data only;
record exact source/dist revisions, transitive requirements, `php`, `ext-*`,
platform constraints, generated/vendor code, scripts and licenses. Never run
install/update or arbitrary hooks for metadata inspection. Resolve each requirement
against a machine-readable registry with `unsupported`, `compile-pass`, and
`differential-pass` distinct. The inventory reader and historical applicability
decisions are not implemented. Single-run evidence and read-only multi-run
views with target/kind/status filters are available; they do not combine runs
into a support verdict. See the [ledger contract](compatibility-ledger.md).

The current harness compares stdout and successful exit status. It does **not**
validate filesystem side effects, DB mutation, PHP exception identity, warnings,
serialization, hook ordering, or arbitrary nonzero PHP outcomes. Runtime-negative
tests verify failure, not equivalent PHP warning behavior. JSON CLI reports record
source/generated hashes and target, with status `generated`, never automatic PASS.
