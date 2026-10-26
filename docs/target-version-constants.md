# Target version constants (bounded)

GoHP recognizes exactly two PHP predefined constants in the closed subset:
`PHP_MAJOR_VERSION` and `PHP_MINOR_VERSION`. They are exact, uppercase,
unqualified identifier tokens, used as ordinary expressions and not followed by
`(`. They resolve to integers from the selected `Program.PHPVersion`: target
`8.3` resolves them to `8` and `3`, while target `8.4` resolves them to `8`
and `4`.

This is target semantics, not a query of the PHP executable which happens to
run the compiler or of the native oracle. The selected major/minor is therefore
independent of the host PHP version and of the minimum native-oracle patch
release. Separate targets compiled in the same Go process resolve independently;
there is no mutable process-global current target.

PHP documents these names among its [predefined constants](https://www.php.net/manual/en/reserved.constants.php).
The general [constants documentation](https://www.php.net/manual/en/language.constants.php)
remains relevant to PHP as a language, but this contract does not implement its
general declaration or lookup mechanisms.

## Lowering and admitted contexts

IR `0.1` gains the additive `version_constant` expression kind. Its `Text`
field keeps the symbolic source name through parsing and semantic validation;
code generation resolves it against `Program.PHPVersion` and emits the
resulting `int64`. No serialized IR schema, runtime operation or global target
state is introduced.

The constants are admitted wherever the existing subset admits an ordinary
integer expression. That includes arithmetic, branches, echo, array values,
array-read keys such as `$items[PHP_MINOR_VERSION]`, global functions, methods
and capture-free arrows. A scalar-frame-selected function may contain them: the
constant remains a scalar expression and does not widen that storage boundary.

Constants and functions have separate namespaces. A same-named function
declaration or call continues through the existing function path; only the
unqualified non-call identifier spelling above is recognized as this predefined
constant.

## Exclusions

This increment intentionally does not add general PHP constant support. It
does not add `PHP_VERSION`, `PHP_VERSION_ID`, `PHP_RELEASE_VERSION`,
`PHP_EXTRA_VERSION` or other predefined constants; namespaced or qualified
constant syntax; user constant declarations; `define()` and `constant()`.
It also does not extend the parser's literal-only parameter defaults, property
defaults, literal array keys, write-key syntax, `isset` keys, coalesce keys or
unary-minus grammar. Existing supported assignment through a variable still
applies, so a constant first stored in a variable can be used as an existing
supported variable array-write key.

## Evidence boundary

`tests/fixtures/version_constants.php` deliberately prints target-dependent
values. It is intended to let the existing read-only two-archive comparison
report a `native-stdout-difference` when complete, compatible PHP 8.3 and 8.4
evidence is available, while each archive separately checks PHP against the Go
candidate. That comparison is an observation, not proof that a PHP minor caused
the difference. Coverage requires successful independent native CI evidence
for both targets; the fixture's presence alone is not a compatibility result.

No report schema, CLI option, target-admission policy, oracle configuration,
runtime API, capability or WebMCP exposure is added. The existing local CLI
capability remains excluded from WebMCP.
