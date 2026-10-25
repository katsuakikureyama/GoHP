# Conditional and array compiler integration

This change combines the published bounded string/array stack (Draft #22,
`5be2e59c3bb05d98ddee2bcab700888a637b3965`) with the full ternary and
string-only stripslashes stack (Draft #23,
`0e49cc995913dfede0332566dd4dc11c5df16ff8`). Neither default branch is merged
and no runtime domain is widened. The separate checkout used for #22 is not
modified. New tests must establish the combined configuration; earlier stack
passes are historical inputs, not proof of this merge.

## Resolution and contract

- The complete builtin reservation set and both native fixture registries are
  retained. Array keys/values/existence/list, strlen, bounded formatting and
  replacement remain present alongside conditional/string-only unescaping.
- Semantic calls keep #22's one shared operand-validation pass. Stripslashes
  checks exact arity before indexing the validated types, then applies the
  existing actual-string/unknown boundary. Both conditional arms remain
  validated; runtime evaluation remains lazy with one condition evaluation.
- The parser, runtime implementations, PHP8.3-only admission policy, immutable
  native image and CI isolation settings are unchanged from their source stacks.

The new owned `conditional_array_integration.php` fixture combines guarded
string processing with array operations, null presence, nested values and
bounded string formatting/replacement. Marker output makes conditional
evaluation visible. This is stdout/normal-exit differential evidence only,
not general side-effect, warning, stderr or exception equivalence.

Existing corpus entries are retained. The new integration check is linked to
`php.syntax.full-ternary` and `php.builtin.stripslashes-string`; it does not turn
unsupported features into supported ones. Invalid domains keep existing static
and runtime guards. References, array writes, recursive callbacks, general
coercion, short ternary/coalescing and object truthiness remain excluded.

## Gossip adoption is separate

Gossip can pin the committed combined compiler/runtime and compile its twelve
selected original helpers in one independent translation unit. Such a test
does not modify the originals, repin historical units or prove whole-file/core,
HTTP, database, theme/plugin or hook integration. Its exact manifest, cases and
typed observations belong to the Gossip repository, not this native fixture.

`gohp.cli.generate` remains WebMCP-excluded local developer tooling under
[the integration contract](integration.md). No UI action, server operation,
authentication route, cookie forwarding, deployment, release or license change
is introduced. No devboard card is identified or marked Done.
