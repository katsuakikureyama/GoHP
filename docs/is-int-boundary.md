# Integer type predicate

`is_int(value)` takes exactly one supported value and returns an actual bool.
Direct builtin calls are case-insensitive and the name is reserved against
redeclaration. Only the runtime's canonical `int64` integer values return true.
Null, booleans, strings (including numeric strings), arrays, class instances and
generated arrows return false. No conversion, traversal or mutation occurs.
The argument expression is evaluated once.

This follows the [PHP is_int contract](https://www.php.net/manual/en/function.is-int.php)
for the existing GoHP value domain. It is a type predicate, not a numeric-string
test. The existing 64-bit integer representation, source literal restrictions,
arithmetic behavior and PHP 8.3 oracle admission are unchanged. Runtime-only
tests cover both int64 endpoints and distinguish Go int/int32/uint64/float64
values from the PHP integer representation; this does not add those source types.

Semantic validation recognizes the result as bool. Invalid argument counts,
case-insensitive redeclarations and known bool-as-array misuse fail before
generation. No branch-sensitive narrowing is added: unsupported expressions in
dead or short-circuited branches remain invalid. Dynamic inputs can use the
existing short-circuit operators to guard runtime arithmetic.

The owned `is_int.php` fixture observes strict bool results, integers versus
numeric strings and other values, single evaluation, guarded arithmetic and
integer/string keys through generated callbacks. Native evidence compares
stdout and normal exit on PHP 8.3 >=8.3.36, not general warning/exception behavior.
The feature remains partial in the function catalog. `is_integer`/`is_long`
aliases, `is_numeric`, float/resource syntax, named arguments and builtin callable
lookup are not introduced. Literal named `array_map` callbacks still require
a declared user function; arrows may call `is_int` through their normal body.

Gossip's original-source pins, previous evidence and pending publication are
unchanged. The existing 19-original bundle is not re-labelled with this compiler
revision; adoption needs a separate pinned comparison. PHP 8.4+ remains separate.
Capability `gohp.cli.generate` stays WebMCP-excluded: local developer translation,
no browser/API operation, authentication change, new dependency or deployment.
