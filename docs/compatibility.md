# Compatibility inventory (Phase 0)

| Area | Current status | Evidence / gap |
|---|---|---|
| PHP 7.4 / 8.3 | Closed shared subset; version-dependent `. / +` precedence | Differential fixtures run against each native version |
| PHP 5.6, other 7.x / 8.x minors | Unsupported target | Native reference matrix needed |
| Scalar values | int64, strings, bool, null | `arrays.php`; no float or interpolation |
| Operators | `+`, `.` and parentheses | Integral coercion only; float promotion unsupported |
| Functions | Named functions, fixed value arguments and returns | `core.php`; no optional/typed/by-reference/variadic args |
| Arrays | Mixed ordered keys/values, reads, nested value copies, foreach | Literal int/string keys; no writes, references or array union |
| Classes | Public scalar defaults, no-arg construction, property reads, methods | `core.php`; no inheritance, constructors, visibility or static members |
| Closures/callbacks/dynamic calls | Unsupported | Negative closure fixture; no claim of coverage |
| Exceptions/warnings | Unsupported PHP behavior | Nonzero runtime failures; never functional PASS |
| Built-ins | `count(array)` only | Everything else fails closed at compile time |
| include/require/autoload/namespaces | Unsupported | Negative include test |
| Composer | No resolver or package support | Never execute Composer hooks to inspect metadata |
| mysqli / PDO / SQLite | Unsupported | No DB writes or credentials |
| mbstring / iconv / cURL / GD | Unsupported | Extension inventory and fixtures needed |
| XML / PCRE / serialization | Unsupported | Go library behavior must not be assumed equivalent |
| WordPress / plugins / themes | Not runnable | Independent adapter and pinned oracle required |

Composer inventory plan: read `composer.json` and `composer.lock` as data only;
record exact source/dist revisions, transitive requirements, `php`, `ext-*`,
platform constraints, generated/vendor code, scripts and licenses. Never run
install/update or arbitrary hooks for metadata inspection. Resolve each requirement
against a machine-readable registry with `unsupported`, `compile-pass`, and
`differential-pass` distinct. The inventory reader and registry are not implemented.

The current harness compares stdout and successful exit status. It does **not**
validate filesystem side effects, DB mutation, PHP exception identity, warnings,
serialization, hook ordering, or arbitrary nonzero PHP outcomes. Runtime-negative
tests verify failure, not equivalent PHP warning behavior. JSON CLI reports record
source/generated hashes and target, with status `generated`, never automatic PASS.
