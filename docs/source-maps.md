# Statement-level source maps

`-report` now includes `output` (the CLI path as supplied) and `source_map`:

```json
{
  "schema": "gohp.sourcemap/v0.1",
  "precision": "statement-start-line",
  "mappings": [
    {"generated_start_line": 12, "generated_end_line": 12, "source_line": 2}
  ]
}
```

This is a GoHP-specific format, not Source Map v3. The example's Go line is
illustrative; actual ranges are computed **after gofmt** from the exact published
bytes. All line numbers are one-based. Generated ranges are inclusive, sorted and
non-overlapping. Each maps to the starting PHP line of an assignment, echo,
return or foreach statement. Multiline expressions map to the enclosing statement
start, not to an expression column or the exact failing operator. Empty programs
produce `mappings: []`. Unknown IR positions are omitted, not invented.

Nested statements override enclosing foreach scopes. The loop header, generated
key/value bindings and closing brace map to the foreach start; its body's
statements map to their own starts. Function/method bodies are covered, but their
synthetic declarations, parameter setup, implicit returns, class constructors,
property access helpers and other runtime scaffolding are not. Comments and blank
lines are unmapped. Never extrapolate a previous entry into an unmapped gap.

Ordinary `// gohp:begin N` / `// gohp:end` scope comments remain in generated Go
to keep mapping construction independent of gofmt layout. They are not Go `//line`
directives and contain neither source paths nor PHP text. PHP strings are quoted
as Go literals, so text resembling a marker cannot inject a scope. Generation
without `-report` emits the same Go; it simply does not publish the map.

## Consumer contract

1. Require a successful CLI exit and verify the report's `source_sha256` and
   `generated_sha256` against the exact files. Paths are labels, not trusted
   filesystem locations; resolve them within the caller-owned build workspace.
2. Recognize both schemas (`gohp.build/v0.1` and `gohp.sourcemap/v0.1`) and the
   precision. Treat a missing map in an older build report as unavailable.
3. For a compiler diagnostic or panic frame in **this generated file**, look up
   its physical Go line in an inclusive range. Display the corresponding
   `input:source_line`, explicitly labeled statement-level. If no range matches,
   preserve the Go location rather than inventing a PHP location.

Do not apply this map after editing/reformatting the generated source, or to a
different build/runtime file. Optimized/inlined Go traces may omit frames; the
map does not guarantee every panic is translatable. It does not rewrite compiler
diagnostics or stack traces automatically. No standard debugger integration,
column spans, multiple PHP files, structured frontend diagnostics or complete
PHP semantic compatibility is claimed.

Tests check nested scopes, functions/methods, multiline strings, marker-like PHP
text, unknown/unmapped locations, report hashes and exact published line numbers.
The runtime-negative fixture additionally resolves a real Go panic frame back
to the failing PHP statement; existing native-PHP differential tests remain the
semantic regression gate. `gohp.cli.generate` remains WebMCP-excluded as documented
in [the integration contract](integration.md); no new operation or auth route is added.
