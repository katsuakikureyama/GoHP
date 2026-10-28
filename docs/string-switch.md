# Nonnumeric string switch

The fixed parser's P1 gate needs a switch on token names. This increment accepts
braced `switch (value) { case 'name': ... default: ... }` for an actual
nonnumeric string selector and literal nonnumeric string labels. PHP loose
comparison equals byte equality in this domain. Numeric strings use a shared
PHP8 decimal/fraction/exponent grammar guard, including whitespace and overflow;
they are rejected before dispatch rather than compared as literal strings.

The selector runs once. First matching label wins, including duplicates. Empty
labels and ordinary fallthrough preserve source order; default may be anywhere
and is chosen only when no label matches. Without default, a miss runs no body.
Dispatch first chooses an integer clause index, then executes the clauses using
Go switch/fallthrough, retaining later duplicate labels as execution positions.
Unconditional return/break tails suppress fallthrough; dead tails are still
validated. Bare break targets the nearest loop/switch. Inner loops retain normal
continue behavior. A continue targeting switch directly (also through an if)
rejects because PHP emits a warning outside the quiet comparison profile.

IR uses `switch.Expr` plus `switch.Body` of `case`/`default` clauses. Each clause
has a statement Body; case has only a literal string Expr, default none. No
additional fields or general expression kind is introduced. Existing traversals
retain Execution routing and reject nested experimental preg output as before.
All scopes containing switch retain legacy map storage.

Flow checking conservatively widens every existing local written in any clause
before checking clauses and after switch. This accounts for earlier fallthrough,
no match, early break and loops. Branch-only locals do not escape, including
initialization solely in an earlier case; class/direct-null facts are not reused
after writes. The selector itself is checked/evaluated before these writes.

Excluded: numeric/nonstring selectors, numeric/dynamic labels, alternate syntax,
case semicolons, multiple defaults, explicit transfer levels and general PHP
loose equality. Static literal wrong types reject; unknown and derived strings
are checked at runtime. These failures are unsupported-input guards, not PHP
warning/exception equivalence. Original parser execution and P2–P5 dependencies
remain separate gates.

Catalog `php.syntax.nonnumeric-string-switch` tracks owned PHP8.3.36/8.4.27
differential evidence and compiler/runtime boundaries. Existing developer-local
`gohp.cli.generate`; WebMCP excluded. No route/authentication/deployment change.

References: [PHP switch](https://www.php.net/manual/en/control-structures.switch.php),
[numeric strings](https://www.php.net/manual/en/language.types.numeric-strings.php).
