# Bounded string empty — evidence (2026-10-29)

Clean code `d723bfdead1f014488213ff9652310a3b473cfce`, tree
`69300e2dac5c2bcd4c10e51f6551a951bd088e65`, based on main5776a7b.
Owned test fixture was recorded separately at758f707 before the parent production
commit; no test ran at that incomplete intermediate workspace. Parent corrected
the ABI draft's stateless function-name lookup before the verification below.
This evidence-only addition reuses unchanged successful code/configuration.

## Focused verification

Cached Go1.23.0 executable first in PATH, Darwin/arm64, CGO0,
`GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOWORK=off`, existing shared
`GOCACHE=/private/tmp/gohp-go-cache`. Default local target8.3:

```sh
go test -count=1 -run '^Test(ArrayEmpty|EmptyString|EmptyValue)' ./internal/codegen ./runtime ./tests
go test -count=1 ./internal/compat
```

PASS once: compiler0.531s/runtime0.892s/generated guard+external ABI3.424s;
catalog0.647s. No test failures. New source `empty_value` checks actual array or
string; only `''` and `'0'` are empty strings. Old raw `empty_array` and
`runtime.ArrayEmpty` stay array-only. Unsupported null/bool/integer/object/
callable/foreign values reject, even though general PHP empty accepts more.
Scopes stay legacy storage. Source maps retain original file/line.

Seven new checks include the owned differential fixture, new IR/map and legacy
raw-IR boundaries, runtime domains and generated external host observations.
Native comparison is pending at this checkpoint and is completed by the PR
result; no local native PHP installation or native-loading success is assumed.

## One unchanged complete original closure probe

The existing Gossip whole-file manifest is unchanged, SHA256
`89f897c64146705559eae41bb50d49aff57abad0a11ee02d16d27fe78d64adaa`.
Its Parser/Block/Frame, license and version bytes remain the original selection
from WordPress54106f6. No concatenation, extraction or replacement algorithm.

```sh
go run ./cmd/gohp -manifest GOSSIP_ROOT/gossip/tests/gohp/parser-source-closure.json \
  -source-root GOSSIP_ROOT -entry wp-includes/class-wp-block-parser.php \
  -php-version 8.3 -experimental-pcre2 -out FRESH/parser.go -report FRESH/report.json
```

At the clean code above, run once under the same local environment. Exit1,
stdout0bytes, stderr114bytes; stderr SHA256
`0aa3c061c7896e4450a840ff8a41e16ac9c042c6b8e496d72eae9ac57bb41dba`:

```text
wp-includes/class-wp-block-parser.php: PHP004 line 394: array cast requires a known generated class
exit status 1
```

The last line is from `go run`. Prior empty-string rejection354 passes; the next
semantic boundary is original `(array) $stack_top->block`394 after an inferred-
unknown stack/object path. No Go/report/staging artifacts were created; only
stdout/stderr logs remain in a task-owned temporary evidence directory.
No original Go build/execution/full12case comparison or G83-PARSER acceptance.
Remaining snapshot/stack/tokenizer/JSON work stays open. No consumer repin,
source/manifest/runtime-loading/auth/route/deployment change or enabled-PCRE
remeasurement. Existing local CLI/WebMCP exclusion retained; no dedicated cache
or persistent build binary. Native tests and the successful original probe are
not rerun for this documentation-only addition.
