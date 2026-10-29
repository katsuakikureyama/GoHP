# Compatibility inventory (Phase 0)

The [verification-ledger design](compatibility-registry.md) separates implemented
contracts from version-specific test evidence. Its initial
[function/feature catalog](../compatibility/catalog.json) is experimental data linked to
existing checks. The [experimental ledger](compatibility-ledger.md) now validates
it and generates a run-scoped PHP-target report from actual test events.
This is the curated, navigable summary of the closed subset, not a blanket
support claim. Start with the project overview in the [README](../README.md)
and use the [roadmap](roadmap.md) for the next priorities; this page records the
current contracts and their evidence boundaries.

## Active targets and evidence

| Target | Current status | Evidence boundary |
|---|---|---|
| PHP 8.3 | Active closed subset; native oracle is >= 8.3.36 | Independently run the bounded differential corpus against the approved 8.3 oracle. |
| PHP 8.4 | Active closed subset; native oracle is >= 8.4.27 | Independently run the bounded corpus against the pinned 8.4 oracle; it is separate from 8.3 evidence. [`array_any`](array-any-boundary.md) and [`array_all`](array-all-boundary.md) are 8.4-only bounded builtins. |
| PHP 7.4, PHP 8.5+, other series | Not active targets | Historical evidence remains readable; enabling a series needs its own policy, matrix and evidence. |

## Syntax and control flow

| Area | Generated-code support | Boundary |
|---|---|---|
| Integer subtraction | Actual-integer binary `-`, with int64 overflow rejection | No scalar coercion, float promotion or general unary negation. See [subtraction](integer-subtraction.md). |
| Literals and target constants | `int64`, strings, bool, null and literal negative decimal constants; exact unqualified `PHP_MAJOR_VERSION` / `PHP_MINOR_VERSION` resolve from compiler target `8.3` or `8.4`; double-quoted `"\f"` is byte `0x0C`; single-quoted backslashes preserve literal bytes except escaped quote/backslash | No float, interpolation, general constants/declarations/dynamic lookup/qualified names, or general double-quoted escape support. See [negative integers](negative-integer-boundary.md), [target constants](target-version-constants.md), [form feed](form-feed-boundary.md), and [single-quoted backslashes](single-quoted-backslashes.md). |
| Operators and conditionals | `+`, `.`, parentheses, full ternary, bounded `??` array addresses; braced `if` / `elseif` / `else if` / `else`; `&&`, `\|\|`, scalar `===` / `!==`, literal-null type probes, `(bool)`, `!` / `!!` | Integral coercion only; ternary is non-associative. `??` needs a defined bare/nested-array address and lazy fallback. No float promotion, `?:`, `??=`, general expression left sides, loose `!=`, keyword logical operators, unbraced/alternative syntax, branch-local escape, object truthiness or general array/object identity. See [ternary/stripslashes](ternary-stripslashes-boundary.md), [nested probes](nested-array-probes.md), and [boolean control](boolean-boundary.md). |
| Arrays and probing | `[]` / `array(...)`, mixed ordered keys/values, reads, nested copies, `foreach`, one-level assignment/append; `empty($definedVariable)` on actual arrays or strings; bounded one-level/nested/bare/multi-address `isset` | No auto-vivification, nested targets, references, unpacking or union. Empty arrays use outer count; strings are empty iff exact `""` or `"0"`. No undefined/address suppression or other scalar domains. Nested `isset` needs a defined root and array/null intermediates. See [legacy syntax](legacy-array-syntax.md), [writes](array-write-boundary.md), [isset](array-isset-boundary.md), [nested probes](nested-array-probes.md), and [empty](array-empty-boundary.md). |
| Loops and calls | Braced `while`; bare nearest-loop `break;` / `continue;` in `foreach` or `while`; direct function/builtin and bare-variable method/arrow calls as statements | `while` re-evaluates its condition and uses conservative loop-carried type guards. No levels, `for`/`do`/`switch`, alternative syntax, new binding escape, general expression statements or chained call receivers. See [while](while-boundary.md), [continue](foreach-continue-boundary.md), [break](foreach-break-boundary.md), and [call statements](call-statement-boundary.md). |

## Built-in functions

[Nonnumeric string switch](string-switch.md) supports literal labels, optional
default, first matching duplicate, fallthrough and nearest break. Numeric values,
dynamic labels and switch-target continue are outside its admitted domain.

[Positional list assignment](positional-list-assignment.md) admits a nonempty
bare-local target list and actual-array RHS with all numeric keys present.
This statement is not general destructuring, reference binding or warning handling.

The additional [relational slice](relational-int-null.md) admits `<` and `>`
for actual integers/null only, with PHP null-to-boolean comparison rules and
non-associative precedence. It does not admit general mixed-type ordering or
make the unchanged full WordPress parser executable.

Every entry below is partial: its linked boundary fixes arity, types, evaluation
and evidence. Other builtins fail closed.

| Group | Bounded support | Boundary |
|---|---|---|
| Array mutation | `array_push` on an existing bare local or direct declared named property, standalone or whole RHS assigned to a bare result local | Zero or more explicit values, left-to-right snapshots, current post-argument array and actual int count. No embedded calls, indexed/nested/dynamic target, autovivification, named/unpacked/reference arguments or general by-reference API. See [array writes/push](array-write-boundary.md#bounded-array_push-calls). |
| Array inspection and transformation | `count(array)`, `array_keys(array)`, `array_values(array)`, `array_key_exists(int\|string, array)`, `array_is_list(array)`, first/last insertion-order key, reversal, merge and positional slice via `array_key_first`, `array_key_last`, `array_reverse`, `array_merge`, `array_slice` | First/last yield an int/string key or null without mutation; reverse/slice return detached arrays; merge accepts zero or more explicit arrays, reindexes integer keys and replaces string keys in place. No implicit coercion, recursive operation or unpack syntax. See [keys](#bounded-array_keys-contract), [values](#bounded-array_values-contract), [key existence](#bounded-array_key_exists-contract), [list](#bounded-array_is_list-contract), [first](array-key-first-boundary.md), [last](array-key-last-boundary.md), [reverse](array-reverse-boundary.md), [merge](array-merge-boundary.md), and [slice](array-slice-boundary.md). |
| Predicates and membership | `is_bool`, `is_string`, `is_int`, `is_null`, `is_array`, `is_object`, `gettype(value)`, and strict `in_array(scalar, scalar-valued array, true)` | `gettype` covers only existing null/bool/int/string/array/generated-object values; `is_object` recognizes generated classes/arrows. No float/resources/foreign hosts/general reflection, object truthiness, loose membership or array/object membership values. See [gettype](gettype-boundary.md), [boolean predicates](boolean-boundary.md), [is_int](is-int-boundary.md), [is_null](is-null-boundary.md), [is_object](is-object-boundary.md), and [strict membership](in-array-strict-boundary.md). |
| Strings and formatting | `strlen(string)`, literal-mask `rtrim`, ASCII `strtolower`, `strcasecmp(string, string)`, `str_contains(string, string)`, `str_starts_with(string, string)`, string-only `stripslashes` / `addslashes`, `sprintf('%0Ns', string\|int64)` (width 0–4096), `str_replace(string, string, string)`, byte `substr`, and `strtok(string, delimiters)` / initialized `strtok(delimiters)` | Actual-string domains unless specified; `strtok` uses byte delimiter sets, string/false results and an execution-owned cursor. No Unicode/locale folding, general formatting, array substitutions, by-reference count, general coercion, warning equivalence or SQL/HTML/shell-safety claim. See [strlen](#bounded-strlen-contract), [rtrim](rtrim-boundary.md), [strcasecmp](strcasecmp-boundary.md), [contains](str-contains-boundary.md), [starts-with](str-starts-with-boundary.md), [ternary/stripslashes](ternary-stripslashes-boundary.md), [sprintf](sprintf-boundary.md), [replace](str-replace-boundary.md), [parser output](parser-output-boundary.md), and [strtok](strtok-boundary.md). |
| Callback mapping | `array_map('declared_function', array)` | One literal named user callback and one array, with required arity one; no dynamic/builtin/method/arrow callbacks or references. See [array_map](array-map-boundary.md). |
| PHP 8.4 predicates | `array_any(array, 'declared_name')` / `array_all(array, 'declared_name')` | Literal callback with exactly two required untyped value parameters and explicit actual-bool return. No optional/typed/reference parameters, dynamic/builtin/method/arrow callbacks or truthiness conversion. PHP 8.3 retains ordinary user declarations. See [array_any](array-any-boundary.md) / [array_all](array-all-boundary.md). |

## Values and types

| Area | Generated-code support | Boundary |
|---|---|---|
| Scalars and arrays | int64, byte strings, bool and null; ordered arrays with numeric-string key normalization and independent nested value copies on assignment, arguments and returns | No float/resource values, general key coercion or reference-containing arrays. Shared storage is generated only in the proved scalar-entry lane below; no general COW/performance claim. Objects preserve pointer identity. See [array keys/indexes](array-index-boundary.md) and [value model](reference-value-model.md). |
| Functions and returns | Named globals with fixed value arguments, exact `int` / `string` / `bool`, `array` / generated-arrow `callable` parameters, global `: array` and actual-bool `: bool` returns, and trailing untyped `= null` defaults | No weak bool-return conversion, typed/non-null defaults, by-reference/variadic arguments or typed methods. See [scalar parameters](scalar-function-parameters.md), [array parameters](array-parameter-boundary.md), [array returns](array-return-boundary.md), [bool returns](bool-return-boundary.md), and [parser output](parser-output-boundary.md). |
| Storage/runtime ownership foundation | Selected scalar functions use ScalarFrame; proved flat-array functions with no parameters or all exact int/string/bool parameters, including if joins and domain-invariant while loops, use ValueSlot/ValueTemporary with operation-scoped cleanup and detached returns | No PHP references. Unproved loops, array/callable/untyped parameters, foreach, nested values and unlisted calls keep legacy array scopes; existing runtime.Array APIs are unchanged. See [reference/value stages](reference-value-model.md), [generated ownership plan](owned-array-slot-values.md#first-generated-scope-and-expression-ownership), and [generated scalar frames](generated-scalar-frames.md). |
| Parser-output prerequisites | Public method null defaults/bare returns, declared array-property append, sole-new method factory snapshots and byte `substr` | A closed parser-output boundary only: no general casts, null-array auto-vivification, nested writes or full parser claim. See [parser output](parser-output-boundary.md). |

## Objects and callables

| Area | Generated-code support | Boundary |
|---|---|---|
| Classes and properties | Public scalar/null defaults and implicit null; bounded public constructors; property reads/methods, declared assignment/array append on known or runtime-guarded unknown bare locals, public-object `(array)` snapshots on known or runtime-guarded unknown expressions | No dynamic properties/names, inheritance, typed/non-null defaults, weak scalar conversion, nonpublic/static members or general casts. Runtime object protocols are explicit host opt-in, not provenance checks or PHP exception equivalence. See [constructors/snapshots](block-data-constructor.md), [constructor scalars](scalar-constructor-boundary.md), and [property writes](public-property-write.md). |
| Nested property updates | Named chains rooted in bare locals, with `=`, actual-string `.=` and actual-array `[] =`; guarded final-owner identity and value copies | Two or more properties; nonnull declared object paths only. No array-index paths, dynamic/effectful steps or autovivification. See [nested updates](public-property-write.md#named-nested-property-updates). |
| Property presence | `isset` of named property paths rooted in defined bare locals; missing/null/nonobject is false, nonnull falsy slots are set; mixed multi-address short circuit | No undefined roots, magic/typed/inherited members, dynamic/indexed/effectful paths or property coalescing. Ordinary reads remain strict. See [property isset](property-isset.md). |
| Declared class parameters | Exact nonnull generated objects on named functions, public methods and constructors; case-insensitive forward declarations and identity-preserving mutation | No inheritance, namespaces, nullable/union/reference parameters or typed defaults. Primitive typed ordinary methods remain excluded. See [class parameters](declared-class-parameters.md). |
| Arrows and callable values | Capture-free arrows, generated-arrow `callable` arguments and variable invocation | No anonymous `function`, captures, string/method callables or typed arrows. See [arrow callables](arrow-callable-boundary.md); `array_map` has its separate [literal callback contract](array-map-boundary.md). |

## Application and extension exclusions

| Area | Status / boundary |
|---|---|
| References, exceptions and warnings | PHP references are unsupported; warnings/exceptions are not modeled. Runtime failures are nonzero failures, never functional PASS. |
| Source loading and dependencies | Explicit [static declaration closure](static-source-loading.md) admits fixed top-level `require_once __DIR__` edges over a checked manifest; dependencies have no top-level effects. General `include` / `require` / autoload / namespaces remain unsupported. Composer has no resolver/package support; metadata inspection never executes hooks. |
| Database and extensions | `mysqli`, PDO, SQLite, `mbstring`, `iconv`, cURL, GD, XML, general PCRE and serialization are unsupported. The [optional PCRE2 bridge](optional-pcre2-bridge.md) and [PHP8.3 adapter](preg83-value-wrapper.md) have a separately opt-in [bounded local-output compiler experiment](preg83-local-output.md), not default or PHP8.4 `preg_*` support. Do not assume a Go library is PHP-equivalent. |
| WordPress applications | WordPress, plugins and themes are not runnable; adoption needs an independent adapter and pinned oracle. |

## Combined subset validation

`feature_integration.php` composes both feature branches on one compiler/runtime:
legacy arrays, array keys/values, nested reads, count, guarded string helpers,
type predicates, conditional assignment and the shared target-profile operator
precedence. The native oracle must admit the selected target's floor before any
fixture runs.
The catalog links this cross-feature check to `php.builtin.array_values`;
individual feature checks and unsupported-boundary checks remain separate.
This is stdout/normal-exit evidence, not typed native return observation or
validation of Gossip/WordPress source pins. Earlier separate-branch PASS results
do not constitute a combined revision PASS.

`string_array_integration.php` additionally composes the bounded sprintf and
str_replace implementations with the array helpers (including key existence and
list checks), guarded string conversion and byte-length observation. The two
string entries link the same integration check; it is one measured case, not
two independent proofs. The generated binary is compared with the admitted
native oracle for its selected target. This does not update or validate Gossip's
source pins.

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

The active compatibility targets are PHP **8.3** and **8.4**. Native
differential tests must run against PHP **8.3.36 or newer** for `8.3`, or PHP
**8.4.27 or newer** for `8.4`. A compiler target is a series label; the patch
version is an oracle requirement and evidence attribute, not a CLI/compiler
target. Each live evidence run records the selected target and actual PHP
version. CI also records its pinned image digest; host-PHP runs have no
container image identity unless explicitly supplied.

Within the closed subset, `PHP_MAJOR_VERSION` and `PHP_MINOR_VERSION` are
compiler-target constants: `8.3` produces `8`/`3` and `8.4` produces `8`/`4`.
They do not inspect host PHP or an oracle patch release. Their deliberately
narrow syntax and expression boundary is documented in
[target version constants](target-version-constants.md).

`internal/phpversion/policy.go` is the compiler-side registry of accepted
series labels and their minimum oracle version. The CI matrix has one explicit
job per target. Its fixed PHP 8.4 oracle is
`php:8.4.27-cli@sha256:debadc2e36b857657846e2e952cbaf52702e55482dd3c2c43806f68d8c534a70`;
the 8.3 image identity remains in the execution configuration. A target PASS
exists only after that target's job produces evidence. The same source corpus
does not permit an 8.3 PASS to satisfy the 8.4 target. The
[PHP 8.3.36 release](https://www.php.net/releases/8_3_36.php),
[PHP 8.4.27 release](https://www.php.net/releases/8_4_27.php), and PHP's
[supported-version policy](https://www.php.net/supported-versions.php) are
review inputs for approved image-digest updates and EOL decisions.

PHP 7.4 and older 8.3 evidence remain readable historical records only; they
are not active test gates or current support claims. To add a later series,
introduce a separate policy/matrix entry with a series name, minimum release
floor, resolved immutable oracle digest, target-specific semantic table/corpus,
native differential job, and reviewed evidence. It must remain unclaimed until
those requirements pass. A version floor alone is not a security guarantee for
GoHP, generated binaries, input, dependencies, or deployment. The
8.4-specific boundaries are in [its target contract](php84-target.md).

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
against the selected active target's approved PHP oracle. Binary/NUL/invalid UTF-8 cases are runtime unit
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
the selected target's pinned oracle. The native comparison observes stdout and successful
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

Unlike native PHP in the active target series, this bounded implementation rejects bool/null/float/resource
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

For pinned original source identity (not Composer metadata or PHP callsite
coverage), see the separate [application corpus verifier](application-corpus.md).
Its [lexical inventory](application-corpus.md#lexical-inventory) uses a selected
native PHP tokenizer without executing source; candidate frequencies, native
environment/selected-declaration name evidence and all
unclassified parentheses remain distinct from resolved calls or support ratios.

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
