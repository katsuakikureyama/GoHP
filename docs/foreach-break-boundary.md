# Bounded foreach break

Bare `break;` (case-insensitive) exits the nearest lexically enclosing `foreach`
or [bounded while](while-boundary.md).
It may appear inside braced `if`/`elseif`/`else`, in nested foreach loops, and in
functions/methods containing their own loop. It preserves effects and assignments
performed before it, skips the rest of that iteration and all later iterations,
then resumes after that loop. An inner break does not exit its outer foreach.
It differs from `continue`, which proceeds to the next iteration.

IR v0.1 gains a payload-free `break` statement. Semantic validation rejects
expression/target/name/binding/body payloads, explicit numeric levels (even 1),
and break outside a function-local foreach/while, including dead branches. A called
function cannot inherit its caller's loop context. Codegen emits ordinary Go
break with statement source mapping; no new runtime control-flow protocol.

The same-block tail is still validated for unsupported operations and undefined
reads but not emitted. Its assignments cannot overwrite the type state at the
reachable break. Existing conservative branch/loop type merging remains in use:
branch-only/new loop bindings do not escape; changed existing types retain
runtime guards. No path-sensitive return/exit inference is added. Existing
foreach snapshot iteration and array copying are unchanged.

This was the first loop-exit increment; bounded while is now tracked separately.
It does not add `for`, `do`, `switch`, labels or multi-level break. The owned token-array fixture selects a
first `wp-elements-` prefix with existing `str_starts_with`; it is not the
unchanged WordPress render function. That source uses stateful `strtok` and a
while loop before calling an HTML processor. Tokenizer and HTML processing
remain missing; while support alone does not certify that caller.
Gossip source/pins/corpora are unchanged, with no consumer adoption implied.

Catalog `php.syntax.foreach_break` remains `partial`, with six differential,
source-map, type-state, source/IR rejection and generated guard checks. Only PHP
target 8.3 is validated (native oracle >= 8.3.36), not later PHP targets or
general warning/exception equivalence.

Capability: existing developer-local `gohp.cli.generate`, WebMCP `excluded`.
No browser/server/MCP operation, route, authentication, deployment or auto-submit.

Reference: [PHP break manual](https://www.php.net/manual/en/control-structures.break.php).
