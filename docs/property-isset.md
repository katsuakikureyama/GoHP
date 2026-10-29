# Named-property isset

`isset($local->name)` and named chains such as `isset($this->child->value)` are
admitted when the root is a defined bare local. They can mix with existing bare
and array addresses inside multi-argument `isset`. IR `isset_property` has a
bare-variable first argument and one or more canonical identifier-string names;
all node metadata is checked before generation. Known callable roots are outside
the subset; no control-flow type narrowing is inferred from a successful probe.

Missing and null slots are false. False, integer zero, empty string, string `0`
and empty arrays in the final slot are true. Defined null/int64/string/bool/Array
roots or intermediate values quietly stop with false. Missing *variables* remain
unsupported: this is not general PHP undefined-variable suppression.

Generated classes add a read-only `PHPPropertyLookup(string) (any, bool)` method
alongside their strict `Get`/`Set` methods. It reports declaration presence and
the raw slot value without copying arrays or throwing on an absent name.
`runtime.PropertyLookupObject` combines that method with the object marker;
`runtime.PropertyIsset` inspects each reached owner once, never calls `Get`/`Set`,
and stops before later path steps if a slot is absent/null/nonobject. Typed-nil
generated pointers report false without dereferencing. Foreign hosts without
the explicit lookup protocol reject rather than being trusted as PHP objects.
Host lookup implementations can have arbitrary side effects; native PHP
compatibility is claimed only for the measured generated-class fixture.

The root is evaluated once. Multi-address `isset` and logical expressions retain
left-to-right short circuit using Go `&&`/`||`. Ordinary missing-property reads
still reject; no array/property mixed paths, dynamic names, calls/indices in
receivers, magic methods, visibility/inheritance/typed properties or property
null-coalescing are added. Class/`isset_property` scopes retain legacy storage.

Catalog `php.syntax.property-isset` records the owned PHP fixture, lowering and
source maps, source/raw-IR rejection, runtime presence/host guards and external
generated ABI observations. The fixed PHP8.3/PHP8.4 CI matrix establishes each
native target independently. Whole WordPress parser acceptance remains separate;
the original source and consumer pins are not changed by this feature.

This extends the existing local developer CLI compile/inspect capability;
WebMCP is explicitly excluded. There is no service/route/auth/deployment change.
