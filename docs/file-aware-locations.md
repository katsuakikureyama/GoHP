# File-aware source locations (L0-b infrastructure)

This infrastructure is now used by the checked
[static declaration closure](static-source-loading.md), its explicit bundle CLI
and v0.2 build reports. Location metadata alone is not verified loading or native
compatibility evidence. Complete original-parser admission remains separate.

`frontend.ParseSource(sourceID, bytes, target, options)` uses the existing
filesystem-free single-file grammar and hashes exactly those original bytes.
It sets `ir.Program.Sources` metadata and identities on global declarations and
top-level statements. Methods/nested statements inherit their containing file;
class declarations also retain their original line. It does not parse load
directives or manufacture a combined PHP source. The ordinary `Parse` APIs and
CLI retain line-only behavior.

Located IR requires canonical relative slash source IDs, canonical SHA256,
nonnegative byte sizes, unique source table entries and listed declaration/body
identities. Nested scopes may inherit but cannot switch source. Metadata is
validated before semantic/codegen work; invalid metadata produces no Go/map.
Manually assembled IR metadata is **not** proof of verified source buffers;
the future loader must supply the checked closure and provenance itself.

Semantic and codegen diagnostics preserve the originating source ID with the
existing original PHP line. A located dependency error cannot be relabeled with
an outer caller's source. Syntax errors from ParseSource likewise carry that
source and return no partial program. Generic build/target errors without a
known source do not invent one. This is not file-aware PHP runtime exceptions.

`GenerateWithMap` selects `gohp.sourcemap/v0.2` only for located IR. Its source
table contains IDs/hashes/sizes, and each mapped statement range contains a
source ID plus original line. `PHPLocation` returns both; the legacy `PHPLine`
helper refuses v0.2 to avoid silently discarding file identity. Legacy maps stay
`gohp.sourcemap/v0.1`, omit all new JSON fields and keep PHPLine behavior.
Precision is statement-start file/line, not columns or expression coverage;
scaffolding and gaps remain unmapped.

Generated scope-marker comments contain only numeric line/source indices,
never filenames or raw PHP text. Source names with spaces/quotes cannot inject
markers. Nested mapping coalescence requires matching file identity as well as
line. No Go `//line` path remapping, source concatenation or runtime load flags
are introduced.

Catalog `gohp.tooling.file-aware-locations` records source/map/diagnostic/raw-IR
boundary tests and legacy report/map regressions. A test explicitly assembling
owned parsed declarations is a location test, **not** verified loading or native
multi-file compatibility. Existing PHP8.3/8.4 matrix remains a regression gate.
The internal adapter adds no user-facing operation; existing local CLI/WebMCP
exclusion stays. No filesystem reader, consumer repin, auth or deployment change.
