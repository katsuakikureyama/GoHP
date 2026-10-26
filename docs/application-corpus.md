# Pinned application source corpus

The first selection is the **complete, unchanged** WordPress
`wp-includes/functions.php`, not a list of functions already supported by GoHP.
The [manifest](../corpus/wordpress-functions.json) freezes source identity and
selection only. A separate [lexical inventory](#lexical-inventory) records token
observations. Semantic function/resolved-call counts, dependency closure,
compilation, execution and application compatibility are **unmeasured**.

## Provenance and scope

- Upstream: [WordPress/WordPress at `54106f6c545fc8b881f6b39b7963002bece6c28c`](https://github.com/WordPress/WordPress/tree/54106f6c545fc8b881f6b39b7963002bece6c28c).
- Version declaration: `7.2-beta1-64693`, a development snapshot, not a stable
  release or a claim about latest WordPress.
- Source: [original functions.php](https://github.com/WordPress/WordPress/blob/54106f6c545fc8b881f6b39b7963002bece6c28c/wp-includes/functions.php),
  305,835 bytes, SHA-256
  `456ab4cc7a6e6f05b061f2b1905f44233f8906774c461f90c5b80b9841303ccc`;
  upstream Git blob `3b9e247fd44fd6edbd2d942c37eb64889f01c936`.
- Notice: upstream `license.txt`, including its WordPress GPL v2-or-later
  declaration, contributor and inherited notices. The complete notice file is
  required and hash-checked; the manifest does not replace it or select GoHP's
  own license.
- Provenance file: `wp-includes/version.php`, hash-checked but never executed.

This is the same upstream pin retained by Gossip's
`gossip/tests/gohp/return-helpers.json`; that earlier helper selection is **not**
the selection here. Source bytes, blob identity and the retained pin were checked
when creating this manifest. The verifier itself does not authenticate GitHub,
Git commits, version declarations or licenses. It checks equality with a
caller-selected manifest. GoHP does not vendor these WordPress source files.

All bytes of `functions.php` are selected, including unsupported/unparsed syntax,
top-level statements and dynamic calls. No extraction, rewriting or owned
substitute is permitted in this selection. Other WordPress source files are
excluded from executable scope; notice/provenance roles do not increase source
counts. Direct/transitive dependencies, extensions, globals and external services
are unresolved. This is not a runnable WordPress bundle.

`source_file_count: 1` counts files, **not** functions, calls, cases or supported
units. The two target labels declare intended separate measurement series; they
do not assert either target can compile or execute this file. Changing the
selection or manifest bytes creates a new manifest identity.

## Read-only verification

Supply a private local checkout containing the three pinned files. For example,
after obtaining the upstream revision separately, from the GoHP repository:

```sh
go run ./cmd/gohp-corpus-verify \
  -manifest corpus/wordpress-functions.json /path/to/pinned/WordPress
```

The command does not invoke PHP, Composer, Git, a shell, the compiler or a
network client. It does not discover or inspect unlisted files, follow application
includes, install dependencies, or write reports/files. Its JSON stdout includes:

- `schema: gohp.corpus.verification/v0.1`, `status: source-verified`;
- SHA-256 of the **exact manifest bytes**, the parsed manifest, source-file and
  source-byte counts (notice/provenance excluded from these counts);
- `compatibility: unverified`; `function_units`, `callsites`, `compile_ratio`
  and `execution_ratio` are `null`, never zero or a fabricated percentage;
- separate target entries with `compile` and `execution` both `not-measured`.

Exit 0 means all listed files matched length and SHA-256, not PHP compatibility.
Exit 1 means invalid input, read/mismatch or output failure; exit 2 is usage error.
Help exits 0 without verification. Validation failure emits stderr with no JSON.
Output-device failures can leave partial stdout: consumers must check exit status.
For audit, retain stdout together with the GoHP tool commit, command and Go/OS
configuration; this schema does not auto-detect or authenticate the running binary.
No compiler/oracle identity is invented because neither is run.

The compatibility runner's source fingerprint includes JSON manifests under
`corpus/`, in addition to its Go/PHP/catalog inputs. Changing the selected
manifest therefore changes the runner's input identity even if compiler code is
unchanged. Keep generated reports outside that input directory (for example in
`build/`, which is excluded). This binding does not turn a corpus-byte check into
a PHP compatibility observation. Offline CI checks that the committed selection
metadata remains valid without downloading the selected WordPress files.

The strict v0.1 manifest requires exactly the fields shown in the committed
example: nonempty id/upstream repository/version, a 40-lowercase-hex upstream
revision, `whole-file` selection with a nonempty exclusion explanation and
`unresolved` dependencies; one or both unique targets `8.3`/`8.4`. Each listed
file has a canonical relative slash path, role (`source`, `notice`, `provenance`),
positive integral byte count and 64-lowercase-hex SHA-256. At least one source and
one notice are mandatory. Unknown/missing fields, duplicate JSON keys/paths and
trailing JSON are rejected. Limits: 1 MiB manifest, 64 files, 16 MiB per file,
64 MiB total declared bytes. Selected files must be regular and resolve inside
the source root; an escaping symlink is rejected. The manifest itself must be a
regular file, not a symlink. Inputs must not be mutated concurrently: these checks
are **not a sandbox against filesystem races or hostile local processes**.

## Lexical inventory

The separate `gohp-corpus-inventory` command reads the same manifest, verifies
**all** selected source/notice/provenance files, and retains the exact checked
source buffers. Only then does it feed those buffers to a fixed, embedded PHP
tokenizer program on stdin; paths are not reopened. The byte-only verifier above
still does not invoke PHP. No parser library or compiler frontend is replaced.

```sh
go run ./cmd/gohp-corpus-inventory -manifest corpus/wordpress-functions.json \
  -php-version 8.3 /path/to/pinned/WordPress
```

Use an explicitly trusted `php` on PATH matching the selected series and oracle
floor (default 8.3; 8.4 is separate). The tool invokes PHP with `-n`,
`short_open_tag=0`, `memory_limit=256M`, `display_errors=stderr`, `log_errors=0`;
input text is **never** included, evaluated or used as executable script.
It does not invoke Composer, autoload, a shell or the compiler, fetch data, or
write source/report files. `go run` itself builds the trusted tool and can populate
Go's build cache. Local filesystem permissions and the private stable-root
requirement still apply; this is not a hostile-process sandbox.

The [native token_get_all](https://www.php.net/manual/en/function.token-get-all.php)
scanner is used with flags **0**, not `TOKEN_PARSE`. This intentionally permits
lexical observations even for invalid grammar; a successful inventory is not a
syntax-validity result. Token names, not numeric identifiers, bind the protocol
because [PHP token numbers can change across versions](https://www.php.net/manual/en/tokens.php).

JSON result `gohp.corpus.lexical-inventory/v0.2` has `status: lexical-only`,
`compatibility: unverified`, selected `php_target`, observed `php_version`,
SHA-256 of the embedded tokenizer, and the unchanged nested verification report
binding the exact manifest and every selected file. v0.2 adds native name-set
evidence and per-candidate name matches; the v0.1 lexical fields and their
meaning are retained. Consumers must explicitly admit v0.2. Each source file records:

- Path, SHA-256, byte/token totals and a sorted histogram of **all** native token
  kinds and bytes, including whitespace, comments and string data. Single-byte
  punctuation uses `CHAR_XX` hexadecimal labels. Lengths must partition the
  complete source; malformed or partial token responses fail rather than omit bytes.
- `grammar_status: not-validated`, `unparsed_bytes` equal to the **entire file**,
  and `semantic_units` / `resolved_calls: null`. Lexical byte coverage and grammar
  parsing are different dimensions, not a 100% compatibility claim.
- A marker for every `T_FUNCTION` / `T_FN`. `named-function-marker` is a name
  following the function keyword (optional by-reference ampersand), not a global
  function count: it can include methods. Anonymous, arrow and unclassified
  function markers remain separate. No scope or body is parsed.
- A location for **every** literal opening parenthesis: function/arrow parameters,
  language-form, constructor-name, member-name, variable-parenthesis,
  name-parenthesis or unclassified. Classification skips only whitespace/comments.
  The unclassified count is explicit and its positions are retained.
- Frequencies of exact name spellings immediately preceding parentheses in the
  `name-parenthesis` category. These are candidates, **not resolved calls** or
  supported builtins: attributes and first-class callable creation can look the
  same lexically. Namespaces, signature/value types, conditional declarations,
  dispatch and dependency resolution are not evaluated. No catalog-name join is
  used to create support claims. Sorting is frequency descending then byte spelling.

### Candidate name evidence (not call resolution)

The same trusted scanner process also observes
[`get_defined_functions()['internal']`](https://www.php.net/manual/en/function.get-defined-functions.php)
and [loaded extension names](https://www.php.net/manual/en/function.get-loaded-extensions.php).
`native_environment` retains the byte-sorted, unique `internal_functions`,
`extensions` and `zend_extensions` lists. Its `sha256` hashes Go's compact JSON
encoding of `[internal_functions, extensions, zend_extensions]`, without a newline.
The native version and these sets must remain identical across selected files;
drift fails the entire report. These are the observed `php -n` process's names,
not all functions/extensions PHP can provide, extension versions, or an
application's production environment. No selected source is loaded to collect them.

Each file's `candidate_name_evidence` preserves **every** original frequency
entry in its existing order and adds these mutually exclusive categories:

| Category | Name-only observation |
|---|---|
| `native-internal-name` | Matches an observed native internal function name only |
| `selected-declaration-name` | Matches named declaration markers in the selected source files only |
| `both-name-sources` | Both matches exist; no precedence or dispatch is inferred |
| `unmatched-name` | Neither name source matches; it may need another file/extension or may not be a call |
| `qualified-name-unresolved` | Contains a namespace separator; no prefix stripping, basename match or fallback attempted |

Matching uses the base64-decoded bytes and **ASCII-only** case folding, never
display text or Unicode folding. Raw spellings, counts and non-UTF-8 identities
remain separate. `native_internal_match` and `selected_declaration_markers`
make the two sources explicit; qualified candidates have neither lookup applied.
Declaration matches span all selected source files, including methods, nested,
conditional and namespaced declarations; they do not prove a callable function
exists. Source roles, not notice/provenance files, supply these markers.

`name_evidence_status: name-only-not-resolved` applies to all five categories.
The [PHP namespace rules](https://www.php.net/manual/en/language.namespaces.rules.php),
imports, runtime availability, attributes, first-class callable creation and
dynamic dispatch are **not resolved**. Even an unqualified `strlen` can be an
attribute name or refer to a namespaced declaration. No GoHP builtin/catalog
support join is performed. `resolved_calls` / `semantic_units` stay null;
all source bytes stay grammar-unparsed. Use frequencies to prioritize an AST
and dependency investigation, not to label WordPress functions or supported calls.

The internal sidecar protocol is now `gohp.php.tokens/v0.2`, requiring all three
name lists; old/missing/unsorted/duplicate lists fail. Limits are 10,000 internal
function names, 512 names per extension list and 1,024 bytes per name, within
the existing total stdout bound. Names are nonempty UTF-8 with no ASCII controls.

### Positions, limits and verification

Positions are zero-based half-open byte offsets for the keyword or opening
parenthesis and one-based LF-counted source lines. Name base64 preserves exact
bytes; a display `name` is also emitted when valid UTF-8. Source bodies, comments
and string literals are not copied into the report. Identifiers are not redacted;
this tool is not a secret scanner. File results retain manifest source order.

Limits in addition to the manifest limits: 250,000 tokens per file, 500,000 for
the selection, 16 MiB stdout / 64 KiB stderr per tokenizer process and a shared
30-second tokenizer deadline. A failed process, any stderr, wrong series/floor,
changed PHP version/name environment between files, invalid token tape or exhausted limit fails
the **entire** inventory. No file or token is silently dropped. CLI exit codes
are 0 success/help, 2 usage, 1 pipeline/output failure; pipeline errors emit only
generic diagnostics and no JSON. Output-device failures may leave partial bytes;
consumers must check exit status. Retain command, GoHP commit/toolchain and PHP
image identity separately with the report; hashes are not authentication.

CI exercises the trusted tokenizer on repository-owned, side-effectful and
syntactically invalid source-as-data, separately per PHP target. CI does not fetch
the WordPress selection; original-source measurements are separate observations.

### Initial lexical observation

Historical v0.1 reports, before candidate name evidence, measured separately
with GoHP `d77fb01922af81aa51722470bcfc516080980004`
(Go 1.23.0, Linux/arm64 binary built on Darwin/arm64) against the unchanged
manifest SHA-256 `826e7dfa8c59975119042e39fbcf3c8ae753d47fa2dddd22ea8e484b47e891e2`.
Read-only, network-disabled containers used the same PHP image digests as CI:
8.3 `98d555ff97b3865140ad31032508c3ca089fa97a8944b14ca9dc03ea05d788cd`,
8.4 `debadc2e36b857657846e2e952cbaf52702e55482dd3c2c43806f68d8c534a70`.

| Lexical observation | PHP 8.3.36 | PHP 8.4.27 |
|---|---:|---:|
| Source bytes / grammar-unparsed bytes | 305835 / 305835 | 305835 / 305835 |
| Tokens / token kinds | 42732 / 96 | 42732 / 96 |
| Named / anonymous / arrow declaration markers | 217 / 1 / 2 | 217 / 1 / 2 |
| All opening parentheses | 2631 | 2631 |
| Name-parenthesis candidates | 1332 | 1332 |
| Member / variable / constructor parentheses | 56 / 8 / 14 | 56 / 8 / 14 |
| Unclassified parentheses (positions retained) | 64 | 64 |

The leading candidate spellings were `apply_filters` (64), `sprintf` (56),
`__` (53), `function_exists` (35), `is_array` and `str_replace` (33 each).
These are prioritization hints, not resolved dependencies, supported calls or
WordPress execution results. Equal observations do not transfer compatibility
between versions. Both reports keep semantic/resolved-call counts null.

Tokenizer SHA-256: `8eb49cfb2b2a3b5a647a14c4aa7beb3ebb62de32f94422f564e4e23120bb58c6`.
Exact JSON stdout SHA-256: 8.3
`aadebddd3e4823d91c9fd98fc3f437f565b204dce6eaa023411286257455bda7`;
8.4 `a75fa13e76b828f7e8264bd7dd445a42a28bffd605b515f3df0edfff6ff1b627`.
Original-source reports are caller-retained local evidence, not artifacts of the
synthetic CI tests. No source files or full reports are vendored here.

### Candidate name observation

Measured independently with GoHP `a3aeb0ed21345bbd05b4dfe43590461afa8174b0`,
Go 1.23.0 Linux/arm64 binary built on Darwin/arm64, using the same unchanged
manifest, read-only mounts, network-disabled containers and fixed image digests
listed above. Both runs retained 305,835 source/grammar-unparsed bytes,
42,732 tokens, 361 exact candidate spellings and 1,332 candidate occurrences.
The 64 unclassified opening parentheses remain visible outside this name table.

| Name-only category | PHP 8.3.36 spellings / occurrences | PHP 8.4.27 spellings / occurrences |
|---|---:|---:|
| Native internal name only | 152 / 759 | 154 / 761 |
| Selected declaration marker name only | 83 / 184 | 83 / 184 |
| Both name sources | 0 / 0 | 0 / 0 |
| Unmatched name | 126 / 389 | 124 / 387 |
| Qualified name unresolved | 0 / 0 | 0 / 0 |

`array_all` and `array_find_key` each occurred once: unmatched on 8.3.36,
native-internal-name on 8.4.27. This is an environment name-set observation,
**not** a resolved-call or GoHP-support change. The observed internal-name counts
were 1,077 / 1,093; both had 32 regular extensions and no loaded Zend extensions.
The unmatched priority list began `apply_filters` (64), `__` (53), `get_option`
(17), `do_action` (14), `_x` (12), `trailingslashit` (11). These identify where
to investigate dependencies, not proof of their definition or availability.

Tokenizer SHA-256: `8a617596ad016e7c79b1289c863ef2714268076ddda4e276d0f3311556fcee30`.
Native-environment SHA-256: 8.3
`a3b1b180f43b5ba2f6ca11afb5c98d0d8273783983e1978fbe9dd8d68cad15df`;
8.4 `1b038361870ffe39958ccbde5b7628fc72317c52b3461472d5a56d822598d56b`.
Exact JSON stdout SHA-256: 8.3
`1eb3a4dc75f5ef987f78ef94faef18ede160a44f9c3fb30f09e980fb2c7e319f`;
8.4 `32fc23f459e3584c40a62907f277f2d57dd9c9aeb63c9240011e7fb96f6af942`.
Reports remain local evidence, not CI artifacts; semantic/resolved counts are
null and compatibility remains unverified on both targets.

### Inventory capability

Stable ID: `gohp.corpus.inventory`. Owner/service: local GoHP development tool;
resource: caller-selected manifest and source root. Input schema has required
nonempty strings `manifest`, `source_root`, optional enum `php_target` (`8.3`,
`8.4`, default `8.3`), no additional properties. CLI maps target to `-php-version`;
all flags precede the one root. Result is the versioned JSON above or a nonzero
failure, never a save/publication state.

`inspect`: verify and retain the full selection. `prepare`: execute only the
trusted tokenizer on the retained buffers, observe its native name environment,
and construct lexical observations plus explicit name-only evidence (v0.2).
`save` / `commit`: **excluded**, no mutation/publication. Authorization is local
read/execute permission; stdout/stderr and caller-retained run metadata provide
local evidence, not an authenticated server audit. WebMCP, AI Assist, browser UI,
MCP and automatic submit are **excluded**: no hosted operation or route exists,
and no browser/S2S/MCP/webhook authentication is changed. The existing verifier
capability below remains distinct. Tooling tests live in the existing catalog;
they are not application differential evidence.

## Dependency triage (separate selection)

The [companion manifest](../corpus/wordpress-dependency-triage.json) freezes
**seven complete source files / 908,791 source bytes** at the same upstream
`54106f6c545fc8b881f6b39b7963002bece6c28c` revision. It retains `functions.php`
and adds `plugin.php`, `class-wp-hook.php`, `class-wp-filter-sentinel.php`,
`l10n.php`, `option.php` and `formatting.php`, all under `wp-includes/`.
The two files required directly by `plugin.php` are both included; this does
**not** complete the dependency closure. The same full notice and version file
are verified separately. No WordPress source is vendored or modified.

This is a definition/dependency **inspection selection**, not a runnable bundle.
The original one-file manifest and its measurement series remain unchanged.
Do not compare counts across the two selections as a coverage improvement, or
rerun the larger selection and silently replace the original denominator.
The upstream Git tree/blob identities and local byte equality were inspected
when pinning; the CLI still only checks the caller's manifest, not authenticity.

```sh
go run ./cmd/gohp-corpus-verify \
  -manifest corpus/wordpress-dependency-triage.json /path/to/pinned/WordPress
```

The existing `gohp.corpus.verify` / `gohp.corpus.inventory` capability contracts
apply unchanged to this explicitly selected input. No automatic discovery,
include execution, new hosted operation, save/commit or WebMCP exposure is added.
Offline metadata checks protect selection boundaries; they are not source,
grammar or application tests. The companion's byte verification is recorded
separately below; no lexical, native-behavior or compiled-Go measurement of the
larger selection is claimed here.

### Highest-frequency unmatched names

The frequencies below are the **original one-file v0.2 observations**, identical
for these six spellings on 8.3.36 and 8.4.27. Source inspection located these
definitions; it did not resolve each candidate occurrence or runtime dispatch.
All source links below are fixed to the inspected upstream revision.

| Candidate / occurrences | Located definition | Direct dependencies and remaining boundary |
|---|---|---|
| `apply_filters` / 64 | [plugin.php:175](https://github.com/WordPress/WordPress/blob/54106f6c545fc8b881f6b39b7963002bece6c28c/wp-includes/plugin.php#L175) | Global registry/counters/current-hook stack, `_wp_call_all_hook`, `WP_Hook::apply_filters`; registered callbacks and their dependencies are unknown |
| `__` / 53 | [l10n.php:339](https://github.com/WordPress/WordPress/blob/54106f6c545fc8b881f6b39b7963002bece6c28c/wp-includes/l10n.php#L339) | `translate` → translation object lookup/method and static/dynamic gettext filters; not an identity-string helper |
| `get_option` / 17 | [option.php:78](https://github.com/WordPress/WordPress/blob/54106f6c545fc8b881f6b39b7963002bece6c28c/wp-includes/option.php#L78) | Filters, installation state, options caches, `$wpdb` queries, recursion and `maybe_unserialize`; services and initialization remain unknown |
| `do_action` / 14 | [plugin.php:499](https://github.com/WordPress/WordPress/blob/54106f6c545fc8b881f6b39b7963002bece6c28c/wp-includes/plugin.php#L499) | Global registry/counters/stack, `all` callbacks and `WP_Hook::do_action`; callback effects must not become no-ops |
| `_x` / 12 | [l10n.php:461](https://github.com/WordPress/WordPress/blob/54106f6c545fc8b881f6b39b7963002bece6c28c/wp-includes/l10n.php#L461) | `translate_with_gettext_context` → translation object lookup/method and contextual filters; same unresolved loading/state boundary |
| `trailingslashit` / 11 | [formatting.php:2883](https://github.com/WordPress/WordPress/blob/54106f6c545fc8b881f6b39b7963002bece6c28c/wp-includes/formatting.php#L2883) | `untrailingslashit` → `rtrim` with the original slash/backslash character list; bounded helper evidence does not compile the whole formatting file |

### Hooks: shared runtime semantics, not two builtin stubs

- [plugin.php:24](https://github.com/WordPress/WordPress/blob/54106f6c545fc8b881f6b39b7963002bece6c28c/wp-includes/plugin.php#L24)
  requires **both** hook and sentinel classes, then initializes four mutable
  globals. `apply_filters` / `do_action` update counters even on an empty
  registry, and manage nested current-hook state and the `all` hook.
- [WP_Hook](https://github.com/WordPress/WordPress/blob/54106f6c545fc8b881f6b39b7963002bece6c28c/wp-includes/class-wp-hook.php#L26)
  is final and implements `Iterator` / `ArrayAccess`; it has dynamic-property
  attributes and mutable callback, priority and iteration state. It does not
  itself use inheritance. The complete class, not a compatible-looking mock,
  remains outside GoHP's admitted class subset.
- [Dispatch at line 339](https://github.com/WordPress/WordPress/blob/54106f6c545fc8b881f6b39b7963002bece6c28c/wp-includes/class-wp-hook.php#L339)
  calls general string/object/method callbacks with `call_user_func` /
  `call_user_func_array`, truncates arguments using `accepted_args`, and chains
  filter return values. Actions share this machinery but do not expose those
  results. Callback identity uses `_wp_filter_build_unique_id`, including
  object identity, not just a function-name string.
- [Active-iteration adjustment](https://github.com/WordPress/WordPress/blob/54106f6c545fc8b881f6b39b7963002bece6c28c/wp-includes/class-wp-hook.php#L128)
  includes by-reference `foreach`, array-pointer operations and priority
  reordering when callbacks change registration during dispatch.
  `do_all_hook(&$args)` adds a by-reference parameter. Reentrancy, add/remove
  timing, order and reference effects are observable contracts, not bookkeeping
  that can be dropped. The [reference milestones](reference-value-model.md#delivery-gates-and-status)
  and general callable/global-state work are prerequisites, not already complete.
- The [sentinel class](https://github.com/WordPress/WordPress/blob/54106f6c545fc8b881f6b39b7963002bece6c28c/wp-includes/class-wp-filter-sentinel.php#L9)
  is a separate required file at this pin. Its empty final class and strict-types
  declaration do not justify deleting it; consumers distinguish objects by identity.

The missing closure includes actual registered callbacks, plugin/theme loading,
their mutable state and possible I/O/errors. No finite graph is inferred from
this inspection, and no original hook path has been executed or translated here.

### Translation, options and next work

`get_translations_for_domain` ([l10n.php:1582](https://github.com/WordPress/WordPress/blob/54106f6c545fc8b881f6b39b7963002bece6c28c/wp-includes/l10n.php#L1582))
uses global `$l10n`, just-in-time loading, a static fallback object and a reference
assignment into the global array. Translation controllers/registry/classes,
locale/theme paths and translation files are outside this selection. Even the
fallback path retains state and aliasing; replacing `__` / `_x` with identity
functions would remove observable hooks and loading behavior.

`get_option` may short-circuit through filters, but otherwise traverses
alloptions/notoptions/options caches and database queries, mutating caches and
temporarily adjusting database error state. Its returned value can be
unserialized and filtered. A default-value-only or in-memory mock is not the
original function's contract; database/cache backends have not been selected.

Priority from this audit: continue the existing reference/value and general
callable/global-state prerequisites, and define original-hook cases for order,
argument counts, nesting, mutation and references before any source admission.
Keep the roadmap's original block-parser → original hooks acceptance sequence;
Gossip owns application drivers and consumer pins. Name/AST resolution and the
unselected dependency closure remain open. No compiler feature, native behavior,
Go compatibility, application percentage or performance result is added by this audit.

## Next evidence, not claimed here

Build a bounded syntax/function/callsite inventory that retains unknown and
unparsed content in its accounting. Do not interpret regex/name matches as
resolved calls. Freeze any function-unit/case denominator explicitly before
publishing ratios, and retain unresolved dependencies as blockers. Future compiler
or execution reports need exact compiler/runtime/tool/driver/oracle pins in
addition to this source manifest. See the [roadmap](roadmap.md#first-application-inventory)
for the separate inventory, compile, execution and acceptance gates.

## Local capability

Stable ID: `gohp.corpus.verify`. Owner/service: local GoHP development tool.
Resource: caller-selected manifest and local source root. Input schema:
`{manifest: nonempty string, source_root: nonempty string}`, required fields,
no additional properties; the CLI maps these to `-manifest` and one positional
argument. Authorization is local filesystem read permission, not a server role.

`inspect`: validate the manifest and selected file locations. `prepare`: check
bytes/hashes and construct the versioned result. `save`/`commit`: **excluded**;
no mutation, publication, persistent audit store, browser operation or server
operation exists. Result status is `source-verified` with compatibility still
`unverified`; error exits never return a success report. Stdout binds the manifest
and source identities for caller-retained audit, not authenticity. Unit tests cover
these boundaries; the [catalog](../compatibility/catalog.json) registers tooling
evidence separately from PHP language evidence.

WebMCP/AI Assist/browser UI/MCP exposure: **excluded**, a local developer CLI only.
No browser-facing, S2S, MCP or webhook authentication route is added or changed;
native/compat browser support and automatic submit are not applicable. There is
no hosted deployment in this change.
