# Property isset — local evidence (2026-10-29)

Clean code `53546a681006a42fe3cbde03bf2b84a15bb19843`, tree
`a8bf5182398c63d83d0f99355b7b7c56be7618b1`, base main `e5d32e2`.
This document is an evidence-only addition after the checks below; successful
checks are reused without rerunning unchanged code/configuration.

## Focused checks

Go1.23.0 Darwin/arm64, CGO0, cached toolchain executable at the front of PATH.
`GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOWORK=off`; default target8.3,
task-shared temporary GOCACHE. Repository-relative commands:

```sh
go test -count=1 -run '^Test(PropertyIsset|IssetForms|ArrayIsset|NestedArrayProbes)' ./internal/codegen ./runtime ./tests
go test -count=1 ./internal/compat
```

PASS once: compiler0.351s/runtime0.532s/generated external ABI2.810s;
catalog0.328s. New checks and affected old isset boundaries are covered. Two old
source-rejection cases now reject mixed property/index paths rather than newly
admitted pure property paths. The owned fixture was adjusted before verification
to keep a poison array receiver statically unknown and to predeclare a branch
local; no production/test failures occurred.

## Complete unchanged original source probe

One invocation at the same clean code, without output/report destinations:

```sh
go run ./cmd/gohp -in ORIGINAL_FULL_FILE -php-version 8.3 -experimental-pcre2
```

WordPress source `54106f6c545fc8b881f6b39b7963002bece6c28c`, whole
`wp-includes/class-wp-block-parser.php` SHA256
`8563621090de5e307431b1d29b78a9f75fd9ee238a0137b563a7db7b2e1f8d3b`.

Exit1:

```text
PHP002 line 403: unsupported statement (at "require_once")
exit status 1
```

The final line is the `go run` wrapper. Property isset384 passes source admission;
the next frontend stop is the original `require_once __DIR__ ...` dependency edge.
No source extraction/concatenation/rewrite or substitute implementation is used.
Full-unit semantic validation, generation/build/execution/native comparison are
not reached. Earlier source admission is not earlier semantic/runtime acceptance.

## Native comparison boundary

Local native PHP is unavailable. The new owned fixture and seven catalog checks
are registered for existing independent pinned PHP8.3.36/PHP8.4.27 CI. Those runs
and saved-artifact readback are pending at this local measurement checkpoint;
the PR result records their completion without changing tested source merely
to append run IDs. No enabled-PCRE or existing consumer/native suite rerun,
original-source modification, consumer repin or deployment change. Existing
developer CLI capability only; WebMCP excluded, no route/auth change.
