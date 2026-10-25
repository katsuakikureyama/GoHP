# Boolean-validation subset

This addition supports a bounded class of scalar/array validation functions,
motivated by Gossip's unchanged WordPress `wp_validate_boolean` fixture.
It is not general PHP control flow or object/coercion compatibility.

## Implemented contract

- `true`, `false` and `null` literals accept ASCII letter-case variants and
  normalize to existing lowercase bool/null IR. This applies in expressions,
  returns and public scalar property defaults. Quoted strings and variable or
  property names retain their case. This does not add general constants,
  globally case-insensitive keywords or new coercions.
- Braced `if (condition) { ... }`, optional `elseif (...) { ... }` / `else if
  (...) { ... }` chains and optional final `else { ... }`, including empty bodies,
  nesting and early return. Each reached condition is evaluated once; the first
  truthy branch prevents evaluation of later conditions/bodies. Both spellings
  may mix and normalize to nested if/ElseBody. No alternative syntax or unbraced bodies.
- `&&` and `||` lower to Go's short-circuit operators: evaluate the left side
  once, and evaluate the right side only if needed. Both sides must still be valid within
  the compiler's supported subset; dead code does not bypass validation.
  `&&` skips its right operand when the left is false; `||` skips it when the
  left is true. Both return bool, not an original operand value. Keyword
  `and`/`or`/`xor` and bitwise `|` remain unsupported.
- Prefix `!` evaluates one operand once, applies the same bounded truthiness
  conversion as `(bool)`, and returns the opposite bool. Repeated `!!` is allowed.
  A one-operand `not` IR expression lowers to Go `!rt.Bool(operand)`; unknown
  unsupported types retain the existing runtime guard. This does not add `!=`,
  a `not` keyword or general constant expressions in property defaults.
- Scalar `===` and `!==` compare type and value for nil, bool, int64 and string. Arrays
  and objects are rejected, including runtime-unknown values of those types.
  `!==` returns the negation of `===` as bool, using `!rt.Identical(left, right)`
  and the same scalar guard, with no type coercion. Each operand is evaluated once.
  Unparenthesized chains of either operator, including mixed chains, are rejected
  as non-associative. Explicit parentheses permit comparing scalar bool results.
- `(bool)` and condition truthiness accept nil, bool, int64, string and GoHP
  arrays only. False values are nil, false, integer zero, empty string, string
  `"0"` and empty arrays. Unsupported dynamic value types fail explicitly.
- `is_bool`, `is_string`, `is_int` and `is_array` take exactly one value and return bool
  type predicates. `is_array` accepts existing scalar/null/array/class-instance
  values: empty, indexed and associative arrays are true; all non-arrays are
  false. It does not convert, walk or mutate the value. Array-like objects are
  not arrays; no float/resource/extension-value syntax is added.
  [is_int](is-int-boundary.md) returns true only for canonical int64 values;
  numeric strings and all noninteger values remain false without conversion.
- `strtolower` takes exactly one actual string, mapping ASCII A–Z bytes to a–z
  and leaving other bytes unchanged. No weak coercion or Unicode case folding.
  PHP 8.3 uses ASCII semantics. The oracle additionally selects and verifies
  `LC_CTYPE=C` for reproducibility; historical PHP 7.4 results are not current support.
- New builtin names are case-insensitive and reserved against redeclaration.

PHP references: [boolean conversion](https://www.php.net/manual/en/language.types.boolean.php),
[null literals](https://www.php.net/manual/en/language.types.null.php),
[else](https://www.php.net/manual/en/control-structures.else.php),
[elseif/else if](https://www.php.net/manual/en/control-structures.elseif.php),
[short-circuit operators](https://www.php.net/manual/en/language.operators.logical.php),
[strict comparisons](https://www.php.net/manual/en/language.operators.comparison.php),
[precedence](https://www.php.net/manual/en/language.operators.precedence.php), and
[strtolower](https://www.php.net/manual/en/function.strtolower.php), and
[is_array](https://www.php.net/manual/en/function.is-array.php).

## Scope and evidence

Assignments to existing variables persist if a branch executes. Both branches
are validated independently from the pre-if scope: a then-only binding or type
change cannot leak into else validation. New variables defined inside a branch
cannot be read outside it, even when both branches define them or the condition
is a literal true. Existing binding types become unknown if either branch changes
them, even if both branches agree on a new type. These are conservative compiler
restrictions, not PHP block scope; no path-sensitive or return-sensitive
definite-assignment/type inference is claimed.
This extends to every arm of an elseif chain, including conditions: earlier
branch-only bindings and type changes do not leak into later arms. Unreachable
conditions/bodies must still be valid in the closed subset.

Operator precedence uses the PHP 8.3 `+` above `.` relationship, above
`===` / `!==`, above `&&`, above `||`, above full ternary; prefix `!` and bool casts bind
more tightly. Calls and index reads belong to the operand of `!`, while
`!0 + 2` means `(!0) + 2`.
Statement maps cover nested
conditional headers and bodies. The internal `Stmt.ElseBody` field is optional;
nil means absent and a non-nil empty slice retains an explicit empty else.
Else headers/closing braces inherit the enclosing if's statement mapping;
statements inside else keep their own source lines. No CLI report schema change.
Normalized elseif conditions retain the elseif token's line; `else if` retains
the if token's line, including when the two words are separated by a newline.

The owned `booleans.php` native differential fixture covers truthiness, guarded
string lowering, strict type distinction, nested returns, existing-variable
assignment, precedence and a visible RHS marker for short-circuit execution.
Explicit negative checks cover unsupported syntax, domains, builtin arities,
reserved declarations and branch-local variable escape. Runtime checks preserve
failure boundaries for unsupported dynamic values. The catalog records partial
contracts and target-specific results, never unconditional support.

The `literal_case.php` fixture compares mixed-case literals in conditions,
arguments/returns, property defaults and strict membership against the native
oracle. Compiler tests retain canonical IR and source lines and verify that
case variants cannot bypass existing key/type/strictness guards. Variable and
quoted-string case are not normalized. Runtime behavior and IR schema are unchanged.

The separate `if_else.php` fixture covers supported truthiness, one condition
evaluation, lazy branch markers, empty/nested branches, early returns, updates
to existing bindings, array copies and foreach composition. Both arms remain
statically checked, including dead arms. Negative checks cover branch/sibling
binding leaks, unsupported forms and known invalid values; generated-binary tests
preserve the dynamic condition and selected-else builtin guards. Malformed IR
with ElseBody on non-if statements is rejected before generation.

The separate `elseif.php` fixture observes both spellings and mixed chains,
first/middle/final/fallback selection, once-only condition markers, skipped
conditions after an empty selected body, nesting and existing-variable updates.
Source-map checks cover normalized headers/bodies. Negative tests cover malformed
chains, unbraced bodies, dead invalid conditions/bodies and sibling scope leaks.
Generated binaries check both evaluated and skipped dynamic object conditions
for both spellings. No new truthiness types or variable inference is introduced.

The separate `logical_not.php` fixture observes `!`/`!!` results as bools across
the supported value domain, prefix/binary/postfix precedence, lazy branches and
operand evaluation markers. Negative tests cover known and dynamic objects,
even a statically invalid operand inside a short-circuited branch; malformed
arity and bool-result misuse remain rejected. No object truthiness, PHP warning
or exception equivalence, or general side-effect parity is claimed.

The separate `logical_or.php` fixture observes bool results across scalar/array
operands, once-only left-to-right markers, skipped right operands, chains,
precedence and guarded array operations. Generated-binary tests reject unknown
object values when evaluated on either side, but do not convert an unused dynamic
right operand. Known invalid objects/calls/variables still fail generation even
on a statically unselected side. Malformed IR arity, incomplete syntax, bool
array-key/result misuse and general property defaults remain rejected.

The separate `strict_not_equal.php` fixture observes scalar type/value differences,
bool results, precedence, parenthesized comparisons, one evaluation per operand
and short-circuited surrounding branches. Static negative tests reject arrays and
objects on either side (even in dead branches), mixed unparenthesized chains and
invalid syntax/result uses. Generated-binary checks exercise unknown array/object
operands on both sides; the existing scalar guard is not bypassed by a type mismatch.
This does not add loose `!=`/`<>`, array/object identity, floating-point values,
PHP exception equivalence or a general PHP operand evaluation-order guarantee.

The separate `is_array.php` fixture observes predicate results using `is_bool`
and strict true/false comparison, covers both array syntaxes and generated
objects, and checks single argument evaluation and short-circuit array guards.
It checks original values after calls. Runtime-only tests also distinguish
native Go slices/maps/pointers from the runtime's Array representation; those
Go values do not imply new PHP source types. The predicate does not add
branch-sensitive type narrowing: both sides of `&&` / `||` are statically validated.

Current native runs use PHP 8.3.36 or newer within 8.3 under the shared policy.
The former 7.4 / older 8.3 evidence is historical; 8.4+ needs separate admission
and validation. Integration does not validate Gossip's original-source pins.

GoHP does not vendor WordPress implementation code in this fixture. Gossip must
pin compiler and runtime together, retain original source hashes and separately
compare actual typed returns for the original function. Float/object/resource
truthiness, array identity, weak scalar conversion, other logical/comparison
operators, warnings/exceptions and full WordPress remain outside this change.

`gohp.cli.generate` remains explicitly excluded from WebMCP under the existing
local-compiler contract. No browser/S2S/MCP route, cookie forwarding, production
operation, dependency, deployment or license change is introduced.
