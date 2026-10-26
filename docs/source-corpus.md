# Pinned source corpus inspection

Status: **partial tooling**, not application compatibility. This is the first
step of the [application inventory](roadmap.md#first-application-inventory):
freeze the unchanged source selection before measuring syntax or execution.

## Initial selection and provenance

The [manifest](../corpora/wordpress-functions/manifest.json) selects the entire
`wp-includes/functions.php`, not a filtered set of translatable functions.
Its WordPress upstream revision is
[`54106f6c545fc8b881f6b39b7963002bece6c28c`](https://github.com/WordPress/WordPress/commit/54106f6c545fc8b881f6b39b7963002bece6c28c),
with version `7.2-beta1-64693`: a development snapshot, **not a stable release**.
This pin was selected from the original-source checkout already retained by
Gossip. The three selected files match that upstream baseline; their byte counts
and SHA-256 hashes are recorded in the manifest. A later Gossip fork HEAD is not
substituted for the upstream commit, and no Gossip consumer pin is changed.

| Role | File | Bytes |
|---|---|---:|
| Subject | `wp-includes/functions.php` | 305835 |
| Provenance | `wp-includes/version.php` | 1113 |
| Provenance | `license.txt` | 19903 |

Only metadata is committed to GoHP; upstream PHP and license text are not copied.
The pinned [version file](https://github.com/WordPress/WordPress/blob/54106f6c545fc8b881f6b39b7963002bece6c28c/wp-includes/version.php)
and [license notice](https://github.com/WordPress/WordPress/blob/54106f6c545fc8b881f6b39b7963002bece6c28c/license.txt)
are provenance, not extra subjects or a license approval. GoHP's
[license decision](licensing.md) remains separate and unresolved.

Other core files, themes, plugins and transitive dependencies are outside this
initial selection, not established as absent or unnecessary. Dependency closure
is `unresolved`. A future repin starts a new measurement series. Earlier extracted
Gossip helper comparisons do not validate this complete file or current GoHP.

## Offline use

Provide an existing private, stable source directory containing the pinned files:

```sh
go run ./cmd/gohp-corpus -manifest corpora/wordpress-functions/manifest.json /path/to/wordpress
go run ./cmd/gohp-corpus -manifest corpora/wordpress-functions/manifest.json -format json /path/to/wordpress
```

All flags precede the one source-root argument. Markdown is the default.
The tool reads the manifest and listed file bytes only: no download, Git command,
PHP parsing/execution, compiler invocation, dependency resolution, plugin loading
or input/output-file writes. A prebuilt inspector does not require PHP or Git;
`go run` builds the trusted tool and may populate Go's build cache.

Success means every listed path, size and SHA-256 matches. It does not authenticate
the manifest's repository, revision or version; the caller reviews provenance.
Unlisted files are not traversed or inspected. Inspection uses local permissions,
not a sandbox for a malicious user concurrently modifying directories.

## Manifest and result contract

Input schema `gohp.source-corpus/v0.1` requires all fields shown in the committed
manifest. Unknown fields, duplicate keys, case-folded aliases, nulls and trailing
JSON are rejected. Limits: UTF-8 regular manifest at most 1 MiB, at most 64 JSON
nesting levels, 1–64 listed files, 1 byte–16 MiB per file, 64 MiB total.
At least one `subject` is required; other files must be `provenance`.

IDs use lowercase ASCII letters/digits/dot/underscore/hyphen, start alphanumeric,
and are at most 96 bytes. Project/version are nonempty printable text up to 128
bytes each; selection description up to 1024. Repository is a bounded HTTPS URL
without credentials/query/fragment, revision a full lowercase 40-hex Git commit,
and each hash a lowercase 64-hex SHA-256. PHP targets form a nonempty unique list
of currently admitted labels; they declare intended scopes, not measured support.
Selection unit is `whole-file`, dependency closure `unresolved` in this schema.

Paths are unique canonical relative slash paths, at most 512 bytes, with ASCII
alphanumeric/dot/underscore/hyphen components and no `.` or `..` component.
Listed sources must be regular files; symlink files or intermediate components
are rejected. The explicitly supplied root may resolve a symlink, but the caller
must keep the resolved root stable. A symlink manifest is rejected. These checks
are not a filesystem snapshot or protection against hostile replacement races.

The common JSON result schema is `gohp.source-corpus.inspection/v0.1`:

- `status: source-identity-match`, `compatibility: unverified`, and
  `analysis_status: not-measured` are separate dimensions.
- `manifest_sha256` hashes the exact input bytes. `manifest` retains declared
  selection/order; observed `files` are sorted by path, with role, bytes, SHA-256
  and `status: match`. No absolute source-root path or timestamp is emitted.
- `measurements` separates subject files/bytes from provenance files/bytes.
  For this pin these are 1 / 305835 and 2 / 21016, respectively.
- `function_units`, `syntax_nodes`, `call_sites`, `compiled_units` and
  `executed_cases` are JSON `null`, Markdown `n/a (not measured)`, never inferred
  zero. Dependencies, application coverage and performance remain unmeasured.
- Notices retain these limits. Declared text is escaped for Markdown/HTML;
  metadata is not a general secret-scanning or redaction boundary.

Exit codes: 0 after successful output (or help), 2 for CLI usage, 1 for inspection,
rendering or output failures. Input/inspection failures produce no report or
partial stdout; an output-writer failure can leave partial bytes and remains
failure. Read errors may identify local paths on stderr. No automatic report
publication occurs. Callers must check successful exit before consuming stdout.

## Capability and evidence

| Field | Contract |
|---|---|
| Stable capability ID | `gohp.corpus.inspect` |
| Owner / resource | Local GoHP developer tool / caller-readable manifest and its selected source files |
| Input schema | Required `manifest: string`, `source_root: string`; optional `format: markdown\|json`, default `markdown`; CLI mapping above |
| `inspect` | Validate manifest and read/hash all selected files under local filesystem permissions |
| `prepare` | Build the complete versioned inspection report in memory |
| `save` / `commit` | Excluded: no file save, source changes, publishing or external mutation |
| Result / status | Versioned JSON or equivalent Markdown only after full inspection; identity match remains compatibility `unverified`; errors use stderr/nonzero exit |
| WebMCP / agent exposure | **Excluded**: offline local CLI, no browser operation or server endpoint; no native/compat WebMCP implementation or auto-submit |
| Authentication / audit | Caller OS permissions; stdout/stderr and caller-retained command/revision are local evidence, not an authenticated server audit. No browser session, S2S token, MCP bearer or webhook route is introduced |

`gohp.tooling.source-corpus` in the [catalog](../compatibility/catalog.json) links
synthetic unit/CLI rejection and identity checks. CI validates committed metadata
without fetching WordPress. These tooling tests do not execute or certify the
selected application source. A real source inspection must separately retain
the command, GoHP revision/configuration and manifest hash/result.
Pinned `corpora/**/manifest.json` inputs also participate in the compatibility
runner's source fingerprint; generated report JSON does not.
