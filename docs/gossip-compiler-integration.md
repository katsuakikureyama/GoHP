# Conditional and array compiler integration

The earlier PR #24 combined the published bounded string/array stack (Draft #22,
`5be2e59c3bb05d98ddee2bcab700888a637b3965`) with the full ternary and
string-only stripslashes stack (Draft #23,
`0e49cc995913dfede0332566dd4dc11c5df16ff8`). That change left default branches
and the separate #22 checkout untouched. Its evidence belongs to that exact
configuration; earlier stack passes were not proof of that merge.

The earlier PR #26 integration combined that published PR #24 revision
`e5ba8c1db81d235cc177cc83699868d812abff7e` with PR #25's `is_array` predicate
`14383426911e42134780e098fec721ce68c0debb`. All earlier fixtures and guards
remain registered. No source-stack runtime or parser domain is widened.
The owned `predicate_conditional_integration.php` fixture uses `is_array` to
choose array values, guarded string unescaping, or an empty fallback. It checks
empty/list/associative/nested arrays, scalar/null/object fallbacks, observable
condition/arm evaluation and original-value preservation in one generated binary.
Both new catalog links identify the same single differential check. Prior PR #24
compiler evidence and Gossip's twelve-helper bundle remain pinned historical
evidence, not validation of this newer compiler revision or automatic adoption.

The current revision additionally integrates PR #27's literal callback mapping
and string-only addslashes. See [the callback boundary](array-map-boundary.md)
for its exact parent revisions, new composition fixture and separate evidence.

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

The retained `conditional_array_integration.php` fixture from PR #24 combines guarded
string processing with array operations, null presence, nested values and
bounded string formatting/replacement. Marker output makes conditional
evaluation visible. This is stdout/normal-exit differential evidence only,
not general side-effect, warning, stderr or exception equivalence.

Existing corpus entries are retained. That earlier integration check is linked to
`php.syntax.full-ternary` and `php.builtin.stripslashes-string`; it does not turn
unsupported features into supported ones. Invalid domains keep existing static
and runtime guards. References, array writes, general callbacks, coercion,
short ternary/coalescing and object truthiness remain excluded. Explicit finite
recursion through a literal declared callback is now admitted only by the
separate [one-array mapping contract](array-map-boundary.md).

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
