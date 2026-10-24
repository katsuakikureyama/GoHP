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

Before Go source generation, a semantic pass validates IR operand shape and
rejects statically known mismatches such as `count(1)`, scalar `foreach`/indexing,
and a known class receiver using an absent property, method, or wrong arity.
Values whose type genuinely depends on runtime input remain protected by explicit
runtime failures; validation success is not a differential compatibility result.

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
