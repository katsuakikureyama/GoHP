# Reconcile the parallel bool-return branches

This integration preserves both published histories as merge parents:

- #49: `3cd91b559fefb88b737b07d0ff4eddad21e5fce3`, on the newer control-flow,
  negative-key isset and case-insensitive literal stack.
- #48: `bab3ade7c07d1c4b96ba55beedadeeac15753089`, independently based on #44.

Both implement the same actual-bool-only contract, canonical return annotation,
static known-type rejection, runtime guard and guarded fallthrough. The merged
compiler keeps #49's implementation and source-map/control-flow coverage. There
is one `BoolReturn` function and one `php.syntax.bool-return-type` feature entry;
no scalar coercion, inferred call types or new callable domain is introduced.

The two branches used some identical check IDs for independently authored tests.
Historical reports remain tied to their original commits; they are not rewritten
or counted as results for this merge. Current check reconciliation is explicit:

| #48 check | Merged check | Treatment |
|---|---|---|
| `diff.bool_return` | `diff.bool_return_composition` | #48 fixture retained byte-for-byte at a distinct path; #49 keeps `diff.bool_return` |
| `unit.bool_return.values` | `unit.bool_return.values` | Canonical test preserves both actual bool values |
| `negative.bool_return.domain` | `negative.bool_return.coercion` | All #48 runtime unit inputs are already covered by the broader #49 test |
| `compiler.bool_return.paths` | `compiler.bool_return.paths` | Canonical lowering/source-map check includes both arms and foreach |
| `negative.bool_return.source` | `negative.bool_return.source` | Supported-boundary rejection cases consolidated, without duplicate test functions |
| `negative.bool_return.ir` | `negative.bool_return.ir` | Noncanonical/unsupported annotation rejection consolidated |
| `negative.bool_return.runtime` | `negative.bool_return.runtime` | Canonical generated guards retain all domains; add unknown integer zero alongside one |

The retained fixture's Git blob is `c128fbc725c1075fef46595823d4f83cec6aa6f0`.
It observes explicit bool casts, empty foreach, true/false parameter passthrough,
a forward evaluation marker and typed-array/literal array_map composition.
The other fixture preserves elseif/else-if/else paths and copy isolation.
Parent PASS results do not substitute for a new merged-configuration run.

No parent PR is closed or overwritten. Main/release/deployment and Gossip source,
pins and prior evidence are unchanged. Typed array/callable arguments, arrow
callbacks and dynamic invocation still block the numeric-array original.
Local `gohp.cli.generate` remains WebMCP-excluded; no service/authentication change.
