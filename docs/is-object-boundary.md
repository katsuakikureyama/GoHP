# Object type predicate

`is_object(value)` is a case-insensitive, one-argument direct builtin returning
an actual bool. Within the existing PHP value subset, generated class instances
(including empty classes) and generated capture-free arrows return true. PHP
arrows are Closure objects; this does not make ordinary class instances callable.
Scalar/null/array values return false without coercion, property access or
mutation. The argument expression is evaluated once, with ordinary read guards.

This implements the [PHP is_object value contract](https://www.php.net/manual/en/function.is-object.php)
only within the source types already admitted by GoHP. Every generated class
pointer explicitly implements runtime `Object` via `PHPObject() bool`; generated
arrows use the same marker. Merely having a `Get` method does not mark a foreign
Go value as a PHP object. Nil/invalid runtime-only pointers are not PHP source
values and return false, rather than introducing PHP null conversion semantics.
The generated marker is a Go implementation detail, not a callable PHP method.

Arity, undefined variables, reserved-name redeclaration and statically invalid
bool use fail before output. Builtin callable lookup remains excluded. No object
truthiness, reference comparison, constructors, property writes, array property
defaults, inheritance, arbitrary host-object admission, object/Closure methods,
resources or incomplete classes are added. Missing reads still fail normally.

The owned differential fixture composes classes/arrows, typed bool/array returns,
ordered arrays, callbacks, short-circuit property access and a visible once-only
argument marker. It compares stdout and successful exits with admitted PHP8.3,
not PHP warning/exception behavior or Gossip original-function compatibility.
Gossip adoption requires a separately pinned original-source observation.

Capability `gohp.cli.generate` remains WebMCP-excluded: local developer compiler,
no browser/API operation or route/auth change. No production target is added.
