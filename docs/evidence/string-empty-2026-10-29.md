# Defined-string empty — 2026-10-29

Source `empty($definedBareLocal)` now accepts actual arrays or strings, with
neutral additive `empty_var` IR and guarded `runtime.VariableEmpty`. Strings
are empty only for exact empty bytes or one ASCII zero; arrays use outer count.
Legacy raw `empty_array` and `ArrayEmpty` stay array-only, including rejecting
strings. No general scalar coercion or undefined/address suppression. See the
[contract and official references](../array-empty-boundary.md).

## Focused local evidence

Clean code `62cc41fdc45bf2126acd7c54d70c28034d4557ec`, tree
`17e6ab69577bd640904405d85759596b9865961a`, base main5776a7b.
Cached Go1.23.0 Darwinarm64/CGO0/offline modules, shared existing
`GOCACHE=/private/tmp/gohp-go-cache`, cached Go also on child-process PATH:

```sh
go test -count=1 ./internal/frontend ./internal/codegen ./runtime ./internal/compat
go test -count=1 ./tests -run '^(TestArrayEmptyRuntimeGuards|TestStringEmptyFixture)$'
```

All PASS once: frontend0.362s, codegen0.465s, runtime0.202s, compat0.703s,
generated domain guards/committed fixture3.805s. New compiler checks cover typed
string returns, original substring/local/!empty form, arrow routing, bool result,
maps, neutral metadata/unsupported forms, and strict legacy rawIR separation.
Runtime tests cover explicit byte/value matrices, NUL/invalid UTF8, array input/
metadata preservation and unsupported foreign domains. Generated unknown null,
bool, integer, object and arrow-input guards reject instead of silently coercing.
No manual successful-suite rerun or verification-only extra agent.

Owned real fixture SHA256
`5bb597705d84495bb3082d43eea9cbac716e9f2ba4a6fb99977c904950e4c2ec` observes
zero-string distinctions, typed/untyped/arrow helpers, unchanged input, derived
HTML, string/array reassignment, widened locals, singleton falsy/nested arrays
and lazy guards. Eight new catalog checks are registered at the retained stable
empty feature ID. Native evidence requires each candidate/target automatic CI;
local fixture lowering is not native PHP PASS.

## One unchanged complete-original closure probe

WordPress/WordPress `54106f6c545fc8b881f6b39b7963002bece6c28c`, same complete
Parser/Block/Frame plus notice/provenance, unchanged
[Gossip manifest](https://github.com/katsuakikureyama/Gossip/blob/master/gossip/tests/gohp/parser-source-closure.json)
SHA256 `89f897c64146705559eae41bb50d49aff57abad0a11ee02d16d27fe78d64adaa`.
One invocation at clean code62cc41f and the above cached/offline configuration:

```sh
go run ./cmd/gohp -manifest GOSSIP_ROOT/gossip/tests/gohp/parser-source-closure.json \
  -source-root GOSSIP_ROOT -entry wp-includes/class-wp-block-parser.php \
  -php-version 8.3 -experimental-pcre2 -out FRESH/parser.go -report FRESH/report.json
```

Exit1, stdout0bytes, stderr114bytes:

```text
wp-includes/class-wp-block-parser.php: PHP004 line 394: array cast requires a known generated class
exit status 1
```

Stderr SHA256 `0aa3c061c7896e4450a840ff8a41e16ac9c042c6b8e496d72eae9ac57bb41dba`;
last line comes from go run. No generated Go/report/staging artifacts; the empty
temporary directory was removed. Prior string-empty354 now passes admission.
The next rejection is the original `(array)` snapshot of the popped frame's
block at394, whose inferred class is unknown. Earlier admission is not proof of
every earlier method's semantics. Full original generation/build/execution,
12-case/24-call comparison and G83-PARSER remain blocked. Remaining stack,
snapshot, tokenizer/JSON and other application dependencies still need proof.

No original/manifest/pin rewrite, extraction, concatenation, interpreter fallback,
new dependency or unrelated builtin. Existing131/older native successes reused.
Local generation capability/WebMCP exclusion updated in the same change set;
no new resource, input schema, route/auth/deployment or auto-submit. Devboard
unidentified/not synchronized. Source maps/report/IR version policies unchanged.
