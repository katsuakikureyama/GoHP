# Roadmap and application acceptance

This is the current priority/exit-criteria map, not a compatibility verdict or
a release promise. GoHP remains a general-purpose, source-generating compiler;
Gossip owns WordPress integration. The [compatibility inventory](compatibility.md)
describes implemented contracts, the [catalog](../compatibility/catalog.json)
links checks, and [run archives](compatibility-ledger.md) supply revision/target
evidence. This roadmap does not add catalog statuses or a new CLI operation.

## Current position

- PHP 8.3 and 8.4 are active closed-subset targets with independent pinned
  oracles. No result covers every patch, PHP feature or application.
- Functions, value arrays, simple classes and capture-free arrows generate Go.
  Internal scalar frames, shared scalar arrays, owned array-valued Slots,
  synchronous Slot-rooted array mutations, value temporaries and an opt-in
  runtime value adapter plus scalar-entry generated array scopes with bounded if/while
  are implemented; PHP references and general shared-array integration are not.
- [Gossip's earlier original-source comparisons](integration.md) are useful
  historical evidence at their recorded compiler/source pins. They are neither
  a current-main application result nor a combined WordPress percentage.
- The first whole-file [source manifest](application-corpus.md) is frozen and
  byte-verifiable, with a bounded native-token [lexical inventory](application-corpus.md#lexical-inventory).
  Candidate names now retain native-environment and selected-declaration name
  evidence, explicitly not dispatch resolution. Resolved syntax/callsite coverage and comparative performance
  are **not measured**; function/callsite/case denominators remain unfrozen. Show
  `n/a`, not zero, 100%, or a ratio derived from compiler test counts.

## Priorities and completion criteria

Owner priority, 2026-10-27: **make a fixed Gossip source work on PHP 8.3 first**.
The first executable gate is the complete original `WP_Block_Parser`, including
Block/Frame and source-loading edges, not another selected helper. Gossip's
[acceptance contract](https://github.com/katsuakikureyama/Gossip/blob/master/docs/php83-application-acceptance.md)
owns the frozen source, scenarios and status. Application semantics remain PHP
semantics, not WordPress-specific branches in GoHP.

| Workstream | Current state | Next observable result / completion criterion |
|---|---|---|
| PHP 8.3 Gossip acceptance (first) | Full original parser not translated; selected-method results remain separate | G83-PARSER: all seven original methods plus Block/Frame and loading closure, native/generated behavior at exact pins; then bootstrap/hooks, DB, HTTP and standard user flows |
| Static source closure (next measured blocker) | PR126 passes property isset384; original require_once403 rejects. [Loading contract](static-source-loading.md) has internal L0-a graph inspection only | Complete file-aware executable IR/diagnostics/maps/reports, then owned multi-file native comparison and one unchanged three-file parser probe; no concatenation or general include/autoload claim |
| Documentation | README and compatibility tables reorganized; historical target claims distinguished | Keep the entry point, implementation limits and evidence links aligned in each affected change |
| Application inventory | Earlier functions.php lexical/name evidence retained; no semantic coverage percentage | Work from the complete parser's actual calls and state mutations first; inventory bootstrap dependencies when that gate is next |
| PHP values/references | R1 partial; existing ownership/runtime adapter and exact-scalar-entry scopes retained | Implement only locations, aliasing and mutation contracts needed by the selected application gate; no independent array/untyped-parameter or COW expansion queue |
| Performance baseline | Deferred; no PHP/Go benchmark | After the first executable application unit, measure a reproducible output-checked workload; performance is not a substitute for correct application behavior |
| Frontend adapter | Assessment only; no candidate installed | Revisit if removing measured parser/application syntax blockers justifies migration; preserve target-aware diagnostics and fail-closed semantics |
| PHP versions | Existing 8.3/8.4 differential CI retained | Fix regressions in both; defer new minors and unrelated 8.4 feature expansion until the PHP 8.3 application gate advances |
| Broader semantics | Captures, inheritance, exceptions and include/autoload remain unsupported | Select from measured application blockers; each needs a closed contract and native comparison before admission |
| Licensing | Undecided | Owner selects terms after the existing dependency/contributor/provenance review; documentation cleanup does not select a license |

Select changes that remove an observed blocker in the current application gate.
Completing the storage model in isolation is no longer the next milestone. A builtin
addition still needs its own bounded contract; frequency alone does not justify
accepting unsupported argument types or call forms. Pair each increment with the
original-source callsite it enables and re-probe the complete unit after the
affected boundary changes. Do not re-run unchanged successful configurations.
Expected compiler rejection is a recorded blocker, never application PASS.

The next loading slice is limited to the complete original parser's two static
`require_once __DIR__ ...` edges and declaration-only dependencies. Its
[L0–L2 delivery gates](static-source-loading.md#delivery-gates)
separate graph/provenance infrastructure, measured closure behavior and original
source admission. L0 is partial; L1/L2 remain pending. Loading does not complete the parser's
remaining stack/object/tokenizer/JSON requirements.

## First application inventory

The **pinned WordPress core `wp-includes/functions.php`** remains the historical
first frequency/inventory candidate, not the next executable acceptance gate.
Its [manifest and byte verifier](application-corpus.md)
fix the complete source-file selection, notice and version provenance; they do
not parse, compile or execute it. Before publishing any ratio, complete the
remaining inventory/run evidence:

- Upstream project/release/revision, source path and byte hash, provenance and
  license notices; distinguish upstream identity from a retained Gossip copy.
- Exact selection/exclusion rules and all selected source units. Preserve
  unsupported and unparsed units in the denominator; do not extract only easy
  functions and describe that selection as the complete file.
- Direct/transitive dependencies, unresolved dynamic calls and external
  services/extensions. An inventory of this file is not a dependency-complete
  executable WordPress bundle. Unknown dependencies remain blockers.
- Compiler/runtime revision, parser/inventory tool revision, corpus/driver
  hashes, PHP target and exact oracle configuration for each run.

The semantic frontend currently rejects out-of-subset syntax. It therefore
cannot, by itself, claim a complete PHP callsite inventory. Use the
[parser evaluation](parser-adapter-assessment.md) or an explicitly bounded
inventory method with unknown/unparsed counts. Regex/name matches are not
resolved PHP calls and cannot yield a support percentage.
The new lexical inventory counts every token byte and opening parenthesis using
the selected native tokenizer. Declaration markers and name/parenthesis
frequencies are prioritization hints only. It records all source bytes as
grammar-unparsed, keeps unclassified parentheses visible, and does not define a
semantic function or resolved-call denominator.

The [top-six dependency triage](application-corpus.md#dependency-triage-separate-selection)
pins a separate seven-file inspection selection, including both classes required
by `plugin.php`. It locates hook, translation, option and path-helper definitions,
but does not resolve callbacks, cache/database/translation services or loading.
Its larger selection must not replace the original one-file denominator.
Hooks require mutable globals, general callables, reference-aware iteration and
reentrant mutation behavior; implementing two names as stubs would not satisfy
that contract. Keep these blockers ahead of unrelated helper-count growth.

Keep these measurements separate:

| Measurement | Denominator and evidence | What it does not mean |
|---|---|---|
| Lexical inventory | Every token and byte; every opening parenthesis, including unclassified forms; native version, flags and source/tool hashes | Function-keyword markers and name/parenthesis frequency are not resolved calls, parsed functions or support percentages |
| Syntax/callsite inventory | All occurrences in the fixed scope, with classified and unknown counts; show occurrence frequency separately from unique symbol counts | Recognizing a name does not establish support for that call's signature, values or effects |
| Original-unit compile ratio | Successfully generated **and built** unchanged units / all selected original units for that target; list blockers and missing dependencies | A successful build is not PHP equivalence; rewritten units and owned substitutes are separate categories |
| Case execution ratio | Passed / all applicable predefined cases at the fixed source/compiler/target; failed, incomplete and unmeasured cases stay visible | A small fixture set does not establish every input or an entire application's behavior |
| Application acceptance | Complete original unit plus dependency closure, all required cases/dimensions measured and passing, exact pins retained | Never derive this verdict from the above ratios or from another PHP target |

Show numerator, denominator, scope identity and unmeasured counts alongside any
percentage. Target-inapplicable cases must be explicitly accounted for, not
silently removed. Changing the manifest creates a new series of measurements;
do not compare differently selected denominators as a single progress curve.
Neither the Composer metadata reader nor source-byte verifier is a PHP
syntax/callsite inventory tool.

For executable application acceptance, retain the reference design's sequence:
the **complete original block parser**, then hooks and broader callers. The
existing `WP_Block_Parser_Block` data-class result is not a parser-algorithm
result. Original `WP_Hook` behavior must not be replaced by an owned hook mock;
the [render dependency audit](render-dependencies.md) similarly requires the
real HTML processor. Gossip owns source bundles, drivers and consumer repins;
this document changes none of them.

## Reference milestone

The R0–R5 model remains the design constraint. Its next implementation slice is
conditional on a demonstrated Gossip blocker, not an obligation to finish every
storage optimization before running an application. In particular, the parser's
property-held stack mutation and nested object identity must not be mistaken
for a requirement to migrate every array to COW first.

Use the existing [R0–R5 gates](reference-value-model.md#delivery-gates-and-status),
not a second storage engine. R1 now has synchronous one-level mutations through
the owning Slot, without a Read/copy/writeback temporary, and separately owned
value snapshots with single consumption/release. A value-only runtime adapter
now imports/exports canonical flat values without exposing storage handles. It
is emitted in proved nullary/all-exact-scalar-parameter scopes with if joins, domain-invariant while loops,
expression-level cleanup and lazy selected-arm acquisition. Broader loop/array-parameter scopes and
persistent locations still need explicit contracts
without changing PHP syntax acceptance. Internal unit results do not count as
reference support.

When an application blocker requires R2, its first end-to-end target is bounded scalar variable reference assignment,
write-through, rebinding and unset. For each admitted PHP target, compare source
programs and compiled Go for alias updates and surviving/recreated bindings;
also reject unsupported locations, malformed IR and by-reference forms. Keep
ordinary value-copy behavior unchanged. If arrays are not yet ready, reject
array/object references explicitly rather than silently using eager copies.
Array-element references and copy-timing migration remain the separate R3 gate;
captures and by-reference calls remain R4. Do not relabel the PHP 8.3-only native
reference specification as an 8.4 or generated-Go test.

## Performance baseline

Deferred behind the first executable application unit. When resumed, begin with
a repository-owned, bounded array-copy/read/write workload whose
deterministic result is checked against each selected PHP oracle. Document input
sizes and an operation count; avoid timing large stdout writes as computation.
Keep this a separate measurement, not a correctness or speedup gate.

Record CPU/OS/architecture, Go version/build flags, compiler/runtime/source hashes,
PHP version/image and relevant INI settings including CLI OPcache/JIT, memory
limits, warmup and repetition counts. Report build time separately from process
startup and execution; use repeated runs with a median and spread, not one best
sample. Measure peak process memory on both sides with the same method; Go-only
allocation counts are supplemental, not interchangeable with PHP peak memory.
If a dimension cannot be observed, mark it unmeasured. Retain slow results and
avoid extrapolating a microbenchmark to WordPress throughput.

## Frontend and dynamic-feature strategy

Keep the current frontend until the [adapter gates](parser-adapter-assessment.md#migration-gates)
pass for both active targets. A new parser recognizes syntax; semantic validation,
runtime behavior and fail-closed handling remain GoHP responsibilities. Candidate
versions and licenses in the assessment are dated observations, not an adoption
decision or a claim about current upstream support.

Continue generating Go functions/structs and explicit runtime operations. Dynamic
dispatch may be implemented in that runtime with separate contracts; an embedded
interpreter or silent source rewrite is not the default fallback. Whole-application
feasibility remains unproven. A future fallback or source-transformation mode
would be an explicit architectural/product decision with separate evidence.

## Documentation maintenance

Keep README short: purpose, active targets, scope, quickstart and links. Curate
function lists in the existing compatibility tables/catalog; put exact limits in
their linked contracts. Update architecture only for structural changes and this
roadmap for milestone/state changes. Preserve historical results at their original
pins, clearly labeled; never "refresh" their versions without new observations.
The [licensing proposal](licensing.md) remains undecided.
