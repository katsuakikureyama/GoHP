# Bounded `str_replace` strings

This compiler subset accepts `str_replace($search, $replacement, $subject)`
only when all three values are strings. Function names are case-insensitive and
`str_replace` is reserved against user redeclaration.

The supported behavior is byte-wise, case-sensitive, non-overlapping
replace-all. An empty search string returns the subject unchanged, matching the
PHP behavior rather than Go's empty-needle insertion behavior. Replacements
are not scanned again. Empty replacements are supported.

The owned native fixture [`tests/fixtures/str_replace.php`](../tests/fixtures/str_replace.php)
compares the supported syntax and values with PHP 8.3 >=8.3.36: ordinary
replace-all, empty search, overlapping source bytes, non-rescanning inserted
bytes, empty replacement, case sensitivity, case-insensitive call spelling and
multibyte UTF-8 text and literal `$1` replacement. The native fixture avoids
unsupported NUL/byte escape syntax; the runtime test covers NUL and invalid UTF-8
bytes directly, not as a native-fixture claim for unsupported literal escapes.

Known bad arity and static non-string arguments fail before generation. Dynamic
non-string values are rejected by the shared runtime boundary in each argument
position. Arrays, optional by-reference `count`, scalar/object coercion,
case-insensitive replacement, regular expressions and PHP warning/TypeError
equivalence are outside this contract.

The catalog item `php.builtin.str-replace-strings` links only these tests. This
partial builtin does not claim general `str_replace`, WordPress compatibility,
a service route, deployment, release or license change. It remains
`gohp.cli.generate`, explicitly excluded from WebMCP under
[the integration contract](integration.md).

Reference: [PHP str_replace manual](https://www.php.net/manual/en/function.str-replace.php).
