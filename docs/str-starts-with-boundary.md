# Bounded str_starts_with contract

`str_starts_with(string haystack, string needle): bool` accepts two positional
actual strings. Direct names are case-insensitive; matching is case-sensitive
over bytes at position zero. Empty needle returns true, including for an empty
haystack. Exact/full prefix matches are true; a longer needle, a match only in
the middle/end, or different case is false. NUL and invalid UTF-8 are bytes,
not terminators or decoding errors. No trimming, Unicode folding/normalization,
token boundaries or grapheme interpretation is performed.

Arguments are each evaluated once. Known nonstrings are rejected before output,
including in unexecuted branches. Unknowns receive runtime string guards before
empty-needle handling. No scalar/null/object coercion, named arguments, argument
unpacking, builtin callable lookup or PHP exception/message equivalence is added.
The result has bool semantic kind and composes with typed bool returns, existing
short-circuit string guards, callbacks and bounded array coalescing.

## Actual dependency and remaining scope

WordPress baseline `wp-includes/block-supports/elements.php:291`, in
`wp_render_elements_class_name`, uses this builtin to recognize the prefix of
an already-tokenized class name. `str_contains` on the complete attribute is
only an earlier guard; the two operations are not interchangeable.

Our fixture uses owned strings/token arrays and guard composition. It does not
execute or certify the unchanged full render function. [Bounded strtok](strtok-boundary.md)
is tracked separately; `WP_HTML_Tag_Processor` remains a missing dependency. [Bounded while](while-boundary.md)
now supports bare nearest-loop break/continue, but does not certify that caller.
No Gossip source, pins, corpora or consumer adoption are changed.

Catalog entry `php.builtin.str_starts_with` remains `partial` with six checks:
native differential, compiler/source mapping, static rejection, byte behavior,
runtime domain rejection and generated dynamic guards. Only PHP target 8.3 is
validated (native oracle >= 8.3.36); later targets remain separate work.

Capability: developer-local `gohp.cli.generate`, explicitly WebMCP `excluded`.
No browser/server/MCP operation, route, authentication, deployment or auto-submit.

Reference: [PHP str_starts_with manual](https://www.php.net/manual/en/function.str-starts-with.php).
