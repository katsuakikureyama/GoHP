# Type names for the existing value subset

`gettype(value)` is a case-insensitive direct builtin with exactly one argument
and a string result. It follows PHP's [type-name contract](https://www.php.net/manual/en/function.gettype.php)
within the values already admitted by GoHP:

| Value | Returned string |
|---|---|
| null | `NULL` |
| bool | `boolean` |
| int64 | `integer` |
| string, including numeric-looking strings | `string` |
| ordered PHP array | `array` |
| generated class instance or capture-free arrow | `object` |

There is no numeric conversion, property lookup, callback invocation or array
mutation. The argument is evaluated once by ordinary expression lowering; a
missing read is not suppressed. PHP arrows are Closure objects, so this function
reports `object` without making ordinary class instances callable.

The compiler propagates the string result through existing argument and return
validation. Wrong arity, case-insensitive builtin redeclaration, undefined reads
and known string-to-array/int/bool misuse fail before generated output. Dynamic
arguments use `runtime.GetType` with the existing explicit `Object` marker used
by generated classes and valid `Callable` values. Foreign host values, invalid
or typed-nil runtime objects, Go `int`/float/slice/map/function values and array
pointers fail closed. They do not become PHP `NULL` or `unknown type`. The
marker is an explicit Go ABI contract, not a reflection-based class registry;
arbitrary host objects are not newly admitted as PHP source values.

Float, resources, closed resources, general object/class reflection,
`get_debug_type`, `get_class`, `settype`, dynamic builtin lookup and exception
equivalence remain outside this increment. No new PHP type, coercion, syntax,
target or oracle patch is enabled.

`gettype` is on the pure scalar-frame allowlist only when every operand already
has the same scalar proof. `gettype($scalar)` can stay in a closed scalar frame;
`gettype([])` or `gettype(new KnownClass())` retains the value-map path. A scalar
result alone does not admit array/object producers into scalar storage.

The owned `gettype.php` fixture checks type labels as observable stdout under
each pinned PHP 8.3 and 8.4 oracle, including values with otherwise similar
string output. This is explicit fixture coverage, not automatic typed-value
serialization of every case. Existing normal-exit and quiet-success checks
apply; arbitrary diagnostics, exceptions and side-effect equivalence remain
unmeasured. The run/view/comparison schema is unchanged.

Existing capability `gohp.cli.generate` remains WebMCP-excluded: this is local
compiler language support, not a new browser/API operation or authentication
route. No deployment target or Gossip compiler pin is changed.
