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

JSON result `gohp.corpus.lexical-inventory/v0.1` has `status: lexical-only`,
`compatibility: unverified`, selected `php_target`, observed `php_version`,
SHA-256 of the embedded tokenizer, and the unchanged nested verification report
binding the exact manifest and every selected file. Each source file records:

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

Positions are zero-based half-open byte offsets for the keyword or opening
parenthesis and one-based LF-counted source lines. Name base64 preserves exact
bytes; a display `name` is also emitted when valid UTF-8. Source bodies, comments
and string literals are not copied into the report. Identifiers are not redacted;
this tool is not a secret scanner. File results retain manifest source order.

Limits in addition to the manifest limits: 250,000 tokens per file, 500,000 for
the selection, 16 MiB stdout / 64 KiB stderr per tokenizer process and a shared
30-second tokenizer deadline. A failed process, any stderr, wrong series/floor,
changed PHP version between files, invalid token tape or exhausted limit fails
the **entire** inventory. No file or token is silently dropped. CLI exit codes
are 0 success/help, 2 usage, 1 pipeline/output failure; pipeline errors emit only
generic diagnostics and no JSON. Output-device failures may leave partial bytes;
consumers must check exit status. Retain command, GoHP commit/toolchain and PHP
image identity separately with the report; hashes are not authentication.

CI exercises the trusted tokenizer on repository-owned, side-effectful and
syntactically invalid source-as-data, separately per PHP target. CI does not fetch
the WordPress selection; original-source measurements are separate observations.

### Initial lexical observation

Measured separately with GoHP `d77fb01922af81aa51722470bcfc516080980004`
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

### Inventory capability

Stable ID: `gohp.corpus.inventory`. Owner/service: local GoHP development tool;
resource: caller-selected manifest and source root. Input schema has required
nonempty strings `manifest`, `source_root`, optional enum `php_target` (`8.3`,
`8.4`, default `8.3`), no additional properties. CLI maps target to `-php-version`;
all flags precede the one root. Result is the versioned JSON above or a nonzero
failure, never a save/publication state.

`inspect`: verify and retain the full selection. `prepare`: execute only the
trusted tokenizer on the retained buffers and construct lexical observations.
`save` / `commit`: **excluded**, no mutation/publication. Authorization is local
read/execute permission; stdout/stderr and caller-retained run metadata provide
local evidence, not an authenticated server audit. WebMCP, AI Assist, browser UI,
MCP and automatic submit are **excluded**: no hosted operation or route exists,
and no browser/S2S/MCP/webhook authentication is changed. The existing verifier
capability below remains distinct. Tooling tests live in the existing catalog;
they are not application differential evidence.

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
