# Gossip builder contract v0.1 (proposal)

For the current compiler scope and next application acceptance criteria, see the
[compatibility tables](compatibility.md) and [roadmap](roadmap.md). Consumer runs
below are historical, pin-specific observations, not current-main certification.

[Associative JSON decoding](json-associative-decode.md) targets the original
parser's attribute expression without changing its source. Exact two-argument
literal-true mode, ordered int64/string/bool/null/Array values and JSON-error
readers use the existing Execution threading; float forms/overflow and flags
remain excluded. This remains local `gohp.cli.generate`, WebMCP **excluded**:
no browser/server/MCP route, input/publication schema or authentication change,
new storage lane, deployment destination or silent consumer repin.

[Standalone call statements](call-statement-boundary.md) now express the
render caller's discarded-result method call without a dummy assignment.
The [source dependency audit](render-dependencies.md) records resolved
[form-feed lexing](form-feed-boundary.md) and remaining original HTML
tag-processor dependencies. No unchanged render PASS,
HTML substitute, host binding or Gossip repin is introduced. Gossip #41's
separate 16-case data-class evidence is recorded there without summing prior
24/334 configurations or inferring adoption of later compiler increments.

[Bounded while](while-boundary.md) adds condition re-evaluation and nearest-loop
break/continue, with conservative guards for loop-carried types. Owned token-array
selection with str_starts_with is not a replacement for the original render
caller's HTML processing. [Bounded strtok](strtok-boundary.md) now adds an
explicit execution-owned cursor; original-source acceptance remains separate.
Consumer pins and source
certification remain separate; local gohp.cli.generate stays WebMCP-excluded,
without route/auth/deployment changes.

For a generated program using strtok, legacy F/M/New host wrappers each create
a fresh execution. A host requiring a start/continue sequence across entries
must pass the same fresh `*rt.Execution` to FContext/MContext/NewContext entries
for that logical request only. Arrows use the invoking execution, not a retained
creator's cursor. Do not copy or concurrently share contexts, reuse them across
requests or infer a production request lifecycle. Existing pinned Gossip runs
remain unchanged; this additive ABI needs explicit consumer adoption.

Bounded [foreach continue](foreach-continue-boundary.md),
[null defaults and isset lists](array-get-prerequisites.md) address the parser
blockers measured in Gossip PR #35. The first real original-source execution
then exposed the [literal-null identity](literal-null-identity.md) gap, fixed
separately. No unchanged-source PASS or consumer adoption is claimed by compiler
tests. All remain inside WebMCP-excluded `gohp.cli.generate`, without route,
authentication or deployment changes.

Consumer evidence update (2026-10-26 JST):
[Gossip PR #36](https://github.com/katsuakikureyama/Gossip/pull/36) reports the
unchanged `_wp_array_get` passing all 24 retained Go/PHP cases using exact GoHP
pin `b8dc1757ed61f459d51362da004aa1d73d024116` and PHP 8.3.36. It checks return
types/values and input/path/default preservation. That was a separate 25th unit.
Subsequently [Gossip PR #37](https://github.com/katsuakikureyama/Gossip/pull/37)
merged as `9abd4b8364ca5e776473f637d76186e5b98f0047`: 25 unchanged originals plus
48 owned drivers were translated/built into one binary and all 294 Go/PHP cases
passed (262 earlier cases + 24 lookup cases + 8 new settings compositions), using
the same exact GoHP #64 pin above. This is a measured combined run, not addition
of old PASS counts. No full WordPress compatibility or adoption of later GoHP
revisions is inferred. This status sync does not repin or rerun Gossip.

[Gossip PR #38](https://github.com/katsuakikureyama/Gossip/pull/38) subsequently
merged as `fd113a179888cf4736c734c8640d795f58243201`, using exact GoHP #67 pin
`a3039414cff22689d214a6c90767f13dccf56ddb`. Its one-binary run passed 334 Go/PHP
cases across 26 unchanged originals and 48 retained drivers: 294 prior cases
plus 40 cases of the real `wp_should_add_elements_class_name` caller and its
`_wp_array_get` dependency. This measures class-necessity decisions, not rendering
hooks, CSS, HTML generation or full WordPress. Later compiler additions do not
inherit consumer adoption from that pin.

[`str_contains`](str-contains-boundary.md) supplies a concrete dependency of the
separate `wp_render_elements_class_name` guard. Owned fixture composition is
not unchanged-source evidence for that entire render function. It extends only
WebMCP-excluded `gohp.cli.generate`; no Gossip repin, route/auth/deploy changes.

[Gossip PR #39](https://github.com/katsuakikureyama/Gossip/pull/39), merged as
`37be02cf71a5f9536b0b646e1ffaab3e1819ed6b`, adds a separate 24-case binary at exact
GoHP #69 pin `731ef251ac907326f060aac5f1b5b994ac8615bb`: three unchanged original
functions plus an owned data-holder class and composition driver exercise output
serialization skip decisions followed by class-necessity decisions. The earlier
26-original/334-case binary remains unchanged. This is not a 27-original/358-case
combined PASS, nor a claim about registry, render hooks, CSS or HTML generation.

[Gossip PR #40](https://github.com/katsuakikureyama/Gossip/pull/40), merged as
`d6d7adfee959826fdc5d3960d55a3ff01920433f`, moves owned holder construction and
property assignment into translated PHP at exact GoHP #71 pin
`ce38319b4245ccea0a7ae627f53b4396acabfce6`. Its separately measured one-binary run
passes the retained 24-case corpus (16 direct/8 composed): three unchanged
originals, two owned functions and one owned data class. Source input and the
other observed inputs remain unchanged. Neither observer injects generated
fields on this path. Earlier 24/334 configurations and pins are unchanged, not
summed or rerun. This is not original WP_Block_Type/registry/render support or
adoption of the later break/while increments; no Gossip files are changed here.

[`str_starts_with`](str-starts-with-boundary.md) adds the byte-prefix dependency
used later in the render caller. Owned token/guard tests do not certify its
tokenizer, loop or HTML processor. The same WebMCP exclusion and no-repin boundary
apply; no route/auth/deployment change is included.

[`array_slice`](array-slice-boundary.md) adds positional selection with actual
int/null/bool bounds to WebMCP-excluded `gohp.cli.generate`. No consumer adoption,
route/auth/deployment change or Gossip pin/source modification is included.

[`array_merge`](array-merge-boundary.md) extends the same WebMCP-excluded local
`gohp.cli.generate` operation with nonrecursive array merging. Explicit positional
arrays only; no route/auth/deployment change, Gossip repin or source certification.

[`array_reverse`](array-reverse-boundary.md) extends WebMCP-excluded local
`gohp.cli.generate` with bounded array reversal and an actual-bool optional flag.
No route, authentication or deployment change; Gossip pins and original-source
evidence require their own adoption checks and are not changed here.

[`array_key_first`](array-key-first-boundary.md) extends the same WebMCP-excluded
local compiler operation with first-key lookup. It does not adopt or replace
Gossip's array_first value-returning polyfill, repin original-source evidence,
or add a service/authentication route.

[`array_key_last`](array-key-last-boundary.md) adds a read-only key lookup to the
same WebMCP-excluded local generation capability. It does not repin Gossip or
certify original array_last functions; no service/authentication route is added.

Bounded [array/string-variable empty](array-empty-boundary.md) is part of local
`gohp.cli.generate`, still WebMCP-excluded. It adds no route/authentication or
server operation; Gossip original-source pins and evidence remain separate.
The string-domain extension keeps this stable excluded capability and both CLI
input schemas unchanged. Preparation guards actual array/string locals and
rejects other domains; fresh publication and generated/unverified result remain
unchanged. No undefined-address suppression, browser/MCP operation, resource,
auth mode or automatic submission is added.

The `_wp_array_get` prerequisites — trailing untyped null defaults and bounded
isset addresses — use the same local `gohp.cli.generate` capability. They add
no browser/server/MCP operation, authentication route, external save, deployment
or automatic Gossip adoption; WebMCP remains explicitly excluded.

Bounded [nested array probes and null coalescing](nested-array-probes.md) are a
separate candidate for the pure array/boolean paths in WordPress
`wp_should_add_elements_class_name` (elements.php lines 33–97). They do not add
an external resolver, class substitute, browser/service route, dependency or
authentication path. Consumer adoption and a pinned unchanged-source comparison
remain unverified; compiler/runtime tests alone do not certify that consumer.

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

`-php-version 8.3` identifies a language series, not an arbitrary PHP patch.
The consumer observations recorded here use PHP 8.3.36 and their own immutable
oracle/source/compiler identities. The compiler now independently admits
[8.3 and 8.4](compatibility.md#active-php-version-policy); that does not extend
these Gossip results to 8.4 or repin their compiler. A new consumer target needs
a Gossip-owned compatibility decision and its own source/driver/oracle evidence.
Historical 7.4 archives are readable but must not approve a new candidate.

This integration combines the predicate/conditional stack through PR #26
with PR #27's literal array_map and string-only addslashes, retaining the shared 8.3 oracle policy.
Earlier separate-stack evidence does not validate the resulting
revision. Gossip's original-source units and compiler pins are unchanged here;
they require independent validation before adopting this integrated revision.

Capture-free [arrows and callable parameters](arrow-callable-boundary.md) remain
inside the WebMCP-excluded `gohp.cli.generate` capability: developer-only local
source translation, no browser operation, service route, authentication or auto-submit.

The bounded [is_int predicate](is-int-boundary.md) uses that same excluded local
operation. It adds no coercion, source type, route or dependency. Existing Gossip
pins/evidence are unchanged; compiler feature checks do not promote original-source
compatibility or adopt the new compiler revision automatically.

The [is_null predicate](is-null-boundary.md) also stays inside that excluded local
operation: no route/auth change, suppressed undefined reads or new source types.
Its owned fixture does not replace Gossip's pinned original-source comparison.

Fixed positional [array parameters](array-parameter-boundary.md) extend the same
local `gohp.cli.generate` operation; WebMCP remains excluded because this is
developer-only source translation, with no browser/server operation or auth route.

Capability/WebMCP scope: `gohp.cli.generate` is explicitly **excluded** from browser
WebMCP exposure because it is a local compiler writing caller-owned build files.

Bounded [array_push](array-write-boundary.md#bounded-array_push-calls) adds only
simulated PHP local/direct-property mutation and a count-result local inside
that same excluded generation capability. No service resource, CLI input schema,
auth route, browser operation or auto-submit is added. The result remains
generated/unverified with fresh/report-last artifact publication; no original
parser adoption, external send or deployment is implied by an admitted push.

Bounded [array_pop](array-write-boundary.md#bounded-array_pop-calls) likewise
changes only simulated local/direct-property array state and a mixed result
inside `gohp.cli.generate`. WebMCP remains explicitly excluded: no new browser/
server operation, input schema, resource permission, auth route, auto-submit or
deployment. Generation is not original-parser compatibility or application
acceptance; the existing fresh/report-last publication contract is unchanged.

The same local capability now has two exclusive input schemas: legacy
`{in, php_version, experimental_pcre2?, experimental_pcre2_profile?}` or bundle
`{manifest, source_root, entry, php_version, experimental_pcre2?, experimental_pcre2_profile?}`. Bundle paths
are explicit caller-owned local resources, with a private immutable source root.
Inspect/prepare reads and validates every selected source/notice/provenance byte,
resolves static declarations and performs semantic/codegen checks in memory.
Save/commit publishes only fresh caller-specified `out` and optional `report`,
using the existing report-last/no-overwrite/failure-cleanup contract. No artifact
is published for a failed closure. The result remains generated/unverified,
never saved-as-compatible, deployed or externally sent. This is not a new server
operation or browser/MCP route; WebMCP remains excluded, no auto-submit or auth
change. Partial/mixed modes reject before loading. General include/autoload is
outside the [bounded closure contract](static-source-loading.md).

`experimental_pcre2_profile` is an optional closed enum: `pcre2-php83-v1`
or `pcre2-php83-match1-v1`, requiring `experimental_pcre2=true` and PHP8.3.
Omission preserves the old default. Inspect/prepare rejects unknown/missing
opt-in/foreign-target profiles before source loading and records the selected
identity in IR/report; save/commit retains fresh report-last publication.
This extends the same `gohp.cli.generate` input contract, not a new service or
browser capability. Result remains generated/unverified; WebMCP excluded,
no auth/route/auto-submit/deployment or mutable budget operation is added.

[Guarded direct property writes](public-property-write.md) extend only simulated
PHP value operations inside this same excluded `gohp.cli.generate` capability.
Defined unknown locals require explicit runtime object/declared-slot checks,
with fresh artifact preparation/publication and generated/unverified result
unchanged. Full PropertyObject hosts explicitly opt in as for nested paths;
the protocol is not a generated-only provenance or authorization certificate.
No new resource/schema, UI/server/MCP route, auth mode or auto-submit is added.

[Guarded public-object snapshots](block-data-constructor.md) extend the same
WebMCP-excluded `gohp.cli.generate` capability to inferred-unknown cast operands.
The existing actual PublicPropertyObject protocol is required at runtime;
complete host implementations explicitly opt in, not as provenance or auth.
This adds no resource/input schema, route, auth mode, auto-submit or deployment.
Fresh/report-last publication and generated/unverified result remain unchanged;
an admitted snapshot is not complete-parser or original-source certification.

Bundle reports use `gohp.build/v0.2` and file-aware source-map v0.2. They bind the
entry, exact manifest hash, full compiled-source table, ordered loading graph,
unreached selection, profile and generated hash. Legacy single-file reports stay
v0.1; consumers must not treat one scalar input hash as a bundle identity.

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
and adds no service/auth route. Gossip's numeric-array original still needs
a separately pinned comparison of the complete dependency stack;
accepting `: bool` alone is not original-source compatibility or adoption.
The [bool-return branch reconciliation](bool-return-integration.md) merges #48
and #49 without duplicating the feature ID or runtime guard, preserving the old
evidence at its original revisions and requiring a new merged-configuration run.
The [array-parameter integration](array-parameter-integration.md) brings #51's
bounded array arguments onto that reconciled stack. Both parent fixture lists
remain intact, with explicit input/output guard composition checks. This does
not itself add callable parameters or promote Gossip's original-source compatibility.
The subsequent [arrow/typed-stack integration](arrow-typed-integration.md)
adds #53's bounded generated-callable implementation while retaining all #52
composition evidence. Original-source adoption remains a separate pinned check.

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
