# Bounded strtok and execution ownership

`strtok(string, delimiters)` initializes or replaces a scan; `strtok(delimiters)`
continues the same scan. Only actual strings and positional calls are accepted.
Each delimiter byte is a separator, repeated/leading separators are skipped,
and returned tokens are nonempty strings (including `"0"`) or actual `false`.
An empty delimiter set returns the remaining bytes. Delimiters may change on
each continuation; the previous call already consumed one boundary byte.
No Unicode segmentation, coercion, null second argument, named/unpacked args or
builtin callable lookup is provided. The compiler treats string|false as unknown
and retains downstream string/array/bool-return guards.

The scan starts only with two arguments. In PHP 8.3, skipping a suffix consisting
entirely of delimiters clears its state; another continuation then warns. GoHP
rejects uninitialized/cleared-state continuation explicitly instead of claiming
warning/message equivalence. Reaching the end immediately after a token, or
starting with an empty input, retains an exhausted scan returning false. A new
two-argument call always starts over. Both domains are checked before mutation.

## One context per logical execution

Runtime `Execution` owns the cursor and input; no package-global mutable PHP
state, implicit goroutine state or shared singleton is used. A fresh instance
is required per request/job. Share its pointer synchronously within that one
execution; do not copy, concurrently use or recycle it across requests. Input
may remain retained until context disposal or a replacement scan. A nested
initializer intentionally replaces its caller's scan, matching PHP, rather than
creating a function-local token stream.

Only programs containing a direct strtok call emit execution-aware lowering:

- `main` creates a fresh context. All generated direct functions, methods,
  named array_map adapters and arrow invocations propagate it.
- Existing `F_name(args...)` / `M_name(args...)` signatures remain callable;
  each host invocation begins a fresh logical execution. Two separate wrapper
  invocations must not be used as a shared start/continue pair.
- New `FContext_name(*rt.Execution, args...)` / `MContext_name(...)` entries
  allow a host to share state deliberately across same-request entry calls.
  Nil context is rejected. These entries are emitted only for stateful programs;
  consumers must pin compiler/generated artifacts, not assume universal presence.
- Contextual arrows receive the invoking context, never retain a creator's scan.
  Returning an arrow across entries therefore does not leak the old request.
  `rt.InvokeWithExecution` shares explicitly; plain `rt.Invoke` starts fresh.
  Legacy stateless arrows and argument/result copying remain supported.

This is an explicit execution-state contract, not an HTTP/request framework,
thread-local PHP runtime, lifecycle manager or general Closure implementation.
Foreign object method dispatch and cross-generated-module ABI compatibility are
not broadened. No termination limits or implicit source replay are added.

## Evidence and remaining scope

Catalog `php.builtin.strtok` is partial. Tests cover native PHP 8.3 comparison,
byte scanning (NUL/invalid UTF-8 included at runtime), cursor/reset semantics,
argument/state rejection, source maps, malformed IR, generated guards, nested
function/method/callback sharing and same-process independent host entries.
The generated host test also interleaves explicit contexts, passes an escaped
arrow to another context and runs independent wrapper calls concurrently.
This is not a benchmark or a concurrency guarantee for a shared context.

Owned whitespace-token/prefix/while/break composition is not an unchanged-source
WordPress render test. HTML processing and exact-pin consumer adoption remain
separate. No Gossip pin/source/corpus changes. Capability remains developer-local
`gohp.cli.generate`, WebMCP `excluded`; no browser/server/MCP operation,
route/authentication, deployment, external save or automatic submission.

References: [PHP strtok manual](https://www.php.net/manual/en/function.strtok.php),
[PHP 8.3.36 implementation](https://github.com/php/php-src/blob/php-8.3.36/ext/standard/string.c#L1037-L1112).
