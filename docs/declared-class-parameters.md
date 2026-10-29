# Declared class parameters

Global functions, public ordinary methods and public constructors admit required
parameters annotated with a bare class name declared in the same generated unit.
Names are case-insensitive; declarations may follow their consumers. Internal
`ParamTypes` uses canonical `class:lowername` alongside existing annotation kinds.
All signatures resolve after collecting classes, before validating bodies.

The entry guard requires an exact, nonnil generated `*C_lowername`. Null, typed
nil, other generated classes, host `runtime.Object` implementations and other
Go values reject before any PHP body statements, even if the parameter is unused
or immediately overwritten. This is a bounded panic contract, not PHP-compatible
`TypeError` handling. Known wrong source arguments reject before generation;
unknown arguments and indirect/callback entries retain runtime guards.

The value passed into the body is the same object pointer: property mutation is
visible to callers. The parameter local can subsequently be reassigned; its
annotation constrains entry, not permanent local storage. Semantic facts follow
reassignment, and generated class-parameter scopes remain in the legacy map lane.
Existing arity, scalar constructor/global parameters, untyped trailing null
defaults and explicit `Execution` threading are preserved.

## Boundaries

- No inheritance/interfaces, namespace/qualified names, aliases, `self`, `parent`
  or `static` types.
- No nullable/union/intersection/reference parameters or typed defaults,
  including class parameters defaulted to null.
- No primitive typed ordinary methods, typed properties or general return-type
  inference. Existing array/callable/scalar contracts remain separate.
- No original WordPress rewrite, consumer-pin change or whole-parser acceptance
  implied. Native evidence is recorded independently for each active target.
- This extends the existing developer CLI compile/inspect operation; WebMCP is
  explicitly excluded. No browser/S2S/MCP/webhook/public route or auth change.

`php.syntax.declared-class-parameters` in the compatibility catalog tracks the
owned PHP fixture, lowering/source-map checks, source/raw-IR rejection checks
and generated external ABI guard checks. CI compares the fixture independently
on pinned PHP 8.3 and 8.4; a compiler check alone is not native compatibility.
