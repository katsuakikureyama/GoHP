# Declared class parameters — evidence (2026-10-29)

Code commit `35bfcaac799f4cce8a3f0c5fae71cc9acd4cdd60`, tree
`bdb6d517712789aade29aea07001246ee06bff67`, based on main `2045ab2`.
Tests below precede this evidence-only addition. Successful checks are reused,
not rerun for documentation changes.

## Local checks

Go 1.23.0, darwin/arm64; cached toolchain executable selected directly via PATH.
`GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOWORK=off CGO_ENABLED=0`,
`GOCACHE=/private/tmp/gohp-pcre2-b1-evidence.SqqeRt/cache`; default PHP target 8.3.
Repository-relative commands (no user-specific paths):

```sh
go test -count=1 -run '^Test(ClassParameter|ScalarConstructor(Boundaries|MalformedIR)|ScalarFunctionParameter(Boundaries|MalformedIR)|ArrayParameter(UnsupportedSource|MalformedIR)|ArrowCallable(UnsupportedSource|MalformedIR))' ./internal/codegen ./internal/compat
go test -count=1 ./internal/compat
```

Compiler checks PASS (0.288s). The first command's compatibility package had no
matching tests, so its 0.463s result is not counted as catalog validation. The
second command ran actual catalog tests once: PASS (0.414s). No failures.

The delegated external generated ABI check
`go test -count=1 ./tests -run '^TestDeclaredClassParameterHostEntry$'`
already passed once (1.208s) in the implementation workspace before commit.
It is reused, not rerun. The later change to the separate PHP fixture removes
an unsupported scalar ordinary-method annotation and does not affect this ABI
test. ABI test source SHA256:
`2e47347d473fe70e452c53cd5acb0f65fd699c6e29259856978cf14bba838451`.

## Unchanged complete WordPress source probe

One CLI probe at the code commit above, against the entire unmodified
`wp-includes/class-wp-block-parser.php` from WordPress source commit
`54106f6c545fc8b881f6b39b7963002bece6c28c`:

```sh
go run ./cmd/gohp -in ORIGINAL_FULL_FILE -php-version 8.3 -experimental-pcre2
```

Source SHA256:
`8563621090de5e307431b1d29b78a9f75fd9ee238a0137b563a7db7b2e1f8d3b`.
Exit 1, diagnostics:

```text
PHP002 line 384: expected , (at "->")
exit status 1
```

The prior class-parameter frontend blocker at line349 is passed. The next
frontend rejection is property-expression input to
`isset($stack_top->leading_html_start)`. Full-unit semantic validation,
generation, build, execution and native comparison have not been reached.
No generated/report output destination was requested; compilation failure is
not parser compatibility. No source rewrite, dependency pin or deployment change.

## Target-specific evidence

Native PHP is not installed in this local environment; independent pinned
PHP8.3.36/PHP8.4.27 differential CI and saved artifact readback are pending.
The new catalog entry contributes five checks, including one owned native
fixture. Existing enabled-PCRE evidence is not rerun or claimed by this change.
