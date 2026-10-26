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
  Internal scalar frames, shared scalar arrays, owned array-valued Slots and
  synchronous Slot-rooted array mutations are implemented, but PHP references
  and generated shared-array integration are not.
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

| Workstream | Current state | Next observable result / completion criterion |
|---|---|---|
| Documentation | README and compatibility tables reorganized; historical target claims distinguished | Keep the entry point, implementation limits and evidence links aligned in each affected change |
| Application inventory | Whole-file identity, native-token/name evidence and top-six dependency triage available; grammar/name resolution and dependency closure unresolved, no percentage | Follow the observed hook/global/reference/callable prerequisites and AST/dependency investigation; retain unknown/unparsed content and separate selection identities |
| Reference/value integration | R1 partial: internal ownership/storage and synchronous Slot-rooted mutations exist | Finish generated temporary lifetime and locations across calls; then a PHP source → Go binary comparison of reference assignment/update/unset (R2), with rejected unsupported forms |
| Performance baseline | Planned; no PHP/Go benchmark | Publish one reproducible, output-checked small workload per target, separating build, startup, execution and memory measurements |
| Frontend adapter | Assessment only; no candidate installed | Audit/pin a candidate and demonstrate target-aware parsing, diagnostics/spans and supported-IR parity in shadow mode before replacement |
| Application acceptance | Earlier bounded Gossip runs only | Translate an unchanged complete application unit and its actual dependency closure; compare behavior at exact application/compiler/oracle pins |
| Broader semantics | Captures, inheritance, exceptions and include/autoload remain unsupported | Select from measured application blockers; each needs a closed contract and native comparison before admission |
| Licensing | Undecided | Owner selects terms after the existing dependency/contributor/provenance review; documentation cleanup does not select a license |

Prefer changes that remove an observed application blocker or complete the
reference model over an unranked sequence of unrelated builtins. A builtin
addition still needs its own bounded contract; frequency alone does not justify
accepting unsupported argument types or call forms. Parser investigation and a
small benchmark can proceed without blocking ownership work, but neither counts
as application compatibility.

## First application inventory

The **pinned WordPress core `wp-includes/functions.php`** is the first
frequency/inventory candidate. Its [manifest and byte verifier](application-corpus.md)
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

Use the existing [R0–R5 gates](reference-value-model.md#delivery-gates-and-status),
not a second storage engine. R1 now has synchronous one-level mutations through
the owning Slot, without a Read/copy/writeback temporary. Persistent locations
and generated temporary lifetimes still need explicit contracts without changing
PHP syntax acceptance. Its internal unit results do not count as reference support.

The first end-to-end R2 target is bounded scalar variable reference assignment,
write-through, rebinding and unset. For each admitted PHP target, compare source
programs and compiled Go for alias updates and surviving/recreated bindings;
also reject unsupported locations, malformed IR and by-reference forms. Keep
ordinary value-copy behavior unchanged. If arrays are not yet ready, reject
array/object references explicitly rather than silently using eager copies.
Array-element references and copy-timing migration remain the separate R3 gate;
captures and by-reference calls remain R4. Do not relabel the PHP 8.3-only native
reference specification as an 8.4 or generated-Go test.

## Performance baseline

Begin with a repository-owned, bounded array-copy/read/write workload whose
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
