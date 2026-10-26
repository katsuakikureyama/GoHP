# Literal callback array_map and string-only addslashes

This bounded extension supports `array_map('declared_name', $array)` with
exactly one input array and a literal name of a declared one-argument global
function. Names are case-insensitive; declarations may appear later and may
refer recursively to themselves. The compiler resolves the callback at build
time and emits a direct Go-function adapter, not a callable registry, PHP
interpreter or dynamic dispatch facility.

The [PHP array_map contract](https://www.php.net/manual/en/function.array-map.php)
preserves input keys for one array. This implementation retains normalized
keys, insertion order and next-index metadata. Each entry is passed once by
value through the existing nested-array Clone boundary; the result is cloned
into a fresh output array. Empty arrays make no callback calls. The input array
is not modified. Native fixtures cover ordered keyed values and finite recursive
lists; runtime tests cover additional copy/metadata boundaries.

Only a global user-declared function with exactly one required value parameter
is admitted. Dynamic callback variables/expressions, builtin callbacks, null,
closures, method/static callbacks, namespaces, references and multiple arrays
remain unsupported, including for empty inputs. There is no automatic recursive
mapping: recursion happens only through ordinary generated calls made by the
selected function. General object aliasing and PHP callback warning/exception
equivalence are not claimed. Arbitrary-depth/cyclic inputs and stack exhaustion
are not compatibility guarantees.

`addslashes` accepts exactly one actual string, with static/dynamic guards.
It processes bytes once, escaping single/double quotes and backslashes and
encoding NUL as backslash plus ASCII zero. Other bytes are unchanged, without
Unicode decoding or weak scalar conversion. The [PHP manual](https://www.php.net/manual/en/function.addslashes.php)
also distinguishes this operation from database escaping. This is not SQL,
HTML or shell sanitization, and this change adds no database execution.

The existing `php.builtin.array_map` catalog ID becomes partial, not general
callback support; `php.builtin.addslashes-string` has its own bounded entry.
The `is_array` stack at `14383426911e42134780e098fec721ce68c0debb` is integrated
with the prior conditional/array compiler, preserving its predicate and guards.
The new combined configuration requires fresh evidence; parent PASS results
do not stand in for it. At this integration's original evidence point, its
PHP 8.3-only admission policy and immutable native 8.3.36 image remained
unchanged. The current compiler targets are PHP 8.3 and 8.4 under the
[active PHP version policy](compatibility.md#active-php-version-policy); this
historical 8.3 statement does not create 8.4 evidence. Gossip separately pins
this compiler to compare original `wp_slash`; that does not establish
`wp_unslash`, `map_deep` or WordPress core compatibility.

`gohp.cli.generate` remains WebMCP-excluded local developer tooling. No UI,
browser/S2S/MCP route, cookie forwarding, authentication, deployment, release or
license change is introduced. No devboard card is identified or marked Done.

## Integration with the predicate/conditional stack

The current integration combines PR #26 at
`71860d8ab7dc562b32f5173559cffea020ee5ba1` with published PR #27 at
`03d433e73503e32a84f310f023fc0f65c9b188cd`. Source implementations and their
failure boundaries are retained; both sets of earlier checks remain registered.
The owned `callback_composition_integration.php` fixture combines finite named
callback recursion with predicates and ternary selection, nested associative
arrays, byte quoting/unquoting, key observations and original input preservation.
Callback markers also expose order and empty-array non-invocation. Native
evidence is stdout/normal-exit comparison, not a general typed observation or
PHP warning/exception protocol. Its two catalog links count one shared check.

Earlier compiler and Gossip wp_slash evidence stays tied to its original pins.
This merged compiler needs its own validation and does not repin or revalidate
Gossip automatically. Arbitrary-depth/cyclic input and general callbacks remain
outside the supported contract.
