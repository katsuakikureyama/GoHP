# Declared public-property assignment

The closed source form `$holder->property = $value;` is now a statement. Its
receiver must be a defined bare variable statically known to hold a declared
generated class, including a known alias or `$this` in a method. The property
name is a literal, case-sensitive, declared public untyped property. Unknown
function-parameter receivers, scalars, arrays and generated arrows are rejected
before output; an `is_object` guard does not add type narrowing.

IR uses `assign_property` with an explicit `property(variable)` target. Target
shape/metadata and unrelated statement payloads fail closed. A generated setter
dispatches only declared names, clones the supported RHS value once and rejects
unknown names. Source code cannot invoke that Go implementation method. The RHS
is an ordinary supported expression and runs once. No dynamic host setter or
new runtime object admission is added.

All generated public properties are untyped mutable slots: a null/scalar default
is not a permanent inferred type. Reads retain declaration checks but become
unknown so existing operation guards check the actual value after writes,
alias changes and method calls. No optimistic branch/alias-sensitive narrowing.
Classes have scalar/null literal or implicit-null defaults. The later
[constructor boundary](block-data-constructor.md) permits fixed untyped public
constructors; this does not broaden unknown-receiver property assignment.

Arrays use the existing recursive value-copy convention at both setter and
getter boundaries; writing the original or a read-back array does not change
the stored array. Object/arrow identifiers retain identity under Clone, so known
aliases observe the same property update. This follows the bounded application
of PHP's [object identifier semantics](https://www.php.net/manual/en/language.oop5.references.php),
not PHP `&` references or a claim of general object/reference equivalence.

For the single-property form, dynamic properties/names, nested/chained/effectful receivers, property-array
indexed writes, compound/assignment expressions, references, general constructors,
inheritance, typed/private/static properties and general warning/exception
behavior remain excluded. Runtime-only direct setter misuse is not a PHP input
contract. A later [bounded property append](parser-output-boundary.md) admits only a declared bare-variable receiver and an actual array slot, not general nested writes.

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

Resolve the final owner and verify the final property before evaluating RHS
once. Keep that owner even if RHS replaces an intermediate property. Concat
and append read the owner's current final value after RHS evaluation. Concat
requires actual strings on both sides; append requires an actual array slot.
Assignment preserves detached array values and shared object identifiers. Paths
through copied arrays are not supported. Missing/null/scalar intermediates and
dynamic properties fail; no property or array is automatically created.

IR uses `assign_nested_property`, `assign_nested_property_concat` or
`assign_nested_property_append` with a chain of `property` nodes in `Target`
and the RHS in `Expr`. The lexer recognizes `.=` as one adjacent token; `. =`
is rejected. Root/property metadata and unrelated payloads are validated. Code
generation resolves a synchronous `runtime.NestedProperty` and uses its explicit
operation. This does not expose PHP reference bindings or migrate array storage;
these scopes stay on the legacy path. Existing one-property unknown-receiver
rejection remains in force.

The owned `nested_property.php` fixture and catalog
`php.syntax.nested-property-update` track object identity through arrays/list,
copies, aliases, multiple path depths and RHS mutation/order. Evidence remains
per target. Typed/magic/nonpublic properties, array-index targets, effectful or
dynamic path steps, coercing concat, null-array append and general exceptions
remain outside this contract. Source parser integration and remaining stack/
regex/JSON/loading dependencies require their own complete-unit results.
