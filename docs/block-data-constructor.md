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
  and [named-class parameters](#named-class-parameters) are separate bounded extensions. No non-null defaults, return statements/type,
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

## Named class parameters

Global functions, public methods and public constructors additionally admit
required, nonnullable **bare named generated-class** parameters, e.g.
`function take(Block $block)`. Resolve names case-insensitively against all
classes declared in the same program, including forward declarations. No
autoload or ambient host class-name registry is consulted.

IR0.1 retains aligned `ParamTypes`, with canonical `class:<lowercase-name>`
entries distinct from existing primitive annotations. Validate tag/name shape,
declaration existence and class-aware body scope before generation. Statically
known wrong classes, scalars, arrays and null arguments reject at global,
constructor and known-receiver method callsites. Unknown values are checked at
entry. Scalar-frame/owned-value admission is unchanged: class-parameter scopes
stay on the legacy map lane.

After existing Execution/arity/default handling, every parameter is checked in
declaration order before any PHP body statement executes, including unused
parameters or ones overwritten immediately. A class argument must be an exact
nonnull `*C_<name>` generated pointer. Typed nil and structural host lookalikes
are rejected. The pointer itself is stored, preserving object identity through
aliases, calls and arrays; nested array values keep their existing copy rules.
This is an entry contract, not a permanent restriction on the local binding:
later rebinding updates normal type facts without rebinding the caller's local.

Existing global primitive annotations, scalar constructor annotations and
trailing **untyped** null defaults remain. Nullable/implicit-null class defaults,
unions/intersections, namespaces, relative self/parent/static types, interfaces,
inheritance, typed properties/returns/arrows and primitive typed ordinary methods
remain excluded. Runtime failures are explicit Go guards, not PHP TypeError,
exception-message or catchability equivalence. No reference syntax, COW or
storage rewrite is introduced.

The owned `named_class_parameter.php` differential and source/IR/compiler,
generated rejection and host-ABI checks are tracked under
`php.syntax.named-class-parameters`. These do not establish full original
WordPress parser admission, loading or application acceptance. Existing local
CLI capability/WebMCP exclusion and route/auth/deployment boundaries are unchanged.
