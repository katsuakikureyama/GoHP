# Charset comparison and typed-array integration

This dedicated branch combines published strcasecmp/logical-comparison #40
at `d9f6c7b4f288ef816c4ee3cb6a3458d2226e00d5` with isset/array-return #36
at `db3051f8a179078e2227c5d97f50a2f1611d1451`. Both are retained as merge
parents. The separate if/else branch #39 is not included.

Conflict resolution preserves the union of reserved names (`isset` and
`strcasecmp`), all original native fixtures, and both specification rows.
The parser's return declaration/isset IR, semantic type/arity guards, all
typed-return paths and existing logical precedence/short-circuit rules remain
intact. No broadened PHP language domain is introduced by integration.

`charset_array_integration.php` is an owned cross-stack fixture: it selects
existing non-null string/array entries with isset and logical conjunction/OR,
returns through `: array`, shallow-unescapes and maps an ASCII charset
predicate. It observes ordered mixed/negative keys, nested-array rejection,
null/missing/falsy exclusions, duplicate selectors, a forward typed return,
copy isolation and a short-circuited missing index. Existing static/runtime
negative tests for each merged feature remain registered. The shared native
check links to the three existing feature IDs and counts once.

Neither parents' PASS alone nor successful generation proves the merged
configuration. The merged head needs its own fixed-PHP8.3.36 CI. The native
schema compares stdout/normal exits, not general stderr, PHP exceptions or
external effects. Gossip separately pins the resulting compiler to execute
unchanged WordPress originals and compare actual typed results/arguments.
That new bundle is independent of all prior Gossip units and evidence.

General isset, typed parameters/methods, other return types, weak comparison
coercions, Unicode folding, objects/references, public option-dependent charset
wrappers and full WordPress remain outside the established boundary. This
change does not deploy, merge to main, release or finalize licensing.
Local `gohp.cli.generate` remains WebMCP-excluded; no service/browser/MCP route,
authentication or cookie forwarding change. Devboard unidentified/unsynchronized.
