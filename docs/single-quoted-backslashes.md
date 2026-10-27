# Single-quoted backslash bytes

For both explicit targets 8.3 and 8.4, single-quoted strings decode only `\\`
and `\'`. Every other backslash stays literal, including before `s`, `/`, `$`,
hexadecimal/octal/Unicode-looking text, arbitrary bytes and physical newlines.
Physical LF (also within CRLF) advances source locations; written `\n` does not.
Unterminated strings, a trailing backslash and an escaped final quote still fail.
These are the [PHP single-quoted string rules](https://www.php.net/manual/en/language.types.string.php#language.types.string.syntax.single).

The double-quoted closed subset is unchanged: unsupported escapes and
interpolation still reject. This does not enable heredoc/nowdoc, any new regex
backend or PHP8.4 `preg_*`. Existing IR string bytes and Go quoting are reused.

Catalog `php.syntax.single-quoted-backslashes` registers exact-byte/source-line
frontend checks, unterminated negative cases, and an owned differential fixture
for the independent PHP8.3/8.4 CI jobs. The fixture compares raw output, byte
lengths, quote/backslash parity, physical newline, UTF8, function passage and
array keys. Byte tests include NUL/invalid UTF8 at the frontend boundary only;
they are not claimed as additional native differential observations.

## Consumer motivation and boundary

Gossip's unchanged `WP_Block_Parser::next_token` method (upstream WordPress
`54106f6c545fc8b881f6b39b7963002bece6c28c`, lines 239–303) contains literal
backslash sequences. A selected-method probe retains those exact lines inside
an owned class wrapper with public document/offset declarations. On GoHP base
`d73e6e23702969fc59263e3a4664f71d5ecb4390`, with target 8.3 and
`-experimental-pcre2`, the first rejection is `PHP002 line 17: unsupported escape`.
The wrapper is not the full original class and is not an execution fixture.
After this lexer change, the same probe advances to
`PHP002 line 18: expected , (at "->")`: the experimental regex statement does
not admit the document property expression. Neither probe generates Go output.

No WordPress code is copied into GoHP. Gossip #50's already measured driver and
its GoHP pin stay unchanged. Accepting literal bytes does not demonstrate
`next_token` compilation or execution; property inputs, list destructuring and
JSON remain separate dependencies. Ordinary CGO0 CI does not test native PCRE2.

Capability: existing developer-local `gohp.cli.generate`; WebMCP `excluded`.
No new user-facing service operation, browser/S2S/MCP route, authentication,
deployment, dependency installation or consumer repin.
