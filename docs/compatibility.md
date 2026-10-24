# Compatibility inventory (Phase 0)

The [verification-ledger design](compatibility-registry.md) separates implemented
contracts from version-specific test evidence. Its initial
[function/feature catalog](../compatibility/catalog.json) is experimental data linked to
existing checks. The [experimental ledger](compatibility-ledger.md) now validates
it and generates a run-scoped PHP-target report from actual test events.
This table is the current curated summary, not a blanket support claim.

| Area | Current status | Evidence / gap |
|---|---|---|
| PHP 7.4 / 8.3 | Closed shared subset; version-dependent `. / +` precedence | Differential fixtures run against each native version |
| PHP 5.6, other 7.x / 8.x minors | Unsupported target | Native reference matrix needed |
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
