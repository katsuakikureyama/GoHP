# Bounded Boolean return declarations

Named global functions may declare `: bool`, with case-insensitive source
spelling normalized to `Function.ReturnType = "bool"` in IR0.1. This extends
the existing optional return-type field without changing untyped or `: array`
behavior. Noncanonical direct IR types still fail closed.

This initial contract accepts **actual Boolean return values only**. Each
explicit return, including nested `if`/`foreach` returns, evaluates its expression
once and calls `rt.BoolReturn`. Known non-Boolean returns are rejected before
generation, including bare/null returns, invalid returns in dead branches,
integers, strings, arrays and objects. Unknown parameter/call results are checked
at runtime. Fallthrough calls `rt.BoolReturn(nil)` and fails explicitly rather
than silently returning null or false. No path-completeness or return-type
inference at callers is introduced.

PHP normally permits some scalar coercions for scalar return declarations;
see the [official type declaration reference](https://www.php.net/manual/en/language.types.declarations.php#language.types.declarations.strict).
Those weak coercions are deliberately **outside** this subset. For example,
`return 1` under `: bool` is rejected by GoHP, not claimed to be invalid PHP.
An explicit `(bool)` expression already in the supported language may produce
an actual bool, but this guard does not perform that cast implicitly. This is
not support for `declare(strict_types=1)`, nullable/union/intersection types,
literal true/false return types, the `boolean` alias, typed parameters or typed
methods. Runtime panic does not reproduce PHP TypeError text, stack or
catchability, and a negative guard test is not native differential evidence.

## Evidence contract

Runtime tests check true/false preservation and non-bool rejection. Compiler
tests retain the normalized annotation, nested/fallthrough guards, forward
call guard and return source mapping, and reject unsupported source/direct IR.
Generated runtime-negative cases check unknown invalid returns and fallthrough.
The owned native `bool_return.php` fixture observes strict Boolean results,
early returns, empty loops, forward/nested calls, evaluation markers and
composition with typed arrays and literal named `array_map` callbacks. These
are case-limited stdout/normal-exit observations, not PHP exception equivalence.

The registry ID is `php.syntax.bool-return-type`, implementation `partial`.
It links actual executed checks; a new native result is needed for this
revision, not inferred from earlier array-return or Boolean-cast results.
Earlier successful configurations are not rewritten as new evidence.

## Gossip dependency boundary

This is the first compiler prerequisite identified by Gossip#25's unchanged
`wp_is_numeric_array` reference. It does **not** yet execute that WordPress
helper: its two-argument arrow callback and `array_all` dependency remain
unsupported, as do the original PHP8.3 polyfill's typed parameters/dynamic calls.
No WordPress source is rewritten, no compatibility polyfill is stubbed and no
PHP target/version policy is changed. Gossip#25 keeps its old compiler pin and
rejection evidence; adoption requires a separately pinned original-source run.

Local capability `gohp.cli.generate` remains explicitly **WebMCP-excluded**:
trusted local developer source translation, not a user-facing server operation,
browser/S2S/MCP route, session forwarding, automatic submit or deployment API.
There is no merge, release, deployment or full-WordPress compatibility claim.
