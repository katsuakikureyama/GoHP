# Exact scalar constructor parameters

Public `__construct` may use `int`, `string` and `bool` parameter annotations,
case-insensitively, mixed with untyped parameters. An optional trailing suffix
of **untyped** `= null` parameters retains the existing min/max arity contract.
Typed defaults (including implicit-null forms), nullable/union/class/float types,
array/callable constructor annotations, nonpublic constructors, return annotations,
direct constructor re-entry and other magic methods remain excluded.

The admitted input domain is exact int64/string/bool values only. PHP's weak
scalar conversion is deliberately outside the subset: e.g. a numeric string
passed to an int parameter is rejected, not converted. PHP strict_types syntax,
TypeError/warning equivalence and invalid-call side-effect equivalence are not
implemented. This is not full PHP typed-argument compatibility.

Known mismatches reject before code generation. Unknown values and host calls
go through `rt.ScalarParameter` before the constructor body. Argument evaluation
still precedes construction; existing defaults and once-only execution remain.
Annotations constrain entry values, not later local assignments or untyped
properties. Static entry facts are updated by assignments; unknown later values
retain runtime guards. Scalars pass unchanged, including empty strings, byte
strings and false/zero. No new persistent property type is introduced.

IR uses existing ParamTypes, now allowing canonical int/string/bool only on
constructors. Global functions retain array/callable parameter support; ordinary
methods and arrows do not gain scalar annotations. Generated New/NewContext and
M/MContext entry paths all reuse the same guards and existing Execution dispatch.

Motivation: original WP_HTML_Span and WP_HTML_Text_Replacement constructors use
int/int and int/int/string, respectively. This removes one dependency of those
data classes without changing their source. Their original-source acceptance,
the tag processor's state machine and complete rendering still require separate
evidence; the owned fixture is not a WordPress mock or an adoption claim.
See the [HTML dependency map](html-processor-dependencies.md).

Catalog checks cover source/IR rejection, entry lowering/source maps, exact
runtime domains, generated dynamic guards and PHP8.3.36 comparison. Existing
differential evidence compares stdout/normal exits, not exceptions or arbitrary
side effects. Later PHP targets remain disabled.

Capability: existing local `gohp.cli.generate`, WebMCP `excluded`; no new route,
authentication, service operation, dependency or deployment.

Reference: [PHP type declarations](https://www.php.net/manual/en/language.types.declarations.php).
