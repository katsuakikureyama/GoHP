# PHP 8.4 target contract

PHP 8.4 is an active GoHP target for the existing documented closed subset. It
uses the compiler label `8.4`; the native oracle admission floor is PHP
**8.4.27**. CI runs the applicable bounded cases independently under the pinned image
`php:8.4.27-cli@sha256:debadc2e36b857657846e2e952cbaf52702e55482dd3c2c43806f68d8c534a70`.
The resulting report is an 8.4-only record. It does not reuse an 8.3 result,
and no target receives a differential-pass verdict until its own CI execution
has produced complete evidence.

The target is governed by the [PHP 8.x compatibility direction](php8-versioned-compatibility.md):
the native PHP process and generated Go process use the same selected target,
and comparison results remain case-limited to the dimensions actually measured.
Current differential evidence observes stdout and successful exits. It does not
claim PHP warning, deprecation, exception, configuration or side-effect
equivalence.

## PHP 8.4-specific profile

PHP 8.4 adds `array_find`, `array_find_key`, `array_any` and `array_all` to core.
GoHP implements [bounded array_any](array-any-boundary.md) and
[bounded array_all](array-all-boundary.md) subsets.
For target `8.4`, all four names are
reserved against user declarations, case-insensitively. A declaration therefore
fails during semantic validation instead of compiling a conflicting function.
Calls to array_find and array_find_key remain unsupported. Both predicates require a literal
declared callback with two untyped value parameters and an explicit bool return.
The profile does not reproduce native PHP diagnostics.

PHP 8.4 also deprecates class `_`. GoHP rejects that class declaration for
target `8.4`, rather than emitting a generated program which silently omits or
changes the native deprecation. The rejection is deliberately not equivalent to
the PHP warning/deprecation behavior and must not be reported as such.

For this target, the bounded predefined expressions `PHP_MAJOR_VERSION` and
`PHP_MINOR_VERSION` resolve to integer `8` and `4` from `Program.PHPVersion`.
They are not derived from the host executable or from the `8.4.27` oracle floor.
The same cross-target contract yields `8` and `3` for target `8.3`; it is not an
8.4-only reservation rule. See [target version constants](target-version-constants.md).

These decisions follow the PHP 8.4 migration references for
[new features](https://www.php.net/manual/en/migration84.new-features.php),
[incompatible changes](https://www.php.net/manual/en/migration84.incompatible.php),
and [deprecations](https://www.php.net/manual/en/migration84.deprecated.php).
They apply only where GoHP has an explicit profile rule; they do not declare the
rest of PHP 8.4 compatible.

## Boundaries that remain unsupported

The 8.4 target retains the existing subset boundaries. In particular, GoHP does
not support nullable typed parameter or property defaults, `exit`/`die`, property
hooks, asymmetric property visibility, arbitrary new PHP 8.4 syntax, general
callbacks, references, dynamic properties, inheritance, Composer packages or
WordPress as a whole. No new runtime coercions, warning behavior or exception
model is introduced by this target.

When later PHP 8.4 features are added, each change needs a bounded contract,
target-specific positive and negative cases, and separate 8.3 and 8.4 evidence.
An intentional native difference is recorded by target; it is never normalized
away or inferred from a result for the other minor.
