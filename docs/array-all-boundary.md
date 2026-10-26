# PHP 8.4 bounded array_all

For target `8.4`, `array_all($array, 'declared_name')` uses the same admission
contract as [bounded array_any](array-any-boundary.md): an actual array, a literal
declared global function name, exactly two required untyped value parameters,
and an explicit `: bool` return declaration. Names are case-insensitive and
forward-resolved; unknown array values remain guarded at runtime.

The [PHP contract](https://www.php.net/manual/en/function.array-all.php) and
[implemented RFC](https://wiki.php.net/rfc/array_find#array_all) call the predicate
with value then key, stop at the first false, and return true if all elements
match, including empty arrays. This subset accepts actual bool results only.
It visits insertion order and normalized keys; each callback array value uses
the existing Clone boundary. Object identities are not deep-copied. Request-local
strtok state is passed through the generated direct-call adapter.

Neither predicate expands the pure scalar-frame allowlist, general callback
resolution, or truthiness coercion. Invalid callback declarations fail even on
empty inputs. Dynamic/arrow/builtin/method callbacks, optional/typed/reference
callback parameters and missing bool return annotations remain unsupported.
The existing bool-return guard can reject unknown nonbool results/fallthrough;
its panic is not evidence of PHP TypeError equivalence.

Target `8.3` retains normal user-function resolution for `array_all`; a missing
declaration fails rather than importing an 8.4 builtin. Target `8.4` still rejects
user redeclarations. The two targets have separate fixtures and independent
evidence; excluded cases stay not-measured. The 8.4 fixture covers empty/alltrue,
early false, key types/order, copy isolation, stateful callbacks and composition
with array_any. Stdout visit markers do not establish arbitrary side-effect,
diagnostic or exception equivalence. `array_find`/`array_find_key` remain outside
the implemented subset.

Existing capability `gohp.cli.generate` remains WebMCP **excluded**: developer-local
source generation with existing inputs and non-overwriting output publication.
No browser/server operation, authentication route, remote write, auto-submit,
release, deployment, Gossip pin or new PHP target is added.
