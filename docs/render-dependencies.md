# Render caller dependency audit (2026-10-26)

Scope: unchanged `wp_render_elements_class_name`, elements.php:279–306, in
[Gossip's retained WordPress source](https://github.com/katsuakikureyama/Gossip/blob/804c6390be43488dddab8935ccbfa920a436b66c/wp-includes/block-supports/elements.php#L279-L306).
This is a source/dependency inspection, not a compile/run PASS for that caller.
No original source, Gossip pin/corpus or consumer adapter is modified.

| Source location | Dependency | Status within GoHP |
|---|---|---|
| 280 | Nested array address and null coalescing | Bounded defined-root form implemented |
| 282 | is_string / str_contains / short circuit | Existing actual-value contracts implemented |
| 288 | Double-quoted whitespace delimiter containing `\f` | [Form-feed escape](form-feed-boundary.md) implemented; owned byte/tokenizer comparison, not original caller execution |
| 289–296 | strtok, while, strict false comparison, prefix test, break | Bounded implementations exist; no consumer adoption inferred |
| 301 | Construct original WP_HTML_Tag_Processor | Class not compiled/admitted; constructors alone do not supply it |
| 302 / 306 | next_tag condition / get_updated_html return calls | Method-expression syntax exists, but actual class behavior is absent |
| 303 | Discarded-result add_class call | Generic [standalone call statement](call-statement-boundary.md) added; no HTML implementation |

The form-feed lexical blocker is resolved within the bounded string subset.
Construction still requires the real class and its dependencies. The
original [tag processor](https://github.com/katsuakikureyama/Gossip/blob/804c6390be43488dddab8935ccbfa920a436b66c/wp-includes/html-api/class-wp-html-tag-processor.php#L411)
is an HTML state machine, not a small record class: it includes constants,
protected/private mutable state (421–840), defaulted/typed method interfaces
such as next_tag (900), add_class (4828), and get_updated_html (4926). Its full
transitive requirements and output equivalence have not been validated.
The [HTML dependency map](html-processor-dependencies.md) breaks out selected
core paths and support classes. Exact scalar constructor parameters are now
available, but that alone does not implement the processor or rendering.

Do not replace it with a regex/owned mock, treat unknown methods as no-ops,
silently bind a host HTML parser, or certify the caller after checking only its
token-prefix branch. Future work should separate generic syntax/runtime needs,
original-class dependency admission and pinned HTML behavior tests. Production
HTTP, hooks, styles, IDs, themes/plugins and complete WordPress remain unverified.

## Separate consumer evidence

[Gossip PR #41](https://github.com/katsuakikureyama/Gossip/pull/41), merged as
`804c6390be43488dddab8935ccbfa920a436b66c`, tested the complete unchanged
WP_Block_Parser_Block data class/constructor plus two original decision/lookup
functions and one owned glue function in a separate 16-case binary, at exact
GoHP #74 pin `eb3b968d447ab7a11af4efd3b1a1976b2cd277e8`.
Actual class, five properties/snapshot, decision (11 true/5 false) and input
preservation passed PHP8.3.36 comparison. The parser's creation/cast shape was
used, not its parsing algorithm. This is not tag processing or rendering;
earlier 24/334 configurations are unchanged and these counts must not be summed.
Neither #75 strtok nor this call-statement increment inherits consumer adoption.

This audit extends documentation for WebMCP-excluded local `gohp.cli.generate`.
No service, authentication route, external operation or deployment is added.
