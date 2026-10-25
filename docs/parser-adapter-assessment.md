# Full PHP parser adapter assessment (2026-10-24, proposed)

## Decision

Keep the closed, repository-owned frontend as the default. Do not add a parser
dependency yet. A parser can recognise syntax and preserve a location; it
cannot make GoHP's runtime, lvalues, references, coercions, warnings,
exceptions, extensions, or include behaviour PHP-compatible. Every newly
lowered construct still needs a native-PHP differential case and an explicit
unsupported boundary.

`rectorphp/php-parser-in-go` is the first native-Go evaluation candidate, not
an approved dependency. At commit
`662f67042360727216ab55bbba0f640353258387` (2026-10-15), its README advertises
PHP 5/7/8 AST parsing and shows `version.New("8.3")`; its version package
accepts 7.0--7.4 and 8.0-or-newer. Nodes carry start/end lines and byte
positions. But that revision declares `go 1.26`, while GoHP declares Go 1.23,
and its README says it is based on z7zmey. Pin, toolchain, generated-grammar,
transitive-dependency, licence/provenance, API-stability, and active-PHP-8.3
corpus results require an owner-approved audit before use. The observed root
licence is MIT; that alone is not a full provenance audit.

`nikic/PHP-Parser` remains the best non-Go reference/sidecar candidate. At
`8d2b672aff65a1a4dbf7936cec27d19f51989412` (2026-10-05), its README says 5.x
runs on PHP >=7.4 and parses PHP 7.0--8.4; its usage documentation specifies
line, token, and file-position attributes. Its `composer.json` requires PHP
>=7.4 plus tokenizer and JSON. A JSON-AST sidecar would therefore add a pinned
PHP/Composer build dependency and protocol to a Go compiler; it is not a
drop-in pure-Go replacement. The observed root licence is BSD-3-Clause; its
dependency and distribution terms remain unaudited.

`z7zmey/php-parser` is not a direct candidate: its latest observed commit is
`367eff9de651fb9479c98f4f3741db3195e549ad` (2021-02-13), its README promises
PHP 5/7 and lists PHP 8 only on the roadmap, despite a 7.4 CLI default and
position-capable dumper. It has an MIT root licence. This is useful lineage for
an audit, not evidence of PHP 8.3 coverage.

Sources are immutable GitHub views for
[rectorphp README](https://github.com/rectorphp/php-parser-in-go/blob/662f67042360727216ab55bbba0f640353258387/README.md),
[version policy](https://github.com/rectorphp/php-parser-in-go/blob/662f67042360727216ab55bbba0f640353258387/pkg/version/version.go),
[positions](https://github.com/rectorphp/php-parser-in-go/blob/662f67042360727216ab55bbba0f640353258387/pkg/position/position.go),
[go.mod](https://github.com/rectorphp/php-parser-in-go/blob/662f67042360727216ab55bbba0f640353258387/go.mod),
[nikic README](https://github.com/nikic/PHP-Parser/blob/8d2b672aff65a1a4dbf7936cec27d19f51989412/README.md),
[nikic usage](https://github.com/nikic/PHP-Parser/blob/8d2b672aff65a1a4dbf7936cec27d19f51989412/doc/2_Usage_of_basic_components.markdown),
and [z7zmey README](https://github.com/z7zmey/php-parser/blob/367eff9de651fb9479c98f4f3741db3195e549ad/README.md).

## Required adapter contract

The backend must consume GoHP-owned data, never an upstream AST directly.
This is a proposed interface, not implemented by this change. The current
[source map](source-maps.md) records statement start lines only; full spans and
structured diagnostics remain migration work.

```go
type Span struct {
	Source string
	StartLine, StartColumn, StartByte int
	EndLine, EndColumn, EndByte int
}

type Frontend interface {
	Parse(source []byte, sourceName, phpTarget string) (ir.Program, []Diagnostic, error)
}
```

Each lowered IR node must retain its `Span`; diagnostics need a stable code,
severity, message and span. The adapter must reject, rather than omit or
silently rewrite, an AST form outside the supported semantic subset. It must
select the requested PHP target and validate missing/malformed positions.

Code generation must emit a separate versioned map from final formatted Go
line/column ranges to PHP spans, including generated-only ranges marked as
such. It must not imply that a Go panic has PHP semantics. CLI JSON,
diagnostics and maps must name and hash the source so a report cannot be
joined to a different input. `//line` directives are optional debugging output,
not a replacement for this artifact contract.

## Migration gates

1. Freeze an active 8.3 parse corpus: accepted syntax, parse-error locations, and
   current supported fixtures. Record parser revision and toolchains.
2. Add the frontend interface, spans and diagnostics while the existing parser
   remains default; run the candidate in shadow mode and compare only the
   supported subset's normalised IR.
3. Audit and pin exactly one parser only after Go-version and licence gates
   pass. Test parser positions and target selection separately from GoHP
   semantics.
4. Ship source-map schema and current-subset lowering parity before enabling
   it by default. Add syntax one construct at a time only with lowering,
   differential behaviour, map, and negative-boundary tests.

No candidate was installed, built, or benchmarked for this assessment; no
claim is made that any candidate fully parses or behaves like active PHP 8.3 on
the GoHP corpus.
