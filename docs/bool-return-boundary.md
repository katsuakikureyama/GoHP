# Actual-bool return declarations

Named global functions may declare `: bool`, case-insensitive in source. The
parser stores canonical `Function.ReturnType = "bool"` in the existing IR 0.1
field. Only actual boolean return values are admitted; this is a deliberately
narrow subset, not PHP's full scalar return-coercion rules or strict_types mode.

Every explicit return, including if/elseif/else if/else and foreach paths, uses
`rt.BoolReturn(value)`. The expression is evaluated once. Known non-bool values
are rejected before generation, including unreachable returns; unknown values
are checked at runtime. Fallthrough uses `rt.BoolReturn(nil)` and cannot silently
return null. No path-completeness proof or call-site return-type inference is
added. Existing untyped returns and the separate array-return contract remain
unchanged. A named one-argument bool-return function can be an existing literal
`array_map` callback; no dynamic callable engine is introduced.

The [PHP type-declaration reference](https://www.php.net/manual/en/language.types.declarations.php)
describes scalar coercions and return type enforcement. PHP may accept a scalar
return that GoHP rejects under this boundary. Runtime rejection is not PHP
TypeError text, stack or catchability equivalence. Typed methods/parameters,
nullable/union/intersection/standalone true or false return types, `boolean`
aliases, general callable values and weak return coercions are excluded.

`bool_return.php` compares owned positive examples against PHP 8.3 >=8.3.36:
true/false and predicate results, branching/foreach, forward calls, typed callback
composition, one evaluation and caller preservation. The harness compares
stdout and normal exit; explicit `is_bool` and strict comparisons expose types
through stdout. Unit, compiler and generated-binary negative checks enforce
the narrower boundary independently of native error text.

This removes only one prerequisite for Gossip's numeric-array candidate. Typed
array/callable parameters, arrow callbacks and dynamic invocation remain open;
the WordPress original/polyfill is not replaced, adopted or declared compatible.
Gossip must pin and verify the complete dependency chain separately.

Local `gohp.cli.generate` remains WebMCP-excluded. No service/authentication route,
external write, new dependency, automatic promotion, release or deployment.
