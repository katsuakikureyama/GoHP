# rtrim with an explicit literal byte mask

This is a deliberately partial builtin contract, motivated by Gossip's original
WordPress `untrailingslashit($value)` → `rtrim($value, '/\\')` dependency.
PHP's [rtrim reference](https://www.php.net/manual/en/function.rtrim.php) describes
both optional default characters and `..` ranges. Those broader forms are not
included in this change.

Accepted domain:

- Exactly two arguments. The first evaluates to a string; no scalar coercion.
- The second is a string literal, possibly empty, without any `..` sequence.
- The mask is a set of individual bytes, not a substring or a Unicode rune set.
  Leading/interior bytes are unchanged; trailing mask bytes are removed.
- Case-insensitive builtin spelling. `rtrim` is reserved against user function
  redeclaration, as `count` already is.

Known non-string inputs, wrong arity, dynamic masks and masks containing `..`
fail before generation. Unknown argument types at function boundaries remain
guarded by an explicit runtime failure on non-string values. That failure is not
PHP TypeError/warning-equivalence evidence; PHP's weak coercion is not implemented.
Even literal `..` masks with no valid range are conservatively rejected.

The runtime uses a byte set and backward scan. A native differential case with
`rtrim('é', '©')` distinguishes byte trimming from Unicode-rune trimming: the
output can contain an incomplete UTF-8 sequence, compared as raw stdout bytes.
This is not a Unicode normalization or path canonicalization operation.

`tests/fixtures/rtrim.php` exercises accepted inputs through the real compiler
and the active PHP 8.3 >=8.3.36 oracle. Older two-target runs remain historical.
The catalog item `php.builtin.rtrim` links its
version-specific differential and rejection/guard checks. Current evidence is
recorded in the Draft PR and CI ledger; generation alone is not compatibility.

The WordPress functions and their parameterized fixture belong in Gossip, not
this compiler repository. Gossip must pin compiler/runtime together and compare
unchanged original excerpts separately. No full WordPress, default-whitespace
trim, range expansion, arbitrary PHP casts or side-effect support is claimed.

This remains `gohp.cli.generate`, explicitly excluded from WebMCP under the
[integration contract](integration.md). No service/authentication route,
dependency, release, deployment or licensing change is added.
