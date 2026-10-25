# Generated arrows on the typed-parameter integration stack

Both published histories are retained as merge parents:

- #52: `212ed3f4a2af6351dacbfa725cebf0386ab4103e`, array parameters on the
  reconciled bool-return stack, including its composition fixture and guards.
- #53: `bd419a2c6c72e645f0f0d7d09aa01eea15f70913`, capture-free arrows,
  generated-callable parameters and variable invocation on its #50/#51 merge.

The #53 compiler/runtime implementation is retained unchanged. The differential
fixture list keeps every entry from both branches, including #52's
`array_parameter_bool_composition`. Existing stable check IDs are preserved;
parent PASS reports remain attached to their original revisions and are not
evidence for this merge.

The owned `arrow_typed_composition.php` exercises array and callable arguments
with actual-bool returns across conditional/foreach paths, explicit casts,
skipped callbacks and caller isolation. Generated-runtime rejection cases verify
that a callback's unknown result still passes through its enclosing bool-return
guard and a downstream array-parameter guard. The shared positive/negative
checks are linked to all three feature entries, counted once per unique ID.

The [arrow/callable boundary](arrow-callable-boundary.md) remains deliberately
partial: generated capture-free arrows only, variable receivers, exact arity,
no outer capture or general string/method/invokable-object callable registry.
Typed/static/default/reference/variadic arrows, anonymous functions, general
PHP error equivalence and automatic native-array_map arrow support are not
introduced. Existing array/bool contracts and both parent fixtures stay intact.

This is compiler-owned evidence, not a WordPress helper/polyfill replacement.
Gossip must separately pin and compare unchanged original source; its existing
pins, rejected-candidate evidence and successful original count are unchanged.
PHP remains 8.3 with native oracle >=8.3.36; 8.4+ needs separate policy/evidence.
No parent PR is closed or overwritten, and no main merge/release/deploy occurs.

Capability `gohp.cli.generate` remains WebMCP-excluded: local developer source
translation, no browser/API operation, authentication change or automatic
adoption. The merged revision requires its own fixed-oracle CI evidence.
