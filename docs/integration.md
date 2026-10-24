# Gossip builder contract v0.1 (proposal)

Gossip can invoke `gohp -in source.php -out candidate.go -php-version 8.3
-report candidate.report.json` in an isolated build workspace. The command exits
nonzero for unsupported syntax or semantic checks and produces no new Go output
in that case. Both destination paths must be fresh and distinct. Existing files,
directories, symlinks (including dangling links) and hard links are rejected without
overwriting them. This also protects the input from output-path aliasing.

The CLI stages complete files in each destination directory, then publishes Go
source followed by the report using exclusive hard links. A reported publication
failure attempts to remove only files published by this invocation. Errors during
cleanup are returned, not hidden. Artifacts use private file permissions (0600).
Destination filesystems must support hard links; no overwriting fallback exists.

This is not crash-atomic across two files: interruption or filesystem failure can
leave an incomplete pair. Builders must use a private, non-concurrently-mutated
workspace, require exit code zero, verify both report hashes, and discard any
failed candidate. A report's presence alone is not a commit marker. The CLI does
not protect against an attacker changing destination directories concurrently.
`-report` requires `-out`; stdout-only generation remains available without a report
and callers must discard any partial stdout on failure. Repeated generation uses
new candidate paths rather than overwriting a previous build.

`-php-version 8.3` identifies the active language series, not an arbitrary PHP
patch. Native verification for this contract uses PHP 8.3.36 or newer within
the 8.3 series and the
approved immutable oracle digest recorded in run evidence. PHP 8.4+ is not an
accepted integration target: adding it requires a separate policy/matrix and
Gossip-owned compatibility decision. Historical 7.4 archives are readable but
must not be used to approve a new candidate.

Capability/WebMCP scope: `gohp.cli.generate` is explicitly **excluded** from browser
WebMCP exposure because it is a local compiler writing caller-owned build files.
No browser operation, server resource, authentication route or auto-submit is added.
The same local generation operation now also accepts bounded `strlen(string)`
calls; its [string-only byte-length contract](compatibility.md#bounded-strlen-contract)
does not add a service operation or change this WebMCP exclusion.
The bounded one-argument `array_keys(array)` operation is covered by this same
exclusion and [array-only contract](compatibility.md#bounded-array_keys-contract);
it adds no browser/server route, permission, publication or automatic execution.
The same exclusion applies to the one-argument `array_values(array)` builtin
and its [bounded value-copy contract](compatibility.md#bounded-array_values-contract).
Changing the PHP oracle policy or adding a later series likewise adds no browser
operation, resource, authentication route, WebMCP capability, or auto-submit.

Report schema `gohp.build/v0.1` includes IR/compiler versions, source and generated
SHA-256, input path, PHP target and `generated` status. Additive fields `output`
and `source_map` identify the generated file and provide a versioned,
[statement-level line map](source-maps.md); consumers must tolerate unknown
fields, and older reports may omit both. Consumer must record the
exact GoHP Git revision and Go toolchain, build source against the same runtime
revision, and execute isolated differential fixtures before promoting status.
No automatic plugin activation, DB migration, source upload or deployment occurs.

Gossip owns WordPress hook semantics, host/plugin manifests, release slots and
compatibility policy; future WordPress adapters belong outside the compiler core.
No whole WordPress core/plugin conversion works yet. Next cross-project requests:
pin oracle PHP/WP versions; define source bundle manifest and fixture protocol;
add stderr, exit, mutation and hook-order comparison; agree compatibility registry
schema and immutable candidate IDs. Gossip source remains unchanged. These requests are also shared in the
[Gossip kickoff issue](https://github.com/katsuakikureyama/Gossip/issues/1).
