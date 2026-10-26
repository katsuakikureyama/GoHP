# Bounded prerequisites for original parser output

This closed extension targets WP_Block_Parser::add_freeform without rewriting
the original method. Compilation/owned fixtures do not establish original
WordPress adoption; Gossip separately pins the compiler and compares the source.

- Public untyped methods, including constructors, admit trailing case-insensitive
  null defaults. Known-receiver calls enforce required minimum through declared
  maximum; existing NullDefaultArgs pads omitted slots before ordinary copies.
  Explicit null/zero and caller Execution are retained. No typed/non-null/default
  expressions, named/unpacked/reference/variadic arguments or method return types.
- Bare return lowers to the existing canonical-null return IR, preserving its
  source line and dead-tail validation. It is admitted only where an untyped
  value return was already admitted; top-level, typed array/bool and constructors
  still reject it. No new void/never/error semantics.
- A declared public property of a statically known generated bare-variable
  receiver may use []=RHS. New assign_property_append IR has the same property
  target shape as assign_property. Receiver and RHS evaluate once; after RHS,
  the current property is read, existing ArrayAppend appends and Set writes its
  value copy. Thus RHS changes to that same slot are retained. The actual slot
  must be an array at append time: null/bool/scalar autovivification is excluded.
  No indexed/nested/chained property targets, references or compound writes.
- Only a known class method whose entire body is one unconditional return of
  new DeclaredClass(args...) contributes a known object-return class. The body
  is still fully validated, including constructor arity and arguments. This
  enables the existing guarded public-object (array) snapshot, not a general
  unknown-value cast, subclass/type inference or arbitrary factory analysis.
  Multi-statement/conditional/passthrough/global functions remain unknown.
- substr accepts an actual string, actual int64 offset, and optional int64|null
  length (two/three source arguments). It slices bytes, clamps negative/outside
  offsets and positive/negative lengths, returns empty string for empty selections,
  and treats omitted/null length as the remaining tail. All inputs are checked
  even for an empty selection. Bounds avoid negating MinInt64 or adding unbounded
  positive lengths. No scalar coercion, Unicode character slicing or PHP warning/
  exception/message equivalence. The runtime API has three arguments; codegen
  supplies nil for omitted length.

The substr contract follows PHP's [official manual](https://www.php.net/manual/en/function.substr.php);
method defaults follow [argument rules](https://www.php.net/manual/en/functions.arguments.php).
The parser method's coalesce left operand is the local length variable, not a
property. No new property-left coalesce operation is needed or admitted here.

Owned parser_output.php observes defaults, bare/fallthrough null, source factory
casts, seed-array separation/order, same-slot RHS-reset append ordering, method
strtok continuation and byte substr cases. The existing admitted-PHP differential
harness compares stdout/normal exits; stderr/exceptions/other effects remain
unmeasured there. Runtime units cover NUL/invalid UTF-8/extreme bounds and actual
domain errors. Source/IR negatives preserve unsupported forms and malformed
targets; statement maps remain attached to the original append line.

Capability gohp.cli.generate remains WebMCP-excluded: local developer compiler,
not a browser/service operation. No route/auth/dependency/CI changes or production
target creation. Full parser/tokenizer/stack, other methods and general PHP
object/reference/error behavior remain outside this claim.
