# Upstream assessment / ADR 0002 (2026-10-24, proposed)

The recommendation is a small independent implementation for Phase 0, with a
replaceable frontend. No upstream fork is approved or created. Source/licensing
claims below are limited to the pinned evidence in `upstream-evidence/*.json`:
repository HEADs, immutable README/LICENSE URLs and SHA-256. These records are
not a full dependency or legal audit. No upstream implementation is incorporated.

| Candidate | Evidence and target | Licensing observed | Decision |
|---|---|---|---|
| moriyoshi/php2go | README: harvest Zend compiled output; ordinary bundles execute it in a Go interpreter; Go 1.26.5, Bazel, PHP embed toolchain | MIT root license; bundled PHP/dependencies need separate audit | Useful semantic/runtime research; current execution model does not meet readable Go source lowering goal |
| i582/php2go | Go source generator; vendored PHP parser; homogeneous arrays; function-only scope documented; latest fetched commit dated 2020-08-24 | MIT root license; vendored parser provenance still needs audit | Closest reusable architecture; real build and sample below; reuse remains possible after frontend/runtime audit |
| VKCOM/kphp | README: native compilation using C++ code generation, static typing constraints; lacks some dynamic features | GPL-3.0 root license text | Semantic/reference research; not a Go backend; substantial retargeting work |
| jphp-group/jphp | README: PHP to JVM bytecode, Java runtime, intentionally different standard library goals | Apache-2.0 root license | Different runtime target; no JVM backend in GoHP v0 |
| CleverCloud/Quercus | Pinned README: Java PHP implementation, Gradle, Resin/Caucho origin | GPL-2.0 root license text; exact distribution terms need audit | Java target; no direct Go source pipeline demonstrated |
| nikic/PHP-Parser | README: PHP-based AST library, version-specific parsing, diagnostics and source positions | BSD-3-Clause root license | Strong future frontend adapter candidate; adds a build-time PHP/Composer dependency, not Go generation |

## Executed i582 experiment

Revision: `6a25f697acc7baf06e3750478095827b69a866f8`. Build with
Go 1.23.12, `go build -buildvcs=false -mod=readonly`. Initial module download
failed TLS verification inside the container; mounting the existing host CA bundle
read-only restored verification. TLS and Go module checksum verification stayed
enabled. `-buildvcs=false` avoids ownership-dependent Git stamping in the mounted
evaluation checkout; it does not disable artifact verification.

The isolated execution container had no network, production secrets, writable
root filesystem or Linux capabilities, and ran as UID 1000. Test input:

```php
<?php
function main() {
    $a = [2, 3];
    echo $a[0] + $a[1];
}
```

With input named `main.php`, generation succeeded. Building just its generated
`main.go` and running the native binary printed `5`; native PHP
`require "main.php"; main();` also printed `5`. One timed generation took
3,595,756 ns (~3.60 ms), measured using `date +%s%N` immediately around the
translator invocation, excluding container startup. This is a single smoke timing,
not a throughput benchmark or a performance comparison.

Observed limitations: input `tiny.php` generates `package tiny`, not a standalone
main package. Building the whole generated directory failed because the auxiliary
`core.go` was empty (`expected 'package', found 'EOF'`); explicitly building
`main.go` avoids that artifact for this sample. Translating this project's
`tests/fixtures/core.php` exited 2 with `panic: var not found` from the upstream
block walker. That fixture includes a function, heterogeneous ordered array,
foreach and a class; the experiment does not isolate which feature caused the
panic and is not a claim that all upstream class support was tested.

Other candidates were assessed through pinned documentation/licenses, not built
or benchmarked here. moriyoshi requires a separate Zend/Bazel environment; the
other targets cannot directly satisfy Go source generation. Broader maintenance,
benchmarking, transitive licensing and parser coverage audits remain follow-ups.

## Fork / reuse / independent implementation

A fork could inherit i582's grammar and lowering but also its scope/type/runtime
assumptions and maintenance burden. Partial reuse of an audited parser is the
most plausible next step. A small independent implementation now makes the
Go-source acceptance criterion executable without importing an interpreter or
prematurely committing to an unaudited dependency. Its grammar is intentionally
much smaller than the upstream parsers; it is not a long-term full-PHP parser
recommendation. Revisit this decision before implementing large language areas.
