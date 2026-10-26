# Experimental PHP 8.3 local regex output (B3)

This connects the [B2 value adapter](preg83-value-wrapper.md) to a deliberately
closed PHP source form. It is **not** default `preg_*` support, PHP 8.4 evidence,
general references, a WordPress parser port, or a native dependency distribution.

## Admission and signature

Pass `-php-version 8.3 -experimental-pcre2` to `gohp`, or use
`frontend.ParseWithOptions(source, "8.3", frontend.Options{ExperimentalPCRE2: true})`.
The resulting IR records `RegexProfile: "pcre2-php83-v1"`; generated source and
CLI provenance identify that profile. Empty/default profiles reject regex calls;
unknown profiles and the experiment on PHP 8.4 fail before artifact publication.

```php
<?php
function match_word(string $subject, int $offset): array {
    $matches = null;
    $found = preg_match('/(?<word>a)(b)?/s', $subject, $matches, PREG_OFFSET_CAPTURE, $offset);
    return ['result' => $found, 'matches' => $matches, 'error' => preg_last_error()];
}
```

The five explicit arguments are mandatory. The pattern is a literal slash-delimited
`/s` string; subject is a string literal or statically string bare local; offset
is an integer literal (including negative) or statically integer bare local.
Ordinary PHP string escape restrictions of this frontend still apply.
Only `PREG_OFFSET_CAPTURE` is admitted, not a numeric/general flag expression.
Output must be a previously directly `null`-initialized bare local, distinct
from the return local. Neither destination can appear among input reads.
Reinitializing output to literal null permits a subsequent call.

Calls must be direct assignment statements at the root of a program, function
or method body, not inside `if`, `while`, or `foreach`. This is no path-sensitive
join contract. No property/index output, undefined local, derived null, generic
by-reference parameter, nested call expression, named/unpacked arguments, input
side effects or dynamic pattern is admitted. Both regex names are reserved PHP
builtins on both active targets. `preg_last_error()` is a zero-argument integer
expression under the same profile, including ordinary nested generated calls.

## IR, publication and state

`preg_match_local` is an explicit statement signature: `Name` is the result
local; `Target` is the bare output variable; `Expr` is canonical `preg_match`
metadata containing only pattern, subject and offset. The statement itself
encodes the fixed flag. Source and manually constructed IR are both checked;
hidden operands, parameters and unrelated payloads are rejected.

The whole scope stays in the existing `map[string]any` lane. Nested capture
arrays never enter scalar frames or canonical flat-array `ValueSlot` storage.
No raw pointer, address IR or implicit value-as-output argument is introduced.
This is a temporary detached-value adapter boundary, not canonical nested COW.

Generated code evaluates each pure input once, calls `Execution.PregMatch83Local`,
then deep-clones completed captures into the output local before assigning the
distinct scalar result. B2 completes the result and updates execution error
state before either local is published. No user code can observe the intervening
steps in this closed subset; this does not establish general PHP evaluation order.
The return retains `int64(1)`, `int64(0)` or `false`; its static type is unknown
because the existing IR has no union type. Existing downstream actual-type
guards still apply; implicit conversion to a declared bool is not promised.

Regex and `strtok` share the existing whole-program Execution propagation.
Global context wrappers, synchronous generated function/method/constructor and
arrow/callback calls pass the same pointer; standalone host entry points create
a fresh execution. Regex state does not share the strtok cursor. No process-global
error cell, persistent native pointer or asynchronous-context contract is added.

## Backend and failure boundary

Admission does not enable native linking. Build generated code with the explicit
`gohp_pcre2` tag, `CGO_ENABLED=1`, headers and a selected static PCRE2 archive as
documented for [B1](optional-pcre2-bridge.md). CGO0 still builds but executing a
match fails explicitly as unavailable. It never fabricates PHP false/no-match.

Limits are fixed per match: 100000 match, 10000 depth, 131072 KiB heap; B1 pattern
body/subject bounds remain 4096 bytes. Offset must be in the signed byte range
`[-len(subject), len(subject)]`. Runtime checks still apply to typed parameters;
compilation does not prove pattern validity, backend availability or input bounds.
The measured B2 match/no-match/backtrack-limit outcomes set error 0/0/2.
Other inputs/compile/engine errors panic explicitly and preserve the previous
error state; warning/exception text is not PHP-equivalent. No JIT, UTF/UCP,
locale folding, all-PCRE syntax equivalence, ini-setting compiler support, or
PHP 8.4 compatibility is claimed.

## Verification and capability boundary

Compiler/profile/malformed-IR tests, CLI publication, adapter guards and generated
CGO0 unavailability are default tooling checks, not native PHP regex evidence.
The opt-in `TestPreg83GeneratedNativeComparison` separately builds actual generated
Go with the enabled backend and compares typed ordered observations to fresh
pinned PHP 8.3.36. Set `GOHP_PREG83_GENERATED_PHP_IMAGE` to the exact image in that
test; absent configuration explicitly skips the comparison. It is not registered
as a supported builtin in the default CGO0 compatibility ledger.

Measured local result: [2026-10-27 evidence](evidence/preg83-b3-2026-10-27.md),
including failed setup attempts, corrected configuration, exact source identities
and eleven typed native/generated observations. This is Darwin ARM64 enabled-backend
evidence; standard Linux CGO0 CI is a separate regression gate.

The compiler flag is a developer-local source-generation option. UI / WebMCP /
MCP capability is **excluded**: no server operation, browser call, authentication
route, external resource mutation or automatic submit is introduced. Existing
exclusive fresh-path CLI publication is unchanged. No Gossip pin, WordPress
source, release, vendor archive or production deployment is changed here.
