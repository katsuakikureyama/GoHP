# Pinned application source corpus

The first selection is the **complete, unchanged** WordPress
`wp-includes/functions.php`, not a list of functions already supported by GoHP.
The [manifest](../corpus/wordpress-functions.json) freezes source identity and
selection only. Function/callsite counts, dependency closure, compilation,
execution and application compatibility are **unmeasured**.

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
