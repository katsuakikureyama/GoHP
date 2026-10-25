# Boolean-validation subset

This addition supports a bounded class of scalar/array validation functions,
motivated by Gossip's unchanged WordPress `wp_validate_boolean` fixture.
It is not general PHP control flow or object/coercion compatibility.

## Implemented contract

- Braced one-arm `if (condition) { ... }`, including nesting and early return.
  No `else`, `elseif`, alternate syntax or unbraced bodies.
- `&&` lowers to Go's short-circuit operator: evaluate the left side once, and
  evaluate the right side only if needed. Both sides must still be valid within
  the compiler's supported subset; dead code does not bypass validation.
- Scalar `===` compares type and value for nil, bool, int64 and string. Arrays
  and objects are rejected, including runtime-unknown values of those types.
  Unparenthesized chains of `===` are rejected as non-associative.
- `(bool)` and condition truthiness accept nil, bool, int64, string and GoHP
  arrays only. False values are nil, false, integer zero, empty string, string
  `"0"` and empty arrays. Unsupported dynamic value types fail explicitly.
- `is_bool`, `is_string` and `is_array` take exactly one value and return bool
  type predicates. `is_array` accepts existing scalar/null/array/class-instance
  values: empty, indexed and associative arrays are true; all non-arrays are
  false. It does not convert, walk or mutate the value. Array-like objects are
  not arrays; no float/resource/extension-value syntax is added.
- `strtolower` takes exactly one actual string, mapping ASCII A–Z bytes to a–z
  and leaving other bytes unchanged. No weak coercion or Unicode case folding.
  PHP 8.3 uses ASCII semantics. The oracle additionally selects and verifies
  `LC_CTYPE=C` for reproducibility; historical PHP 7.4 results are not current support.
- New builtin names are case-insensitive and reserved against redeclaration.

PHP references: [boolean conversion](https://www.php.net/manual/en/language.types.boolean.php),
[short-circuit operators](https://www.php.net/manual/en/language.operators.logical.php),
[precedence](https://www.php.net/manual/en/language.operators.precedence.php), and
[strtolower](https://www.php.net/manual/en/function.strtolower.php), and
[is_array](https://www.php.net/manual/en/function.is-array.php).

## Scope and evidence

Assignments to existing variables persist if a branch executes. New variables
defined only inside a branch cannot be read outside it, even when the condition
is a literal true. Semantic types for existing bindings are weakened if a branch
can change their type. This conservative rule avoids claiming definite assignment
where a branch might not execute; it is not full PHP data-flow inference.

Operator precedence uses the PHP 8.3 `+` above `.` relationship, above
`===`, above `&&`; bool casts bind more tightly. Statement maps cover nested
conditional headers and bodies. No IR schema fields or report schema change.

The owned `booleans.php` native differential fixture covers truthiness, guarded
string lowering, strict type distinction, nested returns, existing-variable
assignment, precedence and a visible RHS marker for short-circuit execution.
Explicit negative checks cover unsupported syntax, domains, builtin arities,
reserved declarations and branch-local variable escape. Runtime checks preserve
failure boundaries for unsupported dynamic values. The catalog records partial
contracts and target-specific results, never unconditional support.

The separate `is_array.php` fixture observes predicate results using `is_bool`
and strict true/false comparison, covers both array syntaxes and generated
objects, and checks single argument evaluation and short-circuit array guards.
It checks original values after calls. Runtime-only tests also distinguish
native Go slices/maps/pointers from the runtime's Array representation; those
Go values do not imply new PHP source types. The predicate does not add
branch-sensitive type narrowing: both sides of `&&` are statically validated.

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
