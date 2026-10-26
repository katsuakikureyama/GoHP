# Original HTML processor dependency map (2026-10-26)

Read-only source inspection at retained Gossip revision
`804c6390be43488dddab8935ccbfa920a436b66c`, not a compile/run PASS or complete
transitive audit. [Original tag processor](https://github.com/katsuakikureyama/Gossip/blob/804c6390be43488dddab8935ccbfa920a436b66c/wp-includes/html-api/class-wp-html-tag-processor.php)
has 5,397 lines; a render-facing subset still reaches its parsing state machine.
No source rewriting, host parser binding or HTML mock is introduced here.

| Entry | Direct dependencies and observed boundaries |
|---|---|
| constructor 849–861 | String check and protected html state; invalid branch calls WordPress diagnostics/translation globals |
| next_tag 900–917 | parse_query 4992, base_class_next_token 966, tag/attribute scanning 1722–2359, matches 5050; nonpublic state/methods and byte scanning |
| add_class 4828–4860 | Queues class changes; actual output requires class-to-attribute conversion 2438–2583 and set/remove_attribute |
| get_updated_html 4926–4982 | Conversion, replacement application 2596–2685, bookmarks and parser rewind; usort with static comparator 2613/2755 |
| Existing class attributes | Attribute decoding 2925–2931 reaches WP_HTML_Decoder::decode_attribute; named character references are another dependency |

## Supporting data classes

Files are under the same pinned `wp-includes/html-api/` directory.

- `class-wp-html-attribute-token.php`: six public untyped properties and required
  untyped constructor. Used at tag processor 2326–2333; existing data-class shape,
  but original class execution has not been certified by this change.
- `class-wp-html-span.php:77`: public constructor `(int $start, int $length)` and
  simple property assignments. Used for duplicate attributes (2346) and bookmarks
  (1364). [Exact scalar constructor parameters](scalar-constructor-boundary.md)
  remove its annotation gap; offset validity is not implied.
- `class-wp-html-text-replacement.php:79`: public constructor
  `(int $start, int $length, string $text)`, simple assignments. Created at
  processor 4706/4724/4802/4810 and consumed by update application.
  Constructor support is not replacement application or HTML escaping.

## Remaining generic requirements

The exact scalar constructor extension does not make the processor compile.
Nonpublic instance state and array defaults (514–840), constants/self references
(514/540/584), private/static/typed/nullable methods (966/1245/2755), generators
(1229), and byte scanning/sorting builtins remain separate work. Examples include
strpos (980), strspn (1027), strcspn (1740), substr_compare (2072), usort (2613),
and strcmp (2755). static::MAX_* and self::class must not be silently treated as
simple constant reads. Visibility must be enforced, not erased to public fields.

Next acceptance candidate: original Span/Text_Replacement/Attribute_Token data
classes with pinned PHP comparison and exact typed observations, before claiming
any processor adoption. Larger private state/constant/method changes require
their own contracts. Decoder, rawtext/script/comment states, bookmark semantics,
warnings/exceptions and full transitive behavior are still unverified.

See the [render caller audit](render-dependencies.md). Existing local
gohp.cli.generate remains WebMCP-excluded; no route/auth/deployment changes.
