# Positional list assignment

P1 of the fixed Gossip parser gate needs a five-variable `list` assignment and
a two-variable capture unpack. This increment admits the generic statement
`list($first, $second) = $array;` with one or more bare local destinations.
The construct is case-insensitive. Its targets use integer keys 0, 1, …,
independently of insertion order; extra entries are ignored and present null
values remain null.

The RHS is evaluated once. A guarded runtime helper selects and clones every
requested value before any local is replaced. Assignments publish left to right:
duplicate destinations end with the last value, and a destination may also be
the RHS variable. Nested arrays are detached, while object pointers retain the
existing identity behavior. This is not a reference implementation or canonical
nested COW. List scopes remain in legacy map storage; storage optimizations are
not expanded.

IR `list_assign` uses only `Expr` for the RHS and `Target` of kind `list` with
nonempty bare-variable `Args`. No new generic expression form or address API is
introduced. Source and manually constructed IR are checked. RHS validation
happens before target definitions; list writes clear prior class and direct-null
facts, and while loops widen all written locals. Existing branch-local binding
limits and Execution propagation remain intact.

## Explicit boundary

Only actual arrays with all requested numeric keys are admitted. A scalar RHS,
ArrayAccess object or missing key fails explicitly at runtime (or statically for
a known wrong type) before destination writes. PHP can assign null when unpacking
scalars and emits warnings for missing keys; these behaviors are **not emulated**.
The guard is an unsupported-input boundary, not PHP-compatible error handling.

No holes, trailing commas, keyed/nested/bracket destructuring, references,
property/index destinations, assignment-expression return values or chaining.
Reserved/global locals are excluded. Runtime helper count must be positive.
The compiler does not infer selected array element types or guarantee general
exception/side-effect equivalence.

Catalog `php.syntax.positional-list-assignment` registers owned native differential
evidence separately for PHP8.3.36 and 8.4.27, plus lowering/source-map, malformed
IR, runtime-domain and scope/direct-null checks. Existing 8.4 regression coverage
is retained without new version expansion. Generic fixtures contain no WordPress
source. Full original parser execution remains blocked by other P1–P5 dependencies;
successful list fixtures are not application acceptance or a consumer repin.

Capability: existing developer-local `gohp.cli.generate`, WebMCP excluded. No new
server/UI/agent operation, route/auth contract, public deployment or dependency.

Reference: [PHP list documentation](https://www.php.net/manual/en/function.list.php).
