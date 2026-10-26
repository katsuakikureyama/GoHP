# Bounded str_contains contract

`str_contains(string haystack, string needle): bool` accepts two positional
actual strings. Direct builtin names are case-insensitive. Each argument is
evaluated once; matching is case-sensitive over bytes, including NUL and invalid
UTF-8. Empty needle returns true, including with an empty haystack. A nonempty
needle cannot match an empty or shorter haystack. The result is an actual bool,
not a position, count or matched text.

Known nonstring arguments fail before output, including in unexecuted branches.
Unknown arguments are checked at runtime before empty-needle handling. There
is no scalar/null/object coercion, named/unpacked argument support, builtin
callable lookup, Unicode folding/normalization, token parsing or PHP exception
and diagnostic equivalence. Existing `is_string` and short-circuit operators can
guard dynamic inputs; no new type narrowing is introduced.

## Concrete dependency, not whole-consumer certification

Gossip baseline WordPress `wp-includes/block-supports/elements.php:282` uses this
builtin in `wp_render_elements_class_name` (lines 279–306): a short-circuited
`is_string` guard followed by membership of `wp-elements-`. This is substring
membership, not a requirement that the full class attribute starts with that
text or contains a whitespace-delimited class token.

The owned differential fixture covers that guard shape, nested array `??`, and
scalar/array/null guard inputs. It is not an unchanged-source run of the render
function. Bounded [`str_starts_with`](str-starts-with-boundary.md) is tracked
separately, as is [bounded while with bare break/continue](while-boundary.md).
[Bounded strtok](strtok-boundary.md) is also separate; `WP_HTML_Tag_Processor`
remains a missing dependency.
No stubs or claim of HTML rendering
support are added. Gossip pins/source/corpora remain unchanged.

Catalog `php.builtin.str_contains` remains `partial`, with six native-differential,
compiler, source rejection and runtime checks. Only PHP target 8.3 is validated
(native oracle >= 8.3.36); later PHP targets remain separate work.

Capability: existing developer-local `gohp.cli.generate`, WebMCP `excluded`.
No browser/server/MCP operation, route, authentication, deployment or auto-submit.

Reference: [PHP str_contains manual](https://www.php.net/manual/en/function.str-contains.php).
