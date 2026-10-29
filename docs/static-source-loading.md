# Static source loading — first implementation contract

Status: **design only; loading is not implemented or supported**. No new CLI
flag, filesystem access, IR admission or compatibility-catalog PASS is added by
this document. This is the next bounded implementation plan for Gossip's fixed
PHP 8.3 application, not a general PHP include/autoload design.

## Measured starting point

[GoHP PR126](https://github.com/katsuakikureyama/GoHP/pull/126), integrated at
`9eec4d502ae396e389f94bf40d986b8d933caac0`, adds named-property `isset`.
Its [saved original-file probe](evidence/property-isset-2026-10-29.md) passes
line384 and rejects `require_once` at403. The successful native fixtures and
the source probe are reused, not rerun for this design-only change.
[Gossip PR58](https://github.com/katsuakikureyama/Gossip/pull/58) records the
same boundary. Full semantic validation, generation, build and execution of
the parser have **not** been reached.

The retained WordPress revision is
`54106f6c545fc8b881f6b39b7963002bece6c28c`. The whole-file source selection is:

| Source under `wp-includes/` | Loading role | SHA256 |
|---|---|---|
| `class-wp-block-parser.php` | Entry; line403 requires Block, line410 requires Frame | `8563621090de5e307431b1d29b78a9f75fd9ee238a0137b563a7db7b2e1f8d3b` |
| `class-wp-block-parser-block.php` | Complete declaration-only dependency; no further loading edges | `7e21f483cf0c9deddd2144e026842bf906c72983908f7bd163c0716e8f6f325e` |
| `class-wp-block-parser-frame.php` | Complete declaration-only dependency; no further loading edges | `e5883d82dff099a4b1a23786e9a53b7e29aa73654db83e7ddc444ba17df5e105` |

The original directives are:

```php
require_once __DIR__ . '/class-wp-block-parser-block.php';
require_once __DIR__ . '/class-wp-block-parser-frame.php';
```

Both dependencies contain a complete class declaration, not top-level
initialization code. Their method bodies still need ordinary subset validation;
declaration-only does not mean those bodies compile or behave correctly.
The [Gossip acceptance contract](https://github.com/katsuakikureyama/Gossip/blob/master/docs/php83-application-acceptance.md)
remains the authority for the complete application gate and its case denominator.
These hashes identify inspected source bytes, not a new executable result.

## Decision and PHP boundary

Implement a generic, opt-in **static declaration closure**, with the source
selection and application driver owned by Gossip. There must be no WordPress
class names, paths or replacement implementations inside GoHP.

PHP `require_once` avoids including a previously included file; this is more
than merely finding its text. See the [PHP manual](https://www.php.net/manual/en/function.require-once.php).
`__DIR__` belongs to the file containing the expression, not the process working
directory or entry file. See [magic constants](https://www.php.net/manual/en/language.constants.magic.php).
Included files can also inherit the calling scope and execute code; see
[include semantics](https://www.php.net/manual/en/function.include.php).
Those general behaviors are why the first profile must reject executable
dependency bodies and loading from inside functions or conditional branches.

The first proposed grammar is a standalone, top-level
`require_once __DIR__ . '/literal/path.php';` directive. The slash starts the
suffix appended to the containing file's directory; it is not a separate
absolute-path input. Other spellings/forms are outside the profile until
explicitly tested. `__DIR__` is recognized only in this directive, not silently
admitted as a general runtime expression.

Admit declaration-only modules with optional trailing directives. An explicitly
owned driver may have a leading load section followed by ordinary executable
statements; no load may follow executable code. Reject an executable statement
before a later load, even if a particular test would never reach it. This keeps
all dependency declarations available before driver execution without changing
the visibility/order of side-effecting source operations.

Every required module must contain only admitted top-level function/class
declarations and admitted trailing load directives. Reject top-level output,
assignments, returns, conditional declarations/loading and initializers with
effects. The existing declaration grammar still decides which defaults,
signatures, properties and bodies are representable. Do not skip method bodies.

Repeated edges and a diamond must resolve a module once within one compilation.
The cache key is its validated root-relative source identity, not its basename,
class name or content hash. Equal bytes at distinct identities must not be
silently merged. Duplicate declarations from distinct modules remain errors.
Cycles are deliberately rejected in the first slice; PHP's broader cyclic-load
behavior is not claimed. No process-global included-file cache is introduced.
This profile has no runtime initialization effects, so it needs no fake runtime
once flag. Supporting observable initialization later requires its own ordered,
execution-owned contract and native measurements.

## Implementation boundaries

1. **Checked source buffers.** Extend the existing manifest/checked-buffer
   boundary rather than inventing an unbounded filesystem reader.
   [`corpus.LoadSources`](../internal/corpus/corpus.go) already validates every
   selected whole file before returning retained source bytes. It is not a
   compiler or dependency resolver today. Keep provenance/notices and compile
   entry/edge selection explicit; do not repurpose historical inventory
   manifests or their denominators. A missing or unlisted dependency fails.
   The loader must consume those verified buffers, never reopen unchecked
   paths after validation. Require a private immutable build root, bounded
   regular-file inputs and existing byte limits; reject unsupported aliases
   and escaping paths. Do not turn trusted-root verification into a claim of
   an operating-system sandbox against hostile concurrent mutation.
2. **Per-file parsing and graph resolution.** Keep the current
   `frontend.ParseWithOptions` string-only and filesystem-free. Introduce a
   file-aware module adapter that parses declarations and explicit load edges
   while retaining original file/line locations. Each edge carries origin,
   source line, literal suffix and resolved destination. Resolve from the
   containing module's directory, validate against the selected allowlist and
   traverse in source order with unseen/visiting/loaded states. Resolve all
   modules with one selected PHP target and regex admission profile; do not
   infer either from the host PHP installation.
3. **Validate closure, then lower declarations.** Only after every module and
   edge passes the closed profile may the adapter construct a combined
   semantic program. Existing semantic/codegen passes already gather all
   classes/functions before checking bodies. Preserve that behavior and
   reject collisions before producing Go. No unresolved load node may reach
   codegen or be silently discarded by a raw-IR path. Do not strip directives,
   extract methods, concatenate PHP strings or run a source preprocessor.
   The dependency graph remains part of provenance even when no runtime load
   instruction is needed. A failed dependency leaves no new output/report.
4. **File-aware locations and reports.** Current
   [`ir.Program`](../internal/ir/ir.go) carries no source-file identity;
   [`SourceMap`](../internal/codegen/sourcemap.go) maps only PHP line numbers,
   and the CLI report has a single input/hash. A multi-file implementation
   must first retain file identity on source locations through semantic
   diagnostics and codegen. Never relabel a dependency's line as the entry's
   line or present synthetic concatenation lines as original locations.
   The proposed bundle report/source map use new versioned schemas (candidate
   names `gohp.build/v0.2` and `gohp.sourcemap/v0.2`), with source IDs, relative
   paths, exact hashes/sizes, manifest identity, ordered edges and entry ID.
   Define a file-and-line lookup; the existing line-only `PHPLine` helper is
   not a complete multi-file location API. Preserve legacy single-file v0.1
   reports and their tests. No schema is changed by this design document.
5. **Publication and capabilities.** A future explicit bundle input mode may
   extend the local `gohp.cli.generate` capability. It must document its exact
   input schema, read scope, validation/prepare and fresh artifact publication
   stages in the [integration contract](integration.md). Keep report-last,
   fresh/distinct destinations, no overwrite and failure cleanup semantics.
   The report must continue to distinguish generated from compatibility PASS.
   WebMCP stays explicitly excluded as local developer tooling; do not add a
   browser/server/MCP operation, auth route or auto-submit path.

Bare filenames/CWD or `include_path` lookup, arbitrary absolute paths, `..` path
components, symlink/hardlink aliases, stream/network wrappers, dynamic expressions,
require return values, `include`, ordinary `require`, autoload and runtime file
discovery are outside the initial profile. Reject, do not approximate them.

## Delivery gates — all pending

| Slice | Required observable result | Status |
|---|---|---|
| L0: source identity and module graph | Closed per-file parser, verified buffers, deterministic edges/once identity and file-aware diagnostics/report/map; reject unresolved/raw load nodes | Not implemented |
| L1: bounded declaration closure | One owned multi-file fixture compiled and built; behavior independently compared with pinned PHP8.3; duplicate/diamond loads resolve once; all unsupported forms fail closed | Not implemented |
| L2: original three-file admission | One probe of the unchanged complete parser plus both dependencies through the real loader; retain all hashes and the first real frontend/semantic blocker | Not run |
| G83-PARSER | All original methods and complete dependencies built/executed with the declared native comparisons and state/error observations | Blocked; loading is only one dependency |

L0 and L1 may ship together when locations and executable evidence are ready;
graph discovery alone must not become a supported `require_once` catalog row.
Do not emit placeholder initializers or success artifacts for a partially loaded
closure. Only the tested profile can be marked implemented, never general PHP
source loading.

Focused verification for L0/L1 must cover the actual contracts, not a second
verification framework:

- A real owned driver and dependency files exercising forward class references,
  nested containing-directory resolution, repeated edges and a diamond. Run
  native PHP on those real files, not a concatenated substitute. Compare typed
  ordered values and exact stdout, exits and the established quiet-stderr profile.
- Unknown/missing/changed source buffers, escaping/alias/nonregular paths,
  dynamic/conditional/local loads, load-after-execution, top-level effects,
  cycles and distinct-module duplicate declarations. Check that failures do
  not publish artifacts or read outside the declared source selection.
- Two dependencies with the same numeric line but different source IDs, including
  a dependency-local error. Check original file/line diagnostics and generated
  mappings; retain single-file report/map behavior.
- Closure/target/profile consistency and ordinary source-only/raw-IR rejection.
  Preserve existing PHP8.4 regression checks. Offering the new loader for8.4
  requires its own native fixture result; PHP8.3 is the first application gate,
  not evidence for another series.

After the affected implementation changes, run each relevant configuration once
and reuse PASS. Record exact commit/tree, manifest/driver/source hashes, target,
oracle digest/configuration, commands and failures/unmeasured phases. Then probe
the original closure once and record the next real blocker, not an artificial
PASS. Remaining stack/object/snapshot/tokenizer/JSON requirements stay open even
if loading succeeds. No successful compiler/native tests or original probe are
rerun for this design-only change.

Use the existing task-shared build cache, temporary test-owned output directories
and one worktree per change. Remove a clean merged worktree after integration;
preserve unmerged work and durable evidence. This plan does not create binaries,
new dependency caches, Docker images, a consumer repin or deployment changes.
