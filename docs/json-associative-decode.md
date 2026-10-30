# Bounded associative JSON decoding

The measured consumer is the unchanged original parser's
`json_decode($matches['attrs'][0], true)` at line283. The retained full-parser
corpus contains nested string/int/bool/null attributes and malformed JSON; its
observer records both last-error code and message. That is why the error readers
are part of this feature, not unrelated builtin-count expansion.

## Closed source and value contract

- Exactly `json_decode(text, true)`: actual string input, literal `true` second
  argument, default nesting limit512 and no flags/extra arguments. Known wrong
  input domains reject before output; inferred-unknown inputs use a runtime guard.
  Object mode, dynamic associative mode, custom depth, flags/constants, weak
  input coercion and builtin string-callable dispatch remain excluded.
- JSON objects and arrays become ordinary ordered runtime Arrays, retaining
  declaration/input order, nested value copies and null. Object keys use existing
  PHP-array numeric-string normalization. A duplicate normalized key replaces
  its value at the original insertion position, not a sorted Go-map position.
- Strings preserve UTF-8/decoded escape bytes, including escaped NUL and valid
  surrogate pairs. The other values are actual bool, null and int64. Root scalar
  values are not forced into arrays; semantic result type remains unknown/mixed.
- **All floating-point-form numeric tokens** (fraction or exponent), and integer
  tokens outside int64, are outside this profile even if later overwritten by
  a duplicate key. Syntactically valid such documents fail explicitly; they are
  never rounded, stringified or mistaken for malformed-JSON null. General float
  support is not introduced. Invalid syntax still takes its parse-error path.
- Ordinary decode failures return canonical null and publish a last-error code;
  valid supported JSON, including literal `null`, clears it. Default depth512
  allows at most511 nested containers. Supported error categories are depth1,
  state mismatch2, control character3, syntax4, UTF-85 and UTF-1610; error-message
  readers follow their PHP strings. A failure never exposes a partial array.

These boundaries follow the official [decode manual](https://www.php.net/manual/en/function.json-decode.php),
[JSON errors](https://www.php.net/manual/en/json.constants.php),
[last-error code](https://www.php.net/manual/en/function.json-last-error.php) and
[last-error message](https://www.php.net/manual/en/function.json-last-error-msg.php).
The PHP8.3 parser/scanner documents depth and error precedence; the Go
implementation is independent, not copied PHP source or a PHP interpreter.
Native target-specific CI evidence, not those references or local lowering,
determines the measured case-limited compatibility result.

## Execution and host boundary

`json_last_error()` and `json_last_error_msg()` take no arguments and read only
the same explicit `rt.Execution` as decode. Its initial JSON error is0, independent
of strtok and optional PCRE state. Context-aware F/M/New and arrows/callbacks
share the invoking Execution; legacy host entry wrappers create fresh contexts.
Use a fresh instance per logical request, never copy or concurrently share it.

Runtime methods are `Execution.JSONDecodeAssociative(text)`, `JSONLastError()`
and `JSONLastErrorMessage()`. A nil context or unsupported input/value profile
panics before JSON-error publication, preserving prior JSON state. These are
boundary failures, **not PHP exception/message equivalence** for unsupported
domains. No global error singleton, HTTP/request manager or dependency is added.

The frontend and IR0.1 shapes are unchanged: ordinary validated call nodes
lower to direct runtime methods. Mixed decoded values stay on existing legacy
storage, not a widened owned/scalar lane. Other constants, json_encode,
json_validate, arbitrary PHP extension behavior, serialization and resource
exhaustion equivalence remain unsupported.

## Evidence and capability scope

The separate owned fixture observes nested attributes, key order/copies, scalar
types, Unicode, malformed/depth boundaries and error-state reset/sharing.
Runtime units and external generated-entry checks cover actual-domain rejection,
unsupported numeric profiles and execution isolation. This is not original
parser generation/build/runtime or Gossip application acceptance; its exact
source/manifest/compiler/oracle comparison remains a separate consumer gate.
Local Docker was unavailable; no daemon/image was created or downloaded. Native
PHP8.3/8.4 observations come from the existing independent pinned CI jobs.

This extends the existing local `gohp.cli.generate` capability, WebMCP
**excluded** because it is developer-only compilation of caller-owned files.
Input/report/publication/status contracts, routes/auth and deployment remain
unchanged. No production target is configured, and no consumer pin is rewritten.
