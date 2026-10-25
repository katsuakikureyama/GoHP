# Bounded foreach continue

Bare `continue;` (case-insensitive) skips the remainder of the current iteration
of the nearest lexically enclosing `foreach`. It is permitted in nested braced
`if` / `elseif` / `else` blocks, nested foreach loops, named functions and methods
containing their own foreach. Calling a function from a loop does not give that
function a loop context. It does not exit the loop or advance an outer loop.

The frontend emits a payload-free `continue` statement in IR v0.1. Semantic
validation tracks foreach context separately for every function/method and
rejects continue outside a loop, even in a dead branch. Expressions, levels,
targets, bodies and other hidden payloads on a continue IR node are rejected.
Codegen emits an ordinary Go continue with the PHP statement source mapping;
no new runtime interpreter or loop protocol is introduced.

Statements after a bare continue in the same block are validated but not emitted.
The generator still traverses that tail to detect undefined variables and other
codegen-only violations. Dead assignments do not overwrite the type state at the
continue point. This is not general control-flow narrowing: branch-only bindings
still do not escape, and all supported/unsupported dead statements are checked.
Existing foreach snapshot iteration and array-copy behavior remain unchanged.

Explicit numeric levels (including `continue 1`), expressions, unbraced bodies,
break, while/for/do loops and switch are not added. No warning/exception
equivalence or PHP 8.4 support is claimed. Tests link through catalog entry
`php.syntax.foreach_continue`, status `partial`; native comparison uses the
PHP 8.3 target with oracle >= 8.3.36.

## Consumer boundary

Gossip PR #35 identified three independent blockers in the original WordPress
`_wp_array_get`: default parameters, bare/multiple isset and foreach continue.
This change addresses only the third. The original function is not certified:
defaults and isset remain unsupported, and adoption requires a new Gossip-owned
compiler pin and unchanged-source differential evidence. No Gossip pins/corpus
or production configuration are changed here. `array_slice` remains deferred.

This extends local `gohp.cli.generate`, still WebMCP `excluded` (developer-local
translation, no browser operation/route/auth change or automatic submission).

Reference: [PHP continue manual](https://www.php.net/manual/en/control-structures.continue.php).
