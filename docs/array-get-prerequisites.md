# Bounded prerequisites for nested array lookup

These additions enable the syntax used by WordPress `_wp_array_get`; they do not
establish compatibility of WordPress itself. Gossip performs the unchanged-source
comparison separately, using an exact compiler pin.

- Global named functions may have trailing **untyped** parameters with literal
  `null` defaults (case-insensitive). Calls accept the required-to-total positional
  argument range. Omission pads with canonical nil; explicit values, including
  null, retain the existing clone-at-parameter boundary. A literal `array_map`
  callback may have trailing null defaults if its range includes one argument.
- `isset` accepts defined bare variables and the existing one-level indexed
  variable addresses, singly or in a list. Lists lower to Go `&&` in source order;
  an unset first operand prevents subsequent runtime guards from executing.
  Bare variables are set unless canonical nil. Static validation still examines
  every operand, including dead ones. Indexed roots/keys retain their strict
  array and int/string runtime domain.
- Bare `continue;` reuses the implementation merged in GoHP #62: nearest lexical
  foreach, including nested conditionals. Additional composition fixtures here
  do not replace that implementation or its evidence.

Non-null/expression defaults, typed/method/arrow defaults, named arguments,
variadics, references, undefined-variable suppression, nested/property isset,
scalar-root PHP coercions, numeric continue levels and general loop support
remain excluded. No control-flow narrowing is implied by an isset result.

The compatibility catalog marks these contracts partial. `gohp.cli.generate`
remains WebMCP-excluded: compiler CLI only, no browser/server capability, route,
authentication or deployment change. Native PHP fixtures and runtime/compiler
boundary tests are registered in the catalog; evidence distinguishes compile
success from a measured PHP comparison.

Primary specifications: [arguments](https://www.php.net/manual/en/functions.arguments.php),
[isset](https://www.php.net/manual/en/function.isset.php),
[continue](https://www.php.net/manual/en/control-structures.continue.php).
