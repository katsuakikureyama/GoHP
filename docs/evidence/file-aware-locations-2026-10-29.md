# L0-b file-aware locations — evidence (2026-10-29)

Clean code `64403be7dc111f80a96980db31b6cea522f68be9`, tree
`b4d663050ece320cf6723e785d495e1ed3d22eb6`, based on integrated main
`01c325845f96b974ddd0f9f9a0b61704f5db6c0f` (L0-a PR128). The worktree was
fast-forwarded from its public graph-feature base after PR128 merged; no graph
implementation/test/artifact duplication. Canonical L0-b scope6486903333.

Cached Go1.23.0 executable directly selected via PATH; Darwin/arm64, CGO0,
`GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOWORK=off`, existing shared
`GOCACHE=/private/tmp/gohp-go-cache`. One affected-package run:

```sh
go test -count=1 ./internal/codegen ./internal/frontend ./internal/semantic \
  ./internal/ir ./internal/compat ./cmd/gohp
```

PASS once: codegen0.562s, frontend0.842s, compat2.347s, CLI4.548s. Semantic
and IR packages compile but contain no direct test files; they are not counted
as separate test PASS. Five new tooling checks cover per-file source identity,
same-numbered lines across files, inherited method origin, numeric-only markers,
dependency semantic/codegen diagnostics, zero artifacts on failure, parse failure,
legacy v0.1 behavior and invalid metadata. No implementation/test failures.
An initial proposed lexical-rejection input was made unambiguously invalid
before verification; no test was executed against that initial draft.

Parents own design/production integration; a worker owned only the location
test file without running tests. Parent added combined entry/dependency
diagnostic cases, parser failure/metadata cases, then ran the affected suite
once at the committed code above. No independent/adversarial re-verifier.

This evidence-only addition changes no tested source/configuration. Existing
PHP8.3/PHP8.4 CI remains a regression gate; results are completed in the PR
comment without another evidence-only source change. These checks do not
establish PHP native multi-file loading or promote require_once to supported.

No original WordPress probe is rerun: this unit adds no load admission, graph
lowering or CLI bundle input. The last measured source-only blocker remains
require_once403 from PR126. Executable closure lowering, bundle build report/
publication, owned native loading comparison, unchanged three-file probe and
G83-PARSER remain pending. Original/pins/runtime/auth/routes/deployment unchanged.
No new dedicated cache directory or persistent build binary. Internal tooling
adds no user-facing operation; existing CLI/WebMCP exclusion remains.
