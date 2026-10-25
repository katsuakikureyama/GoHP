# Full ternary and string-only `stripslashes`

This deliberately bounded addition supports the full conditional spelling
`condition ? yes : no` and the one-argument `stripslashes` call used by the
unchanged WordPress `stripslashes_from_strings_only($value)` shape. It is not
general PHP conditional, coercion, callback, or recursive array support.

## Conditional contract

The condition uses the existing supported scalar/array truthiness domain and is
evaluated once. Code generation uses a local Go closure so only the selected
arm evaluates. Both arms are nevertheless statically validated. A matching
branch type is retained; different branch types deliberately become unknown,
so validation does not pretend that an always-unselected arm determines a
value's type.

Ternary binds below the supported `||`, `&&`, `===`, `.`, and `+` operators. The
middle expression can itself contain an unparenthesized ternary, while the
final expression cannot: PHP 8's ternary operator is non-associative. Explicit
parentheses allow nesting. Short ternary `?:`, null coalescing `??`, an
unparenthesized ternary chain, object truthiness, and all unsupported expression
forms fail closed.

## `stripslashes` contract

`stripslashes` is case-insensitive and reserved against user redeclaration. It
accepts exactly one actual string. Its byte loop removes one backslash before
the following byte; an escaped ASCII `0` becomes NUL, and a trailing backslash
is dropped. Backslash pairs are consumed once, actual NUL bytes are preserved,
and no Unicode decoding is performed. This is intentionally not PHP scalar or
object coercion, recursive array processing, warning/TypeError equivalence, or
support for `stripslashes_deep` or `wp_unslash`. The separate
[literal array_map extension](array-map-boundary.md) now admits bounded named
callbacks; it does not make this string builtin automatically recursive.

The owned native differential fixture
[`tests/fixtures/ternary_stripslashes.php`](../tests/fixtures/ternary_stripslashes.php)
compares supported precedence, nesting, visible lazy side effects, string-only
guarding, and string values with PHP 8.3 >=8.3.36. It represents an escaped
ASCII zero as an escaped slash followed by `0`; invalid UTF-8 is covered only
by the direct runtime test, not by a PHP-source literal claim.

The catalog items `php.syntax.full-ternary` and
`php.builtin.stripslashes-string` link the exact differential, compiler and
runtime checks. This remains `gohp.cli.generate`, explicitly excluded from
WebMCP under [the integration contract](integration.md). It adds no browser or
service operation, authentication route, deployment, release, license change,
or WordPress-wide compatibility claim.

References: [PHP conditional expressions](https://www.php.net/manual/en/language.operators.comparison.php),
[PHP `stripslashes` manual](https://www.php.net/manual/en/function.stripslashes.php),
and [PHP 8.3 `php_stripslashes_impl`](https://github.com/php/php-src/blob/PHP-8.3/ext/standard/string.c).
