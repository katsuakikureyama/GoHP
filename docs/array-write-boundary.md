# Explicit single-index array writes

The compiler accepts a statement `$array[$key] = $value;` on an already-defined
local variable. The base must be an actual runtime Array and the explicit key
an int64 or string. Known non-array bases/non-int-string keys and undefined or
reserved variables fail before generation; unknown parameters/keys have runtime
guards. This is a bounded prerequisite for array traversal, not map_deep support.

The named variable is the only writable location. The parser emits `array_set`
with Name, one non-nil Index operand and the RHS Expr. Semantic validation rejects
malformed shapes and accidental index fields on other statements. Generated Go
materializes the key and RHS once and installs the returned array into that
variable. There is no dynamic lvalue/reference-cell system or interpreter.
Native fixtures measure key/RHS marker behavior on the pinned PHP8.3 oracle;
no general PHP expression-order guarantee is inferred.

ArraySet normalizes integer/string keys using the existing Key implementation.
Replacing a key retains its insertion position; inserting a new explicit key
places it last. The entire input array and inserted nested-array value are cloned
under the existing value model, so assignment copies, caller arguments and a
self-assigned array snapshot are independent. Foreach traverses the existing
snapshot from Entries while writes replace only the named local variable.
See the [PHP array reference](https://www.php.net/manual/en/language.types.array.php).

No `$array[]` append, implicit array creation from null/false/undefined, nested
write targets, property/string-offset writes, assignment expressions, chained or
compound assignment, references, global variables or weak key coercion. Existing
next-index metadata is preserved/updated without overflow, but this does not add
append support or claim PHP8.3 negative-index append behavior. No object alias,
cycle, arbitrary-depth or PHP warning/exception equivalence claim.

The new capability remains `gohp.cli.generate`, explicitly excluded from WebMCP
as local developer tooling. These are writes to generated program-local values,
not user-facing server resources. No route/authentication/cookie/automatic-submit
changes, default-branch merge, deployment, release or license decision. Gossip's
original-source manifests/results are not repinned or rerun by this change;
wp_unslash/map_deep still needs object traversal and dynamic callbacks.
