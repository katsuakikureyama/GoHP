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

This integration combines the predicate/conditional stack through PR #26
with PR #27's literal array_map and string-only addslashes, retaining the shared 8.3 oracle policy.
Earlier separate-stack evidence does not validate the resulting
revision. Gossip's original-source units and compiler pins are unchanged here;
they require independent validation before adopting this integrated revision.

Capability/WebMCP scope: `gohp.cli.generate` is explicitly **excluded** from browser
WebMCP exposure because it is a local compiler writing caller-owned build files.
The [one-level PHP array assignment and append](array-write-boundary.md) extension remains
within this excluded generation capability. Its writes concern simulated local
PHP values, not service resources; no UI operation, route, authentication,
external save/commit or automatic submission is introduced.
Bounded [array return declarations](array-return-boundary.md) enforce simulated
PHP return values within the same excluded local capability; no new user-facing
operation, server resource or authentication path is introduced.
No browser operation, server resource, authentication route or auto-submit is added.
The [literal-only negative integer syntax](negative-integer-boundary.md) also
uses this same excluded capability, with no new operation or authentication path.
The bounded [logical-not and logical-or operators](boolean-boundary.md) likewise remain within
this local compiler capability; they introduce no route, external write or auto-submit.
The [scalar strict-not-equal comparison](boolean-boundary.md) also uses the same
excluded local generation capability, without UI/server operations or new authentication.
Braced [if/elseif/else chains](boolean-boundary.md), including `else if`, remain inside this excluded capability:
branch execution changes simulated local PHP values, not service resources or auth routes.
The same local generation operation now also accepts bounded `strlen(string)`
calls; its [string-only byte-length contract](compatibility.md#bounded-strlen-contract)
does not add a service operation or change this WebMCP exclusion.
The bounded one-argument `array_keys(array)` operation is covered by this same
exclusion and [array-only contract](compatibility.md#bounded-array_keys-contract);
it adds no browser/server route, permission, publication or automatic execution.
The same exclusion applies to the one-argument `array_values(array)` builtin
and its [bounded value-copy contract](compatibility.md#bounded-array_values-contract).
`array_key_exists(int|string, array)` and its [bounded presence contract](compatibility.md#bounded-array_key_exists-contract)
also remain within this excluded local generation capability; no browser/server
operation, authentication route, resource mutation or auto-submit is introduced.
The one-argument `array_is_list(array)` [list-status check](compatibility.md#bounded-array_is_list-contract)
has the same local compiler exclusion; it adds no service operation or route.
Changing the PHP oracle policy or adding a later series likewise adds no browser
operation, resource, authentication route, WebMCP capability, or auto-submit.
The one-argument `is_array(value)` type predicate also remains within the same
excluded local `gohp.cli.generate` capability. It returns bool without changing
the value, and adds no route, authentication or automatic submission. Its
[bounded contract and evidence](boolean-boundary.md) do not validate Gossip pins.
The legacy `array(...)` frontend addition stays within this excluded local
compiler capability; it adds no service operation or authentication change.
The literal-mask `rtrim` addition also stays within this exclusion; it does not
introduce a browser operation, filesystem path resolver or network call.
The three-string `str_replace` addition remains the same excluded local
`gohp.cli.generate` capability, with no UI operation, route or authentication
change. Array/count/coercion forms remain outside its
[explicit subset contract](str-replace-boundary.md).

The full ternary and string-only `stripslashes` addition remains the same
excluded local `gohp.cli.generate` capability. It adds no browser operation,
service resource, authentication route, or auto-submit; short/coalesce ternary,
recursive WordPress unescaping and scalar/object coercion remain outside its
[explicit subset contract](ternary-stripslashes-boundary.md).

The [literal array_map / addslashes extension](array-map-boundary.md) remains
the same WebMCP-excluded local generation capability. Compile-time callback
resolution introduces no new UI/server operation or authentication route.

The [conditional/array compiler integration](gossip-compiler-integration.md)
preserves that exclusion and the existing operation boundary. Combining these
compiler subsets introduces no additional browser/service capability.

The bounded `sprintf('%0Ns', value)` addition remains the same excluded local
`gohp.cli.generate` capability: developer-only source translation, no browser
operation or new service/authentication route. Dynamic formats and values are
runtime-guarded; see [the subset contract](sprintf-boundary.md).

The boolean-control/type-predicate addition remains the same excluded local
compiler capability; no authentication or user-facing service action is added.

The [string-only strcasecmp extension](strcasecmp-boundary.md) retains that
WebMCP exclusion. It introduces no route or authentication change. Its compiler
original pin was based on #38. The current [charset/array integration](charset-array-integration.md)
explicitly combines published #40 and #36; prior Gossip manifests/results
are not upgraded automatically and retain their independent configurations.

The control-flow/array-return integration combines published #42 and #41.
Typed return validation and runtime guards propagate through both conditional
arms, including normalized `elseif` / `else if` chains and nested `foreach`.
`control_flow_array_integration.php` is an owned compiler fixture, not a
WordPress original or evidence for a new Gossip compiler pin. This integration
retains the excluded `gohp.cli.generate` capability and adds no service route,
authentication change, deployment or automatic promotion.

The control-flow membership integration additionally combines published #44
(`49a302e44464623dc4912bafbcf44192759d7878`) with #43
(`d349e433aabd2230a23d3fb2aaa9e76a5173b58b`). Both parent histories and their
fixtures remain intact. [Strict scalar membership](in-array-strict-boundary.md)
does not gain loose comparison or general mixed-array support from this merge.
Owned native composition and static/runtime rejection cases validate the new
combined configuration; previous parent PASS is not its evidence. The local
`gohp.cli.generate` WebMCP exclusion and existing service/auth boundaries remain
unchanged. No Gossip source/pin/evidence, release or deployment is changed.

Report schema `gohp.build/v0.1` includes IR/compiler versions, source and generated
SHA-256, input path, PHP target and `generated` status. Additive fields `output`
and `source_map` identify the generated file and provide a versioned,
[statement-level line map](source-maps.md); consumers must tolerate unknown
fields, and older reports may omit both. Consumer must record the
exact GoHP Git revision and Go toolchain, build source against the same runtime
revision, and execute isolated differential fixtures before promoting status.
No automatic plugin activation, DB migration, source upload or deployment occurs.

The [actual-bool return declaration](bool-return-boundary.md) extends named
global functions only. It preserves the local `gohp.cli.generate` WebMCP exclusion
and adds no service/auth route. Gossip's numeric-array original still needs typed
parameters and callable/arrow dependencies plus a separately pinned comparison;
accepting `: bool` alone is not original-source compatibility or adoption.

Case-insensitive true/false/null literals normalize only those three predefined
values to existing IR. This stays within the excluded local `gohp.cli.generate`
capability, without a service/authentication change, general constant resolution
or dependency. Existing consumer pins/evidence require separate adoption.

Direct negative-literal keys in bounded array-element `isset` reuse the existing
integer IR and runtime presence probe. This parser consistency extension remains
within the excluded local `gohp.cli.generate` capability: no general key-expression
evaluation, service operation, authentication route or external resource write.
Gossip compiler pins and prior evidence are not automatically upgraded.

Gossip owns WordPress hook semantics, host/plugin manifests, release slots and
compatibility policy; future WordPress adapters belong outside the compiler core.
No whole WordPress core/plugin conversion works yet. Next cross-project requests:
pin oracle PHP/WP versions; define source bundle manifest and fixture protocol;
add stderr, exit, mutation and hook-order comparison; agree compatibility registry
schema and immutable candidate IDs. Gossip source remains unchanged. These requests are also shared in the
[Gossip kickoff issue](https://github.com/katsuakikureyama/Gossip/issues/1).
