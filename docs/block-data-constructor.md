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
  constructors still require exact arity. [Public untyped method null defaults](parser-output-boundary.md)
  are now admitted through the same bounded min/max handling.
  Stateful constructors keep the caller's explicit Execution context; filling
  defaults does not create or reset one. Classes without
  a constructor retain the old zero-argument-only boundary; PHP's more permissive
  ignored-argument behavior is not implemented. Unknown/dynamic classes, named/
  unpacked/reference arguments remain excluded.
- `(array)` accepts a statically known generated class expression (including
  direct `new`) or an inferred-unknown expression whose actual value satisfies
  the guarded public-property snapshot protocol. Defined untyped parameters,
  ordinary function returns, indexed stack elements and named property reads
  can reach this guard; undefined or conditionally-only locals still reject.
  It evaluates the expression once and snapshots all public
  properties in declaration order, retaining null. Nested arrays detach through
  NewArray; object identifiers remain shared. Known scalar/array/Closure/null
  operands reject statically, and unsupported actual values reject at runtime.
  No general scalar/array/Closure
  operand casts, dynamic properties or private/protected key mangling.

The frontend normalizes implicit defaults to canonical null, keeps constructor
arguments on `new`, and emits `object_array_cast` IR. Validation admits the
declared class/arity and explicit cast shape before generation. Generated New
functions call the constructor internally; its statements retain source maps.
Generated `PHPProperties` returns a fresh ordered runtime.Array via NewArray.
Runtime ObjectToArray accepts only a valid public-property-object marker and
calls its snapshot once. Source cannot invoke these Go implementation methods.
This reuses the existing `object_array_cast` IR and direct `rt.ObjectToArray`
lowering without new reflection, storage paths or class-return inference.
An external Go host may explicitly implement `PublicPropertyObject`: this is
protocol opt-in, not proof of generated provenance. Its `PHPObject` must report
false for typed nil and its `PHPProperties` must return a fresh ordered NewArray
snapshot honoring the same copy/identity contract. Marker-only/getter-only or
snapshot-only lookalikes do not qualify. Runtime panics are boundary failures,
not PHP-compatible cast exceptions or generic coercion.

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

Owned `unknown_object_snapshot.php` separately exercises the inferred-unknown
path, including an indexed-frame property shape. `array_pop` is still a separate
unimplemented builtin; this owned fixture does not claim stack mutation.
Its native result is distinct
from complete original-parser generation or execution; see the per-change
evidence and consumer application gate. Existing known-object fixture evidence
does not certify the new operand scope.

`constructor_null_defaults.php` additionally observes omitted/explicit null/zero,
all-optional zero arguments, case-insensitive names, argument-before-body order,
fresh objects and caller/constructor/caller strtok continuation. Codegen checks
cover static min/max, unsupported default forms, malformed suffix counts and
constructor statement maps. Generated runtime negatives retain actual-value guards.

Capability `gohp.cli.generate` stays WebMCP-excluded: local developer compiler,
not browser/service operations. Route/auth/deployment/dependencies are unchanged,
and no production target is configured or created.
