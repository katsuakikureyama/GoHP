# Declared public-property assignment

The closed source form `$holder->property = $value;` is now a statement. Its
receiver must be a defined bare variable with a known generated class or unknown
inferred type, including untyped parameters and values extracted from arrays.
The property name is a literal, case-sensitive, declared public untyped property.
Known scalars, arrays, null and generated arrows reject before output; unknown
values need an explicit runtime object/declared-slot guard. An `is_object` guard
does not add type narrowing. Known classes still reject absent properties.

IR uses `assign_property` with an explicit `property(variable)` target. Target
shape/metadata and unrelated statement payloads fail closed. Codegen evaluates
the ordinary supported RHS once, then uses `runtime.ResolveProperty` to resolve
the receiver/slot and publish through its synchronous Assign/Append operation.
The runtime requires nonnil `PropertyObject` (`Object`/`PHPObject`, Get, Set) and
checks the final declared name with Get before publication. Generated getters
and setters reject unknown names and preserve array-copy boundaries. Source PHP
cannot invoke those Go implementation methods. Setter-only and unmarked
getter/setter hosts reject; a host explicitly implementing the full protocol
opts in, exactly as for existing nested paths. This structural protocol is not
an unforgeable generated-class certificate or a hostile-host security boundary.

All generated public properties are untyped mutable slots: a null/scalar default
is not a permanent inferred type. Reads retain declaration checks but become
unknown so existing operation guards check the actual value after writes,
alias changes and method calls. No optimistic branch/alias-sensitive narrowing.
Classes have scalar/null literal or implicit-null defaults. The later
[constructor boundary](block-data-constructor.md) permits fixed untyped public
constructors; constructor admission and this guarded write contract are separate.

Arrays use the existing recursive value-copy convention at both setter and
getter boundaries; writing the original or a read-back array does not change
the stored array. Object/arrow identifiers retain identity under Clone, so known
aliases observe the same property update. This follows the bounded application
of PHP's [object identifier semantics](https://www.php.net/manual/en/language.oop5.references.php),
not PHP `&` references or a claim of general object/reference equivalence.

For the single-property form, dynamic properties/names, nested/chained/effectful receivers, property-array
indexed writes, compound/assignment expressions, references, general constructors,
inheritance, typed/private/static properties and general warning/exception
behavior remain excluded. Direct single-property `.=` remains excluded.
Runtime-only direct setter misuse is not a PHP input contract.
[Bounded property append](parser-output-boundary.md) also accepts a defined
unknown bare-local receiver, but requires a declared actual-array slot after RHS
evaluation; it never creates an array from null. Missing/nonobject targets reject
explicitly after RHS; this is not PHP exception/warning equivalence.

The owned differential fixture checks nested array independence, known object
aliases, `$this`, case sensitivity, once-only RHS, value type transitions and
object/arrow identifiers. It compares stdout and normal exits with admitted
PHP8.3.36. Generated runtime negatives retain count/strlen/array-return guards
after writes. Gossip adoption requires a separate exact compiler/source pin.

Capability `gohp.cli.generate` remains WebMCP-excluded (local developer compiler,
not a browser/service operation); route/auth/deployment are unchanged. No new
production target or dependency is introduced.

## Named nested-property updates

The parser's Frame/Block updates use a separate closed form:
`$frame->block->text = $value;`, `$frame->block->text .= $string;` and
`$frame->block->items[] = $value;`. Two or more literal property names must be
rooted in a defined bare local. Known object and unknown roots are admitted;
known scalar/array/callable roots reject. Every runtime receiver must implement
the nonnull Object/Get/Set protocol, and each accessed property must be declared.

For these named chains, evaluate RHS once before resolving the writable path.
If RHS replaces an intermediate property, update the replacement object's final
property. Verify that property exists, then publish once. Concat
and append read the owner's current final value after RHS evaluation. Concat
requires actual strings on both sides; append requires an actual array slot.
Assignment preserves detached array values and shared object identifiers. Paths
through copied arrays are not supported. Missing/null/scalar intermediates and
dynamic properties fail; no property or array is automatically created.

IR uses `assign_nested_property`, `assign_nested_property_concat` or
`assign_nested_property_append` with a chain of `property` nodes in `Target`
and the RHS in `Expr`. The lexer recognizes `.=` as one adjacent token; `. =`
is rejected. Root/property metadata and unrelated payloads are validated. Code
generation evaluates RHS, resolves a synchronous `runtime.NestedProperty` and uses its explicit
operation. This does not expose PHP reference bindings or migrate array storage;
these scopes stay on the legacy path. Direct unknown-receiver writes use the
separate one-name ResolveProperty entry without relaxing the 2+ nested-path API.

The owned `nested_property.php` fixture and catalog
`php.syntax.nested-property-update` track object identity through arrays/list,
copies, aliases, multiple path depths and RHS mutation/order. Evidence remains
per target. Typed/magic/nonpublic properties, array-index targets, effectful or
dynamic path steps, coercing concat, null-array append and general exceptions
remain outside this contract. Source parser integration and remaining stack/
regex/JSON/loading dependencies require their own complete-unit results.

The real `unknown_property_write.php` fixture covers array-extracted objects,
untyped parameters, aliases, falsy slots, nested value copies, branch/loop-widened
locals and same-slot RHS-reset append ordering. Generated external-ABI checks
cover typed nil, scalar/array/partial protocol/missing-slot rejection, one RHS
before guard, no later write after failure and explicit full-protocol host opt-in.
Results remain per exact compiler commit/target; no full original parser PASS is
implied. The measured complete-source blocker before this increment was209,
`$stack_top->prev_offset` after array_pop, not a missing ordinary include mode.
The [new complete-source probe and focused evidence](evidence/unknown-property-write-2026-10-29.md)
pass that write and next reject string `empty($html)` at354. Full original
generation/execution and application acceptance remain blocked.
