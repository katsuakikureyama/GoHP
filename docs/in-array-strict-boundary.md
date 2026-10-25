# Strict scalar array membership

`in_array(needle, haystack, true)` accepts exactly three arguments. The third
must be the literal `true` in any ASCII letter case (parentheses are transparent),
not a variable, expression, omitted value or loose mode. The parser canonicalizes
`TRUE` and mixed-case spellings to the same bool IR; `FALSE`/`NULL` remain invalid
strictness arguments.
Function names are case-insensitive and reserved against redeclaration.

The needle and every haystack value must be null, bool, int64 or string;
the haystack must be an actual array. Keys do not participate. Comparison is
type-and-value strict, with case-sensitive binary string equality. The result
is bool, including false for an empty array. Each dynamic operand is evaluated
once. The [PHP manual](https://www.php.net/manual/en/function.in-array.php)
defines strict type checks, case-sensitive strings and the Boolean result.

Known invalid needles/haystacks and literal array elements are rejected before
generation, even in dead branches. Element types in variables and call results
are not tracked statically; the runtime validates them. The complete haystack
is validated before searching: an unsupported nested array/object after an
early match is still rejected. This is a deliberately closed domain, not PHP's
general mixed-array membership semantics. Unsupported runtime values panic;
PHP warnings, exception text and catchability are not reproduced.

Owned unit, lowering/source-map, static-negative, generated-runtime guard and
native differential checks cover this boundary. Native compiler evidence
observes stdout and normal exit, not general stderr or external side effects.
The original #44 builds on the published charset/array integration #41.
The control-flow membership integration combines published #44 with #43,
including braced if/elseif/else and typed-array returns. Existing Gossip pins
remain unchanged; this is not automatic adoption by an original-source consumer.

The owned `membership_control_flow.php` fixture composes strict membership with
all conditional arms, typed callbacks, isset selection, negative-key append,
copy isolation and short-circuiting. Reached dynamic membership still validates
the entire haystack, including invalid values after an early match. Skipped
conditions do not execute runtime guards; known invalid dead arms remain
statically rejected. Boolean membership results cannot bypass array-return
guards, whether known statically or returned through an unknown call.

The intended original-source consumer is WordPress
`wp_is_heic_image_mime_type()`, functions.php9465–9474. Gossip must separately
pin and compare that unchanged helper. This classifies a MIME label, not image
contents, upload safety, decoding or conversion. No whole-core compatibility.

`gohp.cli.generate` remains WebMCP-excluded local developer tooling. No
browser/S2S/MCP route, auth contract, cookie forwarding, deployment, release
or license decision changes. Devboard is unidentified/unsynchronized, not Done.
