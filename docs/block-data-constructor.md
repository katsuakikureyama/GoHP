# Fixed public constructors and public-property array snapshots

This closed extension targets simple data classes such as WordPress's unchanged
`WP_Block_Parser_Block`, not an arbitrary object system or parser implementation.

- Untyped public `$property;` initializes to canonical null. Explicit defaults
  remain scalar/null literals; no typed/private/static/array defaults.
- A public `__construct` (case-insensitive) accepts a fixed number of untyped
  positional arguments (with an optional trailing untyped `= null` suffix) and
  executes once after fresh instance defaults exist.
  Arguments use the existing value-copy convention; `$this` writes only declared
  public properties. [Exact scalar annotations](scalar-constructor-boundary.md)
  are a separate bounded extension. No non-null defaults, return statements/type,
  direct constructor re-entry, other magic methods or inheritance.
- `new Class(args...)` accepts the required minimum through declared maximum.
  Omitted suffix slots become canonical null through existing `NullDefaultArgs`
  before parameter copies. Explicit null/zero values are not replaced. Fixed
  constructors still require exact arity; ordinary method defaults remain rejected.
  Stateful constructors keep the caller's explicit Execution context; filling
  defaults does not create or reset one. Classes without
  a constructor retain the old zero-argument-only boundary; PHP's more permissive
  ignored-argument behavior is not implemented. Unknown/dynamic classes, named/
  unpacked/reference arguments remain excluded.
- `(array)` accepts only a statically known generated class expression, including
  a direct `new`. It evaluates that expression once and snapshots all public
  properties in declaration order, retaining null. Nested arrays detach through
  NewArray; object identifiers remain shared. No scalar/array/Closure/unknown
  operand casts, dynamic properties or private/protected key mangling.

The frontend normalizes implicit defaults to canonical null, keeps constructor
arguments on `new`, and emits `object_array_cast` IR. Validation admits the
declared class/arity and explicit cast shape before generation. Generated New
functions call the constructor internally; its statements retain source maps.
Generated `PHPProperties` returns a fresh ordered runtime.Array via NewArray.
Runtime ObjectToArray accepts only a valid public-property-object marker and
calls its snapshot once. Source cannot invoke these Go implementation methods.

These boundaries follow the admitted subset of PHP's [constructor contract](https://www.php.net/manual/en/language.oop5.decon.php)
and [default arguments](https://www.php.net/manual/en/functions.arguments.php#functions.arguments.default),
plus [object-to-array conversion](https://www.php.net/manual/en/language.types.array.php).
They do not add PHP's generic object/reference, error/exception or visibility
semantics. Arrays use existing Clone conventions; untyped mutable property reads
retain actual-value runtime guards instead of permanent default-type inference.

Owned fixture `block_constructor.php` observes argument/body ordering, once-only
effects, implicit null, fresh instances, nested array copies, snapshot keys/
independence and shared object identifiers. Native evidence compares stdout and
normal exits with admitted PHP8.3.36, not stderr/exceptions/other side effects.
Gossip must separately pin and compare original source before claiming adoption.

`constructor_null_defaults.php` additionally observes omitted/explicit null/zero,
all-optional zero arguments, case-insensitive names, argument-before-body order,
fresh objects and caller/constructor/caller strtok continuation. Codegen checks
cover static min/max, unsupported default forms, malformed suffix counts and
constructor statement maps. Generated runtime negatives retain actual-value guards.

Capability `gohp.cli.generate` stays WebMCP-excluded: local developer compiler,
not browser/service operations. Route/auth/deployment/dependencies are unchanged,
and no production target is configured or created.
