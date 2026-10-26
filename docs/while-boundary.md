# Bounded braced while

Case-insensitive `while (condition) { body }` evaluates its condition before
every iteration, including after bare `continue;`. An initially false condition
runs no body. Bare `break;` exits without another condition evaluation. With
nested while/foreach loops, both transfers target the nearest lexical loop;
function/method loop context is independent of callers. Conditions use the
existing scalar/null/array truthiness domain, not object/Closure truthiness.

IR v0.1 adds `while` with `Expr` and `Body` only; hidden binding, target or else
payloads are rejected. Codegen emits an ordinary `for rt.Bool(condition)` with
statement source mapping, not an interpreter or cached condition. Existing
short-circuit evaluation is retained. Dead tails after transfers are validated
but not emitted, and output resumes after the enclosing loop.

## Loop-carried values

The initial condition is checked against incoming types. Existing variable
roots written anywhere in the body (including nested branches/loops, foreach
bindings and dead tails) are then conservatively widened to unknown for the
repeated condition, body entry and post-loop state. This avoids treating the
last textual assignment as the actual exit type across zero iterations,
continue or early break. Dynamic builtin/return/truthiness guards remain in
force. This is conservative validation, not path-sensitive type inference.

New body-only bindings cannot be used in the condition or after the loop.
Property writes do not replace their receiver; mutable reads already have
unknown type. A receiver variable reassigned within the loop loses its known
class at body entry, so bounded property assignment may require a fresh known
class assignment before use. Existing foreach snapshots and array copying are
unchanged.

## Limits and evidence

Only braced bodies and bare nearest-loop transfers are included. Alternative
syntax, empty/unbraced statements, assignment expressions in conditions,
for/do/switch, numeric transfer levels (even 1) and labels are excluded. No
termination proof, iteration cap, general warning/exception equivalence or
support for later PHP targets is claimed. Fixtures terminate by construction;
programs can still loop forever, as in PHP.

Catalog `php.syntax.while` is `partial`, with six checks: native differential,
lowering/source map, conservative type state, source rejection, malformed IR,
and generated runtime guards. Target PHP 8.3 uses oracle >= 8.3.36. Owned token
array selection composes `str_starts_with` with while/break but does not execute
the unchanged WordPress render caller. Stateful `strtok`, HTML processing and
exact-pin consumer adoption remain separate work. No Gossip pin/corpus changes.

Capability: existing developer-local `gohp.cli.generate`, WebMCP `excluded`.
No browser/server/MCP operation, route, authentication, deployment or auto-submit.

References: [PHP while](https://www.php.net/manual/en/control-structures.while.php),
[PHP continue](https://www.php.net/manual/en/control-structures.continue.php).
