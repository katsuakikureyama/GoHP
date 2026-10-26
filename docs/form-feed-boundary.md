# Form-feed string escape

Double-quoted `"\f"` decodes to the single byte 0x0C. Single-quoted `'\f'`
and double-quoted `"\\f"` retain the two bytes backslash + lowercase f.
Repeated/embedded escapes and combinations with existing `\n`, `\t`, `\r`
are supported. Decoding FF does not advance source line numbers; physical
linefeeds still do. Existing IR string storage and Go quoting preserve the byte.

This is a bounded extension, not full PHP string syntax. Other escapes such as
`\v`, `\e`, hexadecimal/octal/Unicode escapes, escaped dollars, unknown escapes,
interpolation and heredoc/nowdoc remain outside the existing closed subset.
Unsupported escapes fail closed, even where PHP would retain a literal backslash.
No tokenizer, string builtin, runtime or source-map format changes are needed.

The owned differential fixture covers byte lengths, literal-vs-decoded identity,
array keys, function argument/return preservation and the exact whitespace
delimiter `" \t\f\r\n"` used at elements.php:288. It is not execution of the
unchanged WordPress render caller. Its actual HTML processor class remains
unavailable; see the [remaining dependencies](render-dependencies.md).

Catalog `php.syntax.form-feed-escape` is partial, checked against PHP8.3.36
only. Frontend tests check exact byte 12 and source lines, plus retained rejection
boundaries; the PHP differential compares stdout and normal exit, not exceptions
or other side effects. No later PHP target is enabled by this increment.

Capability: existing developer-local `gohp.cli.generate`, WebMCP `excluded`.
No new service operation, browser/S2S/MCP route, authentication or deployment.

Reference: [PHP string literal rules](https://www.php.net/manual/en/language.types.string.php).
