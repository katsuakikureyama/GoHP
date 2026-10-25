# Array parameters on the reconciled bool-return stack

This integration retains both published histories as merge parents:

- #50: `58943797038a1c87215acef265fb10951d94ec7e`, the reconciled #48/#49
  actual-bool return stack with both native fixtures and unique rejection cases.
- #51: `918c4eb47b4d7070ed24f26bcfd5b290030bb4e2`, fixed positional array
  parameters on named global functions, originally based on #49.

The compiler/runtime implementation from #51 is unchanged. Each function checks
and copies array arguments before its body; bool returns still require actual
booleans. The fixture registration conflict is resolved by retaining both lists,
not by dropping the #48 composition fixture or the new array-parameter fixture.
All existing check IDs and parent fixtures remain intact. Parent PASS reports
stay tied to their original revisions, not to the merged configuration.

`array_parameter_bool_composition.php` adds owned native coverage for their
combined contract: array-typed callbacks returning bool through conditional and
foreach paths, explicit casts, empty mapping, single evaluation and caller copy
isolation. Its shared check is linked to both feature entries and counted once.
A static rejection case verifies that an array-typed argument cannot be returned
directly as bool from an else branch. A generated-runtime case verifies that a
valid array argument does not bypass the bool-return guard when a callback
returns an unknown integer element. These checks are about input/output type
boundaries, not a claim of general callable or PHP error compatibility.

The [array-parameter boundary](array-parameter-boundary.md) and
[bool-return boundary](bool-return-boundary.md) remain partial. Callable/scalar
parameters, arrow callbacks, dynamic invocation, nullable/union/intersection
types, references, defaults and variadics remain unsupported. No WordPress
original, Gossip pin or successful original-source count is changed; adoption
requires separate pinned verification. PHP 8.4+ requires separate policy and
evidence. No parent PR is closed or overwritten, and no main merge, release or
deployment is performed.

Capability `gohp.cli.generate` remains WebMCP-excluded: local developer source
translation, no browser surface, service operation, authentication change or
automatic adoption. The merged commit must have its own fixed PHP 8.3 evidence.
