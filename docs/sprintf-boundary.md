# Bounded `sprintf` zero-padding

This is a deliberately narrow `sprintf` contract for the original WordPress
`zeroise($number, $threshold)` shape:

```php
return sprintf('%0' . $threshold . 's', $number);
```

Accepted domain:

- Exactly two arguments: a format string and a string or `int64` value.
- The format must be exactly `%0Ns`, where `N` is canonical decimal `0` through
  `4096`. `%00s` represents width zero; widths with leading zeroes such as
  `%005s` are outside this contract.
- Width is a minimum **byte** width. The result is left-padded with ASCII `0`
  bytes and is never truncated. Negative signs and UTF-8 bytes count as bytes.
- A known literal format, arity or value-type violation fails before generation.
  Dynamic formats and values are guarded by the same runtime boundary and fail
  there when outside the accepted domain.
- `sprintf` is case-insensitive and reserved against user function
  redeclaration.

`N <= 4096` is an explicit resource bound for this compiler subset, not a PHP
format-width claim. The full PHP format grammar is intentionally excluded:
other directives, flags, precision, positional arguments, `*` widths, literals
around the directive, multiple substitutions and PHP scalar/object coercions
are not implemented. In particular, a rejection is not PHP warning, TypeError
or coercion equivalence evidence.

`tests/fixtures/sprintf.php` retains the parameterized dynamic format used by
`zeroise` and compares native PHP 8.3 >=8.3.36 stdout against the generated
binary. It covers positive/zero integers, strings including a negative-number
string, an empty string,
width zero, no truncation and a multibyte string with byte width. The runtime and
code-generation tests separately cover the width ceiling and rejected forms.
The frontend now admits [bounded negative decimal constants](negative-integer-boundary.md).
This sprintf fixture still does not observe parsed negative integer inputs:
negative int64 runtime values are tested directly and through Gossip's separate
typed-input observer. The syntax extension does not expand this fixture's evidence.
The compatibility catalog entry `php.builtin.sprintf-zero-pad` records those
checks; a generation-only result is not a PHP compatibility pass.

The integrated `string_array_integration.php` fixture also combines formatting
with replacement and array helpers on one compiler/runtime. This measures only
the owned composition cases, not Gossip's original-source units or pins.

The WordPress original excerpt and its typed observations remain in Gossip,
which must pin this compiler/runtime revision together. This compiler feature
does not claim general `sprintf` support, WordPress compatibility, a service or
authentication route, a deployment, a release, or a license change. It remains
`gohp.cli.generate`, explicitly excluded from WebMCP under
[the integration contract](integration.md).
