# Experimental PHP8.3 regex value adapter (B2)

The runtime now has `(*Execution).PregMatch83` and `PregLastError83`, an
experimental **Go value API**, not compiler-admitted PHP builtins. It builds
on the optional [B1 engine bridge](optional-pcre2-bridge.md). No PHP source is
transpiled by this API, and no caller variable is written by reference.
The separate [B3 local-output compiler experiment](preg83-local-output.md) now
connects a profile-gated source subset to this API. B2 itself still does not
resolve PHP output locations or provide general reference semantics.

## Closed API and outcomes

`PregMatch83(pattern, subject string, offset int64, limits Preg83Limits)` requires
a slash-delimited `/s` pattern and actual Go byte strings. Its single capture
mode is PHP's PREG_OFFSET_CAPTURE. There is no flag/coercion/default-argument
parser. The PHP string delimiters are removed without rewriting regex syntax;
unescaped inner delimiters are rejected. Byte offsets must lie in
`[-len(subject), len(subject)]`; negative offsets normalize against the **whole**
subject before the engine call. Limits are explicit, positive, and bounded by
B1 (match100000, depth10000, heap131072KiB; 4096-byte body/subject, 64 groups).

The input interface can accept a runtime Go string; a future compiler must
independently enforce its source-literal/pure-argument policy. It is not safe
to infer source admission from the existence of an exported runtime method.

| Engine result | `Preg83Result.Return` | `Matches` | Execution last error |
| --- | --- | --- | --- |
| Positive capture count | `int64(1)` | Fresh nested ordered `runtime.Array` | 0 |
| No match (-1) | `int64(0)` | Fresh empty array | 0 |
| Match limit (-47) | `bool(false)` | Fresh empty array | 2 |
| Unsupported input/backend/compile/other engine failure | Non-nil Go error; result is not a compatible value | No publication contract | Previous state preserved |

`EngineCode` remains raw PCRE2 metadata, not the PHP return or error number.
Invalid delimiters/offset/limits return `ErrPreg83Input`; disabled backend errors
match `ErrPreg83Unavailable`. B1's compile/engine errors remain distinct Go
errors. No unknown failure is converted to PHP false or no-match. This is
explicitly not PHP warning/exception behavior or a complete error-code map.
Both methods require a non-nil Execution using the existing runtime guard.

Each numbered capture is a pair `[raw string, int64 byte offset]`. Group0 comes
first; each named entry precedes its numeric entry. Unmatched middle captures
are empty/-1; trailing unmatched entries are absent, while a matched empty
capture has a nonnegative offset. Duplicate-name/number mappings are rejected
as unmeasured. Returned named/numeric pairs and later results have independent
entry storage, constructed through existing `NewArray` deep cloning.

## Ownership and execution state

The new private error field lives on the existing `Execution`, alongside but
independent of strtok state. Fresh executions start at zero. Success/no-match
reset it, and measured match-limit failure sets it to two, only after the
complete result is built. Synchronous nested **Go** helpers share their caller's
Execution pointer. Separate executions may be interleaved without shared state.
An individual Execution must still not be copied or used concurrently.

This is **detached legacy array value construction**, not a new reference/COW
engine. #111's ValueSlot/ValueTemporary adapter is flat-scalar-only and cannot
store these nested captures; neither it nor internal storage, ScalarFrame,
runtime.Array implementation, compiler or IR changes here. B3 must resolve the
writable-local/storage boundary before publishing output into generated PHP.

## Verification boundaries

The [2026-10-27 measurement](evidence/preg83-b2-2026-10-27.md) records the
successful enabled/default checks, 24 retained cases and seven fresh state
operations. It is runtime evidence only, not a generated-code compatibility entry.

- Opt-in `TestPregMatch83RetainedReference` reads the SHA-pinned Gossip PHP8.3.36
  record and directly compares all 24 typed returns, ordered capture arrays and
  last-error values. This reuses stored PHP evidence; it does not rerun PHP or
  compare against B1's fixture mapper. No WordPress source is copied into GoHP.
- Opt-in `TestPreg83NativeExecutionState` runs the owned
  `tests/oracle/preg83_state.php` once per scenario in two fresh pinned native
  PHP8.3.36 CLI containers, then compares seven operations with interleaved Go
  executions. It covers failure→nested success, failure→no-match, nested failure,
  fresh initial state and request isolation. PHP Fibers are not independent
  executions and are not used as an isolation oracle.
- Enabled generic tests cover capture ownership/order, explicit error kinds,
  guarded byte offsets/limits, and unsupported failures preserving state.
  Default tests verify unavailability and input guards without PCRE headers.
- Neither replay nor native-state checks are PHP→Go compiler differentials or
  PHP8.4 regex evidence. Their names are not added to the PHP compatibility
  catalog as supported builtins. Default CI stays CGO0 and does not exercise
  the enabled backend. Ordinary PHP8.3/8.4 CI remains a separate regression gate.

Supply `GOHP_PCRE2_REFERENCE_DIR` as documented for B1 and
`GOHP_PREG83_PHP_IMAGE=php:8.3.36-cli@sha256:98d555ff97b3865140ad31032508c3ca089fa97a8944b14ca9dc03ea05d788cd`
for the fresh state check. Missing variables skip those respective evidence
checks with explicit messages, never establishing PASS for those comparisons.
The image must already be present (`--pull=never`); tests do not install it.
Docker tests run nonroot/read-only/network-none with a read-only fixture mount,
dropped capabilities and bounded memory/CPU/processes. Process timeouts are
outer test guards, not PHP errors or C-call cancellation support.

No binary/vendor/release/production target is added. Dependency licensing and
distribution remain separate gates. UI/WebMCP/MCP capability is **excluded**:
developer-only library and local oracle tests, no service route, browser call,
resource mutation, authentication change or external account operation.

Reference semantics: [PHP preg_match](https://www.php.net/manual/en/function.preg-match.php)
and [preg_last_error](https://www.php.net/manual/en/function.preg-last-error.php).
Measurements, not these general API descriptions, bound the supported evidence.
