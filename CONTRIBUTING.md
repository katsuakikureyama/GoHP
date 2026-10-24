# Development

Work on a feature branch and open a draft PR linked to #1. Do not merge to main,
release, deploy, fork upstream, or finalize licensing without owner approval.
English documents are authoritative; keep the Japanese README summary accurate.

Use Go 1.23.12 and PHP 8.3/7.4, or the Docker Compose commands in README.
Run `gofmt -w cmd internal runtime tests`, `go vet ./...`, and `sh scripts/check.sh`.
Both matrix jobs must execute real differential tests; missing PHP is a failure.
Fixtures must be repository-owned and offline-safe. Do not add network or
production credentials to tests. Go packages currently have no external dependencies.

## Small follow-up roadmap

1. Full PHP frontend adapter spike following the [assessment](docs/parser-adapter-assessment.md),
   structured diagnostics and expression spans beyond the initial
   [statement-level source maps](docs/source-maps.md).
2. Lvalue/reference model and array writes, copy-on-write differential cases.
3. Closures, callbacks, typed values, exceptions and classes beyond the subset.
4. Extend the [experimental compatibility ledger](docs/compatibility-ledger.md)
   with historical/combined-target evidence views and the Composer/extension
   metadata reader. Keep catalog case links and contract limits current.
5. Sandboxed versioned Gossip builder API, WordPress adapter fixtures and hooks.
6. Broaden PHP versions and compare side effects, warning/exception semantics.

A passing parser/build is not a compatibility claim. Each addition needs positive
native-PHP-vs-generated-Go tests and explicit negative boundary tests.
