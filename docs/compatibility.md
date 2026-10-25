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
| Scalar values | int64, strings, bool, null; literal-only negative decimal constants | `arrays.php`, `negative_integers.php`; no float or interpolation; [negative integer boundary](negative-integer-boundary.md) |
| Operators | `+`, `.`, parentheses and full ternary | Integral coercion only; float promotion unsupported. Ternary is non-associative; no `?:`, `??` or object truthiness; [contract](ternary-stripslashes-boundary.md) |
| Boolean control | Braced `if` / `elseif` / `else if` / `else`, short-circuit `&&` / `\|\|`, scalar `===` / `!==`, `(bool)`, prefix `!`/`!!` | Scalar/array truthiness only; no unbraced/alternative syntax, branch-local escape, object truthiness, array/object identity, loose `!=` or keyword logical operators; [contract](boolean-boundary.md) |
| Functions | Named functions, fixed value arguments, bounded `array` parameters and global `: array` / actual-bool `: bool` returns | `core.php`, `array_parameter.php`, `array_return.php`, `bool_return.php`; [parameters](array-parameter-boundary.md) / [array](array-return-boundary.md) / [bool](bool-return-boundary.md); no weak bool-return coercion, optional/non-array-typed/by-reference/variadic args or typed methods |
| Arrays | `[]` and `array(...)`; mixed ordered keys/values, reads, nested value copies, foreach; one-level element assignment and append to existing array variables | Literal construction keys; literal/variable int/string write keys. No auto-vivification, nested targets, references, unpacking or array union; [syntax](legacy-array-syntax.md), [write boundary](array-write-boundary.md) |
| Classes | Public scalar defaults, no-arg construction, property reads, methods | `core.php`; no inheritance, constructors, visibility or static members |
| Closures/dynamic callbacks | Unsupported | Only literal named one-argument user callbacks through one-array `array_map` are admitted; no general callable support; [contract](array-map-boundary.md) |
| Exceptions/warnings | Unsupported PHP behavior | Nonzero runtime failures; never functional PASS |
| Built-ins | `count(array)`, `strlen(string)`, `array_keys(array)`, `array_values(array)`, `array_key_exists(int\|string, array)`, `array_is_list(array)`, restricted `rtrim`, `is_bool`, `is_string`, `is_array`, ASCII `strtolower(string)`, `stripslashes(string)`, `addslashes(string)`, literal-callback `array_map` | Bounded arities/domains; [rtrim](rtrim-boundary.md), [boolean](boolean-boundary.md), [ternary](ternary-stripslashes-boundary.md), [literal callback](array-map-boundary.md); other builtins fail closed |
| String formatting / replacement | `sprintf('%0Ns', string\|int64)` with canonical width 0–4096; `str_replace(string, string, string)` | [Zero-padding boundary](sprintf-boundary.md) / [replacement boundary](str-replace-boundary.md); no general formatting, array substitutions or count reference |
| Case-insensitive string comparison | `strcasecmp(string, string)` | [ASCII byte comparison](strcasecmp-boundary.md); no non-string coercion or Unicode/locale folding |
| Array-element presence | `isset($array[$key])` on a defined actual array | [Bounded isset](array-isset-boundary.md); one literal/variable int/string key, missing/null false; no undefined roots, nested/property/general isset |
| Strict scalar membership | `in_array(scalar, scalar-valued array, true)` | [Closed strict boundary](in-array-strict-boundary.md); no loose mode or array/object values |
| include/require/autoload/namespaces | Unsupported | Negative include test |
| Composer | No resolver or package support | Never execute Composer hooks to inspect metadata |
| mysqli / PDO / SQLite | Unsupported | No DB writes or credentials |
| mbstring / iconv / cURL / GD | Unsupported | Extension inventory and fixtures needed |
| XML / PCRE / serialization | Unsupported | Go library behavior must not be assumed equivalent |
| WordPress / plugins / themes | Not runnable | Independent adapter and pinned oracle required |

## Combined subset validation

`feature_integration.php` composes both feature branches on one compiler/runtime:
legacy arrays, array keys/values, nested reads, count, guarded string helpers,
type predicates, conditional assignment and PHP 8.3 operator precedence.
The native oracle must admit the shared 8.3.36 floor before any fixture runs.
The catalog links this cross-feature check to `php.builtin.array_values`;
individual feature checks and unsupported-boundary checks remain separate.
This is stdout/normal-exit evidence, not typed native return observation or
validation of Gossip/WordPress source pins. Earlier separate-branch PASS results
do not constitute a combined revision PASS.

`string_array_integration.php` additionally composes the bounded sprintf and
str_replace implementations with the array helpers (including key existence and
list checks), guarded string conversion and byte-length observation. Its single
generated binary is compared with the admitted native PHP 8.3 oracle. Both new
string entries link the same integration check; it is one measured case, not
two independent proofs. This does not update or validate Gossip's source pins.

`predicate_conditional_integration.php` combines `is_array` with full ternary
selection, array_values and guarded stripslashes, preserving empty and nested
arrays and avoiding invalid runtime calls on scalar/null/object alternatives.
Visible markers check single condition evaluation and the selected arm only.
The two catalog links count one check; previous separate-stack evidence does
not validate this integrated revision. See the [integration boundary](gossip-compiler-integration.md).

`callback_composition_integration.php` adds finite literal-callback recursion,
ternary type dispatch, string quoting/unquoting and nested key/value observation.
It preserves the original input and observes callback order and empty-map
non-invocation. This single shared check has the same stdout/normal-exit boundary;
see the [callback contract](array-map-boundary.md), not a general callback or
Gossip adoption claim.

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

### Bounded strcasecmp contract

Exactly two actual strings, ASCII-only case folding and integer byte comparison.
No weak scalar conversion, Unicode folding or locale collation. Known bad
arguments are rejected before generation; unknown values have a runtime guard.
See the [contract and selected PHP implementation](strcasecmp-boundary.md).
Original WordPress adoption requires separate pinned differential evidence.

### Bounded strlen contract

`strlen` accepts one positional string and returns an integer byte length, not a
Unicode character or grapheme count, following the [PHP function contract](https://www.php.net/manual/en/function.strlen.php).
The runtime returns `int64(len(s))` without converting non-strings. Calls are
case-insensitive; redeclaring this builtin as a top-level function is rejected.
Statically known non-string arguments and wrong arity fail before generation.
Unknown values (parameters, function results, array reads) are guarded at runtime;
a non-string value fails explicitly, not as a PHP-compatible warning/exception.

`strlen.php` compares empty, ASCII, whitespace, UTF-8 multibyte/emoji and supported
escaped strings, concatenation, dynamic string values and integer arithmetic
against the active approved PHP 8.3 oracle. Binary/NUL/invalid UTF-8 cases are runtime unit
tests only, not end-to-end PHP evidence. The frontend's string syntax is unchanged:
no interpolation or hex/octal escape support is implied. Scalar/null/object
coercions, named arguments, multibyte overload settings, warnings and exceptions
remain outside this contract. The catalog marks the entry partial and reports
only case-limited differential results after all linked checks pass.

### Bounded array_keys contract

The one-argument `array_keys(array)` form returns the existing integer/string keys
in insertion order, as values in a fresh array with consecutive integer indexes
starting at zero. Empty input returns an empty array. Nested values are not
traversed, and the input is unchanged. This follows the unfiltered
[PHP array_keys contract](https://www.php.net/manual/en/function.array-keys.php).
Key normalization and duplicate-key replacement remain those of the existing
GoHP array constructor; the builtin does not convert keys a second time.

Calls are case-insensitive; top-level builtin redeclaration, non-array inputs
known at generation time, and any argument count other than one are rejected.
Unknown values are guarded at runtime. Return type is known to be an array, so
it works with count, foreach and indexing, but not scalar echo. No filter-value
or strict-search overload, named arguments, dynamic function names, new array
write/reference behavior or PHP-compatible warning/exception behavior is added.

`array_keys.php` covers empty, mixed and normalized keys, duplicates/order,
nested values, dynamic inputs/results and composition using the existing native
PHP-vs-generated-Go stdout/exit comparison. Runtime tests additionally check
exact key/index types and storage isolation; binary keys are runtime-only.
The catalog stays partial, with a case-limited result requiring every linked
positive and boundary check to pass, not a general PHP array compatibility claim.

### Bounded array_values contract

`array_values(array)` accepts exactly one positional array argument and returns
the values in insertion order with new integer indexes starting at zero, following
the [PHP function contract](https://www.php.net/manual/en/function.array-values.php).
It does not sort by the old keys, flatten nested arrays, convert scalar values,
or modify its input. Empty input returns an empty array. Nested array values keep
their keys and use GoHP's existing eager value-copy model. This is not a new
implementation of PHP references, array writes or copy-on-write behavior.

Calls are case-insensitive; builtin redeclaration, wrong arity and statically
known non-array arguments fail before generation. Unknown arguments are guarded
at runtime. The known array return works with count, indexing and foreach, but
cannot be echoed as a scalar. Named arguments, dynamic function names, general
object/reference alias semantics and PHP warning/exception equivalence remain
outside the validated contract.

`array_values.php` compares empty arrays, mixed/sparse keys, duplicate-key
replacement, scalar/nested values, dynamic inputs/results and composition against
the pinned PHP 8.3.36 oracle. The native comparison observes stdout and successful
exits, not a typed serialization. Runtime tests additionally assert exact value
types and outer/nested storage independence. The catalog stays partial and only
reports a case-limited result when every linked check passes.

### Bounded array_key_exists contract

`array_key_exists(key, array)` accepts exactly two positional arguments. The key
must be an integer or string; the second argument must be an array. It returns
a boolean for presence in the first dimension, even when the stored value is
null, false, zero or an empty string. Missing keys return false without reading
their value or raising the missing-index guard. Nested keys are not searched
recursively. These presence rules follow the [PHP reference](https://www.php.net/manual/en/function.array-key-exists.php).

Key normalization uses the same implementation as array construction and reads:
canonical decimal strings such as `'2'` address integer keys, while `'02'` and
`'+2'` remain distinct strings. Input arrays and their values are not modified.
Both array syntaxes and dynamic function parameters/returns are supported within
this domain. The builtin is case-insensitive and reserved against redeclaration.

Unlike native PHP 8.3, this bounded implementation rejects bool/null/float/resource
keys instead of coercing them. Known invalid types/arity are rejected before
generation; unknown values have explicit runtime guards for both arguments.
Object inputs, the `key_exists` alias, named arguments, references and warning/
exception equivalence are not implemented by this addition.

`array_key_exists.php` compares native stdout/normal exits, with `is_bool` and
strict true/false observations as well as guarded reads. Runtime unit checks
cover nonmutation and additional binary/int64-boundary keys; these additional
unit inputs are not end-to-end native evidence. Generated-binary negative tests
cover both dynamic argument guards. The catalog remains partial.

### Bounded array_is_list contract

`array_is_list(array)` takes exactly one positional array argument and returns
a boolean. The empty array is a list. Otherwise, outer keys must be integers
`0, 1, ...` in insertion order, as specified by the [PHP reference](https://www.php.net/manual/en/function.array-is-list.php).
The same key set in a different order, a nonzero starting key, a gap or a string
key makes the result false. No sorting, reindexing or key coercion happens here;
numeric strings already normalized by array construction remain integer keys.
Duplicate-key replacement retains its existing insertion position.

Values do not affect the result: null/false values and nested associative arrays
may be elements of a list. Only the outer keys are inspected; neither input nor
nested values are mutated. Calls are case-insensitive, and the builtin name is
reserved against redeclaration. Known non-arrays and wrong arity fail before
generation; unknown values are guarded at runtime. The result is statically bool.

`array_is_list.php` checks these boundaries, both array syntaxes, dynamic calls,
and composition with `array_keys`/`array_values`, using `is_bool` and strict
true/false observations. Native evidence is still stdout/normal exits; additional
negative/extreme/binary-key cases in runtime units are not native PHP evidence.
The catalog stays partial because the compiler's array construction/value domain
is bounded; this does not add references, writes, named arguments, recursive
validation, or PHP warning/exception equivalence.

Composer metadata inventory is available below. It records declarations only;
PHP callsites, generated/vendor code, dependency resolution and historical applicability
decisions are not implemented. Single-run evidence and read-only multi-run
views with target/kind/status filters are available; they do not combine runs
into a support verdict. See the [ledger contract](compatibility-ledger.md).

The current harness compares stdout and successful exit status. It does **not**
validate filesystem side effects, DB mutation, PHP exception identity, warnings,
serialization, hook ordering, or arbitrary nonzero PHP outcomes. Runtime-negative
tests verify failure, not equivalent PHP warning behavior. JSON CLI reports record
source/generated hashes and target, with status `generated`, never automatic PASS.

## Composer metadata inventory

```sh
go run ./cmd/gohp-inventory /path/to/project
go run ./cmd/gohp-inventory -format json /path/to/project
```

`gohp-inventory` reads only `composer.json` and optional `composer.lock`, then
prints Markdown or `gohp.composer.inventory/v0.1` JSON. The program never invokes
PHP, Composer, plugins, scripts or autoload files; it does not fetch repositories,
install packages, inspect `vendor`, write reports or change input files. A
prebuilt binary does not need Go/PHP/Composer. `go run` itself builds the trusted
tool and may populate Go's build cache.

The report separates root declarations, production/development lock groups,
package-link relations, and lock platform declarations. In Composer,
`require-dev` of a dependency is not a root development requirement, and
`suggest`, `provide`, `replace` and `conflict` are not `require`. Source/dist
reference strings and declared licenses are retained, without claiming that a
reference is immutable or that a license has been approved. See the official
[manifest schema](https://getcomposer.org/doc/04-schema.md).

`php`/`php-*`, `ext-*`, `lib-*` and Composer API requirements are classified
separately from package names; unrecognized names remain unknown. Version
constraints are preserved as strings, not evaluated. Configured platform values
(including `false`) are separate from actual runtime observations: no runtime is
probed. See [Composer platform dependencies](https://getcomposer.org/doc/articles/composer-platform-dependencies.md).

A lock records selected versions, but the reader does not resolve a dependency
graph, apply virtual providers/aliases, or check that the lock matches the current
manifest. Its recorded `content-hash` is **not** treated as freshness proof.
Installed package availability is also unknown. See
[Composer's lock usage](https://getcomposer.org/doc/01-basic-usage.md#installing-from-composer-lock)
and [lock schema](https://github.com/composer/composer/blob/main/res/composer-lock-schema.json).

### Output and error contract

- Overall status is always `metadata-only`, compatibility always `unverified`.
  Every listed dependency stays `unverified`, even if its name appears in a
  lock or GoHP's feature catalog. No PHP/extension/application coverage is added.
- Input filenames, presence and exact-byte SHA-256 hashes identify what was read.
  Missing lock is `absent`, freshness `not-available`, with no inferred packages.
  A present lock has freshness `unverified`; malformed/inaccessible lock is an
  error, never quietly treated as absent.
- Root and locked packages retain names, declared versions/types/licenses and
  source/dist type/reference/shasum. URL fields and repository metadata are omitted
  to avoid echoing common credential-bearing locations. Script command bodies and
  autoload mappings are omitted; only event/mode names are listed. This is not a
  general secret scanner: other user-supplied strings can still be sensitive.
- Dependency rows retain owner, scope, relation, name and raw constraint (or
  description for `suggest`). Development declarations are not folded into
  production requirements. Missing source/dist/license fields remain unknown.
- This is a partial metadata reader, not full schema validation. Unknown fields
  are not interpreted. Selected fields must have the documented object/string/
  array types; duplicate JSON keys at any depth are rejected. Current lock input
  requires `content-hash`, `packages` and `packages-dev`. Legacy array-as-empty-map
  encodings are not accepted. No version/constraint/SPDX validation is performed.
- Each file is limited to 16 MiB and JSON nesting to 64 levels. Files must be
  regular files within the resolved project directory; escaping or dangling
  symlinks fail. Use a private directory without concurrent mutation. The two
  reads are not a snapshot transaction or a hostile-filesystem sandbox.
- Exit 0 means metadata was read, not that it is compatible or installable.
  Usage errors exit 2; read/validation/output errors exit 1. No stdout is written
  until input validation completes; output-stream errors can still leave partial
  stdout. Constraints, provenance and notices remain visible in both formats.

### Local inventory capability

Stable ID: `gohp.inventory.composer`. Owner/resource: local GoHP development tool /
caller-readable Composer metadata. Input schema:

```json
{"type":"object","properties":{"project":{"type":"string","minLength":1},"format":{"enum":["markdown","json"]}},"required":["project"],"additionalProperties":false}
```

`inspect`: read and validate fixed metadata files; `prepare`: derive inventory
in memory; `save` / `commit`: none. Result is the versioned metadata-only report
(input hashes, root/packages/links/overrides, lock freshness, notices), or an
error. Local filesystem read permissions are the authority. Hashes identify
inputs, not authenticity; callers retain reports for audit if desired.
WebMCP policy: **excluded**, because this is a local development reader with no
browser or server operation. No session forwarding, MCP route, external message,
publication or auto-submit is introduced.
