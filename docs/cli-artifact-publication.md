# CLI artifact publication

The CLI previously wrote generated Go before attempting the optional report.
A missing report directory could leave a new Go file despite command failure;
an output or report path aliasing the PHP input could overwrite source bytes.

`fix/cli-artifact-publication` addresses these cases by staging bytes before
exclusive publication to fresh destinations. The report schema and compiler
semantics stay unchanged. See [the builder contract](integration.md) for exit,
hash verification, interruption and filesystem constraints.

Regression coverage executes the CLI on repository-owned PHP source and checks
source preservation, existing-file and alias rejection, absent report directories,
hash consistency, unsupported syntax and stdout mode. It requires no native PHP.
The active PHP 8.3.36-or-newer container CI exercises the complete compiler and
CLI workflow. Its resolved immutable oracle digest and actual version are
recorded in run evidence. Legacy two-version results remain historical only.
Validation results and the exact commit are recorded in the PR. This publication
contract neither enables PHP 8.4+ nor makes an oracle-version security guarantee.

Next work: full parser adapter assessment, structured diagnostics/source maps,
lvalues/references and array writes. This change does not establish compatibility
for additional PHP language features.
